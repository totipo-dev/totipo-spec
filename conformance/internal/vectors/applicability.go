package vectors

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const AdvisoryHistory = "advisory-history"

// ReferenceCapabilities explicitly declares optional functionality implemented by
// the included reference model. It is not wire negotiation or a baseline gate.
var ReferenceCapabilities = []string{AdvisoryHistory}

type Applicability struct {
	Kind       string  `json:"kind"`
	Capability *string `json:"capability,omitempty"`
}

func (a *Applicability) UnmarshalJSON(b []byte) error {
	type plain Applicability
	var value plain
	if err := Decode(b, &value); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return err
	}
	if raw, ok := fields["capability"]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return fmt.Errorf("null capability")
	}
	*a = Applicability(value)
	return a.Validate()
}
func (a Applicability) Validate() error {
	if a.Kind == "baseline" && a.Capability == nil {
		return nil
	}
	if a.Kind == "conditional" && a.Capability != nil && *a.Capability == AdvisoryHistory {
		return nil
	}
	return fmt.Errorf("invalid applicability kind/capability")
}
func (a Applicability) Label() string {
	if a.Kind == "conditional" && a.Capability != nil {
		return "capability=" + *a.Capability
	}
	return "baseline"
}
func validateCapabilities(capabilities []string) error {
	seen := map[string]bool{}
	for _, c := range capabilities {
		if c != AdvisoryHistory || seen[c] {
			return fmt.Errorf("unknown/duplicate capability %q", c)
		}
		seen[c] = true
	}
	return nil
}

// Select selects baseline plus explicitly requested optional capability cases.
// All manifest files are still validated and hash-pinned, even when not selected.
func Select(m Manifest, cases []Case, capabilities []string) ([]Case, error) {
	if err := validateCapabilities(capabilities); err != nil {
		return nil, err
	}
	index := map[string]Case{}
	for _, c := range cases {
		if _, ok := index[c.ID]; ok {
			return nil, fmt.Errorf("duplicate case")
		}
		index[c.ID] = c
	}
	out := []Case{}
	seen := map[string]bool{}
	for _, e := range m.Cases {
		if err := e.Applicability.Validate(); err != nil {
			return nil, err
		}
		if seen[e.ID] {
			return nil, fmt.Errorf("duplicate manifest ID")
		}
		seen[e.ID] = true
		c, ok := index[e.ID]
		if !ok {
			return nil, fmt.Errorf("missing case %s", e.ID)
		}
		if e.Applicability.Kind == "baseline" || len(capabilities) > 0 {
			out = append(out, c)
		}
	}
	return out, nil
}
