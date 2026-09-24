package main

import "fmt"

func safetyFixtures() {
	var cs []C
	add := func(id string, changes M, allowed, active, metadata bool) {
		in := M{"degraded": "false", "incomplete": "false", "lifecycle_conflict": "false", "status_values": "LIVE", "credential_values": "k0", "credential_complete": "true", "issuer_conflict": "false", "account_conflict": "false"}
		for k, v := range changes {
			in[k] = v
		}
		disposition := "BLOCK"
		if allowed {
			disposition = "ALLOW"
		}
		cs = append(cs, C{ID: "v0/transitions/use-" + id, Kind: "use_profile", Description: "Application safety: " + id, Provenance: P{"review/phase2/REVIEW.md", "spec-derived-reviewed", "r36 68.1; use-" + id}, Input: in, Expected: M{"disposition": disposition, "active": fmt.Sprint(active), "recovery_visible": "true", "metadata_conflict": fmt.Sprint(metadata)}})
	}
	add("healthy", M{}, true, true, false)
	add("regression", M{"degraded": "true"}, false, false, false)
	add("incomplete", M{"incomplete": "true"}, false, false, false)
	add("lifecycle", M{"lifecycle_conflict": "true"}, false, false, false)
	add("ambiguous-status", M{"status_values": "LIVE,TOMBSTONE"}, false, false, false)
	add("tombstone", M{"status_values": "TOMBSTONE"}, false, false, false)
	add("credential-ambiguous", M{"credential_values": "k0,k1"}, false, false, false)
	add("credential-missing", M{"credential_values": "", "credential_complete": "false"}, false, false, false)
	add("credential-incomplete", M{"credential_complete": "false"}, false, false, false)
	add("equal-credential-values", M{"credential_values": "k0,k0"}, true, true, false)
	add("equal-live-values", M{"status_values": "LIVE,LIVE"}, true, true, false)
	add("issuer-conflict", M{"issuer_conflict": "true"}, true, true, true)
	add("account-conflict", M{"account_conflict": "true"}, true, true, true)
	add("tombstone-conflict-discoverable", M{"status_values": "TOMBSTONE", "lifecycle_conflict": "true"}, false, false, false)
	write("vectors/v0/transitions/use-profile.json", struct {
		Schema int `json:"schema"`
		Cases  []C `json:"cases"`
	}{1, cs})
	fmt.Println("authored", len(cs), "use profile cases")
}
