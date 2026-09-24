package envelope

import (
	"encoding/binary"
	"errors"
)

const PlaintextSize = 2032
const Capacity = 2030
const FileSize = 2048

var ErrInvalid = errors.New("noncanonical envelope")

func Frame(p []byte) ([]byte, error) {
	if len(p) > Capacity {
		return nil, ErrInvalid
	}
	b := make([]byte, PlaintextSize)
	binary.BigEndian.PutUint16(b, uint16(len(p)))
	copy(b[2:], p)
	return b, nil
}
func Unframe(b []byte) ([]byte, error) {
	if len(b) != PlaintextSize {
		return nil, ErrInvalid
	}
	n := int(binary.BigEndian.Uint16(b))
	if n > Capacity {
		return nil, ErrInvalid
	}
	for _, v := range b[2+n:] {
		if v != 0 {
			return nil, ErrInvalid
		}
	}
	return b[2 : 2+n], nil
}
