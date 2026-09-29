package storage

import (
	"errors"
	"strings"
	"testing"
	"totipo/conformance/internal/cryptov1"
)

func TestNeverReadIgnoredEntries(t *testing.T) {
	k, _ := cryptov1.Derive(make([]byte, 32))
	id := strings.Repeat("a", 64)
	entries := []Entry{}
	for _, path := range []string{"objects-v2/" + id, "objects-v999/" + id, "objects-v1/nested/" + id, "objects-v1/" + strings.ToUpper(id), "objects-v1/" + id + ".tmp", "objects-v1/../objects-v1/" + id, "/objects-v1/" + id, "objects-v1\\" + id} {
		entries = append(entries, Entry{Path: path, Kind: "regular", Read: func() ([]byte, error) { t.Fatal("ignored path read"); return nil, nil }})
	}
	for _, kind := range []string{"directory", "symlink", "fifo", "socket", "device"} {
		entries = append(entries, Entry{Path: "objects-v1/" + id, Kind: kind, Read: func() ([]byte, error) { t.Fatal("nonregular entry read"); return nil, nil }})
	}
	out, e := Scan("directory", entries, k)
	if e != nil || len(out) != 0 {
		t.Fatal(out, e)
	}
	if _, e = Scan("symlink", entries, k); e == nil {
		t.Fatal("namespace symlink accepted")
	}
}
func TestStorageAuthenticationBoundary(t *testing.T) {
	k, _ := cryptov1.Derive(make([]byte, 32))
	id, b, _ := k.Seal([]byte{0, 1, 0, 1, 2, 0, 2, 0, 1, 99})
	for _, n := range []int{0, 1023, 1025} {
		out, e := Scan("directory", []Entry{{Path: "objects-v1/" + id, Kind: "regular", Read: func() ([]byte, error) { return make([]byte, n), nil }}}, k)
		if e != nil || out[0].Class != InvalidStorage || out[0].Object != nil {
			t.Fatal("wrong-size became evidence", out, e)
		}
	}
	valid, err := Scan("directory", []Entry{{Path: "objects-v1/" + id, Kind: "regular", Read: func() ([]byte, error) { return b, nil }}}, k)
	if err != nil || valid[0].Class != "INVALID" {
		t.Fatal("invalid authenticated grammar accepted", err)
	}
	b[0] ^= 1
	out, e := Scan("directory", []Entry{{Path: "objects-v1/" + id, Kind: "regular", Read: func() ([]byte, error) { return b, nil }}}, k)
	if e != nil || out[0].Class != InvalidStorage {
		t.Fatal("bad AEAD became evidence")
	}
	_, e = Scan("directory", []Entry{{Path: "objects-v1/" + id, Kind: "regular", Read: func() ([]byte, error) { return nil, errors.New("unreadable") }}}, k)
	if e == nil {
		t.Fatal("missing read diagnostic")
	}
}

func TestIncompleteScanPreservesAcceptedObservations(t *testing.T) {
	k, _ := cryptov1.Derive(make([]byte, 32))
	id, b, _ := k.Seal([]byte{0, 1, 0, 1, 2, 0, 2, 0, 1, 99})
	entries := []Entry{
		{Path: "objects-v1/" + id, Kind: "regular", Read: func() ([]byte, error) { return b, nil }},
		{Path: "objects-v1/" + strings.Repeat("b", 64), Kind: "regular"},
	}
	obs, err := Scan("directory", entries, k)
	if err == nil || len(obs) != 1 || obs[0].Class != "INVALID" {
		t.Fatalf("%v %v", obs, err)
	}
}

func TestPublicationRetriesAndReplacement(t *testing.T) {
	intended := make([]byte, 1024)
	intended[0] = 1
	// An error can still have installed the bytes. The next ordinary observation
	// sees them, and an exact retry succeeds without fresh persistence.
	_, out := Install(nil, intended, "absent", false)
	if out != "FAILED" {
		t.Fatal(out)
	}
	existing := append([]byte{}, intended...)
	after, out := Install(existing, intended, "regular", false)
	if out != "ALREADY_PRESENT_EXACT" || &after[0] != &existing[0] {
		t.Fatal("exact retry mutated")
	}
	other := append([]byte{}, intended...)
	other[0] = 2
	after, out = Install(other, intended, "regular", true)
	if out != "FAILED" || after[0] != 2 {
		t.Fatal("overwrote different")
	}
	if Replace([]byte{1}, []byte{2}, "regular", true, true, true) != "STALE" {
		t.Fatal("stale replaced")
	}
	if Replace([]byte{1}, []byte{1}, "regular", false, true, true) != "FAILED" {
		t.Fatal("no comparison")
	}
}

func TestCanonicalEntryTypes(t *testing.T) {
	k, _ := cryptov1.Derive(make([]byte, 32))
	never := func() ([]byte, error) { t.Fatal("deliberately read wrong-type entry"); return nil, nil }
	for _, kind := range []string{"absent", "symlink", "directory", "fifo", "socket", "device", "other"} {
		if _, e := OpenBootstrap(Entry{Path: "vault", Kind: kind, Read: never}, nil); e == nil {
			t.Fatal("bootstrap type", kind)
		}
		if Replace([]byte{1}, []byte{1}, kind, true, true, true) != "FAILED" {
			t.Fatal("replacement type", kind)
		}
	}
	for _, path := range []string{"VAULT", "Vault", "vault.tmp", "vault/conflict", "other/vault"} {
		if _, e := OpenBootstrap(Entry{Path: path, Kind: "regular", Read: never}, nil); e == nil {
			t.Fatal("bootstrap alias", path)
		}
	}
	root := make([]byte, 32)
	record, e := cryptov1.Wrap(nil, root, make([]byte, 16), make([]byte, 12))
	if e != nil {
		t.Fatal(e)
	}
	got, e := OpenBootstrap(Entry{Path: "vault", Kind: "regular", Read: func() ([]byte, error) { return record, nil }}, nil)
	if e != nil || len(got) != 32 {
		t.Fatal("regular bootstrap rejected", e)
	}
	entries := []Entry{{Path: "objects-v1/" + strings.Repeat("a", 64), Kind: "regular", Read: never}}
	for _, kind := range []string{"symlink", "regular", "fifo", "socket", "device", "other"} {
		if obs, e := Scan(kind, entries, k); e == nil || len(obs) != 0 {
			t.Fatal("namespace traversed", kind)
		}
	}
	if obs, e := Scan("missing", entries, k); e != nil || len(obs) != 0 {
		t.Fatal("missing namespace", e)
	}
}
