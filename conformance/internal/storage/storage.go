// Package storage evaluates an observed filesystem environment for the v1
// envelope family. It does not open host paths: a live adapter must obtain entry
// kinds and bounded bytes from stable, no-follow handles before exposing Read.
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

// Scan never calls Read for siblings, nested entries, wrong names, symlinks, or
// special files. An unsafe family entry or unreadable candidate is an error,
// not authenticated opaque evidence and not a successful complete discovery.
// Missing family directories are empty. Errors require incomplete discovery in
// the caller; they must not be silently classified as ordinary invalid bytes.
func Scan(familyKind string, entries []Entry, k cryptov1.Keys) ([]Observation, error) {
	out := []Observation{}
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
			return out, errors.New("unreadable candidate")
		}
		b, e := entry.Read()
		if e != nil {
			return out, e
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
	return out, nil
}
