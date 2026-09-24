//go:build linux

package integration

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"totipo/conformance/reference/storage"
)

func TestFilesystemHostileCandidates(t *testing.T) {
	for _, kind := range []string{"symlink", "fifo", "directory", "socket"} {
		t.Run(kind, func(t *testing.T) {
			f := fresh(t)
			id := newID(t)
			p := filepath.Join(f.vault, "objects", id)
			target := filepath.Join(t.TempDir(), "target")
			must(t, os.WriteFile(target, []byte("do not touch"), 0600))
			switch kind {
			case "symlink":
				must(t, os.Symlink(target, p))
			case "fifo":
				must(t, unix.Mkfifo(p, 0600))
			case "directory":
				must(t, os.Mkdir(p, 0700))
			case "socket":
				dir, e := storage.OpenDirectory(filepath.Dir(p))
				must(t, e)
				defer dir.Close()
				listener, e := net.Listen("unix", fmt.Sprintf("/proc/self/fd/%d/%s", dir.Fd(), id))
				must(t, e)
				defer listener.Close()
			}
			start := time.Now()
			_, e := f.store.ReadObject(id)
			rejected(t, e)
			if time.Since(start) > time.Second {
				t.Fatal("special entry blocked")
			}
			rejected(t, f.store.PublishImmutable(id, make([]byte, 2048)))
			b, e := os.ReadFile(target)
			must(t, e)
			if string(b) != "do not touch" {
				t.Fatal("outside target changed")
			}
		})
	}
}
func TestFilesystemVaultAndParentSymlinks(t *testing.T) {
	v, l := dirs(t)
	f := connect(t, v, l)
	target := filepath.Join(t.TempDir(), "target")
	must(t, os.WriteFile(target, bytes.Repeat([]byte{3}, 87), 0600))
	must(t, os.Symlink(target, filepath.Join(v, "VAULT")))
	rejected(t, f.c.Create([]byte(password)))
	rejected(t, f.store.InstallInitialVault(make([]byte, 87)))
	rejected(t, f.store.ReplaceVault(make([]byte, 87)))
	b, e := os.ReadFile(target)
	must(t, e)
	if !bytes.Equal(b, bytes.Repeat([]byte{3}, 87)) {
		t.Fatal("bootstrap symlink target modified")
	}
	alias := filepath.Join(t.TempDir(), "alias")
	must(t, os.Symlink(v, alias))
	_, e = storage.Open(alias)
	rejected(t, e)
	// Rebind objects/ after acquisition; anchored handles must never follow the
	// replacement into a different directory.
	original := filepath.Join(v, "objects-held")
	must(t, os.Rename(filepath.Join(v, "objects"), original))
	outside := t.TempDir()
	must(t, os.Symlink(outside, filepath.Join(v, "objects")))
	id := newID(t)
	must(t, f.store.PublishImmutable(id, make([]byte, 2048)))
	if _, e = os.Stat(filepath.Join(outside, id)); !errors.Is(e, os.ErrNotExist) {
		t.Fatal("escaped pinned namespace")
	}
	if _, e = os.Stat(filepath.Join(original, id)); e != nil {
		t.Fatal("lost pinned namespace", e)
	}
}
func TestFilesystemTemporaryCollision(t *testing.T) {
	for _, attack := range []string{"candidate-symlink", "stage-symlink"} {
		t.Run(attack, func(t *testing.T) {
			f := fresh(t)
			target := filepath.Join(t.TempDir(), "target")
			must(t, os.WriteFile(target, []byte("safe"), 0600))
			f.store.Hook = func(point string) error {
				if point != "object.after-temp-directory" {
					return nil
				}
				entries, e := os.ReadDir(filepath.Join(f.vault, "objects"))
				if e != nil {
					return e
				}
				for _, ent := range entries {
					if strings.HasPrefix(ent.Name(), ".totipo-") {
						stage := filepath.Join(f.vault, "objects", ent.Name())
						if attack == "candidate-symlink" {
							return os.Symlink(target, filepath.Join(stage, "candidate"))
						}
						if e := os.Remove(stage); e != nil {
							return e
						}
						return os.Symlink(filepath.Dir(target), stage)
					}
				}
				return errors.New("no stage found")
			}
			rejected(t, f.store.PublishImmutable(newID(t), make([]byte, 2048)))
			b, e := os.ReadFile(target)
			must(t, e)
			if string(b) != "safe" {
				t.Fatal("temp target overwritten")
			}
		})
	}
}
func TestFilesystemGrammarAndBounds(t *testing.T) {
	f := fresh(t)
	for _, name := range []string{"../VAULT", strings.Repeat("a", 63), strings.Repeat("A", 64), ".tmp", strings.Repeat("g", 64)} {
		_, e := f.store.ReadObject(name)
		rejected(t, e)
		rejected(t, f.store.PublishImmutable(name, make([]byte, 2048)))
	}
	id := newID(t)
	raw(t, f, id, make([]byte, 2049))
	_, e := f.store.ReadObject(id)
	rejected(t, e)
	rejected(t, f.store.PublishImmutable(id, make([]byte, 2047)))
	must(t, os.WriteFile(filepath.Join(f.vault, "objects", "conflicted-copy"), nil, 0600))
	ids, e := f.store.ListCandidates()
	must(t, e)
	if len(ids) != 1 || ids[0] != id {
		t.Fatal("candidate grammar", ids)
	}
}
