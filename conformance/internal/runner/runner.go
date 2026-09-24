// Package runner compares actual results with immutable corpus expectations.
package runner

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"totipo/conformance/internal/bootstrap"
	"totipo/conformance/internal/corpus"
	"totipo/conformance/internal/dispatch"
	"totipo/conformance/internal/ed25519profile"
	"totipo/conformance/internal/envelope"
	"totipo/conformance/internal/model"
	"totipo/conformance/internal/objectcrypto"
	"totipo/conformance/internal/presentation"
	"totipo/conformance/internal/referenceoracle"
	"totipo/conformance/internal/securitymemory"
	"totipo/conformance/internal/tlv"
	"totipo/conformance/internal/totp"
)

var ErrBlocked = errors.New("required conformance evidence is blocked")

func decoded(m map[string]string, k string) []byte { b, _ := corpus.Hex(m[k]); return b }
func compare(c corpus.Case, actual map[string]string) error {
	keys := make([]string, 0, len(c.Expected))
	for k := range c.Expected {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := c.Expected[k]
		got, ok := actual[k]
		if !ok || got != v {
			if strings.HasSuffix(k, "_hex") {
				return fmt.Errorf("%s differs (expected %d bytes, actual %d bytes)", k, len(v)/2, len(got)/2)
			}
			return fmt.Errorf("%s: expected %s; actual %s", k, v, got)
		}
	}
	return nil
}
func number(s string) (uint64, error) { return strconv.ParseUint(s, 10, 64) }
func Run(c corpus.Case) error {
	if e := corpus.Validate(c); e != nil {
		return e
	}
	in := c.Input
	out := map[string]string{}
	put := func(k string, b []byte) { out[k+"_hex"] = hex.EncodeToString(b) }
	switch c.Kind {
	case "blocked":
		return fmt.Errorf("%w: %s", ErrBlocked, c.Expected["reason"])
	case "use_profile":
		list := func(s string) []string {
			if s == "" {
				return []string{}
			}
			return strings.Split(s, ",")
		}
		v := model.AssessUse(model.UseProfile{Degraded: in["degraded"] == "true", Incomplete: in["incomplete"] == "true", LifecycleConflict: in["lifecycle_conflict"] == "true", StatusValues: list(in["status_values"]), CredentialValues: list(in["credential_values"]), CredentialComplete: in["credential_complete"] == "true", IssuerConflict: in["issuer_conflict"] == "true", AccountConflict: in["account_conflict"] == "true"})
		out["disposition"] = "BLOCK"
		if v.Allowed {
			out["disposition"] = "ALLOW"
		}
		out["active"] = strconv.FormatBool(v.Active)
		out["recovery_visible"] = strconv.FormatBool(v.RecoveryVisible)
		out["metadata_conflict"] = strconv.FormatBool(v.MetadataConflict)
	case "presentation":
		us := make([]presentation.Update, len(c.Presentation.Updates))
		for i, u := range c.Presentation.Updates {
			us[i] = presentation.Update{ID: u.ID, Signer: u.Signer, Parents: u.Parents, Name: u.Name, Kind: u.Kind}
		}
		got, e := presentation.Evaluate(us)
		if e != nil {
			return e
		}
		b, e := json.Marshal(c.Presentation.Expected)
		if e != nil {
			return e
		}
		var want presentation.View
		if e = json.Unmarshal(b, &want); e != nil {
			return e
		}
		if !reflect.DeepEqual(got, want) {
			return fmt.Errorf("presentation: expected %+v, actual %+v", want, got)
		}
		out["disposition"] = "VERIFIED"
	case "memory":
		events := make([]securitymemory.Event, len(c.Trace))
		for i, e := range c.Trace {
			events[i] = securitymemory.Event{Op: e.Op, Args: e.Args, Expected: e.Expected}
		}
		if err := securitymemory.Run(in["binding"], events); err != nil {
			return err
		}
		if err := referenceoracle.ReviewMemory(in["binding"], events); err != nil {
			return err
		}
		out["disposition"] = "VERIFIED"
	case "ed25519":
		err := ed25519profile.Verify(decoded(in, "public_key_hex"), decoded(in, "message_hex"), decoded(in, "signature_hex"))
		out["disposition"] = "ACCEPT"
		out["reason"] = "ACCEPT"
		if err != nil {
			out["disposition"] = "REJECT"
			out["reason"] = err.Error()
		}
	case "model":
		updates := make([]model.Update, 0, len(c.Scenario.Updates))
		unavailable := map[string]bool{}
		for _, id := range c.Scenario.Unavailable {
			unavailable[id] = true
		}
		selected := map[string]bool{}
		if c.Scenario.Observed != nil {
			todo := append([]string{}, c.Scenario.Observed...)
			for len(todo) > 0 {
				id := todo[0]
				todo = todo[1:]
				if selected[id] || unavailable[id] {
					continue
				}
				selected[id] = true
				for _, u := range c.Scenario.Updates {
					if u.ID == id {
						todo = append(todo, u.Parents...)
					}
				}
			}
		}
		for _, u := range c.Scenario.Updates {
			if unavailable[u.ID] || (c.Scenario.Observed != nil && !selected[u.ID]) {
				continue
			}
			updates = append(updates, model.Update{ID: u.ID, Token: u.Token, Parents: u.Parents, Fields: u.Fields})
		}
		view, e := model.Evaluate(updates)
		if e != nil {
			return e
		}
		literal, e := referenceoracle.Evaluate(updates)
		if e != nil {
			return e
		}
		if !reflect.DeepEqual(view, literal) {
			return fmt.Errorf("primary/reference disagreement: primary=%+v reference=%+v", view, literal)
		}
		if c.Scenario.Coverage != nil {
			primary, e := model.ResolutionCoverage(updates)
			if e != nil {
				return e
			}
			reference, e := referenceoracle.ResolutionCoverage(updates)
			if e != nil {
				return e
			}
			if !reflect.DeepEqual(primary, c.Scenario.Coverage) || !reflect.DeepEqual(reference, c.Scenario.Coverage) {
				return fmt.Errorf("resolution coverage mismatch: expected %v primary %v reference %v", c.Scenario.Coverage, primary, reference)
			}
		}
		b, e := json.Marshal(c.Scenario.Expected)
		if e != nil {
			return e
		}
		var expected model.View
		if e := json.Unmarshal(b, &expected); e != nil {
			return e
		}
		if !reflect.DeepEqual(view, expected) {
			actual, _ := json.Marshal(view)
			return fmt.Errorf("model view differs: expected %s; actual %s", b, actual)
		}
		out["disposition"] = "VERIFIED"
	case "tlv":
		o, e := tlv.Parse(decoded(in, "bytes_hex"), in["form"] == "signed")
		out["disposition"] = "STRUCTURALLY_INVALID"
		if e == nil {
			out["disposition"] = "STRUCTURALLY_VALID"
			put("unsigned", o.Unsigned)
		}
	case "envelope":
		p, e := envelope.Unframe(decoded(in, "bytes_hex"))
		out["disposition"] = "ENVELOPE_REJECTED"
		if e == nil {
			out["disposition"] = "ENVELOPE_VALID"
			put("semantic", p)
			out["semantic_disposition"] = "STRUCTURALLY_INVALID"
			if _, e := tlv.Parse(p, true); e == nil {
				out["semantic_disposition"] = "STRUCTURALLY_VALID"
			}
			framed, e := envelope.Frame(p)
			if e != nil || !bytes.Equal(framed, decoded(in, "bytes_hex")) {
				return fmt.Errorf("framing did not reproduce input")
			}
		}
	case "file_length":
		n, e := number(in["length"])
		if e != nil {
			return e
		}
		out["disposition"] = "FILE_LENGTH_REJECTED"
		if n == 2048 {
			out["disposition"] = "FILE_LENGTH_VALID"
		}
	case "dispatch":
		out["disposition"] = dispatch.Classify(decoded(in, "bytes_hex"))
	case "object_crypto":
		k, e := objectcrypto.Derive(decoded(in, "root_hex"))
		if e != nil {
			return e
		}
		put("prk", k.PRK)
		put("k_id", k.ID)
		put("k_object_root", k.ObjectRoot)
		put("k_signature_context", k.SignatureContext)
		typ := byte(1)
		if in["object_type"] == "DEVICE_UPDATE" {
			typ = 2
		}
		unsigned := decoded(in, "unsigned_hex")
		if _, e := tlv.Parse(unsigned, false); e != nil {
			return e
		}
		domain, e := objectcrypto.SignatureDomain(typ)
		if e != nil {
			return e
		}
		put("signature_domain", domain)
		msg, e := k.SignatureInput(typ, unsigned)
		if e != nil {
			return e
		}
		put("signature_input", msg)
		seed := decoded(in, "signing_seed_hex")
		if len(seed) != 32 {
			return fmt.Errorf("seed length")
		}
		private := ed25519.NewKeyFromSeed(seed)
		public := private.Public().(ed25519.PublicKey)
		put("public_key", public)
		sig := ed25519.Sign(private, msg)
		put("signature", sig)
		signed := append(append(append([]byte{}, unsigned...), 0xff, 1, 0, 64), sig...)
		put("signed_plaintext", signed)
		// Independently consume the external signed plaintext, not just our construction.
		external := decoded(c.Expected, "signed_plaintext_hex")
		o, e := tlv.Parse(external, true)
		if e != nil || !bytes.Equal(o.Unsigned, unsigned) {
			return fmt.Errorf("external signed/unsigned relation failed")
		}
		if !ed25519.Verify(decoded(c.Expected, "public_key_hex"), decoded(c.Expected, "signature_input_hex"), decoded(c.Expected, "signature_hex")) {
			return fmt.Errorf("standard Ed25519 positive diagnostic failed")
		}
		id, file, e := k.Seal(signed)
		if e != nil {
			return e
		}
		put("object_id", id)
		key, nonce, aad, e := k.Parameters(id)
		if e != nil {
			return e
		}
		put("object_key", key)
		put("nonce", nonce)
		put("aad", aad)
		frame, e := envelope.Frame(signed)
		if e != nil {
			return e
		}
		put("envelope_plaintext", frame)
		put("semantic_length", frame[:2])
		put("ciphertext", file[:2032])
		put("tag", file[2032:])
		put("file", file)
		p, stage := k.Open(decoded(c.Expected, "object_id_hex"), decoded(c.Expected, "file_hex"), nil)
		if stage != "STRUCTURALLY_VALID" || !bytes.Equal(p, external) {
			return fmt.Errorf("external object open: %s", stage)
		}
		out["disposition"] = "VERIFIED"
	case "bootstrap":
		file, key, e := bootstrap.Wrap(decoded(in, "root_hex"), decoded(in, "password_utf8_hex"), decoded(in, "argon2_salt_hex"), decoded(in, "wrap_nonce_hex"), nil)
		if e != nil {
			return e
		}
		put("vault_file", file)
		put("k_wrap", key)
		put("header_aad", file[:39])
		put("wrapped_root", file[39:71])
		put("wrap_tag", file[71:])
		root, s := bootstrap.Unwrap(decoded(c.Expected, "vault_file_hex"), decoded(in, "password_utf8_hex"), nil)
		if s != "VERIFIED" || !bytes.Equal(root, decoded(in, "root_hex")) {
			return fmt.Errorf("external bootstrap unwrap: %s", s)
		}
		out["disposition"] = "VERIFIED"
	case "bootstrap_check":
		calls := 0
		root, s := bootstrap.Unwrap(decoded(in, "file_hex"), decoded(in, "password_utf8_hex"), func(p, s []byte) []byte { calls++; return bootstrap.Derive(p, s) })
		out["disposition"] = s
		out["kdf_calls"] = strconv.Itoa(calls)
		put("root", root)
	case "password":
		var e error
		if in["encoding"] == "utf8-hex" {
			p, err := corpus.Hex(in["value"])
			if err != nil {
				return err
			}
			e = bootstrap.Password(p)
		} else {
			var runes []rune
			for _, v := range strings.Fields(in["value"]) {
				n, err := strconv.ParseInt(v, 16, 32)
				if err != nil {
					return err
				}
				runes = append(runes, rune(n))
			}
			_, e = bootstrap.EncodeRunes(runes)
		}
		out["disposition"] = "PASSWORD_REJECTED"
		if e == nil {
			out["disposition"] = "PASSWORD_VALID"
		}
	case "totp":
		algorithm, e := strconv.Atoi(in["algorithm"])
		if e != nil {
			return e
		}
		digits, e := strconv.Atoi(in["digits"])
		if e != nil {
			return e
		}
		period, e := number(in["period"])
		if e != nil {
			return e
		}
		seconds, ok := new(big.Int).SetString(in["seconds"], 10)
		if !ok {
			return fmt.Errorf("seconds")
		}
		code, e := totp.Generate(decoded(in, "secret_hex"), algorithm, digits, period, seconds)
		out["disposition"] = "INVALID"
		if e == nil {
			out["disposition"] = "VERIFIED"
			out["code"] = code
		}
	case "pipeline":
		k, e := objectcrypto.Derive(decoded(in, "root_hex"))
		if e != nil {
			return e
		}
		var stages []string
		_, out["disposition"] = k.Open(decoded(in, "object_id_hex"), decoded(in, "file_hex"), func(s string) { stages = append(stages, s) })
		out["stages"] = strings.Join(stages, ",")
	default:
		return fmt.Errorf("unimplemented kind %s", c.Kind)
	}
	return compare(c, out)
}
