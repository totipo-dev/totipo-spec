// Package dispatch accepts only authenticated, envelope-validated, identity-matching plaintext.
package dispatch

import "totipo/conformance/internal/tlv"

func Classify(p []byte) string {
	tag, v, r, e := tlv.Next(p)
	if e != nil || tag != 1 || len(v) != 1 {
		return "INVALID_PREFIX"
	}
	version := v[0]
	tag, v, _, e = tlv.Next(r)
	if e != nil || tag != 2 || len(v) != 1 {
		return "INVALID_PREFIX"
	}
	if version != 0 {
		return "AUTHENTICATED_UNSUPPORTED_FUTURE"
	}
	if _, e := tlv.Parse(p, true); e != nil {
		return "STRUCTURALLY_INVALID"
	}
	return "STRUCTURALLY_VALID"
}
