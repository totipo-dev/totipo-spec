// Package publication describes the required publication ordering. This is an
// abstract interface, not an OS implementation or a filesystem security claim.
package publication

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type Record struct {
	Bytes   []byte
	Binding string
}
type Store interface {
	CreateExclusiveNoFollow(temp string) error
	WriteComplete(temp string, record Record) error
	SyncFile(temp string) error
	ValidateBootstrap(temp string) error
	// Install checks the writer gate and performs the namespace transition under
	// the caller's serialized publication boundary. It rejects symlinks/special
	// files, preserves no-replace semantics, and permits bootstrap replacement
	// only when the existing and new bindings match.
	Install(temp, destination string, replace bool, gate func() bool) error
	SyncDirectory(destination string) error
	RemoveTemp(temp string)
}

func Publish(s Store, temp, destination string, record Record, bootstrap, replace bool, gate func() bool) error {
	for _, name := range []string{temp, destination} {
		if name == "" || name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00") {
			return errors.New("unsafe namespace component")
		}
	}
	size := 2048
	if bootstrap {
		size = 87
		if record.Binding == "" {
			return errors.New("bootstrap binding required")
		}
	}
	if len(record.Bytes) != size {
		return errors.New("incomplete record")
	}
	if replace && !bootstrap {
		return errors.New("immutable object replacement forbidden")
	}
	if e := s.CreateExclusiveNoFollow(temp); e != nil {
		return e
	}
	defer s.RemoveTemp(temp)
	if e := s.WriteComplete(temp, record); e != nil {
		return e
	}
	if e := s.SyncFile(temp); e != nil {
		return e
	}
	if bootstrap {
		if e := s.ValidateBootstrap(temp); e != nil {
			return e
		}
	}
	if e := s.Install(temp, destination, replace, gate); e != nil {
		return e
	}
	return s.SyncDirectory(destination)
}

// MemoryStore is a deterministic namespace/durability simulator. Entry kind is
// an explicit hostile-namespace input. It makes no platform syscall assumptions.
type Entry struct {
	Kind   string
	Record Record
	Synced bool
}
type MemoryStore struct {
	Visible, Durable map[string]Entry
	Log              []string
	FailAt           string
}

func New() *MemoryStore {
	return &MemoryStore{Visible: map[string]Entry{}, Durable: map[string]Entry{}, Log: []string{}}
}
func (s *MemoryStore) step(name string) error {
	s.Log = append(s.Log, name)
	if s.FailAt == name {
		return fmt.Errorf("injected %s failure", name)
	}
	return nil
}
func (s *MemoryStore) CreateExclusiveNoFollow(temp string) error {
	if e := s.step("create-exclusive"); e != nil {
		return e
	}
	if _, ok := s.Visible[temp]; ok {
		return errors.New("temporary destination exists")
	}
	s.Visible[temp] = Entry{Kind: "regular"}
	return nil
}
func (s *MemoryStore) WriteComplete(temp string, r Record) error {
	if e := s.step("write-complete"); e != nil {
		return e
	}
	v, ok := s.Visible[temp]
	if !ok || v.Kind != "regular" {
		return errors.New("unsafe temporary entry")
	}
	v.Record = Record{append([]byte{}, r.Bytes...), r.Binding}
	s.Visible[temp] = v
	return nil
}
func (s *MemoryStore) SyncFile(temp string) error {
	if e := s.step("sync-file"); e != nil {
		return e
	}
	v, ok := s.Visible[temp]
	if !ok || v.Kind != "regular" {
		return errors.New("unsafe file")
	}
	v.Synced = true
	s.Visible[temp] = v
	return nil
}
func (s *MemoryStore) Install(temp, destination string, replace bool, gate func() bool) error {
	if e := s.step("linearize"); e != nil {
		return e
	}
	v, ok := s.Visible[temp]
	if !ok || v.Kind != "regular" || !v.Synced {
		return errors.New("temporary file not complete and synced")
	}
	if old, ok := s.Visible[destination]; ok {
		if old.Kind != "regular" {
			return errors.New("unsafe canonical entry")
		}
		if !replace {
			return errors.New("canonical destination exists")
		}
		if old.Record.Binding == "" || old.Record.Binding != v.Record.Binding {
			return errors.New("replacement root differs")
		}
	} else if replace {
		return errors.New("replacement requires existing bootstrap")
	}
	if gate != nil && !gate() {
		return errors.New("writer gate changed")
	}
	s.Visible[destination] = v
	delete(s.Visible, temp)
	return nil
}
func (s *MemoryStore) SyncDirectory(destination string) error {
	if e := s.step("sync-directory"); e != nil {
		return e
	}
	v, ok := s.Visible[destination]
	if !ok || !v.Synced {
		return errors.New("destination not synced")
	}
	s.Durable[destination] = v
	return nil
}
func (s *MemoryStore) RemoveTemp(temp string) { delete(s.Visible, temp) }
func (s *MemoryStore) Crash() {
	s.Visible = map[string]Entry{}
	for p, v := range s.Durable {
		s.Visible[p] = v
	}
}
func (s *MemoryStore) EqualBytes(path string, b []byte) bool {
	return bytes.Equal(s.Visible[path].Record.Bytes, b)
}

// ValidateBootstrap treats Binding as the result of the separately tested
// unwrap/binding derivation. A real Store must authenticate the temporary record;
// this simulator permits deterministic validation-failure injection.
func (s *MemoryStore) ValidateBootstrap(temp string) error {
	if e := s.step("validate-bootstrap"); e != nil {
		return e
	}
	v, ok := s.Visible[temp]
	if !ok || len(v.Record.Bytes) != 87 || v.Record.Binding == "" {
		return errors.New("bootstrap validation failed")
	}
	return nil
}
