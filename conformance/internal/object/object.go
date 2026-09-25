// Package object separates frozen routing from the supported version body.
// Dispatch accepts already authenticated semantic bytes; callers reading storage
// must first use cryptov1.Open to authenticate the envelope and keyed identity.
package object

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"totipo/conformance/internal/tlv"
	"unicode/utf8"
)

const (
	Supported = "SUPPORTED_VALID"
	Opaque    = "OPAQUE_ROUTABLE"
	Unscoped  = "OPAQUE_UNSCOPED"
	Invalid   = "INVALID"
	Token     = 1
	Device    = 2
	Capacity  = 1006
)

var ErrGrammar = errors.New("invalid semantic grammar")

type Routing struct {
	Version    byte     `json:"version"`
	Type       byte     `json:"type"`
	Parents    [][]byte `json:"parents"`
	AuthorTime uint64   `json:"author_time"`
	Identity   []byte   `json:"identity"`
	Author     []byte   `json:"author,omitempty"`
}
type Object struct {
	Routing
	Status      byte   `json:"status,omitempty"`
	Issuer      string `json:"issuer,omitempty"`
	Account     string `json:"account,omitempty"`
	Algorithm   byte   `json:"algorithm,omitempty"`
	Digits      byte   `json:"digits,omitempty"`
	Period      uint32 `json:"period,omitempty"`
	Secret      []byte `json:"secret,omitempty"`
	PublicKey   []byte `json:"public_key,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Signature   []byte `json:"signature"`
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

// ParseRouting never inspects any byte after the last frozen identity field.
func ParseRouting(p []byte) (Routing, []byte, error) {
	c := cursor{p: p}
	r := Routing{Version: c.u8(1), Type: c.u8(2)}
	count := c.get(4, 2)
	if c.err != nil {
		return r, nil, c.err
	}
	n := int(binary.BigEndian.Uint16(count))
	if n > 32 {
		return r, nil, ErrGrammar
	}
	for i := 0; i < n; i++ {
		b := c.get(5, 32)
		if c.err != nil {
			return r, nil, c.err
		}
		if i > 0 && bytes.Compare(r.Parents[i-1], b) >= 0 {
			return r, nil, ErrGrammar
		}
		r.Parents = append(r.Parents, bytes.Clone(b))
	}
	t := c.get(6, 8)
	if c.err != nil {
		return r, nil, c.err
	}
	r.AuthorTime = binary.BigEndian.Uint64(t)
	switch r.Type {
	case Token:
		r.Identity = bytes.Clone(c.get(0x101, 32))
		r.Author = bytes.Clone(c.get(0x102, 32))
	case Device:
		r.Identity = bytes.Clone(c.get(0x200, 32))
	default:
		return r, nil, ErrGrammar
	}
	return r, c.p, c.err
}
func DeviceID(key []byte) []byte {
	p := append([]byte("totipo/v1/device-id"), key...)
	h := sha256.Sum256(p)
	return h[:]
}
func validString(s string) bool { return len(s) <= 256 && utf8.ValidString(s) }
func Dispatch(p []byte) (string, *Object) {
	if len(p) > Capacity {
		return Invalid, nil
	}
	c := cursor{p: p}
	v := c.u8(1)
	typ := c.u8(2)
	if c.err != nil {
		return Unscoped, nil
	}
	if typ != Token && typ != Device {
		return Unscoped, nil
	}
	r, tail, e := ParseRouting(p)
	if e != nil {
		if v == 1 {
			return Invalid, nil
		}
		return Unscoped, nil
	}
	o := &Object{Routing: r}
	if v != 1 {
		return Opaque, o
	}
	c = cursor{p: tail}
	switch typ {
	case Token:
		o.Status = c.u8(0x103)
		o.Issuer = string(c.get(0x104, -1))
		o.Account = string(c.get(0x105, -1))
		nested := cursor{p: c.get(0x106, -1)}
		o.Algorithm = nested.u8(0x301)
		o.Digits = nested.u8(0x302)
		period := nested.get(0x303, 4)
		if nested.err == nil {
			o.Period = binary.BigEndian.Uint32(period)
		}
		o.Secret = bytes.Clone(nested.get(0x304, -1))
		if nested.err != nil || len(nested.p) != 0 || o.Status < 1 || o.Status > 2 || !validString(o.Issuer) || !validString(o.Account) || o.Algorithm < 1 || o.Algorithm > 3 || o.Digits < 6 || o.Digits > 8 || o.Period == 0 || len(o.Secret) < 1 || len(o.Secret) > 128 {
			return Invalid, nil
		}
	case Device:
		o.PublicKey = bytes.Clone(c.get(0x201, 65))
		o.DisplayName = string(c.get(0x202, -1))
		if !validString(o.DisplayName) || !bytes.Equal(o.Identity, DeviceID(o.PublicKey)) {
			return Invalid, nil
		}
	}
	o.Signature = bytes.Clone(c.get(0xff01, -1))
	if c.err != nil || len(c.p) != 0 || len(o.Signature) > 72 {
		return Invalid, nil
	}
	return Supported, o
}
func (r Routing) Encode() ([]byte, error) {
	if len(r.Parents) > 32 {
		return nil, ErrGrammar
	}
	var p []byte
	add := func(tag uint16, b []byte) { f, _ := tlv.Encode(tag, b); p = append(p, f...) }
	add(1, []byte{r.Version})
	add(2, []byte{r.Type})
	add(4, tlv.U16(uint16(len(r.Parents))))
	for _, b := range r.Parents {
		add(5, b)
	}
	add(6, tlv.U64(r.AuthorTime))
	switch r.Type {
	case Token:
		add(0x101, r.Identity)
		add(0x102, r.Author)
	case Device:
		add(0x200, r.Identity)
	default:
		return nil, ErrGrammar
	}
	_, _, err := ParseRouting(p)
	return p, err
}
func (o Object) Encode() ([]byte, error) {
	if o.Version != 1 {
		return nil, ErrGrammar
	}
	p, e := o.Routing.Encode()
	if e != nil {
		return nil, e
	}
	add := func(tag uint16, b []byte) {
		if e != nil {
			return
		}
		var f []byte
		f, e = tlv.Encode(tag, b)
		p = append(p, f...)
	}
	if o.Type == Token {
		add(0x103, []byte{o.Status})
		add(0x104, []byte(o.Issuer))
		add(0x105, []byte(o.Account))
		var cred []byte
		for _, f := range []tlv.Field{{Tag: 0x301, Value: []byte{o.Algorithm}}, {Tag: 0x302, Value: []byte{o.Digits}}, {Tag: 0x303, Value: tlv.U32(o.Period)}, {Tag: 0x304, Value: o.Secret}} {
			b, err := tlv.Encode(f.Tag, f.Value)
			if err != nil {
				return nil, err
			}
			cred = append(cred, b...)
		}
		add(0x106, cred)
	} else {
		add(0x201, o.PublicKey)
		add(0x202, []byte(o.DisplayName))
	}
	add(0xff01, o.Signature)
	if e != nil {
		return nil, e
	}
	s, _ := Dispatch(p)
	if s != Supported {
		return nil, ErrGrammar
	}
	return p, nil
}
func (o Object) Unsigned() ([]byte, error) {
	o.Signature = nil
	p, e := o.Encode()
	if e != nil {
		return nil, e
	}
	return p[:len(p)-4], nil
}

// ReservedSize is independent of the actual signature; validate all other
// fields on a parentless shape so oversized frontiers can still be planned.
func (o Object) ReservedSize() (int, error) {
	n := len(o.Parents)
	if n > 32 {
		return 0, ErrGrammar
	}
	if _, e := o.Routing.Encode(); e != nil {
		return 0, e
	}
	o.Parents = nil
	o.Signature = nil
	p, e := o.Encode()
	if e != nil {
		return 0, e
	}
	return len(p) + 72 + 36*n, nil
}
func (o Object) FanIn() (int, error) {
	o.Parents = nil
	n, e := o.ReservedSize()
	if e != nil {
		return 0, e
	}
	return min(32, (Capacity-n)/36), nil
}

// ValueBytes excludes routing, author time and provenance metadata.
func (o Object) ValueBytes() []byte {
	p, e := o.Encode()
	if e != nil {
		return nil
	}
	_, tail, _ := ParseRouting(p)
	return bytes.Clone(tail[:len(tail)-4-len(o.Signature)])
}
