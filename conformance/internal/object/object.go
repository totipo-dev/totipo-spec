// Package object implements the exact r16 TOKEN grammar. Parse only receives
// authenticated semantic bytes; storage readers first authenticate the envelope.
package object

import (
	"bytes"
	"encoding/binary"
	"errors"
	"totipo/conformance/internal/tlv"
	"unicode/utf8"
)

const (
	Supported  = "SUPPORTED_VALID"
	Invalid    = "INVALID"
	Capacity   = 1006
	MaxParents = 4
)

var ErrGrammar = errors.New("invalid TOKEN grammar")

type Object struct {
	Identity   []byte   `json:"identity"`
	Parents    [][]byte `json:"parents"`
	Status     byte     `json:"status"`
	Issuer     string   `json:"issuer"`
	Account    string   `json:"account"`
	Algorithm  byte     `json:"algorithm"`
	Digits     byte     `json:"digits"`
	Period     uint32   `json:"period"`
	Secret     []byte   `json:"secret"`
	ClientName *string  `json:"client_name,omitempty"`
	ClientTime *uint64  `json:"client_time,omitempty"`
}
type cursor struct {
	p   []byte
	err error
}

func (c *cursor) get(tag uint16, n int) []byte {
	if c.err != nil {
		return nil
	}
	f, r, e := tlv.Take(c.p)
	if e != nil || f.Tag != tag || (n >= 0 && len(f.Value) != n) {
		c.err = ErrGrammar
		return nil
	}
	c.p = r
	return f.Value
}
func (c *cursor) u8(tag uint16) byte {
	b := c.get(tag, 1)
	if b == nil {
		return 0
	}
	return b[0]
}
func (c *cursor) next(tag uint16) bool { return len(c.p) >= 2 && binary.BigEndian.Uint16(c.p) == tag }
func validText(s string, max int) bool { return len(s) <= max && utf8.ValidString(s) }
func Dispatch(p []byte) (string, *Object) {
	if len(p) > Capacity {
		return Invalid, nil
	}
	c := cursor{p: p}
	o := &Object{Identity: bytes.Clone(c.get(1, 32)), Parents: [][]byte{}}
	count := c.get(2, 2)
	if c.err != nil {
		return Invalid, nil
	}
	n := int(binary.BigEndian.Uint16(count))
	if n > MaxParents {
		return Invalid, nil
	}
	for i := 0; i < n; i++ {
		b := c.get(3, 32)
		if c.err != nil || (i > 0 && bytes.Compare(o.Parents[i-1], b) >= 0) {
			return Invalid, nil
		}
		o.Parents = append(o.Parents, bytes.Clone(b))
	}
	o.Status = c.u8(4)
	o.Issuer = string(c.get(5, -1))
	o.Account = string(c.get(6, -1))
	o.Algorithm = c.u8(7)
	o.Digits = c.u8(8)
	period := c.get(9, 4)
	if c.err != nil {
		return Invalid, nil
	}
	o.Period = binary.BigEndian.Uint32(period)
	o.Secret = bytes.Clone(c.get(10, -1))
	if c.next(11) {
		s := string(c.get(11, -1))
		o.ClientName = &s
		if !validText(s, 128) {
			return Invalid, nil
		}
	}
	if c.next(12) {
		b := c.get(12, 8)
		if c.err != nil {
			return Invalid, nil
		}
		v := binary.BigEndian.Uint64(b)
		o.ClientTime = &v
	}
	if c.err != nil || len(c.p) != 0 || o.Status < 1 || o.Status > 2 || !validText(o.Issuer, 256) || !validText(o.Account, 256) || o.Algorithm < 1 || o.Algorithm > 3 || o.Digits < 6 || o.Digits > 8 || o.Period == 0 || len(o.Secret) < 1 || len(o.Secret) > 128 {
		return Invalid, nil
	}
	return Supported, o
}
func (o Object) Encode() ([]byte, error) {
	if len(o.Parents) > MaxParents {
		return nil, ErrGrammar
	}
	var p []byte
	var err error
	add := func(tag uint16, b []byte) {
		if err != nil {
			return
		}
		var f []byte
		f, err = tlv.Encode(tag, b)
		p = append(p, f...)
	}
	add(1, o.Identity)
	add(2, tlv.U16(uint16(len(o.Parents))))
	for _, b := range o.Parents {
		add(3, b)
	}
	add(4, []byte{o.Status})
	add(5, []byte(o.Issuer))
	add(6, []byte(o.Account))
	add(7, []byte{o.Algorithm})
	add(8, []byte{o.Digits})
	add(9, tlv.U32(o.Period))
	add(10, o.Secret)
	if o.ClientName != nil {
		add(11, []byte(*o.ClientName))
	}
	if o.ClientTime != nil {
		add(12, tlv.U64(*o.ClientTime))
	}
	if err != nil {
		return nil, err
	}
	if s, _ := Dispatch(p); s != Supported {
		return nil, ErrGrammar
	}
	return p, nil
}

// ValueBytes includes STATUS and every credential field, and excludes identity,
// parents and optional client metadata.
func (o Object) ValueBytes() []byte {
	p, e := o.Encode()
	if e != nil {
		return nil
	}
	c := cursor{p: p}
	c.get(1, 32)
	n := c.get(2, 2)
	for i := 0; i < int(binary.BigEndian.Uint16(n)); i++ {
		c.get(3, 32)
	}
	start := c.p
	for tag := uint16(4); tag <= 10; tag++ {
		c.get(tag, -1)
	}
	return bytes.Clone(start[:len(start)-len(c.p)])
}
