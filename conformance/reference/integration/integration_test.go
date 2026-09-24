//go:build linux

package integration

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
	"time"
	"totipo/conformance/internal/bootstrap"
	"totipo/conformance/internal/objectcrypto"
	"totipo/conformance/reference/client"
	"totipo/conformance/reference/localstate"
	"totipo/conformance/reference/storage"
)

const password = "reference-test-password"

type fixture struct {
	vault, local string
	store        *storage.Linux
	state        *localstate.Files
	c            *client.Client
}

func dirs(t *testing.T) (string, string) {
	t.Helper()
	base := t.TempDir()
	v, l := filepath.Join(base, "vault"), filepath.Join(base, "local")
	must(t, os.Mkdir(v, 0700))
	must(t, os.Mkdir(l, 0700))
	return v, l
}
func connect(t *testing.T, v, l string) *fixture {
	t.Helper()
	s, e := storage.Open(v)
	must(t, e)
	db, e := localstate.Open(l)
	must(t, e)
	t.Cleanup(func() { s.Close(); db.Close() })
	return &fixture{v, l, s, db, client.New(s, db, v)}
}
func fresh(t *testing.T) *fixture {
	v, l := dirs(t)
	f := connect(t, v, l)
	must(t, f.c.Create([]byte(password)))
	return f
}
func restart(t *testing.T, f *fixture) *fixture {
	g := connect(t, f.vault, f.local)
	must(t, g.c.Open([]byte(password), nil))
	return g
}
func must(t *testing.T, e error) {
	t.Helper()
	if e != nil {
		t.Fatal(e)
	}
}
func snapshot(t *testing.T, f *fixture) localstate.State {
	t.Helper()
	var s localstate.State
	must(t, f.state.With(func(tx localstate.Transaction) error {
		b, e := json.Marshal(tx.State())
		if e != nil {
			return e
		}
		return json.Unmarshal(b, &s)
	}))
	return s
}
func fields(t *testing.T) map[string]string {
	t.Helper()
	b, e := client.NewCredential()
	must(t, e)
	return map[string]string{"STATUS": "LIVE", "ISSUER": "Example", "ACCOUNT": "alice", "CREDENTIAL": hex.EncodeToString(b)}
}
func keys(t *testing.T, f *fixture) objectcrypto.Keys {
	t.Helper()
	b, e := f.store.ReadVault()
	must(t, e)
	r, stage := bootstrap.Unwrap(b, []byte(password), nil)
	if stage != "VERIFIED" {
		t.Fatal(stage)
	}
	k, e := objectcrypto.Derive(r)
	must(t, e)
	return k
}
func encode(t *testing.T, f *fixture, typ byte, token string, parents []string, fields map[string]string) (string, []byte) {
	t.Helper()
	id, b, e := client.Encode(keys(t, f), f.c.Device().Private, typ, token, parents, fields)
	must(t, e)
	return id, b
}
func raw(t *testing.T, f *fixture, id string, b []byte) {
	t.Helper()
	must(t, os.WriteFile(filepath.Join(f.vault, "objects", id), b, 0600))
}
func newID(t *testing.T) string { t.Helper(); id, e := client.RandomID(); must(t, e); return id }
func token(t *testing.T, f *fixture) (string, string) {
	t.Helper()
	tok, id, e := f.c.CreateToken(fields(t))
	must(t, e)
	return tok, id
}
func rejected(t *testing.T, e error) {
	t.Helper()
	if e == nil {
		t.Fatal("expected rejection")
	}
}

func TestObjectIngestPendingRestart(t *testing.T) {
	f := fresh(t)
	tok := newID(t)
	parent, pb := encode(t, f, 1, tok, nil, fields(t))
	child, cb := encode(t, f, 1, tok, []string{parent}, map[string]string{"ISSUER": "later"})
	raw(t, f, child, cb)
	v, e := f.c.Scan()
	must(t, e)
	if v.Disposition[child] != "PENDING" {
		t.Fatal(v.Disposition)
	}
	if snapshot(t, f).Pending[child] != "token:"+tok {
		t.Fatal("pending evidence missing")
	}
	must(t, os.Remove(filepath.Join(f.vault, "objects", child)))
	g := restart(t, f)
	rejected(t, g.c.WriterEligibility("token:"+tok))
	for _, bad := range [][]byte{{1}, bytes.Repeat([]byte{1}, 2048)} {
		raw(t, g, child, bad)
		_, e = g.c.Scan()
		must(t, e)
		if snapshot(t, g).Pending[child] == "" {
			t.Fatal("pathname failure erased evidence")
		}
	}
	raw(t, g, child, cb)
	raw(t, g, parent, pb)
	v, e = g.c.Scan()
	must(t, e)
	if v.Disposition[child] != "FULLY_VALID" {
		t.Fatal(v.Disposition)
	}
	s := snapshot(t, g)
	if len(s.Pending) != 0 || !reflect.DeepEqual(s.Heads["token:"+tok], []string{child}) {
		t.Fatal(s)
	}
	h := restart(t, g)
	after, e := h.c.Scan()
	must(t, e)
	if !reflect.DeepEqual(v.Tokens, after.Tokens) {
		t.Fatal("restart differs")
	}
}
func TestRecoveryAbandonmentAndTerminal(t *testing.T) {
	f := fresh(t)
	tok, root := token(t, f)
	missing := newID(t)
	id, b := encode(t, f, 1, tok, []string{missing}, map[string]string{"ACCOUNT": "pending"})
	raw(t, f, id, b)
	_, e := f.c.Scan()
	must(t, e)
	rejected(t, f.c.WriterEligibility("token:"+tok))
	must(t, f.c.Recovery([]string{id}, "acknowledge"))
	rejected(t, f.c.WriterEligibility("token:"+tok))
	must(t, f.c.Recovery([]string{id}, "abandon"))
	must(t, os.Remove(filepath.Join(f.vault, "objects", id)))
	g := restart(t, f)
	must(t, g.c.WriterEligibility("token:"+tok))
	s := snapshot(t, g)
	if !s.Abandoned[id] || !s.Acknowledged[id] || s.Pending[id] == "" {
		t.Fatal("recovery lost evidence")
	}
	must(t, os.Remove(filepath.Join(f.vault, "objects", root)))
	rejected(t, g.c.WriterEligibility("token:"+tok))
	// A signed descendant of a deterministically invalid, identity-matching root
	// is terminal. Merely bad current path bytes above were not terminal.
	f2 := fresh(t)
	tok2 := newID(t)
	bad, bb := encode(t, f2, 1, tok2, nil, map[string]string{"STATUS": "TOMBSTONE"})
	dep, db := encode(t, f2, 1, tok2, []string{bad}, map[string]string{"ACCOUNT": "x"})
	raw(t, f2, dep, db)
	_, e = f2.c.Scan()
	must(t, e)
	raw(t, f2, bad, bb)
	v, e := f2.c.Scan()
	must(t, e)
	s = snapshot(t, f2)
	if v.Disposition[dep] != "INVALID" || !s.Terminal[dep] || s.Pending[dep] != "" {
		t.Fatal("terminal handoff", v.Disposition, s)
	}
}
func TestFutureEvidenceAndPrefix(t *testing.T) {
	f := fresh(t)
	tok, _ := token(t, f)
	k := keys(t, f)
	p := append(append(client.TLV(1, []byte{9}), client.TLV(2, []byte{255})...), []byte{255, 0, 1}...)
	id, b, e := k.Seal(p)
	must(t, e)
	name := hex.EncodeToString(id)
	raw(t, f, name, b)
	v, e := f.c.Scan()
	must(t, e)
	if v.Disposition[name] != "AUTHENTICATED_UNSUPPORTED_FUTURE" {
		t.Fatal(v.Disposition)
	}
	rejected(t, f.c.WriterEligibility("token:"+tok))
	must(t, os.Remove(filepath.Join(f.vault, "objects", name)))
	g := restart(t, f)
	rejected(t, g.c.WriterEligibility("token:"+tok))
	must(t, g.c.Recovery([]string{name}, "bypass"))
	h := restart(t, g)
	must(t, h.c.WriterEligibility("token:"+tok))
	s := snapshot(t, h)
	if s.Future[name] != 9 || !s.Bypass[name] {
		t.Fatal(s)
	}
	bad := append(client.TLV(1, []byte{9}), 0, 2, 0, 3, 1)
	bid, bb, e := k.Seal(bad)
	must(t, e)
	bn := hex.EncodeToString(bid)
	raw(t, h, bn, bb)
	v, e = h.c.Scan()
	must(t, e)
	if v.Disposition[bn] != "INVALID_PREFIX" || snapshot(t, h).Future[bn] != 0 {
		t.Fatal("malformed prefix reached future dispatch")
	}
}
func TestRestartRegressionAndRetainedCopies(t *testing.T) {
	for _, retain := range []bool{false, true} {
		t.Run(fmt.Sprint(retain), func(t *testing.T) {
			f := fresh(t)
			cv, cl := dirs(t)
			cache := connect(t, cv, cl)
			if retain {
				f.c.Cache = cache.store
			}
			tok, root := token(t, f)
			op, e := f.c.Prepare("token:"+tok, map[string]string{"ACCOUNT": "second"})
			must(t, e)
			head, e := f.c.Publish(op)
			must(t, e)
			before, e := f.c.Scan()
			must(t, e)
			must(t, os.Remove(filepath.Join(f.vault, "objects", root)))
			must(t, os.Remove(filepath.Join(f.vault, "objects", head)))
			g := restart(t, f)
			if retain {
				g.c.Cache = cache.store
			}
			after, e := g.c.Scan()
			must(t, e)
			if retain {
				if !reflect.DeepEqual(before.Tokens, after.Tokens) {
					t.Fatal("retained bytes failed reconstruction")
				}
				must(t, g.c.WriterEligibility("token:"+tok))
			} else {
				if !after.Degraded["token:"+tok] {
					t.Fatal("missing regression")
				}
				rejected(t, g.c.WriterEligibility("token:"+tok))
				if !reflect.DeepEqual(snapshot(t, g).Heads["token:"+tok], []string{head}) {
					t.Fatal("lost remembered frontier")
				}
			}
		})
	}
}
func TestPresentationForkAndRecovery(t *testing.T) {
	f := fresh(t)
	dev := f.c.Device()
	signer := hex.EncodeToString(dev.Private.Public().(ed25519.PublicKey))
	scope := "device:" + signer
	root, rb := encode(t, f, 2, "", nil, map[string]string{"DISPLAY_NAME": "phone"})
	a, ab := encode(t, f, 2, "", []string{root}, map[string]string{"DISPLAY_NAME": "one"})
	b, bb := encode(t, f, 2, "", []string{root}, map[string]string{"DISPLAY_NAME": "two"})
	raw(t, f, a, ab)
	_, e := f.c.Scan()
	must(t, e)
	rejected(t, f.c.WriterEligibility(scope))
	must(t, f.c.Recovery([]string{a}, "abandon"))
	must(t, f.c.WriterEligibility(scope))
	tok, _ := token(t, f)
	must(t, f.c.WriterEligibility("token:"+tok))
	raw(t, f, root, rb)
	raw(t, f, b, bb)
	v, e := f.c.Scan()
	must(t, e)
	if len(v.Presentation.Heads[signer]) != 2 || len(snapshot(t, f).Pending) != 0 {
		t.Fatal(v)
	}
	g := connect(t, f.vault, f.local)
	must(t, g.c.Open([]byte(password), &dev))
	v, e = g.c.Scan()
	must(t, e)
	if len(v.Presentation.Heads[signer]) != 2 {
		t.Fatal("fork lost")
	}
	must(t, os.Remove(filepath.Join(f.vault, "objects", b)))
	v, e = g.c.Scan()
	must(t, e)
	if !v.Degraded[scope] {
		t.Fatal("presentation rollback lost")
	}
	must(t, g.c.WriterEligibility("token:"+tok))
}
func TestVaultCreateRewrapAndRootPin(t *testing.T) {
	f := fresh(t)
	old, e := f.store.ReadVault()
	must(t, e)
	binding := snapshot(t, f).Binding
	must(t, f.c.Rewrap([]byte("new-password")))
	newBytes, e := f.store.ReadVault()
	must(t, e)
	if bytes.Equal(old, newBytes) {
		t.Fatal("rewrap reused bytes")
	}
	g := connect(t, f.vault, f.local)
	rejected(t, g.c.Open([]byte(password), nil))
	must(t, g.c.Open([]byte("new-password"), nil))
	if snapshot(t, g).Binding != binding {
		t.Fatal("changed binding")
	}
	other := fresh(t)
	sub, e := other.store.ReadVault()
	must(t, e)
	must(t, os.WriteFile(filepath.Join(f.vault, "VAULT"), sub, 0600))
	h := connect(t, f.vault, f.local)
	rejected(t, h.c.Open([]byte(password), nil))
	if snapshot(t, h).Binding != binding {
		t.Fatal("root substitution")
	}
	// Existing protocol candidates forbid silent reinitialization even without VAULT.
	v, l := dirs(t)
	z := connect(t, v, l)
	raw(t, z, newID(t), []byte{1})
	rejected(t, z.c.Create([]byte(password)))
	if snapshot(t, z).Binding != "" {
		t.Fatal("initialized over recoverable objects")
	}
}
func TestLocalDatabaseRollbackLimitation(t *testing.T) {
	f := fresh(t)
	backup, e := os.ReadFile(filepath.Join(f.local, "state.json"))
	must(t, e)
	k := keys(t, f)
	p := append(client.TLV(1, []byte{2}), client.TLV(2, []byte{2})...)
	id, b, e := k.Seal(p)
	must(t, e)
	name := hex.EncodeToString(id)
	raw(t, f, name, b)
	_, e = f.c.Scan()
	must(t, e)
	if len(snapshot(t, f).Future) != 1 {
		t.Fatal("missing S1 evidence")
	}
	must(t, os.Remove(filepath.Join(f.vault, "objects", name)))
	must(t, os.WriteFile(filepath.Join(f.local, "state.json"), backup, 0600))
	g := restart(t, f)
	if len(snapshot(t, g).Future) != 0 {
		t.Fatal("test expected old backup to lose evidence")
	}
	t.Log("external local database rollback loses prior evidence; no trusted continuity mechanism is implemented")
}
func TestFreshnessAndLinearization(t *testing.T) {
	for _, change := range []string{"head", "pending", "future", "restart"} {
		t.Run(change, func(t *testing.T) {
			f := fresh(t)
			tok, root := token(t, f)
			// Build deletion versus concurrent credential rotation: a real lifecycle conflict.
			del, db := encode(t, f, 1, tok, []string{root}, map[string]string{"STATUS": "TOMBSTONE"})
			cred, e := client.NewCredential()
			must(t, e)
			rot, rb := encode(t, f, 1, tok, []string{root}, map[string]string{"CREDENTIAL": hex.EncodeToString(cred)})
			raw(t, f, del, db)
			raw(t, f, rot, rb)
			op, e := f.c.Prepare("token:"+tok, map[string]string{"STATUS": "LIVE"})
			must(t, e)
			op.Confirm()
			switch change {
			case "head":
				id, b := encode(t, f, 1, tok, []string{del, rot}, map[string]string{"ISSUER": "new"})
				raw(t, f, id, b)
			case "pending":
				id, b := encode(t, f, 1, tok, []string{newID(t)}, map[string]string{"ACCOUNT": "pending"})
				raw(t, f, id, b)
			case "future":
				id, b, e := keys(t, f).Seal(append(client.TLV(1, []byte{2}), client.TLV(2, []byte{9})...))
				must(t, e)
				name := hex.EncodeToString(id)
				raw(t, f, name, b)
				_, e = f.c.Scan()
				must(t, e)
				must(t, f.c.Recovery([]string{name}, "bypass"))
				must(t, os.Remove(filepath.Join(f.vault, "objects", name)))
			case "restart":
				f = restart(t, f)
			}
			before, e := f.store.ListCandidates()
			must(t, e)
			_, e = f.c.Publish(op)
			rejected(t, e)
			after, e := f.store.ListCandidates()
			must(t, e)
			if !reflect.DeepEqual(before, after) {
				t.Fatal("stale operation published")
			}
		})
	}
	t.Run("after-linearization", func(t *testing.T) {
		f := fresh(t)
		tok, root := token(t, f)
		op, e := f.c.Prepare("token:"+tok, map[string]string{"ACCOUNT": "local"})
		must(t, e)
		id, e := f.c.Publish(op)
		must(t, e)
		remote, b := encode(t, f, 1, tok, []string{root}, map[string]string{"ACCOUNT": "remote"})
		raw(t, f, remote, b)
		v, e := f.c.Scan()
		must(t, e)
		if v.Disposition[id] != "FULLY_VALID" || len(v.Tokens.Tokens[tok].Fields["ACCOUNT"]) != 2 {
			t.Fatal("later fork invalidated publication")
		}
	})
}
func TestMigrationAndRandomGeneration(t *testing.T) {
	f := fresh(t)
	tok, root := token(t, f)
	k := keys(t, f)
	for i := 0; i < 33; i++ {
		id, b, e := client.Encode(k, f.c.Device().Private, 1, tok, []string{root}, map[string]string{"ACCOUNT": fmt.Sprint(i)})
		must(t, e)
		raw(t, f, id, b)
	}
	v, e := f.c.Scan()
	must(t, e)
	if len(v.Tokens.Tokens[tok].Heads) != 33 {
		t.Fatal("heads truncated")
	}
	rejected(t, f.c.WriterEligibility("token:"+tok))
	before, e := f.store.ListCandidates()
	must(t, e)
	dv, dl := dirs(t)
	dest := connect(t, dv, dl)
	newToken, newRoot, e := f.c.Migrate(dest.c, []byte(password), fields(t))
	must(t, e)
	if newToken == tok || bytes.Equal(dest.c.Device().Private, f.c.Device().Private) || snapshot(t, dest).Binding == snapshot(t, f).Binding {
		t.Fatal("identities reused")
	}
	after, e := f.store.ListCandidates()
	must(t, e)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("migration changed source")
	}
	nv, e := dest.c.Scan()
	must(t, e)
	if nv.Disposition[newRoot] != "FULLY_VALID" || len(nv.Tokens.Tokens[newToken].Fields) != 4 {
		t.Fatal(nv)
	}
	b, e := dest.store.ReadVault()
	must(t, e)
	if len(b) != 87 || len(b[11:27]) != 16 || len(b[27:39]) != 12 {
		t.Fatal("bootstrap generation lengths")
	}
	credential, e := client.NewCredential()
	must(t, e)
	if len(credential) < 20 {
		t.Fatal("credential too short")
	}
}

// All process tests invoke this binary with a single helper test. Exit(86)
// intentionally bypasses all defers and discards the entire client/Go heap.
func TestProcessHelper(t *testing.T) {
	mode := os.Getenv("TOTIPO_HELPER")
	if mode == "" {
		return
	}
	v, l := os.Getenv("TOTIPO_VAULT"), os.Getenv("TOTIPO_LOCAL")
	s, e := storage.Open(v)
	if e != nil {
		t.Fatal(e)
	}
	db, e := localstate.Open(l)
	if e != nil {
		t.Fatal(e)
	}
	c := client.New(s, db, v)
	point := os.Getenv("TOTIPO_CRASH")
	hook := func(p string) error {
		if p == point {
			os.Exit(86)
		}
		if p == os.Getenv("TOTIPO_HOLD_POINT") {
			must(t, os.WriteFile(os.Getenv("TOTIPO_HOLD_READY"), nil, 0600))
			waitReady(t, os.Getenv("TOTIPO_HOLD_RELEASE"))
		}
		return nil
	}
	s.Hook = hook
	db.Hook = hook
	c.Hook = hook
	if ready := os.Getenv("TOTIPO_READY"); ready != "" { // Both creators inspect absence before the release barrier.
		_, _ = s.ReadVault()
		must(t, os.WriteFile(ready, []byte("ready"), 0600))
		until := time.Now().Add(10 * time.Second)
		for {
			if _, e = os.Stat(os.Getenv("TOTIPO_START")); e == nil {
				break
			}
			if time.Now().After(until) {
				t.Fatal("barrier timeout")
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
	switch mode {
	case "create":
		e = c.Create([]byte(password))
	case "raw":
		b, err := os.ReadFile(os.Getenv("TOTIPO_INPUT"))
		must(t, err)
		e = s.PublishImmutable(os.Getenv("TOTIPO_ID"), b)
	default:
		e = c.Open([]byte(password), nil)
		if e != nil {
			break
		}
		switch mode {
		case "scan":
			_, e = c.Scan()
		case "publish":
			op, err := c.Prepare(os.Getenv("TOTIPO_SCOPE"), map[string]string{"ACCOUNT": "process"})
			if err != nil {
				e = err
			} else {
				_, e = c.Publish(op)
			}
		case "rewrap":
			e = c.Rewrap([]byte("new-password"))
		case "abandon", "bypass":
			e = c.Recovery([]string{os.Getenv("TOTIPO_ID")}, mode)
		default:
			e = errors.New("unknown helper")
		}
	}
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(23)
	}
	os.Exit(0)
}
func command(t *testing.T, f *fixture, mode, point string, extra ...string) *exec.Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestProcessHelper$")
	cmd.Env = append(os.Environ(), "TOTIPO_HELPER="+mode, "TOTIPO_CRASH="+point, "TOTIPO_VAULT="+f.vault, "TOTIPO_LOCAL="+f.local)
	cmd.Env = append(cmd.Env, extra...)
	return cmd
}
func crashed(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	out, e := cmd.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(e, &exit) || exit.ExitCode() != 86 {
		t.Fatalf("expected process death at hook, got %v: %s", e, out)
	}
}
func waitReady(t *testing.T, paths ...string) {
	t.Helper()
	end := time.Now().Add(10 * time.Second)
	for _, p := range paths {
		for {
			if _, e := os.Stat(p); e == nil {
				break
			}
			if time.Now().After(end) {
				t.Fatal("process ready timeout")
			}
			time.Sleep(5 * time.Millisecond)
		}
	}
}
func TestProcessEstablishmentRace(t *testing.T) {
	for round := 0; round < 3; round++ {
		t.Run(fmt.Sprint(round), func(t *testing.T) {
			v, l := dirs(t)
			f := connect(t, v, l)
			base := t.TempDir()
			r1, r2, start := filepath.Join(base, "r1"), filepath.Join(base, "r2"), filepath.Join(base, "start")
			a := command(t, f, "create", "", "TOTIPO_READY="+r1, "TOTIPO_START="+start)
			b := command(t, f, "create", "", "TOTIPO_READY="+r2, "TOTIPO_START="+start)
			must(t, a.Start())
			must(t, b.Start())
			waitReady(t, r1, r2)
			must(t, os.WriteFile(start, nil, 0600))
			ae, be := a.Wait(), b.Wait()
			if (ae == nil) == (be == nil) {
				t.Fatalf("expected one winner: %v / %v", ae, be)
			}
			g := restart(t, f)
			if snapshot(t, g).Establishment != "ESTABLISHED" {
				t.Fatal("winner not durable")
			}
		})
	}
}
func TestProcessObjectPublicationRace(t *testing.T) {
	for _, scenario := range []string{"identical", "differing", "independent"} {
		t.Run(scenario, func(t *testing.T) {
			f := fresh(t)
			base := t.TempDir()
			id := newID(t)
			other := id
			aBytes := bytes.Repeat([]byte{1}, 2048)
			bBytes := append([]byte{}, aBytes...)
			if scenario == "differing" {
				bBytes[0] = 2
			}
			if scenario == "independent" {
				other = newID(t)
			}
			p1, p2 := filepath.Join(base, "one"), filepath.Join(base, "two")
			must(t, os.WriteFile(p1, aBytes, 0600))
			must(t, os.WriteFile(p2, bBytes, 0600))
			r1, r2, start := filepath.Join(base, "r1"), filepath.Join(base, "r2"), filepath.Join(base, "start")
			a := command(t, f, "raw", "", "TOTIPO_INPUT="+p1, "TOTIPO_ID="+id, "TOTIPO_READY="+r1, "TOTIPO_START="+start)
			b := command(t, f, "raw", "", "TOTIPO_INPUT="+p2, "TOTIPO_ID="+other, "TOTIPO_READY="+r2, "TOTIPO_START="+start)
			must(t, a.Start())
			must(t, b.Start())
			waitReady(t, r1, r2)
			must(t, os.WriteFile(start, nil, 0600))
			ae, be := a.Wait(), b.Wait()
			if scenario == "differing" {
				if (ae == nil) == (be == nil) {
					t.Fatalf("expected conflict loser %v %v", ae, be)
				}
			} else {
				must(t, ae)
				must(t, be)
			}
			got, e := f.store.ReadObject(id)
			must(t, e)
			if !bytes.Equal(got, aBytes) && !bytes.Equal(got, bBytes) {
				t.Fatal("corrupt publication")
			}
		})
	}
}

func TestProcessWriterSerialization(t *testing.T) {
	f := fresh(t)
	tok, _ := token(t, f)
	base := t.TempDir()
	r1, r2, start := filepath.Join(base, "r1"), filepath.Join(base, "r2"), filepath.Join(base, "start")
	a := command(t, f, "publish", "", "TOTIPO_SCOPE=token:"+tok, "TOTIPO_READY="+r1, "TOTIPO_START="+start)
	b := command(t, f, "publish", "", "TOTIPO_SCOPE=token:"+tok, "TOTIPO_READY="+r2, "TOTIPO_START="+start)
	must(t, a.Start())
	must(t, b.Start())
	waitReady(t, r1, r2)
	must(t, os.WriteFile(start, nil, 0600))
	ae, be := a.Wait(), b.Wait()
	if ae != nil && be != nil {
		t.Fatalf("both writers failed: %v %v", ae, be)
	}
	g := restart(t, f)
	v, e := g.c.Scan()
	must(t, e)
	if len(v.Tokens.Tokens[tok].Heads) != 1 {
		t.Fatal("local writers forked same context")
	}
	t.Log("two local processes: serialize or stale-retry; no avoidable fork")
}
func TestProcessObservationLinearization(t *testing.T) {
	f := fresh(t)
	tok, _ := token(t, f)
	k := keys(t, f)
	rawID, future, e := k.Seal(append(client.TLV(1, []byte{8}), client.TLV(2, []byte{9})...))
	must(t, e)
	id := hex.EncodeToString(rawID)
	base := t.TempDir()
	ready, release := filepath.Join(base, "ready"), filepath.Join(base, "release")
	writer := command(t, f, "publish", "", "TOTIPO_SCOPE=token:"+tok, "TOTIPO_HOLD_POINT=publish.before-install", "TOTIPO_HOLD_READY="+ready, "TOTIPO_HOLD_RELEASE="+release)
	must(t, writer.Start())
	waitReady(t, ready)
	// Remote bytes can arrive while the writer holds the gate lock, but another
	// local process cannot accept their evidence before the linearized write.
	raw(t, f, id, future)
	scanner := command(t, f, "scan", "")
	must(t, scanner.Start())
	done := make(chan error, 1)
	go func() { done <- scanner.Wait() }()
	select {
	case err := <-done:
		t.Fatalf("observation acceptance escaped writer lock: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	must(t, os.WriteFile(release, nil, 0600))
	must(t, writer.Wait())
	must(t, <-done)
	g := restart(t, f)
	v, e := g.c.Scan()
	must(t, e)
	if len(v.Tokens.Tokens[tok].Heads) != 1 || len(v.Tokens.Validation) != 2 || snapshot(t, g).Future[id] != 8 {
		t.Fatal("lost publication or later observation")
	}
	rejected(t, g.c.WriterEligibility("token:"+tok))
}
func TestResourceBudgetAndGarbageIsolation(t *testing.T) {
	f := fresh(t)
	tok, _ := token(t, f)
	raw(t, f, newID(t), make([]byte, 17))
	must(t, f.c.WriterEligibility("token:"+tok))
	f.c.MaxCandidates = 1
	v, e := f.c.Scan()
	rejected(t, e)
	if !v.Incomplete {
		t.Fatal("resource exhaustion not surfaced")
	}
	rejected(t, f.c.WriterEligibility("token:"+tok))
	f.c.MaxCandidates = 4096
	must(t, f.c.WriterEligibility("token:"+tok))
}

func TestConfirmedResolutionAndUsePolicy(t *testing.T) {
	f := fresh(t)
	tok, root := token(t, f)
	use, e := f.c.AssessUse(tok)
	must(t, e)
	if !use.Allowed {
		t.Fatal("healthy token blocked")
	}
	del, db := encode(t, f, 1, tok, []string{root}, map[string]string{"STATUS": "TOMBSTONE"})
	credential, e := client.NewCredential()
	must(t, e)
	rot, rb := encode(t, f, 1, tok, []string{root}, map[string]string{"CREDENTIAL": hex.EncodeToString(credential)})
	raw(t, f, del, db)
	raw(t, f, rot, rb)
	use, e = f.c.AssessUse(tok)
	must(t, e)
	if use.Allowed || !use.RecoveryVisible {
		t.Fatal("conflict use policy")
	}
	op, e := f.c.Prepare("token:"+tok, map[string]string{"STATUS": "LIVE"})
	must(t, e)
	_, e = f.c.Publish(op)
	rejected(t, e)
	op.Confirm()
	id, e := f.c.Publish(op)
	must(t, e)
	v, e := f.c.Scan()
	must(t, e)
	if v.Disposition[id] != "FULLY_VALID" || len(v.Tokens.Tokens[tok].Conflicts) != 0 {
		t.Fatal("resolution not complete")
	}
	use, e = f.c.AssessUse(tok)
	must(t, e)
	if !use.Allowed {
		t.Fatal("resolved token not usable")
	}
	future, fb, e := keys(t, f).Seal(append(client.TLV(1, []byte{8}), client.TLV(2, []byte{1})...))
	must(t, e)
	fid := hex.EncodeToString(future)
	raw(t, f, fid, fb)
	_, e = f.c.Scan()
	must(t, e)
	must(t, f.c.Recovery([]string{fid}, "bypass"))
	must(t, f.c.WriterEligibility("token:"+tok))
	use, e = f.c.AssessUse(tok)
	must(t, e)
	if use.Allowed {
		t.Fatal("bypass falsely established ordinary completeness")
	}
	v, e = f.c.Scan()
	must(t, e)
	if v.Complete || len(v.Future) != 1 {
		t.Fatal("future incompleteness hidden")
	}
}
func TestIngestStagesDoNotCreateEvidence(t *testing.T) {
	f := fresh(t)
	k := keys(t, f)
	tok := newID(t)
	id, b := encode(t, f, 1, tok, nil, fields(t))
	rawID, e := hex.DecodeString(id)
	must(t, e)
	plain, stage := k.Open(rawID, b, nil)
	if stage != "STRUCTURALLY_VALID" {
		t.Fatal(stage)
	}
	plain[len(plain)-1] ^= 1
	invalid, invalidFile, e := k.Seal(plain)
	must(t, e)
	name := hex.EncodeToString(invalid)
	raw(t, f, name, invalidFile)
	v, e := f.c.Scan()
	must(t, e)
	if v.Disposition[name] != "SIGNATURE_INVALID" {
		t.Fatal(v.Disposition)
	}
	s := snapshot(t, f)
	if len(s.Pending) != 0 || len(s.Heads) != 0 || len(s.Future) != 0 {
		t.Fatal("invalid signature acquired authority")
	}
	raw(t, f, id, b)
	v, e = f.c.Scan()
	must(t, e)
	if v.Disposition[id] != "FULLY_VALID" {
		t.Fatal(v.Disposition)
	}
}
