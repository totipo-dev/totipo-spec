package vectors

import (
	"bytes"
	"fmt"
	"totipo/conformance/internal/storage"
)

type LocalCase struct {
	Trials []LocalTrial `json:"trials"`
}
type LocalTrial struct {
	Operation     string `json:"operation"`
	Existing      string `json:"existing"`
	Acknowledged  bool   `json:"acknowledged"`
	Authenticated bool   `json:"authenticated"`
	Expected      string `json:"expected"`
}

func runLocal(c Case) error {
	for i, t := range c.Local.Trials {
		var got string
		if t.Operation == "install" {
			intended := bytes.Repeat([]byte{0x42}, 1024) // abstract transport bytes, not a crypto fixture
			kind := "regular"
			var existing []byte
			switch t.Existing {
			case "absent":
				kind = "absent"
			case "exact":
				existing = bytes.Clone(intended)
			case "different":
				existing = bytes.Repeat([]byte{0x43}, 1024)
			case "symlink":
				kind = "symlink"
			default:
				return fmt.Errorf("unknown target")
			}
			before := bytes.Clone(existing)
			result, status := storage.Install(existing, intended, kind, t.Acknowledged)
			got = status
			if kind != "absent" && !bytes.Equal(result, before) {
				return fmt.Errorf("existing object overwritten")
			}
			if got == "PUBLISHED_NEW" && !bytes.Equal(result, intended) {
				return fmt.Errorf("partial publication")
			}
		} else if t.Operation == "binding" {
			derived := bytes.Repeat([]byte{0x11}, 32)
			var existing []byte
			switch t.Existing {
			case "absent":
			case "exact":
				existing = bytes.Clone(derived)
			case "different":
				existing = bytes.Repeat([]byte{0x12}, 32)
			case "corrupt":
				existing = []byte{1}
			default:
				return fmt.Errorf("unknown binding")
			}
			got = storage.Establish(existing, derived, t.Existing != "absent", t.Authenticated, t.Acknowledged)
		} else {
			return fmt.Errorf("unknown local operation")
		}
		if got != t.Expected {
			return fmt.Errorf("local trial %d: %s want %s", i, got, t.Expected)
		}
	}
	return nil
}
