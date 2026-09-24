//go:build linux

package integration

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"totipo/conformance/reference/client"
)

func TestCrashVaultEstablishment(t *testing.T) {
	points := []string{"establish-pending.before-write", "establish-pending.during-write", "establish-pending.after-file-sync", "establish-pending.after-install", "establish-pending.after-directory-sync", "create.after-pending", "vault.before-write", "vault.after-temp-directory", "vault.during-write", "vault.before-file-sync", "vault.after-file-sync", "vault.after-install", "vault.after-directory-sync", "established.before-write", "established.during-write", "established.after-file-sync", "established.after-install", "established.after-directory-sync"}
	for _, point := range points {
		t.Run(point, func(t *testing.T) {
			v, l := dirs(t)
			f := connect(t, v, l)
			crashed(t, command(t, f, "create", point))
			g := connect(t, v, l)
			s := snapshot(t, g)
			installed := point == "vault.after-install" || point == "vault.after-directory-sync" || strings.HasPrefix(point, "established.")
			pending := !(point == "establish-pending.before-write" || point == "establish-pending.during-write" || point == "establish-pending.after-file-sync")
			if (s.Binding != "") != pending {
				t.Fatalf("pending durable=%v state=%+v", pending, s)
			}
			e := g.c.Open([]byte(password), nil)
			if installed {
				must(t, e)
				if snapshot(t, g).Establishment != "ESTABLISHED" {
					t.Fatal("canonical matching root did not recover")
				}
			} else {
				rejected(t, e)
				if s.Establishment == "ESTABLISHED" {
					t.Fatal("established without canonical")
				}
				if pending {
					rejected(t, g.c.Create([]byte(password)))
				}
			}
			t.Logf("Section 10.1: killed at %s; binding persisted=%v, canonical installed=%v; ordinary open allowed only after matching canonical recovery", point, pending, installed)
		})
	}
	t.Run("pending-different-root", func(t *testing.T) {
		v, l := dirs(t)
		f := connect(t, v, l)
		crashed(t, command(t, f, "create", "create.after-pending"))
		before := snapshot(t, f)
		other := fresh(t)
		b, e := other.store.ReadVault()
		must(t, e)
		must(t, os.WriteFile(filepath.Join(v, "VAULT"), b, 0600))
		g := connect(t, v, l)
		rejected(t, g.c.Open([]byte(password), nil))
		if snapshot(t, g).Binding != before.Binding {
			t.Fatal("pending binding overwritten")
		}
	})
}
func TestCrashObjectPublication(t *testing.T) {
	points := []string{"object.before-write", "object.after-temp-directory", "object.during-write", "object.before-file-sync", "object.after-file-sync", "object.after-install", "object.after-directory-sync", "publish.before-install", "publish.after-install", "frontier.before-write", "frontier.during-write", "frontier.after-file-sync", "frontier.after-install", "frontier.after-directory-sync"}
	for _, point := range points {
		t.Run(point, func(t *testing.T) {
			f := fresh(t)
			tok, root := token(t, f)
			crashed(t, command(t, f, "publish", point, "TOTIPO_SCOPE=token:"+tok))
			g := restart(t, f)
			before := snapshot(t, g)
			ids, e := g.store.ListCandidates()
			must(t, e)
			installed := point == "object.after-install" || point == "object.after-directory-sync" || point == "publish.after-install" || strings.HasPrefix(point, "frontier.")
			if installed && len(ids) != 2 || !installed && len(ids) != 1 {
				t.Fatalf("installed=%v IDs=%v", installed, ids)
			}
			accepted := point == "frontier.after-install" || point == "frontier.after-directory-sync"
			if !accepted && !reflect.DeepEqual(before.Heads["token:"+tok], []string{root}) {
				t.Fatal("old evidence lost before durable replacement")
			}
			view, e := g.c.Scan()
			must(t, e)
			must(t, g.c.WriterEligibility("token:"+tok))
			if len(view.Tokens.Tokens[tok].Heads) != 1 {
				t.Fatal("unexpected heads")
			}
			if installed && view.Tokens.Tokens[tok].Heads[0] == root {
				t.Fatal("failed restart acceptance of published object")
			}
			t.Logf("Sections 63.1/71.1: no acknowledgement survived process death; installed=%v, new acceptance present before restart scan=%v", installed, accepted)
		})
	}
}
func TestCrashRewrap(t *testing.T) {
	for _, point := range []string{"rewrap.before-write", "rewrap.after-temp-directory", "rewrap.during-write", "rewrap.before-file-sync", "rewrap.after-file-sync", "rewrap.after-install", "rewrap.after-directory-sync"} {
		t.Run(point, func(t *testing.T) {
			f := fresh(t)
			binding := snapshot(t, f).Binding
			crashed(t, command(t, f, "rewrap", point))
			g := connect(t, f.vault, f.local)
			pass := password
			if point == "rewrap.after-install" || point == "rewrap.after-directory-sync" {
				pass = "new-password"
			}
			must(t, g.c.Open([]byte(pass), nil))
			if snapshot(t, g).Binding != binding {
				t.Fatal("rewrap changed root")
			}
			b, e := g.store.ReadVault()
			must(t, e)
			if len(b) != 87 {
				t.Fatal("torn canonical bootstrap")
			}
		})
	}
}
func TestCrashPendingAndAcceptedHandoff(t *testing.T) {
	t.Run("observation-survives-disappearance", func(t *testing.T) {
		f := fresh(t)
		tok, root := token(t, f)
		_ = root
		id, b := encode(t, f, 1, tok, []string{newID(t)}, map[string]string{"ACCOUNT": "pending"})
		raw(t, f, id, b)
		crashed(t, command(t, f, "scan", "observations.after-directory-sync"))
		must(t, os.Remove(filepath.Join(f.vault, "objects", id)))
		g := restart(t, f)
		rejected(t, g.c.WriterEligibility("token:"+tok))
		if snapshot(t, g).Pending[id] == "" {
			t.Fatal("durable pending block vanished")
		}
	})
	for _, point := range []string{"frontier.before-write", "frontier.after-file-sync", "frontier.after-directory-sync", "handoff.before-write", "handoff.during-write", "handoff.after-file-sync", "handoff.after-install", "handoff.after-directory-sync"} {
		t.Run(point, func(t *testing.T) {
			f := fresh(t)
			tok := newID(t)
			parent, pb := encode(t, f, 1, tok, nil, fields(t))
			child, cb := encode(t, f, 1, tok, []string{parent}, map[string]string{"ISSUER": "child"})
			raw(t, f, child, cb)
			_, e := f.c.Scan()
			must(t, e)
			raw(t, f, parent, pb)
			crashed(t, command(t, f, "scan", point))
			g := restart(t, f)
			s := snapshot(t, g)
			newDurable := point != "frontier.before-write" && point != "frontier.after-file-sync"
			removed := point == "handoff.after-install" || point == "handoff.after-directory-sync"
			if newDurable && !reflect.DeepEqual(s.Heads["token:"+tok], []string{child}) {
				t.Fatal("new evidence not durable")
			}
			if (s.Pending[child] == "") != removed {
				t.Fatal("pending removal ordering", s)
			}
			if removed && !newDurable {
				t.Fatal("invalid test")
			}
			_, e = g.c.Scan()
			must(t, e)
			must(t, g.c.WriterEligibility("token:"+tok))
			if len(snapshot(t, g).Pending) != 0 {
				t.Fatal("handoff did not finish")
			}
			t.Logf("Section 63.1: replacement durable=%v; older pending removed=%v; restart completes handoff", newDurable, removed)
		})
	}
	for _, point := range []string{"handoff.before-write", "handoff.after-directory-sync"} {
		t.Run("terminal/"+point, func(t *testing.T) {
			f := fresh(t)
			tok := newID(t)
			bad, bb := encode(t, f, 1, tok, nil, map[string]string{"STATUS": "TOMBSTONE"})
			child, cb := encode(t, f, 1, tok, []string{bad}, map[string]string{"ISSUER": "child"})
			raw(t, f, child, cb)
			_, e := f.c.Scan()
			must(t, e)
			raw(t, f, bad, bb)
			crashed(t, command(t, f, "scan", point))
			g := restart(t, f)
			s := snapshot(t, g)
			if point == "handoff.before-write" {
				if s.Pending[child] == "" || s.Terminal[child] {
					t.Fatal("premature terminal handoff")
				}
			} else {
				if s.Pending[child] != "" || !s.Terminal[child] {
					t.Fatal("terminal proof not atomic with removal")
				}
			}
		})
	}
}
func TestCrashFutureAndRecoveryDecisions(t *testing.T) {
	for _, operation := range []string{"observe", "bypass", "abandon"} {
		t.Run(operation, func(t *testing.T) {
			f := fresh(t)
			tok, _ := token(t, f)
			var id string
			if operation == "abandon" {
				var b []byte
				id, b = encode(t, f, 1, tok, []string{newID(t)}, map[string]string{"ACCOUNT": "pending"})
				raw(t, f, id, b)
			} else {
				rawID, b, e := keys(t, f).Seal(append(client.TLV(1, []byte{7}), client.TLV(2, []byte{99})...))
				must(t, e)
				id = hex.EncodeToString(rawID)
				raw(t, f, id, b)
			}
			mode, point := "scan", "observations.after-directory-sync"
			if operation != "observe" {
				_, e := f.c.Scan()
				must(t, e)
				mode = operation
				point = "recovery-" + operation + ".after-directory-sync"
			}
			crashed(t, command(t, f, mode, point, "TOTIPO_ID="+id))
			must(t, os.Remove(filepath.Join(f.vault, "objects", id)))
			g := restart(t, f)
			s := snapshot(t, g)
			switch operation {
			case "observe":
				if s.Future[id] != 7 {
					t.Fatal(s)
				}
				rejected(t, g.c.WriterEligibility("token:"+tok))
			case "bypass":
				if s.Future[id] != 7 || !s.Bypass[id] {
					t.Fatal("bypass erased future observation")
				}
				must(t, g.c.WriterEligibility("token:"+tok))
			case "abandon":
				if s.Pending[id] == "" || !s.Abandoned[id] {
					t.Fatal("abandonment erased pending fact")
				}
				must(t, g.c.WriterEligibility("token:"+tok))
			}
		})
	}
}
func TestAtomicLocalStateCrashReplacement(t *testing.T) {
	// The JSON snapshot is one transaction, not a shared flush masquerading as
	// atomicity: fork heads and recovery records are encoded into one rename.
	for _, point := range []string{"observations.before-write", "observations.during-write", "observations.before-file-sync", "observations.after-file-sync", "observations.after-install", "observations.after-directory-sync"} {
		t.Run(point, func(t *testing.T) {
			f := fresh(t)
			tok, _ := token(t, f)
			old := snapshot(t, f)
			pending, pb := encode(t, f, 1, tok, []string{newID(t)}, map[string]string{"ACCOUNT": "p"})
			raw(t, f, pending, pb)
			rawID, fb, e := keys(t, f).Seal(append(client.TLV(1, []byte{2}), client.TLV(2, []byte{2})...))
			must(t, e)
			future := hex.EncodeToString(rawID)
			raw(t, f, future, fb)
			crashed(t, command(t, f, "scan", point))
			g := restart(t, f)
			s := snapshot(t, g)
			newState := point == "observations.after-install" || point == "observations.after-directory-sync"
			if (s.Pending[pending] != "") != newState || (s.Future[future] != 0) != newState {
				t.Fatal("mixed transaction")
			}
			if !reflect.DeepEqual(s.Heads, old.Heads) {
				t.Fatal("unrelated evidence changed")
			}
			must(t, os.Remove(filepath.Join(f.vault, "objects", pending)))
			must(t, os.Remove(filepath.Join(f.vault, "objects", future)))
			if newState {
				rejected(t, g.c.WriterEligibility("token:"+tok))
			} else {
				must(t, g.c.WriterEligibility("token:"+tok))
			}
			t.Log("process crash reads a complete old or new snapshot; this does not emulate power-loss cache loss")
		})
	}
}
func TestImmutableCollisionAndSameMetadataReplacement(t *testing.T) {
	f := fresh(t)
	tok, root := token(t, f)
	path := filepath.Join(f.vault, "objects", root)
	original, e := os.ReadFile(path)
	must(t, e)
	st, e := os.Stat(path)
	must(t, e)
	bad := append([]byte{}, original...)
	bad[99] ^= 1
	must(t, os.WriteFile(path, bad, 0600))
	must(t, os.Chtimes(path, st.ModTime(), st.ModTime()))
	v, e := f.c.Scan()
	must(t, e)
	if v.Disposition[root] != "AEAD_REJECTED" || !v.Degraded["token:"+tok] {
		t.Fatal("trusted unchanged size/mtime")
	}
	rejected(t, f.store.PublishImmutable(root, original))
	got, e := os.ReadFile(path)
	must(t, e)
	if !bytes.Equal(got, bad) {
		t.Fatal("immutable object overwritten")
	}
	raw(t, f, root, original)
	must(t, f.c.WriterEligibility("token:"+tok))
}

func TestCrashPresentationHandoff(t *testing.T) {
	for _, point := range []string{"frontier.after-file-sync", "frontier.after-directory-sync", "handoff.before-write", "handoff.after-directory-sync"} {
		t.Run(point, func(t *testing.T) {
			f := fresh(t)
			parent, pb := encode(t, f, 2, "", nil, map[string]string{"DISPLAY_NAME": "parent"})
			child, cb := encode(t, f, 2, "", []string{parent}, map[string]string{"DISPLAY_NAME": "child"})
			raw(t, f, child, cb)
			_, e := f.c.Scan()
			must(t, e)
			scope := snapshot(t, f).Pending[child]
			raw(t, f, parent, pb)
			crashed(t, command(t, f, "scan", point))
			g := restart(t, f)
			s := snapshot(t, g)
			removed := point == "handoff.after-directory-sync"
			if (s.Pending[child] == "") != removed {
				t.Fatal("presentation evidence ordering")
			}
			if point != "frontier.after-file-sync" && !reflect.DeepEqual(s.Heads[scope], []string{child}) {
				t.Fatal("presentation successor not durable")
			}
			_, e = g.c.Scan()
			must(t, e)
			if snapshot(t, g).Pending[child] != "" {
				t.Fatal("restart failed presentation handoff")
			}
		})
	}
}
