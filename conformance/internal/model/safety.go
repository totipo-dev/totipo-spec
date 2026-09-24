package model

// UseProfile is application policy, not consensus validity. Values are already
// reconstructed from complete atomic credential revisions, not merged subfields.
type UseProfile struct {
	Degraded, Incomplete, LifecycleConflict bool
	StatusValues, CredentialValues          []string
	CredentialComplete                      bool
	IssuerConflict, AccountConflict         bool
}
type UseDecision struct{ Allowed, Active, RecoveryVisible, MetadataConflict bool }

func AssessUse(p UseProfile) UseDecision {
	distinct := func(xs []string) map[string]bool {
		m := map[string]bool{}
		for _, x := range xs {
			m[x] = true
		}
		return m
	}
	status := distinct(p.StatusValues)
	credentials := distinct(p.CredentialValues)
	live := len(status) == 1 && status["LIVE"]
	allowed := !p.Degraded && !p.Incomplete && !p.LifecycleConflict && live && p.CredentialComplete && len(credentials) == 1
	return UseDecision{Allowed: allowed, Active: allowed, RecoveryVisible: true, MetadataConflict: p.IssuerConflict || p.AccountConflict}
}
