package requirements

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"totipo/conformance/internal/corpus"
)

// LoadCases verifies the pinned specification and manifest, every artifact in
// that manifest, and the current corpus's own complete manifest. vectorRoot may
// override the corpus location; it does not override any integrity requirement.
func (p *Profile) LoadCases(repoRoot, vectorRoot string) ([]corpus.Case, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if vectorRoot == "" {
		vectorRoot = filepath.Join(repoRoot, filepath.FromSlash(p.Vectors.Root))
	}
	if err := verifyHash(filepath.Join(repoRoot, filepath.FromSlash(p.Spec.Path)), p.Spec.SHA256); err != nil {
		return nil, err
	}
	manifestPath := filepath.Join(repoRoot, filepath.FromSlash(p.Vectors.ManifestPath))
	if err := verifyHash(manifestPath, p.Vectors.ManifestSHA256); err != nil {
		return nil, err
	}
	manifest, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	if len(manifest) == 0 || manifest[len(manifest)-1] != '\n' {
		return nil, errors.New("pinned manifest must be nonempty and newline-terminated")
	}
	var pinned []corpus.Case
	previous := ""
	for _, line := range strings.Split(string(manifest[:len(manifest)-1]), "\n") {
		if len(line) < 67 || !hashPattern.MatchString(line[:64]) || line[64:66] != "  " {
			return nil, errors.New("invalid pinned manifest entry")
		}
		name := line[66:]
		if !relativePath(name) || !strings.HasSuffix(name, ".json") || name <= previous {
			return nil, errors.New("pinned manifest paths must be sorted, unique, relative JSON paths")
		}
		previous = name
		path := filepath.Join(vectorRoot, filepath.FromSlash(name))
		if err := verifyHash(path, line[:64]); err != nil {
			return nil, fmt.Errorf("pinned vector artifact: %w", err)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		cases, err := corpus.Decode(b)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		pinned = append(pinned, cases...)
	}
	// A required ID must originate in the pinned files, not an unpinned extra.
	if _, err := p.Select(pinned); err != nil {
		return nil, err
	}
	cases, err := corpus.Load(vectorRoot)
	if err != nil {
		return nil, err
	}
	return p.Select(cases)
}
