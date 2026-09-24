// Package client integrates frozen codecs, strict signatures and Phase 2 models
// with serialized durable local evidence. It is a reference, not a product API.
package client

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"totipo/conformance/internal/bootstrap"
	"totipo/conformance/internal/ed25519profile"
	"totipo/conformance/internal/model"
	"totipo/conformance/internal/objectcrypto"
	"totipo/conformance/internal/presentation"
	"totipo/conformance/internal/referenceoracle"
	"totipo/conformance/internal/tlv"
	"totipo/conformance/reference/fault"
	"totipo/conformance/reference/localstate"
	"totipo/conformance/reference/storage"
)

var ErrStale = errors.New("prepared operation is stale; retry and reconfirm")

type Device struct {
	Binding string
	Private ed25519.PrivateKey
}
type Client struct {
	Objects  storage.Objects
	Security localstate.Store
	// Cache is optional trusted retained encrypted immutable storage outside the
	// synchronized namespace. Cached bytes still pass every authentication gate.
	Cache    storage.Objects
	Location string
	// MaxCandidates is an implementation work limit, never a protocol limit.
	// Zero selects 4096. Exceeding it aborts before accepting a partial view.
	MaxCandidates int
	Hook          fault.Hook
	root          []byte
	keys          objectcrypto.Keys
	device        Device
	session       string
}

func random(n int) ([]byte, error) { b := make([]byte, n); _, e := rand.Read(b); return b, e }
func RandomID() (string, error)    { b, e := random(32); return hex.EncodeToString(b), e }
func NewCredential() ([]byte, error) {
	secret, e := random(20)
	if e != nil {
		return nil, e
	}
	return append(append(append(TLV(0x301, []byte{1}), TLV(0x302, []byte{6})...), TLV(0x303, []byte{0, 0, 0, 30})...), TLV(0x304, secret)...), nil
}
func Binding(root []byte) string {
	return hex.EncodeToString(objectcrypto.MAC(root, []byte("TOTP-Vault/v0/local-vault-binding")))
}
func New(objects storage.Objects, state localstate.Store, location string) *Client {
	return &Client{Objects: objects, Security: state, Location: location}
}
func (c *Client) initialize(root []byte, device *Device) error {
	var e error
	c.keys, e = objectcrypto.Derive(root)
	if e != nil {
		return e
	}
	c.root = append([]byte{}, root...)
	binding := Binding(root)
	if device != nil {
		if device.Binding != binding || len(device.Private) != ed25519.PrivateKeySize {
			return errors.New("device binding mismatch")
		}
		c.device = Device{binding, append(ed25519.PrivateKey{}, device.Private...)}
	} else {
		_, key, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}
		c.device = Device{binding, key}
	}
	c.session, e = RandomID()
	return e
}
func (c *Client) Device() Device {
	return Device{c.device.Binding, append(ed25519.PrivateKey{}, c.device.Private...)}
}
func (c *Client) check(s *localstate.State) error {
	if s.Location != c.Location || s.Establishment != "ESTABLISHED" || s.Binding != Binding(c.root) {
		return errors.New("not established for this location/root")
	}
	return nil
}
func candidate(root, password []byte) ([]byte, error) {
	salt, e := random(16)
	if e != nil {
		return nil, e
	}
	nonce, e := random(12)
	if e != nil {
		return nil, e
	}
	b, _, e := bootstrap.Wrap(root, password, salt, nonce, nil)
	if e != nil {
		return nil, e
	}
	r, stage := bootstrap.Unwrap(b, password, nil)
	if stage != "VERIFIED" || !bytes.Equal(r, root) {
		return nil, errors.New("candidate failed self-validation")
	}
	return b, nil
}
func (c *Client) Create(password []byte) error {
	root, e := random(32)
	if e != nil {
		return e
	}
	b, e := candidate(root, password)
	if e != nil {
		return e
	}
	e = c.Security.With(func(t localstate.Transaction) error {
		s := t.State()
		if s.Establishment != "" || s.Binding != "" {
			return errors.New("establishment already started")
		}
		if _, err := c.Objects.ReadVault(); err == nil {
			return errors.New("canonical VAULT exists")
		} else if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		ids, err := c.Objects.ListCandidates()
		if err != nil {
			return err
		}
		if len(ids) > 0 {
			return errors.New("recoverable object candidates preclude creation")
		}
		s.Location = c.Location
		s.Binding = Binding(root)
		s.Establishment = "PENDING"
		if err = t.Commit("establish-pending"); err != nil {
			return err
		}
		if err = c.Hook.Hit("create.after-pending"); err != nil {
			return err
		}
		if err = c.Objects.InstallInitialVault(b); err != nil {
			return err
		}
		return c.finishOpen(t, password)
	})
	if e != nil {
		return e
	}
	return c.initialize(root, nil)
}
func (c *Client) finishOpen(t localstate.Transaction, password []byte) error {
	b, e := c.Objects.ReadVault()
	if e != nil {
		return e
	}
	root, stage := bootstrap.Unwrap(b, password, nil)
	if stage != "VERIFIED" {
		return errors.New(stage)
	}
	s := t.State()
	binding := Binding(root)
	if s.Binding != "" && (s.Binding != binding || s.Location != c.Location) {
		return errors.New("canonical root/location differs from durable binding")
	}
	if s.Establishment != "ESTABLISHED" {
		s.Location = c.Location
		s.Binding = binding
		s.Establishment = "ESTABLISHED"
		if e = t.Commit("established"); e != nil {
			return e
		}
	}
	c.root = append([]byte{}, root...)
	return nil
}
func (c *Client) Open(password []byte, device *Device) error {
	e := c.Security.With(func(t localstate.Transaction) error { return c.finishOpen(t, password) })
	if e != nil {
		return e
	}
	return c.initialize(c.root, device)
}
func (c *Client) Rewrap(password []byte) error {
	b, e := candidate(c.root, password)
	if e != nil {
		return e
	}
	return c.Security.With(func(t localstate.Transaction) error {
		if e := c.check(t.State()); e != nil {
			return e
		}
		return c.Objects.ReplaceVault(b)
	})
}

func TLV(tag uint16, b []byte) []byte {
	out := make([]byte, 4, len(b)+4)
	binary.BigEndian.PutUint16(out, tag)
	binary.BigEndian.PutUint16(out[2:], uint16(len(b)))
	return append(out, b...)
}

// Encode creates canonical signed bytes; semantic validation belongs to the
// writer pipeline. Parents must already be the complete selected frontier.
func Encode(keys objectcrypto.Keys, private ed25519.PrivateKey, typ byte, token string, parents []string, fields map[string]string) (string, []byte, error) {
	if len(private) != ed25519.PrivateKeySize || len(parents) > 32 {
		return "", nil, errors.New("invalid signing key or capacity")
	}
	b := append(TLV(1, []byte{0}), TLV(2, []byte{typ})...)
	b = append(b, TLV(3, private.Public().(ed25519.PublicKey))...)
	count := make([]byte, 2)
	binary.BigEndian.PutUint16(count, uint16(len(parents)))
	b = append(b, TLV(4, count)...)
	ps := append([]string{}, parents...)
	sort.Strings(ps)
	for _, p := range ps {
		x, e := hex.DecodeString(p)
		if e != nil || len(x) != 32 {
			return "", nil, errors.New("invalid parent ID")
		}
		b = append(b, TLV(5, x)...)
	}
	if typ == 1 {
		x, e := hex.DecodeString(token)
		if e != nil || len(x) != 32 {
			return "", nil, errors.New("invalid token ID")
		}
		b = append(b, TLV(0x101, x)...)
	}
	tags := map[string]uint16{"STATUS": 0x102, "ISSUER": 0x103, "ACCOUNT": 0x104, "CREDENTIAL": 0x105, "DISPLAY_NAME": 0x201}
	order := []string{"STATUS", "ISSUER", "ACCOUNT", "CREDENTIAL", "DISPLAY_NAME"}
	for name := range fields {
		if tags[name] == 0 {
			return "", nil, errors.New("unknown field")
		}
	}
	for _, name := range order {
		v, ok := fields[name]
		if !ok {
			continue
		}
		data := []byte(v)
		if name == "STATUS" {
			switch v {
			case "LIVE":
				data = []byte{1}
			case "TOMBSTONE":
				data = []byte{2}
			default:
				return "", nil, errors.New("invalid status")
			}
		}
		if name == "CREDENTIAL" {
			var e error
			data, e = hex.DecodeString(v)
			if e != nil {
				return "", nil, e
			}
		}
		b = append(b, TLV(tags[name], data)...)
	}
	if _, e := tlv.Parse(b, false); e != nil {
		return "", nil, e
	}
	msg, e := keys.SignatureInput(typ, b)
	if e != nil {
		return "", nil, e
	}
	b = append(b, TLV(0xff01, ed25519.Sign(private, msg))...)
	id, file, e := keys.Seal(b)
	return hex.EncodeToString(id), file, e
}

type decoded struct {
	id, scope string
	token     model.Update
	device    presentation.Update
	typ       byte
	parents   []string
}
type View struct {
	Tokens       model.View
	Presentation presentation.View
	Disposition  map[string]string
	Degraded     map[string]bool
	Incomplete   bool
	Complete     bool
	Future       map[string]byte
	Bypass       map[string]bool
	Pending      map[string]string
	Abandoned    map[string]bool
	objects      map[string]decoded
}

func (c *Client) decode(id string, b []byte) (decoded, string, byte) {
	raw, _ := hex.DecodeString(id)
	p, stage := c.keys.Open(raw, b, nil)
	d := decoded{id: id}
	if stage == "AUTHENTICATED_UNSUPPORTED_FUTURE" {
		return d, stage, p[4]
	}
	if stage != "STRUCTURALLY_VALID" {
		return d, stage, 0
	}
	o, e := tlv.Parse(p, true)
	if e != nil {
		return d, "STRUCTURALLY_INVALID", 0
	}
	f := map[uint16][]byte{}
	for _, v := range o.Fields {
		f[v.Tag] = v.Value
		if v.Tag == 5 {
			d.parents = append(d.parents, hex.EncodeToString(v.Value))
		}
	}
	msg, _ := c.keys.SignatureInput(o.Type, o.Unsigned)
	if ed25519profile.Verify(f[3], msg, f[0xff01]) != nil {
		return d, "SIGNATURE_INVALID", 0
	}
	d.typ = o.Type
	if o.Type == 1 {
		token := hex.EncodeToString(f[0x101])
		d.scope = "token:" + token
		fields := map[string]string{}
		for tag, name := range map[uint16]string{0x102: "STATUS", 0x103: "ISSUER", 0x104: "ACCOUNT", 0x105: "CREDENTIAL"} {
			v, ok := f[tag]
			if !ok {
				continue
			}
			value := string(v)
			if tag == 0x102 {
				value = "LIVE"
				if v[0] == 2 {
					value = "TOMBSTONE"
				}
			}
			if tag == 0x105 {
				value = hex.EncodeToString(v)
			}
			fields[name] = value
		}
		d.token = model.Update{ID: id, Token: token, Parents: d.parents, Fields: fields}
	} else {
		signer := hex.EncodeToString(f[3])
		d.scope = "device:" + signer
		d.device = presentation.Update{ID: id, Signer: signer, Parents: d.parents, Name: string(f[0x201]), Kind: "DEVICE_UPDATE"}
	}
	return d, "SIGNATURE_VERIFIED", 0
}
func terminal(stage string) bool {
	return stage == "INVALID_PREFIX" || stage == "STRUCTURALLY_INVALID" || stage == "SIGNATURE_INVALID" || stage == "INVALID"
}
func (c *Client) Scan() (View, error) {
	var v View
	e := c.Security.With(func(t localstate.Transaction) error { var e error; v, e = c.scan(t); return e })
	return v, e
}
func (c *Client) scan(t localstate.Transaction) (View, error) {
	v := View{Disposition: map[string]string{}, Degraded: map[string]bool{}, objects: map[string]decoded{}}
	s := t.State()
	if e := c.check(s); e != nil {
		return v, e
	}
	ids, e := c.Objects.ListCandidates()
	if e != nil {
		return v, e
	}
	all := map[string]bool{}
	for _, id := range ids {
		all[id] = true
	}
	if c.Cache != nil {
		cached, e := c.Cache.ListCandidates()
		if e != nil {
			return v, e
		}
		for _, id := range cached {
			all[id] = true
		}
	}
	futures := map[string]byte{}
	limit := c.MaxCandidates
	if limit <= 0 {
		limit = 4096
	}
	if len(all) > limit {
		v.Incomplete = true
		return v, errors.New("validation incomplete: candidate work limit")
	}
	files := map[string][]byte{}
	for id := range all {
		b, err := c.Objects.ReadObject(id)
		d, stage, version := c.decode(id, b)
		if err != nil {
			stage = "UNAVAILABLE"
		}
		if c.Cache != nil && stage != "SIGNATURE_VERIFIED" && stage != "AUTHENTICATED_UNSUPPORTED_FUTURE" {
			if cached, err := c.Cache.ReadObject(id); err == nil {
				cd, cs, cv := c.decode(id, cached)
				if cs == "SIGNATURE_VERIFIED" || cs == "AUTHENTICATED_UNSUPPORTED_FUTURE" {
					d, stage, version, b = cd, cs, cv, cached
				}
			}
		}
		v.Disposition[id] = stage
		if stage == "AUTHENTICATED_UNSUPPORTED_FUTURE" {
			futures[id] = version
		}
		if stage == "SIGNATURE_VERIFIED" {
			v.objects[id] = d
			files[id] = b
		}
	}
	updates := []model.Update{}
	devices := []presentation.Update{}
	// Identity-grounded invalid objects and wrong-kind parents are supplied as
	// invalid sentinels so dependents cannot be mistaken for missing dependencies.
	for id, stage := range v.Disposition {
		d, ok := v.objects[id]
		if ok && d.typ == 1 {
			updates = append(updates, d.token)
		} else if ok || terminal(stage) {
			updates = append(updates, model.Update{ID: id, Token: "invalid", Fields: map[string]string{}})
		}
		if ok && d.typ == 2 {
			devices = append(devices, d.device)
		} else if ok || terminal(stage) {
			devices = append(devices, presentation.Update{ID: id, Signer: "invalid", Kind: "INVALID"})
		}
	}
	v.Tokens, e = model.Evaluate(updates)
	if e != nil {
		return v, e
	}
	oracle, e := referenceoracle.Evaluate(updates)
	if e != nil || !reflect.DeepEqual(v.Tokens, oracle) {
		return v, errors.New("token oracle disagreement")
	}
	v.Presentation, e = presentation.Evaluate(devices)
	if e != nil {
		return v, e
	}
	for id, d := range v.objects {
		if d.typ == 1 {
			v.Disposition[id] = v.Tokens.Validation[id]
		} else {
			v.Disposition[id] = v.Presentation.Validation[id]
		}
	}
	changed := false
	for id, version := range futures {
		if old, ok := s.Future[id]; !ok || old != version {
			s.Future[id] = version
			changed = true
		}
	}
	for id, d := range v.objects {
		if v.Disposition[id] == "PENDING" && s.Pending[id] != d.scope {
			s.Pending[id] = d.scope
			changed = true
		}
	}
	if changed {
		if e = t.Commit("observations"); e != nil {
			return v, e
		}
	}
	heads := map[string][]string{}
	for token, state := range v.Tokens.Tokens {
		heads["token:"+token] = state.Heads
	}
	for signer, hs := range v.Presentation.Heads {
		heads["device:"+signer] = hs
	}
	// Keep any older head that cannot be proven covered by currently valid history.
	changed = false
	for scope, old := range s.Heads {
		for _, x := range old {
			covered := false
			for _, h := range heads[scope] {
				covered = covered || v.before(x, h)
			}
			if !covered {
				v.Degraded[scope] = true
			}
		}
	}
	for scope, hs := range heads {
		combined := append([]string{}, hs...)
		for _, old := range s.Heads[scope] {
			covered := false
			for _, h := range hs {
				covered = covered || v.before(old, h)
			}
			if !covered {
				combined = append(combined, old)
			}
		}
		sort.Strings(combined)
		if !reflect.DeepEqual(s.Heads[scope], combined) {
			s.Heads[scope] = combined
			changed = true
		}
	}
	if c.Cache != nil {
		for id, b := range files {
			if v.Disposition[id] == "FULLY_VALID" {
				if e = c.Cache.PublishImmutable(id, b); e != nil {
					return v, e
				}
			}
		}
	}
	if changed {
		if e = t.Commit("frontier"); e != nil {
			return v, e
		}
	}
	// Removal is a separate transaction, deliberately after the new evidence.
	changed = false
	for id := range s.Pending {
		stage := v.Disposition[id]
		if stage == "FULLY_VALID" || terminal(stage) {
			if terminal(stage) {
				s.Terminal[id] = true
			}
			delete(s.Pending, id)
			delete(s.Abandoned, id)
			delete(s.Acknowledged, id)
			changed = true
		}
	}
	if changed {
		if e = t.Commit("handoff"); e != nil {
			return v, e
		}
	}
	// Invalid/unreadable path representations do not gain security authority.
	// Previously authenticated evidence is handled by its own scope above.
	v.Future, v.Bypass, v.Pending, v.Abandoned = s.Future, s.Bypass, s.Pending, s.Abandoned
	v.Complete = !v.Incomplete && len(s.Future) == 0 && len(s.Pending) == 0 && len(v.Degraded) == 0
	return v, nil
}
func (v View) before(a, b string) bool {
	seen := map[string]bool{}
	todo := []string{b}
	for len(todo) > 0 {
		x := todo[0]
		todo = todo[1:]
		if x == a {
			return true
		}
		if seen[x] {
			continue
		}
		seen[x] = true
		todo = append(todo, v.objects[x].parents...)
	}
	return false
}
func gate(s *localstate.State, v View, scope string) error {
	if v.Incomplete {
		return errors.New("validation incomplete")
	}
	if v.Degraded[scope] {
		return errors.New("known-history regression")
	}
	for id := range s.Future {
		if !s.Bypass[id] {
			return errors.New("future observation blocks writers")
		}
	}
	for id, p := range s.Pending {
		if p == scope && !s.Abandoned[id] {
			return errors.New("pending history blocks writer")
		}
	}
	if len(s.Heads[scope]) > 32 {
		return errors.New("capacity blocked")
	}
	return nil
}
func (c *Client) Recovery(ids []string, action string) error {
	return c.Security.With(func(t localstate.Transaction) error {
		s := t.State()
		if e := c.check(s); e != nil {
			return e
		}
		for _, id := range ids {
			switch action {
			case "abandon", "acknowledge":
				if _, ok := s.Pending[id]; !ok {
					return errors.New("unknown pending ID")
				}
				if action == "abandon" {
					s.Abandoned[id] = true
				} else {
					s.Acknowledged[id] = true
				}
			case "bypass":
				if _, ok := s.Future[id]; !ok {
					return errors.New("unknown future ID")
				}
				s.Bypass[id] = true
			default:
				return errors.New("unknown recovery action")
			}
		}
		return t.Commit("recovery-" + action)
	})
}

type Operation struct {
	scope                 string
	fields                map[string]string
	parents               []string
	generation            uint64
	session               string
	fingerprint           [32]byte
	resolution, confirmed bool
}

func (c *Client) Prepare(scope string, fields map[string]string) (*Operation, error) {
	var op *Operation
	e := c.Security.With(func(t localstate.Transaction) error {
		v, e := c.scan(t)
		if e != nil {
			return e
		}
		s := t.State()
		if e = gate(s, v, scope); e != nil {
			return e
		}
		op = &Operation{scope: scope, fields: map[string]string{}, parents: append([]string{}, s.Heads[scope]...), generation: s.Generation, session: c.session}
		for k, v := range fields {
			op.fields[k] = v
		}
		if strings.HasPrefix(scope, "token:") {
			op.resolution = len(v.Tokens.Tokens[strings.TrimPrefix(scope, "token:")].Conflicts) > 0 && fields["STATUS"] != ""
		} else if scope != "device:"+hex.EncodeToString(c.device.Private.Public().(ed25519.PublicKey)) {
			return errors.New("wrong presentation signer")
		}
		data, _ := json.Marshal(struct {
			View   View
			Scope  string
			Fields map[string]string
		}{v, scope, fields})
		op.fingerprint = sha256.Sum256(data)
		return nil
	})
	return op, e
}

// Confirm represents explicit user confirmation of the prepared resolution and
// its desired STATUS. Operation internals cannot be mutated by callers.
func (op *Operation) Confirm() { op.confirmed = true }
func (c *Client) Publish(op *Operation) (string, error) {
	id := ""
	if op == nil {
		return id, errors.New("nil operation")
	}
	e := c.Security.With(func(t localstate.Transaction) error {
		v, e := c.scan(t)
		if e != nil {
			return e
		}
		s := t.State()
		if e = gate(s, v, op.scope); e != nil {
			return e
		}
		data, _ := json.Marshal(struct {
			View   View
			Scope  string
			Fields map[string]string
		}{v, op.scope, op.fields})
		if op.session != c.session || op.generation != s.Generation || op.fingerprint != sha256.Sum256(data) {
			return ErrStale
		}
		if op.resolution && !op.confirmed {
			return errors.New("explicit lifecycle confirmation required")
		}
		typ := byte(1)
		token := strings.TrimPrefix(op.scope, "token:")
		if strings.HasPrefix(op.scope, "device:") {
			typ = 2
			token = ""
		}
		var b []byte
		id, b, e = Encode(c.keys, c.device.Private, typ, token, op.parents, op.fields)
		if e != nil {
			return e
		}
		d, stage, _ := c.decode(id, b)
		if stage != "SIGNATURE_VERIFIED" {
			return errors.New(stage)
		}
		if typ == 1 {
			us := []model.Update{d.token}
			for _, o := range v.objects {
				if o.typ == 1 {
					us = append(us, o.token)
				}
			}
			check, e := model.Evaluate(us)
			if e != nil || check.Validation[id] != "FULLY_VALID" {
				return errors.New("invalid semantic operation")
			}
		}
		// Linearization: this successful final check and immutable installation are
		// inside the same cross-process lock as all observation acceptance/recovery.
		if e = c.Hook.Hit("publish.before-install"); e != nil {
			return e
		}
		if e = c.Objects.PublishImmutable(id, b); e != nil {
			return e
		}
		if e = c.Hook.Hit("publish.after-install"); e != nil {
			return e
		}
		// Publication success additionally requires persisted accepted evidence. Read
		// back through the full pipeline; namespace substitution cannot earn success.
		result, e := c.scan(t)
		if e != nil {
			return e
		}
		if result.Disposition[id] != "FULLY_VALID" {
			return errors.New("published object no longer reconstructible")
		}
		return nil
	})
	return id, e
}
func (c *Client) CreateToken(fields map[string]string) (string, string, error) {
	token, e := RandomID()
	if e != nil {
		return "", "", e
	}
	op, e := c.Prepare("token:"+token, fields)
	if e != nil {
		return "", "", e
	}
	id, e := c.Publish(op)
	return token, id, e
}
func (c *Client) Migrate(destination *Client, password []byte, fields map[string]string) (string, string, error) {
	if fields["STATUS"] != "LIVE" || len(fields) != 4 {
		return "", "", errors.New("migration requires locally selected complete LIVE creation values")
	}
	// Source need not publish an unencodable resolution. No source write occurs.
	if e := destination.Create(password); e != nil {
		return "", "", e
	}
	return destination.CreateToken(fields)
}
func (c *Client) WriterEligibility(scope string) error {
	return c.Security.With(func(t localstate.Transaction) error {
		v, e := c.scan(t)
		if e != nil {
			return e
		}
		return gate(t.State(), v, scope)
	})
}
func (c *Client) String() string { return fmt.Sprintf("Totipo reference client at %s", c.Location) }

// AssessUse keeps ordinary OTP/export eligibility separate from object validity.
// Future bypass and pending abandonment permit selected writers, but cannot make
// the unresolved history a complete ordinary current-token view.
func (c *Client) AssessUse(token string) (model.UseDecision, error) {
	var decision model.UseDecision
	err := c.Security.With(func(t localstate.Transaction) error {
		v, e := c.scan(t)
		if e != nil {
			return e
		}
		state := v.Tokens.Tokens[token]
		scope := "token:" + token
		values := func(field string) []string {
			out := []string{}
			for _, id := range state.Fields[field] {
				out = append(out, v.objects[id].token.Fields[field])
			}
			return out
		}
		incomplete := v.Incomplete || len(t.State().Future) > 0
		for _, pending := range t.State().Pending {
			if pending == scope {
				incomplete = true
			}
		}
		distinct := func(xs []string) int {
			m := map[string]bool{}
			for _, x := range xs {
				m[x] = true
			}
			return len(m)
		}
		decision = model.AssessUse(model.UseProfile{Degraded: v.Degraded[scope], Incomplete: incomplete, LifecycleConflict: len(state.Conflicts) > 0, StatusValues: values("STATUS"), CredentialValues: values("CREDENTIAL"), CredentialComplete: len(state.Fields["CREDENTIAL"]) > 0, IssuerConflict: distinct(values("ISSUER")) > 1, AccountConflict: distinct(values("ACCOUNT")) > 1})
		return nil
	})
	return decision, err
}
