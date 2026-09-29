package vectors

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"totipo/conformance/internal/cryptov1"
)

func TestCorpus(t *testing.T) {
	m, cases, e := Read("../../..")
	if e != nil {
		t.Fatal(e)
	}
	if e = VerifyProfile("../../..", m); e != nil {
		t.Fatal(e)
	}
	for _, c := range cases {
		t.Run(c.ID, func(t *testing.T) {
			if e := Run(c); e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestStrictDecode(t *testing.T) {
	var c Case
	for _, s := range []string{`{"unknown":1}`, `{} {}`, `{"id":"a","id":"b"}`, `{"graph":{"steps":[],"steps":[]}}`} {
		if Decode([]byte(s), &c) == nil {
			t.Fatal("accepted", s)
		}
	}
}
func TestManifestTampering(t *testing.T) {
	root := t.TempDir()
	if e := os.CopyFS(filepath.Join(root, "vectors"), os.DirFS("../../../vectors")); e != nil {
		t.Fatal(e)
	}
	p := filepath.Join(root, "vectors/manifest.json")
	b, e := os.ReadFile(p)
	if e != nil {
		t.Fatal(e)
	}
	b = []byte(strings.Replace(string(b), `"r17"`, `"r99"`, 1))
	os.WriteFile(p, b, 0600)
	if _, _, e := Read(root); e == nil {
		t.Fatal("revision")
	}
}

func TestConsumerDetectsHeadMetadataLoss(t *testing.T) {
	for _, id := range []string{"equal-concurrent", "cycle-equal"} {
		p, e := os.ReadFile("../../../vectors/cases/graph/v1.graph." + id + ".001.json")
		if e != nil {
			t.Fatal(e)
		}
		for _, field := range []string{"name", "time"} {
			var c Case
			if e := Decode(p, &c); e != nil {
				t.Fatal(e)
			}
			for _, step := range c.Graph.Steps {
				if step.Want != nil && len(step.Want.HeadObjects) > 1 {
					if field == "name" {
						step.Want.HeadObjects[0].ClientName = nil
					} else {
						step.Want.HeadObjects[0].ClientTime = nil
					}
					break
				}
			}
			if Run(c) == nil {
				t.Fatal("consumer accepted metadata loss", id, field)
			}
		}
	}
}

func TestPostAEADFixturesMustAuthenticateAndHaveDefect(t *testing.T) {
	for _, defect := range []string{"nonzero-padding", "object-id-mismatch", "semantic-length-invalid"} {
		b, e := os.ReadFile("../../../vectors/cases/crypto/v1.crypto." + defect + ".001.json")
		if e != nil {
			t.Fatal(e)
		}
		var c Case
		if e := Decode(b, &c); e != nil {
			t.Fatal(e)
		}
		if e := Run(c); e != nil {
			t.Fatal(e)
		}
		// A negative case with a broken tag must not pass merely because Open rejects.
		encrypted, e := unhex(c.PostAEAD.Object)
		if e != nil {
			t.Fatal(e)
		}
		encrypted[len(encrypted)-1] ^= 1
		c.PostAEAD.Object = hex.EncodeToString(encrypted)
		if e := Run(c); e == nil || !strings.Contains(e.Error(), "must pass AEAD") {
			t.Fatal("failed authentication counted as post-AEAD rejection", e)
		}
		// An ordinary valid object mislabeled with any defect also must not pass.
		if e := Decode(b, &c); e != nil {
			t.Fatal(e)
		}
		root, _ := unhex(c.Root)
		p, _ := unhex(c.Semantic)
		k, _ := cryptov1.Derive(root)
		name, valid, e := k.Seal(p)
		if e != nil {
			t.Fatal(e)
		}
		plain, _ := cryptov1.Padded(p)
		c.PostAEAD.ObjectID = name
		c.PostAEAD.Object = hex.EncodeToString(valid)
		c.PostAEAD.Plaintext = hex.EncodeToString(plain)
		if Run(c) == nil {
			t.Fatal("missing defect accepted", defect)
		}
	}
}
