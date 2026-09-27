package main

import (
	"os"
	"path/filepath"
	"strings"
	"totipo/conformance/internal/vectors"
)

// Preserve reviewed r13 workflows verbatim. References are resolved after the
// complete manifest is written, before the generator reports success.
func recoveryCases() {
	for _, item := range []struct {
		id       string
		sections []string
	}{
		{"v1.future.unscoped-retains-exact-object.001", []string{"24", "24.1", "24.7", "34"}},
		{"v1.future.unscoped-reprocess-retained.001", []string{"24.2", "37.3"}},
		{"v1.provenance.late-device-reclassify.001", []string{"16.1", "55"}},
		{"v1.graph.rebaseline-history-reappears.001", []string{"34.3", "34.4"}},
		{"v1.graph.rebaseline-resource-incomplete.001", []string{"24.3", "34.3"}},
	} {
		path := casePath(item.id)
		b, e := os.ReadFile(filepath.Join(rootDir, "vectors", path))
		must(e)
		var c vectors.Case
		must(vectors.Decode(b, &c))
		if c.ID != item.id || c.Format != "totipo-case-v1" || c.Expected != "PASS" {
			panic("invalid r13 case identity")
		}
		must(vectors.ValidateShape(c, "semantic"))
		manifest.Cases = append(manifest.Cases, vectors.Entry{ID: c.ID, Category: strings.Split(c.ID, ".")[1], Kind: "semantic", Normative: true, Path: path, Expected: c.Expected, Sections: item.sections, SHA256: vectors.Hash(b)})
	}
}
