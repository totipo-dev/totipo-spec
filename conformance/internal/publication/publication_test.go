package publication

import (
	"bytes"
	"reflect"
	"testing"
)

func TestObjectPublicationOrdering(t *testing.T) {
	s := New()
	b := bytes.Repeat([]byte{1}, 2048)
	if e := Publish(s, "temp", "object", Record{Bytes: b}, false, false, func() bool { return true }); e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(s.Log, []string{"create-exclusive", "write-complete", "sync-file", "linearize", "sync-directory"}) {
		t.Fatal(s.Log)
	}
	s.Crash()
	if !s.EqualBytes("object", b) {
		t.Fatal("durable object lost")
	}
}
func TestCrashPoints(t *testing.T) {
	for _, step := range []string{"create-exclusive", "write-complete", "sync-file", "linearize", "sync-directory"} {
		t.Run(step, func(t *testing.T) {
			s := New()
			s.FailAt = step
			if e := Publish(s, "temp", "object", Record{Bytes: make([]byte, 2048)}, false, false, nil); e == nil {
				t.Fatal("missing injected failure")
			}
			s.Crash()
			if _, ok := s.Visible["object"]; ok {
				t.Fatal("undurable object survived simulator crash")
			}
		})
	}
}
func TestHostileNamespace(t *testing.T) {
	for _, kind := range []string{"symlink", "special", "directory", "regular"} {
		for _, path := range []string{"temp", "object"} {
			t.Run(kind+"/"+path, func(t *testing.T) {
				s := New()
				s.Visible[path] = Entry{Kind: kind, Record: Record{Bytes: []byte("original")}}
				if e := Publish(s, "temp", "object", Record{Bytes: make([]byte, 2048)}, false, false, nil); e == nil {
					t.Fatal("clobbered hostile/existing entry")
				}
				if !s.EqualBytes(path, []byte("original")) {
					t.Fatal("changed existing entry")
				}
			})
		}
	}
}
func TestBootstrapNoReplaceAndSameRootReplacement(t *testing.T) {
	s := New()
	old := Record{bytes.Repeat([]byte{1}, 87), "root-a"}
	if e := Publish(s, "t1", "VAULT", old, true, false, nil); e != nil {
		t.Fatal(e)
	}
	if e := Publish(s, "t2", "VAULT", Record{make([]byte, 87), "root-b"}, true, false, nil); e == nil {
		t.Fatal("second initializer won")
	}
	if e := Publish(s, "t3", "VAULT", Record{make([]byte, 87), "root-b"}, true, true, nil); e == nil {
		t.Fatal("cross-root replacement")
	}
	replacement := Record{bytes.Repeat([]byte{2}, 87), "root-a"}
	s.FailAt = "sync-directory"
	if e := Publish(s, "t4", "VAULT", replacement, true, true, nil); e == nil {
		t.Fatal("expected directory durability failure")
	}
	s.Crash()
	if !s.EqualBytes("VAULT", old.Bytes) {
		t.Fatal("crash did not retain old bootstrap")
	}
	s.FailAt = ""
	if e := Publish(s, "t5", "VAULT", replacement, true, true, nil); e != nil {
		t.Fatal(e)
	}
	s.Crash()
	if !s.EqualBytes("VAULT", replacement.Bytes) {
		t.Fatal("durable same-root replacement lost")
	}
}
func TestWriterGateAtLinearization(t *testing.T) {
	s := New()
	if e := Publish(s, "temp", "object", Record{Bytes: make([]byte, 2048)}, false, false, func() bool { return false }); e == nil {
		t.Fatal("published through changed gate")
	}
	if _, ok := s.Visible["object"]; ok {
		t.Fatal("installed blocked object")
	}
}
func TestIncompleteWriteRejected(t *testing.T) {
	s := New()
	if e := Publish(s, "temp", "object", Record{Bytes: make([]byte, 2047)}, false, false, nil); e == nil || len(s.Log) != 0 {
		t.Fatal("incomplete object reached storage")
	}
}

func TestUnsafePaths(t *testing.T) {
	for _, path := range []string{"../outside", "/absolute", "a/b", "a\\b", "..", "a\x00b"} {
		s := New()
		if e := Publish(s, path, "object", Record{Bytes: make([]byte, 2048)}, false, false, nil); e == nil || len(s.Log) > 0 {
			t.Fatalf("unsafe path %q reached storage", path)
		}
	}
}
func TestBootstrapCandidateValidatedBeforeInstall(t *testing.T) {
	s := New()
	s.FailAt = "validate-bootstrap"
	if e := Publish(s, "temp", "VAULT", Record{make([]byte, 87), "binding"}, true, false, nil); e == nil {
		t.Fatal("invalid bootstrap installed")
	}
	if _, ok := s.Visible["VAULT"]; ok {
		t.Fatal("canonical name changed")
	}
}
