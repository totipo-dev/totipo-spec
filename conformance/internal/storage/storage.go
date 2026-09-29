// Package storage evaluates an observed filesystem environment for the v1
// envelope family. It does not open host paths: a live adapter must honor observed
// entry kinds, bound reads, and handle ordinary churn conservatively. Hostile local
// syscall-race immunity is optional.
package storage

import (
	"errors"
	"strings"
	"totipo/conformance/internal/cryptov1"
	"totipo/conformance/internal/object"
)

const Namespace = "objects-v1"
const InvalidStorage = "INVALID_STORAGE"

// Entry paths are exact slash-separated paths relative to a configured root.
// Kind describes the entry itself, not a followed symlink target.
type Entry struct {
	Path, Kind string
	Read       func() ([]byte, error)
}
type Observation struct {
	Path, Class string
	Semantic    []byte
	Object      *object.Object
}

// BootstrapCandidate accepts the observed entry, not a followed link target.
func BootstrapCandidate(path, kind string) bool { return path == "vault" && kind == "regular" }

// OpenBootstrap models type checking before reading or interpreting canonical
// bytes. A live adapter still supplies bounded reads; no inode/race proof is assumed.
func OpenBootstrap(entry Entry, password []byte) ([]byte, error) {
	if !BootstrapCandidate(entry.Path, entry.Kind) || entry.Read == nil {
		return nil, errors.New("invalid canonical bootstrap entry")
	}
	record, e := entry.Read()
	if e != nil {
		return nil, e
	}
	return cryptov1.Unwrap(password, record)
}

func Candidate(path, kind string) bool {
	if kind != "regular" || !strings.HasPrefix(path, Namespace+"/") {
		return false
	}
	name := strings.TrimPrefix(path, Namespace+"/")
	if len(name) != 64 {
		return false
	}
	for _, c := range name {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

// Scan retains validated objects while accumulating diagnostic errors. Failure
// does not globally invalidate observed objects or prohibit protocol operations.
func Scan(familyKind string, entries []Entry, k cryptov1.Keys) ([]Observation, error) {
	out := []Observation{}
	var diagnostics []error
	if familyKind == "missing" {
		return out, nil
	}
	if familyKind != "directory" {
		return nil, errors.New("unsafe objects-v1 namespace")
	}
	for _, entry := range entries {
		if !Candidate(entry.Path, entry.Kind) {
			continue
		}
		if entry.Read == nil {
			diagnostics = append(diagnostics, errors.New("unreadable candidate"))
			continue
		}
		b, e := entry.Read()
		if e != nil {
			diagnostics = append(diagnostics, e)
			continue
		}
		obs := Observation{Path: entry.Path, Class: InvalidStorage}
		if len(b) == 1024 {
			p, e := k.Open(strings.TrimPrefix(entry.Path, Namespace+"/"), b)
			if e == nil {
				obs.Class, obs.Object = object.Dispatch(p)
				obs.Semantic = p
			}
		}
		out = append(out, obs)
	}
	return out, errors.Join(diagnostics...)
}
