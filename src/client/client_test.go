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
		qrpc.RegisterHandler(&shs[i])
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
	k := common.QuorumSize / 2
	sec.SetRedundancy(k)

	// upload sector to quorum
	err = uploadSector(sec, q)
	if err != nil {
		t.Fatal("Failed to upload file:", err)
	}

	// rebuild file from first k segments
	segments := make([]common.Segment, k)
	for i := 0; i < k; i++ {
		segments[i] = shs[i].seg
	}

	err = erasure.RebuildSector(sec, segments)
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
	sec.SetRedundancy(common.QuorumSize / 2)

	// encode sector
	segments, err := erasure.EncodeRing(sec)
	if err != nil {
		t.Fatal("Failed to encode sector data:", err)
	}

	// create TCPServer
	rpcs, err := network.NewRPCServer(9988)
	if err != nil {
		t.Fatal("Failed to initialize TCPServer:", err)
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
		sh.seg = segments[i]
		qrpc.RegisterHandler(sh)
	}

	// download file from quorum
	err = downloadSector(sec, q)
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
