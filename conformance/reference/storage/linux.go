//go:build linux

package storage

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"totipo/conformance/reference/fault"
)

// OpenDirectory walks every component without following symlinks. The returned
// descriptor pins the selected directory even if its pathname is later rebound.
func OpenDirectory(path string) (*os.File, error) {
	absolute, e := filepath.Abs(path)
	if e != nil {
		return nil, e
	}
	fd, e := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if e != nil {
		return nil, e
	}
	for _, p := range strings.Split(strings.TrimPrefix(absolute, "/"), "/") {
		if p == "" {
			continue
		}
		next, err := unix.Openat(fd, p, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
		unix.Close(fd)
		if err != nil {
			return nil, err
		}
		fd = next
	}
	return os.NewFile(uintptr(fd), absolute), nil
}
func ReadAt(dir *os.File, name string, limit int) ([]byte, error) {
	return readAt(dir, name, limit, false)
}
func readAt(dir *os.File, name string, limit int, sync bool) ([]byte, error) {
	if name == "" || strings.ContainsAny(name, "/\\\x00") || name == "." || name == ".." {
		return nil, errors.New("invalid component")
	}
	fd, e := unix.Openat(int(dir.Fd()), name, unix.O_PATH|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if e != nil {
		return nil, e
	}
	f := os.NewFile(uintptr(fd), name)
	defer f.Close()
	st, e := f.Stat()
	if e != nil {
		return nil, e
	}
	if !st.Mode().IsRegular() {
		return nil, errors.New("not an ordinary file")
	}
	// Reopen this pinned regular inode through Linux procfs, never the hostile
	// pathname. O_PATH above cannot trigger device/FIFO open side effects.
	readFD, e := unix.Open(fmt.Sprintf("/proc/self/fd/%d", fd), unix.O_RDONLY|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if e != nil {
		return nil, e
	}
	r := os.NewFile(uintptr(readFD), name)
	defer r.Close()
	b, e := io.ReadAll(io.LimitReader(r, int64(limit)+1))
	if e != nil {
		return nil, e
	}
	if len(b) > limit {
		return nil, errors.New("file exceeds bound")
	}
	if sync {
		if e = r.Sync(); e != nil {
			return nil, e
		}
	}
	return b, nil
}

type Linux struct {
	vault, objects *os.File
	Hook           fault.Hook
}

func Open(path string) (*Linux, error) {
	v, e := OpenDirectory(path)
	if e != nil {
		return nil, e
	}
	err := unix.Mkdirat(int(v.Fd()), "objects", 0700)
	if err != nil && !errors.Is(err, unix.EEXIST) {
		v.Close()
		return nil, err
	}
	fd, e := unix.Openat(int(v.Fd()), "objects", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if e != nil {
		v.Close()
		return nil, e
	}
	if e = v.Sync(); e != nil {
		unix.Close(fd)
		v.Close()
		return nil, e
	}
	return &Linux{vault: v, objects: os.NewFile(uintptr(fd), "objects")}, nil
}
func (s *Linux) Close() error { a := s.objects.Close(); b := s.vault.Close(); return errors.Join(a, b) }
func (s *Linux) ListCandidates() ([]string, error) {
	fd, e := unix.Openat(int(s.objects.Fd()), ".", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if e != nil {
		return nil, e
	}
	f := os.NewFile(uintptr(fd), "objects")
	defer f.Close()
	names, e := f.Readdirnames(-1)
	if e != nil {
		return nil, e
	}
	out := []string{}
	for _, n := range names {
		if Candidate(n) {
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out, nil
}
func (s *Linux) ReadObject(id string) ([]byte, error) {
	if !Candidate(id) {
		return nil, errors.New("invalid ID")
	}
	return ReadAt(s.objects, id, 2048)
}
func (s *Linux) ReadVault() ([]byte, error) { return ReadAt(s.vault, "VAULT", 87) }
func (s *Linux) PublishImmutable(id string, b []byte) error {
	if !Candidate(id) || len(b) != 2048 {
		return errors.New("invalid object publication")
	}
	e := s.install(s.objects, id, b, false, "object")
	if errors.Is(e, unix.EEXIST) {
		old, err := readAt(s.objects, id, 2048, true)
		if err != nil || !bytes.Equal(old, b) {
			return ErrConflict
		}
		return s.objects.Sync()
	}
	return e
}
func (s *Linux) InstallInitialVault(b []byte) error {
	if len(b) != 87 {
		return errors.New("invalid bootstrap length")
	}
	return s.install(s.vault, "VAULT", b, false, "vault")
}
func (s *Linux) ReplaceVault(b []byte) error {
	if len(b) != 87 {
		return errors.New("invalid bootstrap length")
	}
	if _, e := s.ReadVault(); e != nil {
		return e
	}
	return s.install(s.vault, "VAULT", b, true, "rewrap")
}

// A fresh private staging directory prevents pre-existing temp entries from
// redirecting writes. All accesses remain relative to pinned directory handles.
func (s *Linux) install(dir *os.File, name string, b []byte, replace bool, op string) error {
	if e := s.Hook.Hit(op + ".before-write"); e != nil {
		return e
	}
	r := make([]byte, 16)
	if _, e := rand.Read(r); e != nil {
		return e
	}
	stage := ".totipo-" + hex.EncodeToString(r)
	if e := unix.Mkdirat(int(dir.Fd()), stage, 0700); e != nil {
		return e
	}
	defer unix.Unlinkat(int(dir.Fd()), stage, unix.AT_REMOVEDIR)
	if e := s.Hook.Hit(op + ".after-temp-directory"); e != nil {
		return e
	}
	fd, e := unix.Openat(int(dir.Fd()), stage, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if e != nil {
		return e
	}
	stageFile := os.NewFile(uintptr(fd), stage)
	defer stageFile.Close()
	tf, e := unix.Openat(fd, "candidate", unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
	if e != nil {
		return e
	}
	f := os.NewFile(uintptr(tf), "candidate")
	defer f.Close()
	defer unix.Unlinkat(fd, "candidate", 0)
	half := len(b) / 2
	if _, e = f.Write(b[:half]); e != nil {
		return e
	}
	if e = s.Hook.Hit(op + ".during-write"); e != nil {
		return e
	}
	if _, e = f.Write(b[half:]); e != nil {
		return e
	}
	if e = s.Hook.Hit(op + ".before-file-sync"); e != nil {
		return e
	}
	if e = f.Sync(); e != nil {
		return e
	}
	// A byte-for-byte readback of a cryptographically validated bootstrap
	// validates the actual complete temp, not only its former in-memory input.
	copy, err := ReadAt(stageFile, "candidate", len(b))
	if err != nil || !bytes.Equal(copy, b) {
		return ErrConflict
	}
	if e = s.Hook.Hit(op + ".after-file-sync"); e != nil {
		return e
	}
	flags := uint(unix.RENAME_NOREPLACE)
	if replace {
		flags = 0
	}
	if e = unix.Renameat2(fd, "candidate", int(dir.Fd()), name, flags); e != nil {
		return fmt.Errorf("install %s: %w", name, e)
	}
	if e = s.Hook.Hit(op + ".after-install"); e != nil {
		return e
	}
	if e = dir.Sync(); e != nil {
		return e
	}
	if e = s.Hook.Hit(op + ".after-directory-sync"); e != nil {
		return e
	}
	installed, e := ReadAt(dir, name, len(b))
	if e != nil || !bytes.Equal(installed, b) {
		return ErrConflict
	}
	return nil
}
