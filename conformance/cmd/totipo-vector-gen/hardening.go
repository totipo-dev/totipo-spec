package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"totipo/conformance/internal/vectors"
)

// r12 workflows have hand-authored expectations; carry their reviewed bytes
// forward unchanged. They are intentionally not derived from the evaluator.
func hardeningCases() {
	for _, item := range []struct {
		id, kind string
		sections []string
	}{
		{"v1.future.unscoped-sticky.001", "semantic", []string{"24.2", "37.3"}},
		{"v1.future.unscoped-rebaseline-absent.001", "semantic", []string{"24.2", "34.4", "37.3"}},
		{"v1.future.unscoped-rebaseline-present.001", "semantic", []string{"24.2", "34.4", "37.3"}},
		{"v1.device.rename-incorporates-rejected-head.001", "semantic", []string{"31", "49"}},
		{"v1.graph.known-id-corrupt-bytes-retain-node.001", "semantic", []string{"24.7"}},
		{"v1.graph.local-security-memory-corruption.001", "semantic", []string{"24.7", "34"}},
		{"v1.provenance.initial-device-before-token.001", "semantic", []string{"16.2"}},
		{"v1.provenance.signature-context-cross-vault.001", "bytes", []string{"19"}},
	} {
		category := strings.Split(item.id, ".")[1]
		path := "cases/" + category + "/" + item.id + ".json"
		b, e := os.ReadFile(filepath.Join(rootDir, "vectors", path))
		must(e)
		var c vectors.Case
		must(json.Unmarshal(b, &c))
		if c.ID != item.id || c.Format != "totipo-case-v1" || c.Expected != "PASS" {
			panic("invalid r12 case identity")
		}
		must(vectors.ValidateShape(c, item.kind))
		must(vectors.Run(c))
		manifest.Cases = append(manifest.Cases, vectors.Entry{ID: c.ID, Category: category, Kind: item.kind, Normative: true, Path: path, Expected: c.Expected, Sections: item.sections, SHA256: vectors.Hash(b)})
	}
}
