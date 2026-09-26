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
	if err != nil || valid[0].Class != "OPAQUE_UNSCOPED" {
		t.Fatal("authenticated unscoped future evidence lost", err)
	}
	b[0] ^= 1
	out, e := Scan("directory", []Entry{{Path: "objects-v1/" + id, Kind: "regular", Read: func() ([]byte, error) { return b, nil }}}, k)
	if e != nil || out[0].Class != InvalidStorage {
		t.Fatal("bad AEAD became evidence")
	}
	_, e = Scan("directory", []Entry{{Path: "objects-v1/" + id, Kind: "regular", Read: func() ([]byte, error) { return nil, errors.New("unreadable") }}}, k)
	if e == nil {
		t.Fatal("unreadable candidate treated as complete")
	}
}
