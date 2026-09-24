// Package tlv implements the immutable r36 section 69 grammar, not history validity.
package tlv

import (
	"bytes"
	"encoding/binary"
	"errors"
	"unicode/utf8"
)

var ErrInvalid = errors.New("noncanonical v0 TLV")

type Field struct {
	Tag    uint16
	Value  []byte
	Offset int
}
type Object struct {
	Type     byte
	Fields   []Field
	Unsigned []byte
}

// Next performs bounded extraction only. Returned values alias the input.
func Next(b []byte) (tag uint16, value, rest []byte, err error) {
	if len(b) < 4 {
		return 0, nil, nil, ErrInvalid
	}
	n := int(binary.BigEndian.Uint16(b[2:4]))
	if n > len(b)-4 {
		return 0, nil, nil, ErrInvalid
	}
	return binary.BigEndian.Uint16(b[:2]), b[4 : 4+n], b[4+n:], nil
}
func Framing(b []byte) error {
	for len(b) > 0 {
		_, _, rest, e := Next(b)
		if e != nil {
			return e
		}
		b = rest
	}
	return nil
}
func Credential(b []byte) error {
	for tag := uint16(0x301); tag <= 0x304; tag++ {
		t, v, r, e := Next(b)
		if e != nil || t != tag {
			return ErrInvalid
		}
		b = r
		switch tag {
		case 0x301:
			if len(v) != 1 || v[0] < 1 || v[0] > 3 {
				return ErrInvalid
			}
		case 0x302:
			if len(v) != 1 || v[0] < 6 || v[0] > 8 {
				return ErrInvalid
			}
		case 0x303:
			if len(v) != 4 || binary.BigEndian.Uint32(v) == 0 {
				return ErrInvalid
			}
		case 0x304:
			if len(v) < 1 || len(v) > 128 {
				return ErrInvalid
			}
		}
	}
	if len(b) != 0 {
		return ErrInvalid
	}
	return nil
}
func Parse(b []byte, signed bool) (*Object, error) {
	if len(b) > 2030 {
		return nil, ErrInvalid
	}
	o := &Object{}
	remaining := b
	seen := map[uint16]bool{}
	var prev uint16
	parents, count := 0, -1
	var lastParent []byte
	assertions := 0
	signatureAt := -1
	for len(remaining) > 0 {
		off := len(b) - len(remaining)
		tag, v, r, e := Next(remaining)
		if e != nil {
			return nil, e
		}
		remaining = r
		if tag < prev || (tag == prev && tag != 5) {
			return nil, ErrInvalid
		}
		prev = tag
		seen[tag] = true
		o.Fields = append(o.Fields, Field{tag, v, off})
		switch tag {
		case 1:
			if off != 0 || len(v) != 1 || v[0] != 0 {
				return nil, ErrInvalid
			}
		case 2:
			if !seen[1] || len(v) != 1 || (v[0] != 1 && v[0] != 2) {
				return nil, ErrInvalid
			}
			o.Type = v[0]
		case 3:
			if len(v) != 32 {
				return nil, ErrInvalid
			}
		case 4:
			if len(v) != 2 {
				return nil, ErrInvalid
			}
			count = int(binary.BigEndian.Uint16(v))
			if count > 32 {
				return nil, ErrInvalid
			}
		case 5:
			if len(v) != 32 || count < 0 || parents >= count || (lastParent != nil && bytes.Compare(lastParent, v) >= 0) {
				return nil, ErrInvalid
			}
			parents++
			lastParent = v
		case 0x101:
			if o.Type != 1 || len(v) != 32 {
				return nil, ErrInvalid
			}
		case 0x102:
			if o.Type != 1 || len(v) != 1 || (v[0] != 1 && v[0] != 2) {
				return nil, ErrInvalid
			}
			assertions++
		case 0x103, 0x104:
			if o.Type != 1 || len(v) > 256 || !utf8.Valid(v) {
				return nil, ErrInvalid
			}
			assertions++
		case 0x105:
			if o.Type != 1 || Credential(v) != nil {
				return nil, ErrInvalid
			}
			assertions++
		case 0x201:
			if o.Type != 2 || len(v) > 256 || !utf8.Valid(v) {
				return nil, ErrInvalid
			}
		case 0xff01:
			if !signed || len(v) != 64 || len(remaining) != 0 {
				return nil, ErrInvalid
			}
			signatureAt = off
		default:
			return nil, ErrInvalid
		}
	}
	if !seen[1] || !seen[2] || !seen[3] || !seen[4] || parents != count || signed != (signatureAt >= 0) {
		return nil, ErrInvalid
	}
	if o.Type == 1 && (!seen[0x101] || assertions == 0) || o.Type == 2 && !seen[0x201] {
		return nil, ErrInvalid
	}
	o.Unsigned = b
	if signed {
		o.Unsigned = b[:signatureAt]
	}
	return o, nil
}
