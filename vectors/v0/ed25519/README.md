# Strict Ed25519 profile

`phase2.json` contains 33 fixed Section 16 cases: 4 ACCEPT and 29 REJECT. The explicit mathematical review and read-only independent arithmetic checker are in [../../../review/ed25519/REVIEW.md](../../../review/ed25519/REVIEW.md).

The strict consumer checks point canonicality, small order, true unreduced-integer [L] subgroup membership for A/R, S<L, and the cofactorless equation. Go's standard verifier accepts three supplied cases that Totipo rejects; see the review for details. Normal tests never generate expectations.

The missing Phase 1 corpus blocker has been addressed by independently establishing new expectations as authorized by Phase 2. This is explicit author review with separate arithmetic, not a claim of an external human audit.
