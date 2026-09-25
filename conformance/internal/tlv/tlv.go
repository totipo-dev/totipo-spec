// Package tlv implements the fixed u16be tag/length framing in section 41.
package tlv

import (
	"encoding/binary"
	"errors"
)

var ErrEncoding = errors.New("invalid TLV encoding")

type Field struct {
	Tag   uint16
	Value []byte
}

// Take bounds-checks before returning a view into the input. No allocation is
// proportional to an untrusted declared length.
func Take(p []byte) (Field, []byte, error) {
	if len(p) < 4 {
		return Field{}, nil, ErrEncoding
	}
	n := int(binary.BigEndian.Uint16(p[2:4]))
	if n > len(p)-4 {
		return Field{}, nil, ErrEncoding
	}
	return Field{binary.BigEndian.Uint16(p[:2]), p[4 : 4+n]}, p[4+n:], nil
}
func Encode(tag uint16, v []byte) ([]byte, error) {
	if len(v) > 65535 {
		return nil, ErrEncoding
	}
	p := make([]byte, 4+len(v))
	binary.BigEndian.PutUint16(p, tag)
	binary.BigEndian.PutUint16(p[2:], uint16(len(v)))
	copy(p[4:], v)
	return p, nil
}
func U16(n uint16) []byte { b := make([]byte, 2); binary.BigEndian.PutUint16(b, n); return b }
func U32(n uint32) []byte { b := make([]byte, 4); binary.BigEndian.PutUint32(b, n); return b }
func U64(n uint64) []byte { b := make([]byte, 8); binary.BigEndian.PutUint64(b, n); return b }
