# Go v1/r15 reference/conformance consumer

This Go 1.23+ module is the single in-repository reference consumer, not a production
client. Run `make check`, `make race`, and `make fuzz` from the repository root.
The CLI also accepts `-root /path/to/repository` and `-verify-only`.

`make conformance` executes the 93 baseline requirements with advisory history
support disabled. `make conformance-all` executes baseline plus the 12 conditional
cases using the reference model's explicit `ReferenceCapabilities` declaration.
For example, `go run ./conformance/cmd/totipo-conformance -root . -capability advisory-history`
reports baseline and `capability=advisory-history` results separately. Unknown
capability names reject. The complete manifest has 105 schema-valid, hash-pinned
cases; `make check` tests both selections. Downstream implementations without an
advisory-history feature need only baseline; they need not retain cross-run graph
history or emit HISTORY_MEMORY_LOST merely because no feature exists. Capability
selection is a conformance concept, not wire negotiation.

Packages separate unchanged TLV, crypto, object parsing, and TOTP from snapshot
graph interpretation and vector IO. External bytes pass `cryptov1.Keys.Open`
before dispatch. Invalid supported bodies never become future/opaque evidence.

`graph.State` contains one explicitly supplied accepted observation. `Snapshot`
copies planning evidence; `Plan` chooses exact parents from it. Later observations
do not stale ordinary publication. Confirmed conflict decisions track exact TOKEN
heads, complete values, desired value, and intent. TOKEN-relevant changes learned
before publication require reconfirmation; unrelated events do not.

When the advisory-history capability is selected, remembered IDs only derive diagnostics and cannot add nodes/edges/values.
Disappearance removes current evidence in a subsequent observation. `author` means
ordinary authorship eligibility (including empty new-token state), while
`requires_confirmation` identifies a visible supported conflict. Cache health,
incomplete discovery, and unscoped evidence are warnings, not global gates.

`internal/storage` evaluates exact namespace, filename, ordinary-file, bounded
reader, and cryptographic obligations. It also models complete/no-overwrite local
publication and binding establishment. These abstract API outcomes do not prove
physical crash durability or hidden remote completeness. Go race testing remains
useful for implementation data races, not synchronized transaction claims.

The first DEVICE workflow requires matching valid self-signed advertisement and
TOKEN publication acknowledgements. Optional caches are independent. Provenance
re-evaluation changes attribution without changing TOKEN values or causality.

Tests never regenerate crypto fixtures. The obsolete monolithic r14 generator was
removed because it encoded the retired state model; fixed byte fixtures remain
committed and independently pinned. See [FORMAT.md](../vectors/FORMAT.md) for the
strict language-neutral contract and [r15 review](../review/V1_R15_STATE_MODEL_SIMPLIFICATION_REPORT.md)
for every semantic delta. An independent live implementation is still required
before RC freeze. No downstream implementation is modified here.
