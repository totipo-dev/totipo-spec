package vectors

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"
	"totipo/conformance/internal/cryptov1"
	"totipo/conformance/internal/object"
	"totipo/conformance/internal/storage"
)

// runPostAEAD proves the fixture reaches a particular post-authentication check,
// then feeds the unmodified bytes through the strict reader and storage boundary.
// Raw AEAD here is fixture verification, never an alternative accepting reader.
func runPostAEAD(c Case) error {
	x := c.PostAEAD
	root, e := unhex(c.Root)
	if e != nil {
		return e
	}
	k, e := cryptov1.Derive(root)
	if e != nil {
		return e
	}
	id, e := unhex(x.ObjectID)
	if e != nil {
		return e
	}
	b, e := unhex(x.Object)
	if e != nil {
		return e
	}
	if len(id) != 32 || len(b) != 1024 {
		return fmt.Errorf("fixture filename/size")
	}
	block, e := aes.NewCipher(k.ObjectKey(id))
	if e != nil {
		return e
	}
	a, e := cipher.NewGCM(block)
	if e != nil {
		return e
	}
	plain, e := a.Open(nil, id[:12], b, cryptov1.AAD(id))
	if e != nil {
		return fmt.Errorf("fixture must pass AEAD: %w", e)
	}
	if e := equalHex("authenticated encryption plaintext", x.Plaintext, plain); e != nil {
		return e
	}
	p, e := unhex(c.Semantic)
	if e != nil {
		return e
	}
	if class, _ := object.Dispatch(p); class != object.Supported {
		return fmt.Errorf("fixture semantic P must be canonical TOKEN")
	}
	want, e := cryptov1.Padded(p)
	if e != nil {
		return e
	}
	correctID := cryptov1.MAC(k.ID, p)
	n := int(binary.BigEndian.Uint16(plain))
	switch x.Defect {
	case "nonzero-padding":
		if n != len(p) || !bytes.Equal(id, correctID) || !bytes.Equal(plain[:2+n], want[:2+n]) || bytes.Equal(plain[2+n:], want[2+n:]) {
			return fmt.Errorf("not an isolated nonzero-padding fixture")
		}
	case "object-id-mismatch":
		if !bytes.Equal(plain, want) || bytes.Equal(id, correctID) {
			return fmt.Errorf("not an isolated keyed-ID mismatch")
		}
	case "semantic-length-invalid":
		if n <= object.Capacity || !bytes.Equal(plain[2:], want[2:]) || !bytes.Equal(id, correctID) {
			return fmt.Errorf("not an isolated length failure")
		}
	default:
		return fmt.Errorf("unknown post-AEAD defect")
	}
	if p, e := k.Open(x.ObjectID, b); e == nil || p != nil {
		return fmt.Errorf("strict reader accepted malformed envelope")
	}
	obs, e := storage.Scan("directory", []storage.Entry{{Path: "objects-v1/" + x.ObjectID, Kind: "regular", Read: func() ([]byte, error) { return b, nil }}}, k)
	if e != nil || len(obs) != 1 || obs[0].Class != c.Expected || obs[0].Object != nil || obs[0].Semantic != nil {
		return fmt.Errorf("post-AEAD failure contributed semantic state")
	}
	return nil
}
