package runner

import (
	"testing"
	"totipo/conformance/internal/corpus"
)

func TestCorpus(t *testing.T) {
	cases, e := corpus.Load("../../../vectors/v0")
	if e != nil {
		t.Fatal(e)
	}
	for _, c := range cases {
		t.Run(c.ID, func(t *testing.T) {
			if e := Run(c); e != nil {
				t.Fatalf("%s (%s): %v", c.Provenance.Source, c.Provenance.Locator, e)
			}
		})
	}
}
