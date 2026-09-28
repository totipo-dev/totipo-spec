# Requirements Profiles

[`v1-pre-rc.json`](v1-pre-rc.json) is the **moving baseline** profile for v1/r15.
Its `required_cases` is exactly the manifest's 93 baseline IDs and case-file hashes.
It also pins the specification, entire manifest, manifest schema, and case schema.
The manifest contains 105 cases, including 12 conditional `advisory-history` cases.
Conditional cases remain hash-pinned and validated but are not baseline requirements.

`make conformance` runs baseline. `make conformance-all` (equivalently
`go run ./conformance/cmd/totipo-conformance -root . -capability advisory-history`)
runs baseline plus that capability. The included reference model explicitly declares
support and `make check` executes both suites. No separate capability profile or
wire negotiation is needed. A downstream client with no advisory-history feature
can conform to the baseline profile without emitting history-derived diagnostics.

`make verify` checks all manifest files, applicability, checksums, exact physical
case coverage, and baseline profile membership. Unknown capabilities and duplicate
IDs reject. Pre-RC case changes require review; checks never regenerate fixtures.

Do not freeze an RC profile until an independent live implementation consumes the
corpus and required security/platform reviews are complete. Optional capability
selection does not change wire compatibility.
