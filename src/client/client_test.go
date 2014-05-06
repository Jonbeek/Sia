package main

import (
	"common"
	"common/crypto"
	"common/erasure"
	"network"
	"testing"
)

type ServerHandler struct {
	seg common.Segment
}

func (sh *ServerHandler) UploadSegment(seg common.Segment, arb *struct{}) error {
	sh.seg = seg
	return nil
}

func (sh *ServerHandler) DownloadSegment(hash crypto.Hash, seg *common.Segment) error {
	*seg = sh.seg
	return nil
}

// TestRPCUploadSector tests the NewTCPServer and uploadFile functions.
// NewRPCServer must properly initialize a RPC server.
// uploadSector must succesfully distribute a Sector among a quorum.
// The uploaded Sector must be successfully reconstructed.
func TestRPCuploadSector(t *testing.T) {
	// create TCPServer
	tcp, err := network.NewTCPServer(9988)
	if err != nil {
		t.Fatal("Failed to initialize TCPServer:", err)
	}
	defer tcp.Close()

	// create quorum
	var q [common.QuorumSize]common.Address
	var shs [common.QuorumSize]ServerHandler
	for i := 0; i < common.QuorumSize; i++ {
		q[i] = common.Address{0, "localhost", 9000 + i}
		qrpc, err := network.NewRPCServer(9000 + i)
		defer qrpc.Close()
		if err != nil {
			t.Fatal("Failed to initialize RPCServer:", err)
		}
		q[i].ID = qrpc.RegisterHandler(&shs[i])
	}

	// create sector
	secData, err := crypto.RandomByteSlice(70000)
	if err != nil {
		t.Fatal("Could not generate test data:", err)
	}

	sec, err := common.NewSector(secData)
	if err != nil {
		t.Fatal("Failed to create sector:", err)
	}

	// upload sector to quorum
	k := common.QuorumSize / 2
	ring, err := uploadSector(sec, k, q)
	if err != nil {
		t.Fatal("Failed to upload file:", err)
	}

	// rebuild file from first k segments
	ring.Segs = []common.Segment{}
	for i := 0; i < k; i++ {
		ring.AddSegment(&shs[i].seg)
	}

	sec, err = erasure.RebuildSector(ring)
	if err != nil {
		t.Fatal("Failed to rebuild file:", err)
	}

	// check hash
	rebuiltHash, err := crypto.CalculateHash(sec.Data)
	if err != nil {
		t.Fatal("Failed to calculate hash:", err)
	}

	if sec.Hash != rebuiltHash {
		t.Fatal("Failed to recover file: hashes do not match")
	}
}

// TestRPCdownloadSector tests the NewRPCServer and downloadSector functions.
// NewRPCServer must properly initialize a RPC server.
// downloadSector must successfully retrieve a Sector from a quorum.
// The downloaded Sector must match the original Sector.
func TestRPCdownloadSector(t *testing.T) {
	// create sector
	secData, err := crypto.RandomByteSlice(70000)
	if err != nil {
		t.Fatal("Could not generate test data:", err)
	}

	sec, err := common.NewSector(secData)
	if err != nil {
		t.Fatal("Failed to create sector:", err)
	}

	// encode sector
	k := common.QuorumSize / 2
	ring, err := erasure.EncodeRing(sec, k)
	if err != nil {
		t.Fatal("Failed to encode sector data:", err)
	}

	// create RPCServer
	rpcs, err := network.NewRPCServer(9988)
	if err != nil {
		t.Fatal("Failed to initialize RPCServer:", err)
	}
	defer rpcs.Close()

	// create quorum
	var q [common.QuorumSize]common.Address
	for i := 0; i < common.QuorumSize; i++ {
		q[i] = common.Address{0, "localhost", 9000 + i}
		qrpc, err := network.NewRPCServer(9000 + i)
		if err != nil {
			t.Fatal("Failed to initialize RPCServer:", err)
		}
		sh := new(ServerHandler)
		sh.seg = ring.Segs[i]
		q[i].ID = qrpc.RegisterHandler(sh)
	}

	// download file from quorum
	ring.Segs = []common.Segment{}
	sec, err = downloadSector(sec.Hash, ring, q)
	if err != nil {
		t.Fatal("Failed to download file:", err)
	}

	// check hash
	rebuiltHash, err := crypto.CalculateHash(sec.Data)
	if err != nil {
		t.Fatal("Failed to calculate hash:", err)
	}

	if sec.Hash != rebuiltHash {
		t.Fatal("Failed to recover file: hashes do not match")
	}
}
