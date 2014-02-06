package disk

import (
	"bytes"
	//"io"
	"os"
	"testing"
)

func fill_block(size uint, input *os.File) []byte {
	out := make([]byte, size)
	_, _ = input.Read(out)
	return out

}

func Test_encoding(t *testing.T) {
	fi, err := os.Open("/dev/urandom")
	if err != nil {
		// panic()
	}

	defer fi.Close()
	in := make([][]byte, 3)
	for i := 0; i < 3; i++ {
		in[i] = fill_block(1024, fi)
	}
	//Should encode two blocks into two coding blocks. I think
	//Seems to specify the indices of the blocks. One of which will be kept as original input[0]
	//Also Any indices on the output which are above the indices of the input seem to be where the actual
	//encoding is done
	i := NewErasureCoder([]byte{0, 1, 2}, []byte{0, 1, 2, 3, 4})
	out := i.Code(in)
	c2 := NewErasureCoder([]byte{0, 3, 4}, []byte{1, 2})
	var in2 = [][]byte{out[0], out[3], out[4]}
	out2 := c2.Code(in2)
	if !bytes.Equal(out2[0], in[1]) {
		t.Error(out2, "!=", in[1])
	}
	if !bytes.Equal(out2[1], in[2]) {
		t.Error(out2, "!=", in[2])
	}
}
