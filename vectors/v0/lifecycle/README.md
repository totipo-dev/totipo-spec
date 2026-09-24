# Initial symbolic lifecycle corpus

The 15 cases in `cases.json` are manually specified from r36 Sections 22–41 and 48–56. They assume authenticated, structurally valid, signature-verified token updates. Symbolic IDs and credential labels are not wire encodings.

They cover creation, disjoint composition, equal-value lineage conflict, inherited conflicts beneath a single causal head, ordinary field resolution, deletion/rotation races, persistent historical witnesses, invalid mixed-field resolution routing, provisional/final status-only coverage, stale witness reappearance, missing dependencies, redundant ancestry, tombstone credential restrictions, and ordinary restore+credential writes when PRE is empty.

These are initial spec-derived examples, not a complete externally reviewed semantic oracle corpus. The current model does not claim writer/publication/recovery conformance.
