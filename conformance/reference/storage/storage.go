// Package storage defines the byte-storage boundary, independent of semantics.
package storage

import "errors"

var ErrConflict = errors.New("immutable destination differs or is unsafe")

type Objects interface {
	ListCandidates() ([]string, error)
	ReadObject(id string) ([]byte, error)
	PublishImmutable(id string, b []byte) error
	ReadVault() ([]byte, error)
	InstallInitialVault(b []byte) error
	ReplaceVault(b []byte) error // caller must validate the established root
}

func Candidate(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
