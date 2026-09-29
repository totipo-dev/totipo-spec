package storage

import "bytes"

// Install classifies backend outcomes, not host syscalls. On failure the
// returned observation does not assert absence or describe effects of an
// ambiguous backend operation. Durable means required persistence succeeded.
func Install(existing, intended []byte, kind string, durable bool) ([]byte, string) {
	if len(intended) != 1024 {
		return existing, "FAILED"
	}
	if kind != "absent" {
		if kind != "regular" || !bytes.Equal(existing, intended) {
			return existing, "FAILED"
		}
		return existing, "ALREADY_PRESENT_EXACT"
	}
	if !durable {
		return existing, "FAILED"
	}
	return bytes.Clone(intended), "PUBLISHED_NEW"
}

// Create needs canonical lowercase `vault` absence, not any inventory of object-looking files.
func Create(kind string, complete, durable bool) string {
	if kind != "absent" || !complete || !durable {
		return "FAILED"
	}
	return "CREATED"
}

// Replace compares exact BASE to freshly observed CURRENT at canonical `vault`. It does not model
// atomic CAS and cannot rule out a race after this observation.
func Replace(base, current []byte, kind string, readable, complete, durable bool) string {
	if !BootstrapCandidate("vault", kind) || !readable {
		return "FAILED"
	}
	if !bytes.Equal(base, current) {
		return "STALE"
	}
	if !complete || !durable {
		return "FAILED"
	}
	return "REPLACED"
}
