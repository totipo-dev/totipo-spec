//go:build linux

package localstate

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"golang.org/x/sys/unix"
	"os"
	"totipo/conformance/reference/fault"
	"totipo/conformance/reference/storage"
)

// Files uses a trusted, pre-created private directory outside the vault. A
// stable lock inode is never replaced. Every transaction reloads after flock.
type Files struct {
	dir  *os.File
	Hook fault.Hook
}

func Open(path string) (*Files, error) {
	d, e := storage.OpenDirectory(path)
	if e != nil {
		return nil, e
	}
	st, e := d.Stat()
	if e != nil || st.Mode().Perm()&0077 != 0 {
		d.Close()
		return nil, errors.New("local state directory must be private (0700)")
	}
	return &Files{dir: d}, nil
}
func (f *Files) Close() error { return f.dir.Close() }

type tx struct {
	owner *Files
	state State
}

func (t *tx) State() *State { return &t.state }
func (f *Files) With(fn func(Transaction) error) error {
	fd, e := unix.Openat(int(f.dir.Fd()), "lock", unix.O_RDWR|unix.O_CREAT|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0600)
	if e != nil {
		return e
	}
	defer unix.Close(fd)
	var st unix.Stat_t
	if e = unix.Fstat(fd, &st); e != nil {
		return e
	}
	if st.Mode&unix.S_IFMT != unix.S_IFREG {
		return errors.New("unsafe lock")
	}
	if e = unix.Flock(fd, unix.LOCK_EX); e != nil {
		return e
	}
	defer unix.Flock(fd, unix.LOCK_UN)
	s := Empty()
	b, e := storage.ReadAt(f.dir, "state.json", 16<<20)
	if e != nil && !errors.Is(e, os.ErrNotExist) {
		return e
	}
	if e == nil {
		if e = json.Unmarshal(b, &s); e != nil {
			return e
		}
		if s.Heads == nil || s.Pending == nil || s.Future == nil || s.Abandoned == nil || s.Acknowledged == nil || s.Bypass == nil || s.Terminal == nil {
			return errors.New("incomplete local state")
		}
		// A prior process may have died after rename but before directory fsync.
		// Finish that durability barrier before any caller relies on the record.
		if e = f.dir.Sync(); e != nil {
			return e
		}
	}
	return fn(&tx{f, s})
}
func (t *tx) Commit(op string) error {
	hit := func(p string) error { return t.owner.Hook.Hit(op + "." + p) }
	if e := hit("before-write"); e != nil {
		return e
	}
	t.state.Generation++
	b, e := json.Marshal(t.state)
	if e != nil {
		return e
	}
	if len(b) > 16<<20 {
		return errors.New("local security state capacity exceeded")
	}
	r := make([]byte, 16)
	if _, e = rand.Read(r); e != nil {
		return e
	}
	name := ".state-" + hex.EncodeToString(r)
	fd, e := unix.Openat(int(t.owner.dir.Fd()), name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
	if e != nil {
		return e
	}
	f := os.NewFile(uintptr(fd), name)
	defer f.Close()
	defer unix.Unlinkat(int(t.owner.dir.Fd()), name, 0)
	half := len(b) / 2
	if _, e = f.Write(b[:half]); e != nil {
		return e
	}
	if e = hit("during-write"); e != nil {
		return e
	}
	if _, e = f.Write(b[half:]); e != nil {
		return e
	}
	if e = hit("before-file-sync"); e != nil {
		return e
	}
	if e = f.Sync(); e != nil {
		return e
	}
	if e = hit("after-file-sync"); e != nil {
		return e
	}
	if e = unix.Renameat(int(t.owner.dir.Fd()), name, int(t.owner.dir.Fd()), "state.json"); e != nil {
		return e
	}
	if e = hit("after-install"); e != nil {
		return e
	}
	if e = t.owner.dir.Sync(); e != nil {
		return e
	}
	return hit("after-directory-sync")
}
