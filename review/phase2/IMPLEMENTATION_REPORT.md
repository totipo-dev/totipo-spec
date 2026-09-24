# Totipo r36 Phase 2 implementation report

The Phase 2 abstract conformance layer is implemented and reviewed against `spec/totipo-vault-format-v0.md`. The runner reports **388 PASS / 0 FAIL / 0 BLOCKED** for the supplied corpus. Formatting, vet, deterministic tests, race tests, the independent arithmetic review and integrity checks pass. This does not claim a production vault implementation, an external security audit, or complete OS/filesystem conformance.

## Baseline and preserved evidence

Before changes, the requested baseline commands passed: 270 cases, zero failures, and a passing race run. The Phase 1 report recorded one blocked deliverable (missing strict Ed25519 acceptance corpus), while its old CLI did not print BLOCKED counts. Stateful work was explicitly incomplete, not counted as passing empty categories.

All 11 pre-existing JSON vector files still match the saved Phase 1 manifest byte-for-byte. All 17 historical review artifacts still match the unchanged `conformance/review-inventory.sha256`. The spec, frozen literals and historical review inputs were not edited or regenerated.

## Final corpus counts

| Category | PASS | FAIL | BLOCKED |
|---|---:|---:|---:|
| TLV | 65 | 0 | 0 |
| Envelope/file-length gates | 13 | 0 | 0 |
| Object crypto | 2 | 0 | 0 |
| Bootstrap/password | 37 | 0 | 0 |
| Dispatch/processing order | 56 | 0 | 0 |
| TOTP | 82 | 0 | 0 |
| Strict Ed25519 | 33 | 0 | 0 |
| Retained lifecycle scenarios | 15 | 0 | 0 |
| Transitions and application-use policy | 46 | 0 | 0 |
| Recovery/security-memory traces | 28 | 0 | 0 |
| DEVICE_UPDATE presentation | 11 | 0 | 0 |
| **Total** | **388** | **0** | **0** |

The transition category comprises 32 token-history scenarios and 14 application-use policy cases. Thus **47 token/lifecycle scenarios** (15 retained + 32 new) run through both token evaluators. The 28 recovery traces contain **229 events**, each with a fixed outcome and selected checkpoint assertions, consumed by both memory evaluators. Pending DEVICE_UPDATE abandonment/re-entry and presentation rollback are additionally covered in those recovery traces.

Five new normative JSON files contribute 118 cases:

```text
vectors/v0/ed25519/phase2.json            33
vectors/v0/transitions/phase2.json       32
vectors/v0/transitions/use-profile.json  14
vectors/v0/recovery/phase2.json          28
vectors/v0/presentation/phase2.json      11
```

Every new case and provenance class is enumerated in [review/phase2/CASE_INDEX.md](CASE_INDEX.md): 115 `spec-derived-reviewed`, two `reviewed-pinned`, one `published-standard`. “Reviewed” records the agent's explicit author review with separate calculations, not an outside human reviewer or organization. No random/property/fuzz result was automatically promoted into normative expectations. Explicit derivations and assumptions are in [the semantic/state review](REVIEW.md) and [the Ed25519 mathematical review](../ed25519/REVIEW.md).

## Strict Ed25519

The fixed corpus has **4 positive and 29 negative cases**. It includes the RFC 8032 empty-message vector, both pinned Totipo signatures, and one signature produced by the isolated affine arithmetic signer. Negative families cover lengths, nonsquare point recovery, noncanonical y and negative-zero encodings, identity/order-two points, mixed-order A/R, unreduced [L] membership, S==L/S>L/maximal S, mutations and canonical-but-invalid equations.

The isolated reviewer under `review/ed25519/` uses affine `math/big` arithmetic and SHA-512. It imports neither the consumer nor a curve library nor standard Ed25519 verification. It reproduces the RFC signing anchor and independently checks each manually specified rejection predicate. Mixed-order signatures are explicitly constructed to satisfy the cofactorless equation while failing subgroup membership, so vacuous modulo-L subgroup checks cannot pass the corpus.

The strict consumer uses point decoding/addition from **`filippo.io/edwards25519 v1.1.0`**, checks exact canonical round-trips, [8]P, true integer [L]P for A/R, and S<L, then uses the standard PureEd25519 equation verifier. `golang.org/x/crypto/ed25519` is a wrapper around the standard API and does not expose the required point arithmetic. This small pinned dependency is the only new external dependency; the existing x/crypto v0.36.0 and indirect x/sys v0.31.0 remain unchanged.

On Go 1.26.7, the standard verifier accepts **three** cases rejected by Totipo: `identity-forgery`, `mixed-equation-0`, and `mixed-equation-1`. Both accept all four positive cases and reject the other 26 negative cases. Wrong public-key lengths are guarded before the standard API call to avoid its panic contract. The comparison is diagnostic, never the source of expected results.

## Semantics, independent oracles and properties

The primary token model retains direct implicit field-parent traversal and JOIN. The separate reference oracle reconstructs explicit causal closure, projects maximal field assertions, scans historical credential witnesses, calculates confirmation/coverage and filters maximal unresolved witnesses without calling primary traversal/JOIN/validation helpers. Both also check explicit resolution-coverage maps. New fixture features include symbolic signer labels, unavailable dependencies and selected observation frontiers.

The expanded corpus covers required ordinary field operations/conflicts, equal-value lineage, inherited conflicts, deletion/restoration races, multiple status/witness branches, historical non-head witnesses, same-value resolutions, inherited final coverage, invalid mixed-field candidates, PRE-empty ordinary updates and credential/tombstone restrictions. Existing invalid-candidate provisional-coverage mutation checks remain active.

Non-normative fixed-seed tests executed:

- **24 differential histories**, each with 24 candidate attempts: **576 candidate comparisons**, **523 accepted** and **53 rejected** candidates.
- Hidden/revealed views and arrival-order permutations bring the total to **1,680 primary/reference comparisons**.
- **216 randomized JOIN triples**, exercising **648 JOIN-law assertions** (idempotence, commutativity, associativity), plus **543 typed-field-implies-causal-ancestry checks**.
- All **three deliberately broken oracle variants** are distinguished by named corpus cases: current-head-only witnesses, reversed coverage direction, and failure to inherit coverage through STATUS ancestry.
- A forced nonempty candidate POST test verifies no failed provisional coverage leaks into accepted history. A separate semantic 32/33-parent test checks that unencodable operations do not truncate 33 token/credential heads.

Generated histories are test evidence only; no generated outputs become vector expectations.

## Memory, presentation, establishment and publication

The primary memory machine uses explicit durable evidence maps. Its independent reviewer uses a flat durable-fact ledger, a transitive-closure relation and fixed-point reconstruction of available dependencies. They share only event DTOs and independently consume every authored trace.

Traces cover token-local pending evidence, durability before acknowledged processing, disappearance/restart, exact-ID/vault/scope-bound abandonment and warning acknowledgement, new unacknowledged pending IDs, later-valid re-entry, durable frontier/terminal-invalidity handoff, vault-wide future evidence/bypass, compatible durable handoff, known-history regression/restoration, conservative remembered-head replacement, non-atomic partial batches, presentation abandonment/re-entry/rollback, confirmation context and future-epoch freshness, before/after publication linearization, establishment CAS and crash points, and capacity-blocked migration.

Additional focused tests check a changed canonical bootstrap binding immediately before establishment persistence, desired-status binding of confirmation, and fresh parentless migration-root materialization with all required fields and fresh device/token/binding labels. Migration leaves source history intact and makes no continuity/retirement/erasure claim. Cryptographic freshness and exact binding derivation are upstream preconditions of these symbolic labels.

The separate DEVICE_UPDATE evaluator enforces same-key/type dependencies, pending propagation, duplicate/redundant parent rejection and maximal presentation heads. Identical displayed values may coalesce while separate heads remain. Token validity is not derived from presentation state.

Fourteen application policy fixtures independently test ordinary OTP use/export and unqualified active presentation. Degraded/incomplete/lifecycle-conflicted, ambiguous/non-LIVE status or ambiguous/incomplete credential state blocks ordinary use. Equal complete credential values may coalesce without changing lineage. Metadata conflicts remain surfaced without necessarily blocking use, and tombstoned conflicts remain recovery-discoverable.

The abstract publication Store has **eight focused test groups**, including five injected interruption points and eight preexisting/hostile-entry combinations. It checks safe exclusive creation, no-follow/special-file policy, path-component rejection, complete temporary writes, file and directory durability, candidate bootstrap validation before installation, immutable/no-replace installation, same-root replacement and the writer gate at linearization. It does not claim real syscall or power-loss guarantees. User-visible token acceptance and vault establishment still require the separate durable local-memory steps; successful storage publication alone does not imply them.

## Validation and reproduction

Local toolchain: **Go 1.26.7 linux/amd64**, **GCC 15.3.0**. The module language floor remains Go 1.23. Other platforms/versions are configured in CI, not claimed as locally executed.

All required final checks passed:

```sh
gofmt -l conformance tools
go -C conformance fmt ./...
go -C conformance vet ./...
go -C conformance test -count=1 ./...
go -C conformance test -race ./...
go run ./conformance/cmd/totipo-conformance ./vectors/v0
(cd vectors/v0 && sha256sum -c manifest.sha256)
sha256sum -c conformance/review-inventory.sha256
```

Additional read-only review/property checks:

```sh
go run ./review/ed25519/review.go
sha256sum -c review/phase2/inventory.sha256
go -C conformance test ./internal/runner ./internal/model -run 'TestEd25519StandardComparison|TestBrokenOracleVariants|TestDeterministicDifferentialHistories|TestRandomizedJoinAndAncestry' -count=1 -v
```

The environment used `GOMODCACHE=/tmp/totipo-gomod` and `GOCACHE=/tmp/totipo-gocache`. The point-library dependency was downloaded explicitly with approval. Normal tests and CI only read fixed expectations. The manifest was updated with `go run ./tools/manifest/main.go --write-manifest`, which hashes existing JSON without rewriting vectors. The historical review inventory is unchanged; new review documents/source have their own inventory.

Fuzz smoke results (two seconds requested, two workers; not exhaustive):

| Target | Result | Executions |
|---|---|---:|
| FuzzTLVFraming | PASS | 90,488 |
| FuzzV0Grammar | PASS | 81,729 |
| FuzzCredentialGrammar | PASS | 85,274 |
| FuzzEnvelopeUnframe | PASS | 97 |
| FuzzCommonPrefixDispatch | PASS | 102,531 |

Reproduce each with `go -C conformance test ./internal/runner -run '^$' -fuzz '^TARGET$' -fuzztime=2s -parallel=2`. No additional stateful fuzz target was added; deterministic stateful and fixed-seed differential coverage precede further fuzzing.

CI now also runs the independent arithmetic reviewer without its write flag and verifies both historical and Phase 2 review inventories on Linux. Formatting includes the new tools/reviewer. The CLI prints PASS/FAIL/BLOCKED separately, returns nonzero for blocked cases or missing required categories, and never reports a missing category as PASS with zero cases.

## Remaining limits and issues

There is no remaining missing-source BLOCKED fixture in the supplied Phase 2 abstract corpus. The former Ed25519 missing-input blocker was closed by explicitly authorized independent mathematical construction/review. No genuine normative contradiction was found, so no `SPEC_ISSUES.md` contradiction entry was fabricated.

The following remain outside the demonstrated layer and must not be inferred from zero blocked corpus cases:

- Real OS storage adapters, path-rebinding/syscall behavior, multi-process races and actual filesystem/power-loss experiments. Phase 2 explicitly permits these after the abstract deterministic layer.
- End-to-end integration of authenticated files, durable local-memory adapters, user interaction and production writer operations. Symbolic `valid`, binding and identity-grounded proof events assume the corresponding upstream checks.
- CSPRNG-based destination-vault/device/token creation during migration; fixtures establish required fresh identities/root-history semantics using distinct labels.
- Exhaustive Section 74 or cryptographic assurance, external human review, independent organization audit, performance/side-channel certification, and execution of the configured CI matrix on remote runners.

Review confidence comes from explicit spec derivations, separate calculations, fixed vectors, mutation sensitivity and independent consumption. It is not a claim of full Totipo v0 conformance across an implemented client and storage platform.
