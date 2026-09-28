package storage

import "bytes"

// Install models the complete/no-overwrite local API boundary. It intentionally
// has no fsync-existing-file or history-journal parameters. A real adapter must
// supply atomic/complete visibility and report acknowledgement truthfully.
func Install(existing, intended []byte, kind string, acknowledged bool) ([]byte, string) {
	if len(intended) != 1024 {
		return existing, "FAILED"
	}
	if kind != "absent" {
		if kind != "regular" || !bytes.Equal(existing, intended) {
			return existing, "FAILED"
		}
		return existing, "ALREADY_PRESENT_EXACT"
	}
	if !acknowledged {
		return existing, "UNKNOWN"
	}
	return bytes.Clone(intended), "PUBLISHED_NEW"
}

// Establish separates absence from corrupt/mismatching authoritative anchors.
// Authenticated and durable are supplied trusted local API outcomes.
func Establish(existing, derived []byte, present, authenticated, durable bool) string {
	if !authenticated || len(derived) != 32 {
		return "REJECTED"
	}
	if present {
		if len(existing) != 32 {
			return "ANCHOR_FAILURE"
		}
		if !bytes.Equal(existing, derived) {
			return "REJECTED"
		}
		return "ESTABLISHED"
	}
	if !durable {
		return "INCOMPLETE"
	}
	return "ESTABLISHED"
}
