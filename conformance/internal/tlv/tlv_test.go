package tlv

import (
	"bytes"
	"testing"
)

func TestFraming(t *testing.T) {
	p, e := Encode(0x1234, []byte{0xab, 0xcd})
	if e != nil || !bytes.Equal(p, []byte{0x12, 0x34, 0, 2, 0xab, 0xcd}) {
		t.Fatal("wire encoding")
	}
	for i := 0; i < len(p); i++ {
		if _, _, e := Take(p[:i]); e == nil {
			t.Fatalf("truncation %d", i)
		}
	}
	f, rest, e := Take(p)
	if e != nil || len(rest) != 0 || f.Tag != 0x1234 {
		t.Fatal("decode")
	}
	if _, e = Encode(1, make([]byte, 65536)); e == nil {
		t.Fatal("length wrapped")
	}
}
