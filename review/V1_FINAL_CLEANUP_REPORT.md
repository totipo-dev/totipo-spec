# V1 final cleanup report

Baseline: `main` at `8042729fe6b1c00fc57e284000f39c060d76d773` (`v1 work`).
Only the local cleanup instructions were untracked before this pass. No historical
refs, tags, releases, or Git history were changed. Changes are left for the user's
manual commit; this report is a durable artifact to include. The local
`TOTIPO_SPEC_CLEANUP_AGENT_INSTRUCTIONS.md` remains excluded from staging/commits.

## Removed transition artifacts

Removed these exact files after confirming their transition-only purpose:

```text
AGENT_INSTRUCTIONS.md
PROPOSED_TREE.md
SHA256SUMS.txt
copy-to-repo/README.md
copy-to-repo/requirements/README.md
copy-to-repo/review/V1_R8_FINAL_CONSISTENCY_ADVERSARIAL_REVIEW.md
copy-to-repo/review/V1_R8_SIMPLIFICATION_REVIEW.md
copy-to-repo/review/V1_R9_FORWARD_COMPAT_REVIEW.md
copy-to-repo/spec/totipo-vault-format-v1.md
copy-to-repo/tools/check_spec.py
copy-to-repo/vectors/CASE_PLAN.md
copy-to-repo/vectors/README.md
copy-to-repo/vectors/manifest.json
copy-to-repo/vectors/manifest.schema.json
review/V1_VECTOR_FILE_CHANGES.tsv
vectors/CASE_PLAN.md
```

The copied seed files and proposed tree matched the bundle checksums. The old
instruction file's checksum was stale; its content was inspected and still
contained only the reset/implementation phases, not repurposed project guidance.
All 62 planned case IDs were already represented by manifest entries and existing
files; useful encoding/fixture guidance was already in `vectors/FORMAT.md`.
The manifest remains the live contract. The root/tree review found no further
transition artifacts needing removal.

## Documentation/report cleanup

The implementation report now separates historical agent-session circumstances
from the accepted implementation commit. The exact-commit
[GitHub conformance run](https://github.com/totipo-dev/totipo-spec/actions/runs/36197834195)
was verified as successful through the Actions API. The remote refs API confirms
`archive/v0` at `5157f14e1af9d28b6f15db20a51cd384f97c8a37`, matching local and
remote-tracking refs. Remote tags/releases were not enumerated or modified.
The giant file appendix was replaced by reset/implementation commit references.

Exact existing files modified in this cleanup:

```text
README.md
conformance/README.md
conformance/cmd/totipo-vector-gen/main.go
conformance/internal/vectors/runner.go
conformance/internal/vectors/types.go
requirements/v1-pre-rc.json
review/V1_VECTOR_IMPLEMENTATION_REPORT.md
spec/totipo-vault-format-v1.md
vectors/FORMAT.md
vectors/README.md
vectors/case.schema.json
vectors/manifest.json
```

## Spec editorial changes

Only three rationale paragraphs changed: sections 14 (envelope tradeoffs),
30 (whole-state lifecycle rationale), and 51 (semantic redundancy/privacy).
Internal v0/r1 comparisons became version-independent explanations. Compatibility,
migration, normative requirements, wire constants, and section numbering remain.
The moving profile's spec checksum was updated for this textual change.

- r9 wire bytes changed: **NO**.
- r9 semantics changed: **NO**.
- Existing vector case bytes/hashes changed: **NO**; all 68 match the baseline.
- Spec ambiguity found: **none**.

## TOTP vectors added/skipped

Added three cases, each covering the six standard timestamps (18 known answers):

- `v1.totp.rfc6238-sha1.001`
- `v1.totp.rfc6238-sha256.001`
- `v1.totp.rfc6238-sha512.001`

Expected codes come from [RFC 6238 Appendix B](https://www.rfc-editor.org/rfc/rfc6238.html#appendix-B),
using Appendix A's explicit 20/32/64-byte algorithm-specific secrets. Each fixture
records provenance, secret bytes, T0, period, digits, timestamp, integer counter,
u64be counter hex, and expected eight-digit code. Expectations are transcribed
constants, not generated from the tested implementation. The same Go consumer
executes them using standard-library HMAC/SHA primitives; no dependency was added.
Tests also cover leading zeros, time-step boundaries, credential bounds, and the
unsigned timestamp domain. The manifest, case schema, and moving profile were
updated. Final corpus: **71 cases** (28 byte, 6 negative, 37 semantic).

Exact new files, including this report:

```text
conformance/cmd/totipo-vector-gen/totp.go
conformance/internal/totp/totp.go
conformance/internal/totp/totp_test.go
review/V1_FINAL_CLEANUP_REPORT.md
vectors/cases/totp/v1.totp.rfc6238-sha1.001.json
vectors/cases/totp/v1.totp.rfc6238-sha256.001.json
vectors/cases/totp/v1.totp.rfc6238-sha512.001.json
```

## Validation results

- PASS: structural spec check, Go tests, Go vet, `make check`, `make conformance`
  (71/71), `make verify`, `make race`, and the existing three bounded fuzz targets.
- PASS: schema contracts, moving-profile pins, `git diff --check`, formatting,
  documentation links, and baseline comparison of all 68 existing case hashes.
- CI-facing command set still passes locally. The accepted baseline's remote CI
  is green; this unpushed cleanup has not run remotely.
- Nix changes: **none**. Existing flake, lockfile, `.envrc`, and tooling were
  preserved. `nix` is unavailable in this agent environment, so a fresh
  `nix develop --command make check` could not be executed. Checks used the existing
  development environment and a writable `.direnv/go-build` cache.

## Remaining pre-RC work

Start the first live Totipo implementation as the independent vector consumer.
Retain the documented storage/crash, operation-freshness, native P-256, broader
adversarial coverage, and external security-review work before RC freeze.
No second implementation, RC profile, tag, release, or push was created here.
