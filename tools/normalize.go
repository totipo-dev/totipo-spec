// Command normalize copies reviewed expectations; it never imports the conformance implementation.
package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type provenance struct {
	Source  string `json:"source"`
	Status  string `json:"status"`
	Locator string `json:"locator"`
}
type vector struct {
	ID          string            `json:"id"`
	Kind        string            `json:"kind"`
	Description string            `json:"description"`
	Provenance  provenance        `json:"provenance"`
	Input       map[string]string `json:"input"`
	Expected    map[string]string `json:"expected"`
}
type bundle struct {
	Schema int      `json:"schema"`
	Cases  []vector `json:"cases"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func read(p string) []byte  { b, e := os.ReadFile(p); must(e); return b }
func hx(b []byte) string    { return hex.EncodeToString(b) }
func unhex(s string) []byte { b, e := hex.DecodeString(s); must(e); return b }
func write(category string, cases []vector) {
	sort.Slice(cases, func(i, j int) bool { return cases[i].ID < cases[j].ID })
	b, e := json.MarshalIndent(bundle{1, cases}, "", "  ")
	must(e)
	must(os.WriteFile("vectors/v0/"+category+"/cases.json", append(b, '\n'), 0644))
}
func sections(p string) (map[string]string, map[string]map[string]string) {
	head := map[string]string{}
	all := map[string]map[string]string{}
	cur := head
	s := bufio.NewScanner(strings.NewReader(string(read(p))))
	s.Buffer(make([]byte, 4096), 100000)
	for s.Scan() {
		l := s.Text()
		if strings.HasPrefix(l, "[") {
			name := strings.Trim(l, "[]")
			cur = map[string]string{}
			all[name] = cur
		} else if k, v, ok := strings.Cut(l, "="); ok {
			cur[k] = v
		}
	}
	must(s.Err())
	return head, all
}
func main() {
	writeFlag := flag.Bool("write-vectors", false, "explicitly normalize and hash vectors")
	flag.Parse()
	if !*writeFlag {
		panic("requires --write-vectors")
	}
	for p, want := range map[string]string{
		"review/source-vectors/totp-vault-v0-r31-tlv-corpus.json":       "c6c6639413f605ce18ecdb42d6e3d64b4ccd1709148eae43646dbcadf009b48c",
		"review/source-vectors/r31-object-vectors.txt":                  "641727de4921db28851eef5d92c552837cb48f9167fbcd5e2238a7bf5516ae88",
		"review/source-vectors/totp-vault-v0-r32-bootstrap-vectors.txt": "c62ad5b08fdc7c86da806a5aea65019fda74167b5911f9c94dcb1a8634d2d3f9",
		"review/source-code/R31Review.java":                             "50fc8bb40a9bd2253d27cbafa70617c3a6de8d4455d822d959704fb4a9ac3283",
	} {
		sum := sha256.Sum256(read(p))
		if fmt.Sprintf("%x", sum) != want {
			panic("reviewed source checksum mismatch: " + p)
		}
	}
	source := "review/source-vectors/totp-vault-v0-r31-tlv-corpus.json"
	var original struct {
		Cases []struct {
			ID, Kind, Name, Form, Expect, Hex string
			SemanticParseExpect               string `json:"semantic_parse_expect"`
			Length                            int
		}
	}
	must(json.Unmarshal(read(source), &original))
	groups := map[string][]vector{}
	for _, c := range original.Cases {
		kind, category, ok, bad := "tlv", "tlv", "STRUCTURALLY_VALID", "STRUCTURALLY_INVALID"
		in := map[string]string{"bytes_hex": c.Hex}
		ex := map[string]string{}
		switch c.Kind {
		case "semantic_tlv":
			in["form"] = c.Form
			if c.Expect == "valid" && c.Form == "signed" {
				ex["unsigned_hex"] = c.Hex[:len(c.Hex)-136]
			}
		case "envelope_plaintext":
			kind = "envelope"
			category = "envelope"
			ok = "ENVELOPE_VALID"
			bad = "ENVELOPE_REJECTED"
			if c.Expect == "valid" {
				b := unhex(c.Hex)
				n := int(b[0])*256 + int(b[1])
				ex["semantic_hex"] = hx(b[2 : 2+n])
				if c.SemanticParseExpect == "valid_signed" {
					ex["semantic_disposition"] = "STRUCTURALLY_VALID"
				} else {
					ex["semantic_disposition"] = "STRUCTURALLY_INVALID"
				}
			}
		case "object_file_length_gate":
			kind = "file_length"
			category = "envelope"
			ok = "FILE_LENGTH_VALID"
			bad = "FILE_LENGTH_REJECTED"
			in = map[string]string{"length": fmt.Sprint(c.Length)}
		default:
			panic(c.Kind)
		}
		ex["disposition"] = bad
		if c.Expect == "valid" {
			ex["disposition"] = ok
		}
		groups[category] = append(groups[category], vector{"v0/" + category + "/" + strings.ToLower(c.ID), kind, c.Name, provenance{source, "reviewed-pinned-vector", c.ID}, in, ex})
	}
	for cat, cases := range groups {
		write(cat, cases)
	}
	source = "review/source-vectors/r31-object-vectors.txt"
	_, secs := sections(source)
	var objects []vector
	for name, s := range secs {
		in := map[string]string{}
		ex := map[string]string{"disposition": "VERIFIED"}
		for k, v := range s {
			if k == "root" || k == "signing_seed" || k == "unsigned" {
				in[k+"_hex"] = v
			} else {
				ex[k+"_hex"] = v
			}
		}
		in["object_type"] = name
		ex["file_hex"] = s["ciphertext"] + s["tag"]
		ex["semantic_length_hex"] = s["envelope_plaintext"][:4]
		domain := "TOTP-Vault/v0/" + strings.ReplaceAll(strings.ToLower(name), "_", "-")
		ex["signature_domain_hex"] = hx([]byte(domain))
		objects = append(objects, vector{"v0/object-crypto/" + strings.ToLower(strings.ReplaceAll(name, "_", "-")), "object_crypto", "Exact reviewed cryptographic intermediates", provenance{source, "reviewed-pinned-vector", name}, in, ex})
	}
	write("object-crypto", objects)
	source = "review/source-vectors/totp-vault-v0-r32-bootstrap-vectors.txt"
	head, secs := sections(source)
	var boots []vector
	for name, s := range secs {
		in := map[string]string{"root_hex": head["root"]}
		ex := map[string]string{"disposition": "VERIFIED"}
		for k, v := range s {
			if k == "password_utf8" || k == "argon2_salt" || k == "wrap_nonce" {
				in[k+"_hex"] = v
			} else {
				ex[k+"_hex"] = v
			}
		}
		boots = append(boots, vector{"v0/bootstrap/" + strings.ToLower(name), "bootstrap", "Exact reviewed bootstrap record", provenance{source, "reviewed-pinned-vector", name}, in, ex})
	}
	write("bootstrap", boots)
	// Pinned constants in the historical review, not results from its implementation.
	source = "review/source-code/R31Review.java"
	var otp []vector
	for _, period := range []uint64{1, 30, 2147483648, 4294967295} {
		for counter, code := range []string{"755224", "287082", "359152"} {
			for _, edge := range []string{"start", "end"} {
				seconds := period * uint64(counter)
				if edge == "end" {
					seconds += period - 1
				}
				id := fmt.Sprintf("v0/totp/review-p%d-c%d-%s", period, counter, edge)
				otp = append(otp, vector{id, "totp", "Reviewed SHA-1 period boundary", provenance{source, "reviewed-pinned-vector", "testTotp: hotp constants and start/end assertions"}, map[string]string{"secret_hex": hx([]byte("12345678901234567890")), "algorithm": "1", "digits": "6", "period": fmt.Sprint(period), "seconds": fmt.Sprint(seconds)}, map[string]string{"disposition": "VERIFIED", "code": code}})
			}
		}
	}
	write("totp", otp)
	supplements()
	modelFixtures()
	// Inventory records all review inputs without modifying them.
	var inventory []string
	must(filepath.WalkDir("review", func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() {
			h := sha256.Sum256(read(p))
			inventory = append(inventory, fmt.Sprintf("%x  %s", h, p))
		}
		return nil
	}))
	sort.Slice(inventory, func(i, j int) bool { return inventory[i][66:] < inventory[j][66:] })
	must(os.WriteFile("conformance/review-inventory.sha256", []byte(strings.Join(inventory, "\n")+"\n"), 0644))
	var lines []string
	must(filepath.WalkDir("vectors/v0", func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() && strings.HasSuffix(p, ".json") {
			h := sha256.Sum256(read(p))
			rel, e := filepath.Rel("vectors/v0", p)
			must(e)
			lines = append(lines, fmt.Sprintf("%x  %s", h, filepath.ToSlash(rel)))
		}
		return nil
	}))
	sort.Slice(lines, func(i, j int) bool { return lines[i][66:] < lines[j][66:] })
	must(os.WriteFile("vectors/v0/manifest.sha256", []byte(strings.Join(lines, "\n")+"\n"), 0644))
}

func supplemental(category string, cases []vector) {
	sort.Slice(cases, func(i, j int) bool { return cases[i].ID < cases[j].ID })
	b, e := json.MarshalIndent(bundle{1, cases}, "", "  ")
	must(e)
	must(os.WriteFile("vectors/v0/"+category+"/supplemental.json", append(b, '\n'), 0644))
}
func specCase(category, id, kind, description, section string, in, ex map[string]string) vector {
	return vector{"v0/" + category + "/" + id, kind, description, provenance{"spec/totipo-vault-format-v0.md", "spec-derived", section}, in, ex}
}
func supplements() {
	var b bundle
	must(json.Unmarshal(read("vectors/v0/tlv/cases.json"), &b))
	byID := map[string]vector{}
	for _, c := range b.Cases {
		byID[c.ID] = c
	}
	var extra []vector
	parent := unhex(byID["v0/tlv/p003"].Input["bytes_hex"]) // first parent begins at offset 52, value at 56
	if hx(parent[52:56]) != "00050020" {
		panic("source layout")
	}
	parent[55] = 33
	parent = append(append(append([]byte{}, parent[:88]...), 0), parent[88:]...)
	extra = append(extra, specCase("tlv", "parent-33-bytes", "tlv", "33-byte parent ID rejected", "69.3", map[string]string{"bytes_hex": hx(parent), "form": "signed"}, map[string]string{"disposition": "STRUCTURALLY_INVALID"}))
	supplemental("tlv", extra)
	dev := unhex(byID["v0/tlv/p014"].Input["bytes_hex"])
	frame := make([]byte, 2032)
	binary.BigEndian.PutUint16(frame, uint16(len(dev)))
	copy(frame[2:], dev)
	supplemental("envelope", []vector{specCase("envelope", "maximum-device", "envelope", "Maximum actual DEVICE_UPDATE envelope", "14, 69.6", map[string]string{"bytes_hex": hx(frame)}, map[string]string{"disposition": "ENVELOPE_VALID", "semantic_hex": hx(dev), "semantic_disposition": "STRUCTURALLY_VALID"})})
	prefix := func(version, typ byte) []byte { return []byte{0, 1, 0, 1, version, 0, 2, 0, 1, typ} }
	var dispatches []vector
	addDispatch := func(id string, p []byte, want string) {
		dispatches = append(dispatches, specCase("dispatch", id, "dispatch", "Dispatch after authenticated identity validation", "15, 59, 69.1", map[string]string{"bytes_hex": hx(p)}, map[string]string{"disposition": want}))
	}
	addDispatch("future-opaque", append(prefix(1, 255), 0xff, 0x00, 0xff), "AUTHENTICATED_UNSUPPORTED_FUTURE")
	addDispatch("future-empty-tail", prefix(255, 0), "AUTHENTICATED_UNSUPPORTED_FUTURE")
	addDispatch("future-capacity", append(prefix(1, 255), make([]byte, 2020)...), "AUTHENTICATED_UNSUPPORTED_FUTURE")
	addDispatch("v0-unknown-type", prefix(0, 255), "STRUCTURALLY_INVALID")
	addDispatch("v0-missing-body", prefix(0, 1), "STRUCTURALLY_INVALID")
	addDispatch("v0-token", unhex(byID["v0/tlv/p001"].Input["bytes_hex"]), "STRUCTURALLY_VALID")
	for n := 0; n < 10; n++ {
		addDispatch(fmt.Sprintf("truncated-%02d", n), prefix(1, 1)[:n], "INVALID_PREFIX")
	}
	for _, off := range []int{0, 1, 2, 3, 5, 6, 7, 8} {
		p := prefix(1, 1)
		p[off] ^= 0x40
		addDispatch(fmt.Sprintf("bad-prefix-byte-%d", off), p, "INVALID_PREFIX")
	}
	supplemental("dispatch", dispatches)
	// Construct hostile encrypted inputs using reviewed keys and the standard AES-GCM primitive.
	// Expected stages below are authored from the spec, never calculated by the runner.
	_, objects := sections("review/source-vectors/r31-object-vectors.txt")
	obj := objects["TOKEN_UPDATE"]
	hmac256 := func(k, p []byte) []byte { h := hmac.New(sha256.New, k); h.Write(p); return h.Sum(nil) }
	seal := func(id, plain []byte) []byte {
		info := append([]byte("TOTP-Vault/v0/object-key"), id...)
		info = append(info, 1)
		key := hmac256(unhex(obj["k_object_root"]), info)
		block, e := aes.NewCipher(key)
		must(e)
		g, e := cipher.NewGCM(block)
		must(e)
		return g.Seal(nil, id[:12], plain, append([]byte("TOTP-Vault/v0/object"), id...))
	}
	var pipelines []vector
	addPipeline := func(id string, oid, file []byte, want, stages string) {
		pipelines = append(pipelines, specCase("dispatch", id, "pipeline", "Authenticated pipeline ordering", "15", map[string]string{"root_hex": obj["root"], "object_id_hex": hx(oid), "file_hex": hx(file)}, map[string]string{"disposition": want, "stages": stages}))
	}
	oid := unhex(obj["object_id"])
	file := unhex(obj["ciphertext"] + obj["tag"])
	for _, n := range []int{0, 1, 2047, 2049} {
		addPipeline(fmt.Sprintf("pipeline-length-%d", n), oid, make([]byte, n), "FILE_LENGTH_REJECTED", "FILE_LENGTH")
	}
	corrupted := append([]byte{}, file...)
	corrupted[0] ^= 1
	addPipeline("pipeline-aead", oid, corrupted, "AEAD_REJECTED", "FILE_LENGTH,AEAD")
	for _, mutation := range []string{"length", "padding", "identity"} {
		plain := unhex(obj["envelope_plaintext"])
		want, stages := "ENVELOPE_REJECTED", "FILE_LENGTH,AEAD,ENVELOPE"
		switch mutation {
		case "length":
			plain[0] = 255
			plain[1] = 255
		case "padding":
			plain[2031] = 1
		case "identity":
			plain[2] ^= 1
			want = "OBJECT_ID_REJECTED"
			stages += ",OBJECT_ID"
		}
		addPipeline("pipeline-"+mutation, oid, seal(oid, plain), want, stages)
	}
	for _, c := range dispatches {
		p := unhex(c.Input["bytes_hex"])
		plain := make([]byte, 2032)
		binary.BigEndian.PutUint16(plain, uint16(len(p)))
		copy(plain[2:], p)
		id := hmac256(unhex(obj["k_id"]), p)
		addPipeline("pipeline-"+strings.TrimPrefix(c.ID, "v0/dispatch/"), id, seal(id, plain), c.Expected["disposition"], "FILE_LENGTH,AEAD,ENVELOPE,OBJECT_ID,DISPATCH")
	}
	supplemental("dispatch", append(dispatches, pipelines...))
	_, boots := sections("review/source-vectors/totp-vault-v0-r32-bootstrap-vectors.txt")
	initial := boots["INITIAL"]
	vault := unhex(initial["vault_file"])
	var bootstrapCases []vector
	addBoot := func(id string, record []byte, password string, want string, calls int) {
		bootstrapCases = append(bootstrapCases, specCase("bootstrap", id, "bootstrap_check", "Bootstrap preflight/authentication rejection", "6, 8", map[string]string{"file_hex": hx(record), "password_utf8_hex": password}, map[string]string{"disposition": want, "kdf_calls": fmt.Sprint(calls)}))
	}
	for _, n := range []int{0, 1, 10, 86, 88, 100} {
		addBoot(fmt.Sprintf("length-%d", n), make([]byte, n), initial["password_utf8"], "FILE_LENGTH_REJECTED", 0)
	}
	for i := 0; i < 10; i++ {
		v := append([]byte{}, vault...)
		v[i] ^= 1
		addBoot(fmt.Sprintf("magic-%d", i), v, initial["password_utf8"], "MAGIC_REJECTED", 0)
	}
	for _, v := range []byte{1, 127, 255} {
		b := append([]byte{}, vault...)
		b[10] = v
		addBoot(fmt.Sprintf("version-%d", v), b, initial["password_utf8"], "BOOTSTRAP_VERSION_REJECTED", 0)
	}
	for _, p := range []struct{ id, value string }{{"password-overlong", strings.Repeat("61", 1025)}, {"password-invalid-utf8", "c080"}, {"password-surrogate", "eda080"}} {
		addBoot(p.id, vault, p.value, "PASSWORD_REJECTED", 0)
	}
	addBoot("wrong-password", vault, "77726f6e67", "AEAD_REJECTED", 1)
	for name, off := range map[string]int{"salt": 11, "nonce": 27, "ciphertext": 39, "tag": 71} {
		v := append([]byte{}, vault...)
		v[off] ^= 1
		addBoot("mutated-"+name, v, initial["password_utf8"], "AEAD_REJECTED", 1)
	}
	for id, p := range map[string]string{"empty": "", "max-utf8": strings.Repeat("f09f9880", 256), "invalid-continuation": "80", "invalid-codepoint": "f4908080"} {
		want := "PASSWORD_VALID"
		if strings.HasPrefix(id, "invalid") {
			want = "PASSWORD_REJECTED"
		}
		bootstrapCases = append(bootstrapCases, specCase("bootstrap", "domain-"+id, "password", "Strict password domain", "6", map[string]string{"encoding": "utf8-hex", "value": p}, map[string]string{"disposition": want}))
	}
	for _, p := range []string{"d800", "dc00", "110000", "-1"} {
		bootstrapCases = append(bootstrapCases, specCase("bootstrap", "source-codepoint-"+strings.ReplaceAll(p, "-", "minus"), "password", "Invalid source character must not be replaced", "6", map[string]string{"encoding": "codepoints", "value": p}, map[string]string{"disposition": "PASSWORD_REJECTED"}))
	}
	supplemental("bootstrap", bootstrapCases)
	var otp []vector
	times := []string{"59", "1111111109", "1111111111", "1234567890", "2000000000", "20000000000"}
	codes := [][3]string{{"94287082", "46119246", "90693936"}, {"07081804", "68084774", "25091201"}, {"14050471", "67062674", "99943326"}, {"89005924", "91819424", "93441116"}, {"69279037", "90698825", "38618901"}, {"65353130", "77737706", "47863826"}}
	for i, seconds := range times {
		for a, n := range []int{20, 32, 64} {
			secret := strings.Repeat("1234567890", 7)[:n]
			for digits := 6; digits <= 8; digits++ {
				id := fmt.Sprintf("v0/totp/rfc6238-t%s-a%d-d%d", seconds, a+1, digits)
				otp = append(otp, vector{id, "totp", "RFC 6238 value, decimal reduction to selected digit count", provenance{"https://www.rfc-editor.org/rfc/rfc6238.html", "published-standard", "Appendix B table; Appendix A algorithm-specific secret lengths; 6/7 digits are the last digits of the published 8-digit result"}, map[string]string{"secret_hex": hx([]byte(secret)), "algorithm": fmt.Sprint(a + 1), "digits": fmt.Sprint(digits), "period": "30", "seconds": seconds}, map[string]string{"disposition": "VERIFIED", "code": codes[i][a][8-digits:]}})
			}
		}
	}
	for _, v := range []struct{ id, period, seconds string }{{"negative-time", "30", "-1"}, {"period-zero", "0", "0"}, {"period-overflow", "4294967296", "0"}, {"counter-overflow", "1", "18446744073709551616"}} {
		otp = append(otp, specCase("totp", v.id, "totp", "TOTP input bounds", "68", map[string]string{"secret_hex": "01", "algorithm": "1", "digits": "6", "period": v.period, "seconds": v.seconds}, map[string]string{"disposition": "INVALID"}))
	}
	supplemental("totp", otp)
}

func modelFixtures() {
	type update struct {
		ID      string            `json:"id"`
		Token   string            `json:"token"`
		Parents []string          `json:"parents"`
		Fields  map[string]string `json:"fields"`
	}
	type witness struct {
		Status      string   `json:"status"`
		Credential  string   `json:"credential"`
		Transitions []string `json:"transitions"`
	}
	type token struct {
		Heads     []string            `json:"heads"`
		Fields    map[string][]string `json:"fields"`
		Conflicts []witness           `json:"conflicts"`
	}
	type view struct {
		Validation  map[string]string `json:"validation"`
		Tokens      map[string]token  `json:"tokens"`
		Resolutions []string          `json:"resolutions"`
	}
	type scenario struct {
		Updates  []update `json:"updates"`
		Expected view     `json:"expected"`
	}
	type modelCase struct {
		vector
		Scenario scenario `json:"scenario"`
	}
	u := func(id string, parents []string, fields map[string]string) update {
		return update{id, "t", parents, fields}
	}
	root := u("r", []string{}, map[string]string{"STATUS": "LIVE", "ISSUER": "issuer", "ACCOUNT": "account", "CREDENTIAL": "credential-0"})
	state := func(status, issuer, account, credential []string) map[string][]string {
		return map[string][]string{"STATUS": status, "ISSUER": issuer, "ACCOUNT": account, "CREDENTIAL": credential}
	}
	one := func(s string) []string { return []string{s} }
	var cases []modelCase
	add := func(id, description string, updates []update, heads []string, fields map[string][]string, conflicts []witness, resolutions []string, overrides map[string]string) {
		validation := map[string]string{}
		for _, u := range updates {
			validation[u.ID] = "FULLY_VALID"
		}
		for id, s := range overrides {
			validation[id] = s
		}
		tokens := map[string]token{}
		if len(heads) > 0 {
			tokens["t"] = token{heads, fields, conflicts}
		}
		c := specCase("lifecycle", id, "model", description, "22-41, 48, 50-56", map[string]string{}, map[string]string{"disposition": "VERIFIED"})
		cases = append(cases, modelCase{c, scenario{updates, view{validation, tokens, resolutions}}})
	}
	none := []witness{}
	nores := []string{}
	add("creation", "Parentless creation initializes all four fields", []update{root}, one("r"), state(one("r"), one("r"), one("r"), one("r")), none, nores, nil)
	a := u("a", one("r"), map[string]string{"ISSUER": "A"})
	b := u("b", one("r"), map[string]string{"ACCOUNT": "B"})
	add("disjoint-compose", "Concurrent disjoint fields compose", []update{root, a, b}, []string{"a", "b"}, state(one("r"), one("a"), one("b"), one("r")), none, nores, nil)
	b = u("b", one("r"), map[string]string{"ISSUER": "A"})
	add("equal-values-retain-lineage", "Equal concurrent values retain separate revisions", []update{root, a, b}, []string{"a", "b"}, state(one("r"), []string{"a", "b"}, one("r"), one("r")), none, nores, nil)
	m := u("m", []string{"a", "b"}, map[string]string{"CREDENTIAL": "credential-1"})
	add("single-head-inherits-conflict", "A sole causal head can carry multiple field heads", []update{root, a, b, m}, one("m"), state(one("r"), []string{"a", "b"}, one("r"), one("m")), none, nores, nil)
	z := u("z", one("m"), map[string]string{"ISSUER": "selected"})
	add("ordinary-resolution", "Field assertion replaces every incorporated field alternative", []update{root, a, b, m, z}, one("z"), state(one("r"), one("z"), one("r"), one("m")), none, nores, nil)
	d := u("d", one("r"), map[string]string{"STATUS": "TOMBSTONE"})
	w := u("w", one("r"), map[string]string{"CREDENTIAL": "credential-1"})
	conflict := []witness{{"d", "w", one("d")}}
	add("delete-rotation", "Deletion races with credential rotation", []update{root, d, w}, []string{"d", "w"}, state(one("d"), one("r"), one("r"), one("w")), conflict, nores, nil)
	m = u("m", []string{"d", "w"}, map[string]string{"ISSUER": "edited"})
	add("issuer-preserves-witness", "Ordinary edit carries lifecycle evidence", []update{root, d, w, m}, one("m"), state(one("d"), one("m"), one("r"), one("w")), conflict, nores, nil)
	x := u("x", []string{"d", "w"}, map[string]string{"STATUS": "LIVE", "CREDENTIAL": "credential-2"})
	add("mixed-resolution-invalid", "STATUS plus another field is not a lifecycle candidate", []update{root, d, w, x}, []string{"d", "w"}, state(one("d"), one("r"), one("r"), one("w")), conflict, nores, map[string]string{"x": "INVALID"})
	x = u("x", []string{"d", "w"}, map[string]string{"STATUS": "LIVE"})
	add("status-only-resolution", "Candidate provisional coverage becomes final only after empty POST", []update{root, d, w, x}, one("x"), state(one("x"), one("r"), one("r"), one("w")), none, one("x"), nil)
	y := u("y", one("w"), map[string]string{"CREDENTIAL": "credential-2"})
	add("stale-witness-reappears", "A stale credential descendant is outside resolution coverage", []update{root, d, w, x, y}, []string{"x", "y"}, state(one("x"), one("r"), one("r"), one("y")), []witness{{"x", "y", []string{"d", "x"}}}, one("x"), nil)
	s := u("s", one("d"), map[string]string{"STATUS": "LIVE"})
	c := u("c", []string{"s", "w"}, map[string]string{"CREDENTIAL": "credential-2"})
	add("historical-nonhead-witness", "Credential history retains a witness below a causally informed current head", []update{root, d, w, s, c}, one("c"), state(one("s"), one("r"), one("r"), one("c")), []witness{{"s", "w", []string{"d", "s"}}}, nores, nil)
	p := u("p", one("missing"), map[string]string{"ISSUER": "pending"})
	add("pending-no-root-rewrite", "Missing dependencies remain pending and do not create state", []update{p}, []string{}, nil, none, nores, map[string]string{"p": "PENDING"})
	bad := u("bad", []string{"d", "r"}, map[string]string{"ISSUER": "bad"})
	add("redundant-context-invalid", "A parent cannot be an ancestor of another listed parent", []update{root, d, bad}, one("d"), state(one("d"), one("r"), one("r"), one("r")), none, nores, map[string]string{"bad": "INVALID"})
	bad = u("bad", one("d"), map[string]string{"CREDENTIAL": "bad"})
	add("tombstone-credential-invalid", "Tombstone credential writes require LIVE in the same ordinary update", []update{root, d, bad}, one("d"), state(one("d"), one("r"), one("r"), one("r")), none, nores, map[string]string{"bad": "INVALID"})
	restore := u("restore", one("d"), map[string]string{"STATUS": "LIVE", "CREDENTIAL": "restored"})
	add("ordinary-restore-credential", "PRE empty permits an ordinary combined restore and credential write", []update{root, d, restore}, one("restore"), state(one("restore"), one("r"), one("r"), one("restore")), none, nores, nil)
	sort.Slice(cases, func(i, j int) bool { return cases[i].ID < cases[j].ID })
	raw, e := json.MarshalIndent(struct {
		Schema int         `json:"schema"`
		Cases  []modelCase `json:"cases"`
	}{1, cases}, "", "  ")
	must(e)
	must(os.WriteFile("vectors/v0/lifecycle/cases.json", append(raw, '\n'), 0644))
}
