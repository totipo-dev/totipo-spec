# Totipo v0-rc1 requirements preparation

The working tree prepares a reviewable release candidate. No commit, push, tag,
GitHub Release, repository-setting change, dependency upgrade or PR merge was
performed. The incoming [instructions](V0_RC1_AGENT_INSTRUCTIONS.md) are retained
with their contents unchanged.

## r37 promotion

| Artifact | SHA-256 |
|---|---|
| Original canonical r36 | `90bb937fc376517b16cee48b07c64c32df1e687fa32354a19ea1b851bab3eff7` |
| Incoming `spec/totipo-vault-format-v0-r37.md` | `aa281a757be6c324e11b492d195f1b4e5c1604fe9bdf7d20cff9adbfe62bc409` |
| Final [canonical r37](../../../spec/totipo-vault-format-v0.md) | `aa281a757be6c324e11b492d195f1b4e5c1604fe9bdf7d20cff9adbfe62bc409` |

The incoming r37 matched the expected reviewed hash exactly. The complete diff
was reviewed before promotion. Sections 1–73 were compared byte-for-byte and
are identical. Section 74 differs only in its introductory temporal wording:
“Before byte-level v0 freeze, executable transition/conformance evidence MUST
cover at least:” becomes “The v0 conformance publication MUST cover at least:”.
The required evidence list is unchanged. Section 75 replaces stale outstanding
work with completed evidence, assurance limits and remaining release work.
Header/status/revision wording changes correspond to that bookkeeping.

No wire literal, tag, limit, enum, byte layout, cryptographic construction,
semantic transition, accepted history, writer rule, recovery rule or durability
rule changed. The new case/crash counts in Section 75 describe evidence, not
protocol limits. No promotion issue was found. The separate untracked r37 copy
was moved onto the canonical path; no duplicate or new r36 archive remains.

The exact supplied r37 Markdown includes two trailing spaces on its status line
for a hard line break. Unqualified `git diff --check` reports that line; it was
preserved to retain the reviewed hash. The rest of the diff passes whitespace
checking. Historical r36 references remain historical; current project-facing
documentation identifies r37.

## Frozen requirements profile

| Item | Value |
|---|---|
| Profile | [requirements/v0-rc1.json](../../../requirements/v0-rc1.json) |
| Profile SHA-256 | `d476f17f0f02cbee796ea74963ab35ca7e171a7b9f63ed9ccd471a6ce5372eea` |
| Spec SHA-256 | `aa281a757be6c324e11b492d195f1b4e5c1604fe9bdf7d20cff9adbfe62bc409` |
| Original and final current vector-manifest SHA-256 | `bf6a3a670ddf1acab773e44fe563f2e3d9392299fb9229bddb64042194e474a0` |
| Frozen manifest copy SHA-256 | `bf6a3a670ddf1acab773e44fe563f2e3d9392299fb9229bddb64042194e474a0` |
| Required IDs | **388**, lexicographically sorted, unique, each present exactly once |
| Expected results | **388 PASS / 0 FAIL / 0 BLOCKED** |

The maintenance tool uses the actual `corpus.Load` output: recursively loaded
schema-1 JSON bundles, validated against the complete current manifest. It
derives categories from IDs. The generated ID list and final CLI selection were
compared with the pre-change CLI's exact sorted 388-ID list. No artificial
status/documentation cases were included.

| Category | Count |
|---|---:|
| bootstrap | 37 |
| dispatch | 56 |
| ed25519 | 33 |
| envelope | 13 |
| lifecycle | 15 |
| object-crypto | 2 |
| presentation | 11 |
| recovery | 28 |
| tlv | 65 |
| totp | 82 |
| transitions | 46 |
| **Total** | **388** |

### Manifest pinning and corpus growth

The instructions ask both to pin an exact manifest hash and to tolerate future
corpus additions. The existing loader requires its manifest to list every JSON
file, so adding a file necessarily changes that moving manifest. To satisfy
both goals, `vectors.manifest_path` points to the exact retained copy
[requirements/v0-rc1.manifest.sha256](../../../requirements/v0-rc1.manifest.sha256).
This is an intentional departure from the example's moving manifest path.

Validation checks the pinned spec, the exact frozen manifest hash, every listed
artifact's bytes and required IDs, then the full current corpus through the
unchanged loader and its current manifest. Future additions belong in new JSON
files and require updating the current manifest for review. Existing pinned
files cannot change, even to append cases. A changed required expectation fails
with either a stale or recalculated current manifest. Invalid extra artifacts
still fail the full corpus loader; valid extra cases do not execute in profile
mode. A later spec revision likewise cannot silently satisfy the r37 hash.

The frozen copy is release metadata, not a changed normative vector or a
replacement for the current manifest. No current/historical inventory was
updated. All **23 existing files** under `vectors/v0/`, including its README and
manifest, remain byte-for-byte identical, and no files were added there. The
request to update the vector README was subordinated to the explicit constraint
freezing that entire tree; the distinction is documented in the root,
conformance and requirements READMEs instead.

## CLI, maintenance and CI

```sh
# Moving current corpus, preserving the original interface.
go run ./conformance/cmd/totipo-conformance ./vectors/v0

# Complete frozen portable profile.
go run ./conformance/cmd/totipo-conformance --requirements requirements/v0-rc1.json

# Integrity and membership only: not a conformance execution.
go run ./conformance/cmd/totipo-conformance --requirements requirements/v0-rc1.json --verify-only

# Diagnostic subset: explicitly PARTIAL, never CONFORMANT.
go run ./conformance/cmd/totipo-conformance --requirements requirements/v0-rc1.json --filter v0/totp/
```

Full runs execute exactly the frozen IDs and require expected counts to match.
Integrity/schema failures, absent/duplicate IDs, case failures, blocked cases or
empty filters return nonzero. Blocked execution retains exit code 2; failures
use 1. Even an explicitly empty filter is labeled partial and cannot print
`CONFORMANT`. Filtering does not bypass artifact validation. `--verify-only`
cannot be combined with filtering. Paths resolve from the repository containing
the profile under `requirements/`; a positional corpus override must satisfy
the same integrity checks.

The reusable package is `conformance/internal/requirements`. The maintenance
command is `conformance/cmd/totipo-requirements`, located inside the Go module
so it can use that internal package without a new module/dependency. Previewing
requires no write flag and produces JSON only. The original explicit invocation
was:

```sh
go run ./conformance/cmd/totipo-requirements --profile v0-rc1 --write
```

Writing creates a candidate JSON profile plus its exact manifest copy. Existing
output files are never overwritten. Once tagged, keep the profile immutable;
later candidates use new names. Generation asserts expected counts, not that
execution passed; running the candidate remains a separate check.

`make conformance-v0-rc1` and `make verify-requirements` were added; `make check`
includes them alongside existing tests, race tests and inventories. CI adds a
separate full profile execution in every portable OS/Go job. The existing
Linux integration/race distinction, minimum Go 1.23/stable matrix, dependency
versions and immutable action SHAs are preserved. CI never generates profiles
or vectors. `.gitattributes` preserves exact bytes for profile JSON/manifests.

## Validation

The baseline and final normal suites include Linux filesystem, process-crash,
restart, hostile-entry and race integration tests using the repository-local
`TMPDIR`. Initial sandboxed baseline normal/race attempts failed only because
Unix socket `setsockopt` was denied. They were rerun with the needed access;
no test expectation or storage implementation was changed to recover them.

| Check | Result |
|---|---|
| Baseline full normal suite, Go 1.26.7 | PASS, including Linux integration |
| Baseline full race suite, Go 1.26.7 | PASS, including Linux integration |
| Baseline current corpus | 388 / 0 / 0 |
| Final full normal suite, Go 1.26.7 | PASS, including Linux integration |
| Final full race suite, Go 1.26.7 | PASS, including Linux integration |
| Final current corpus | 388 / 0 / 0 |
| Final v0-rc1 profile | 388 / 0 / 0; `CONFORMANT v0-rc1` |
| Integrity-only verification | VALID, 388 required IDs |
| New package and both CLI packages, Go 1.23.12 | PASS |
| `gofmt -l conformance tools` | No output |
| `go -C conformance vet ./...` | PASS |
| Current vector manifest | All 16 JSON files OK |
| Entire vector tree before/after comparison | All 23 files unchanged |
| Historical review inventory | All 17 files OK |
| Phase 2 review inventory | All 4 files OK |
| Phase 3 source inventory | All 9 files OK |
| Updated Markdown links | All checked local targets exist |
| Actual Make execution | Unavailable: GNU Make is not installed in this shell |
| Remote CI | Not run by this task |

New regression tests cover valid r37/profile execution; changed spec, frozen
manifest and current manifest; missing file/case; duplicate and unknown IDs;
wrong totals/categories; malformed/ambiguous JSON; changed expected normative
data with and without recalculated manifest; an additional valid vector in a
temporary corpus; CLI filters, verification, nonzero failure/blocked results;
and maintenance preview, explicit writing and overwrite refusal. Mutations
and generated test fixtures are confined to temporary directories.

Final commands used plain Go and its default caches. The initial baseline used
`GOMODCACHE=/tmp/totipo-gomod GOCACHE=/tmp/totipo-gocache` to work within the
sandbox; those are agent-environment workarounds, not project prerequisites.
The user may run the Make targets in an environment with GNU Make. No Go cache
overrides are added to the Makefile or documentation commands.

### Commands recorded

From the repository root:

```sh
git status --short
git diff -- spec/
sha256sum spec/totipo-vault-format-v0.md vectors/v0/manifest.sha256
diff -u spec/totipo-vault-format-v0.md spec/totipo-vault-format-v0-r37.md
# Expected diff exit 1; reviewed header/Sections 74–75 changes only.

mkdir -m 700 -p .phase3-test-tmp
# Baseline (both rerun successfully with required Unix-socket access):
TMPDIR="$PWD/.phase3-test-tmp" GOMODCACHE=/tmp/totipo-gomod GOCACHE=/tmp/totipo-gocache go -C conformance test -count=1 ./...
TMPDIR="$PWD/.phase3-test-tmp" GOMODCACHE=/tmp/totipo-gomod GOCACHE=/tmp/totipo-gocache go -C conformance test -race -count=1 ./...
GOMODCACHE=/tmp/totipo-gomod GOCACHE=/tmp/totipo-gocache go run ./conformance/cmd/totipo-conformance ./vectors/v0

# After promotion and implementation:
gofmt -l conformance tools
go -C conformance vet ./...
TMPDIR="$PWD/.phase3-test-tmp" go -C conformance test -count=1 ./...
TMPDIR="$PWD/.phase3-test-tmp" go -C conformance test -race -count=1 ./...
go -C conformance test -count=1 ./internal/requirements ./cmd/...
/tmp/totipo-gomod/golang.org/toolchain@v0.0.1-go1.23.12.linux-amd64/bin/go -C conformance test -count=1 ./internal/requirements ./cmd/...
go run ./conformance/cmd/totipo-conformance ./vectors/v0
go run ./conformance/cmd/totipo-conformance --requirements requirements/v0-rc1.json
go run ./conformance/cmd/totipo-conformance --requirements requirements/v0-rc1.json --verify-only

# Run before and after the changes:
(cd vectors/v0 && sha256sum -c manifest.sha256)
sha256sum -c conformance/review-inventory.sha256
sha256sum -c review/phase2/inventory.sha256
sha256sum -c review/phase3/source.sha256

# Exact-tree and pinned-artifact checks:
find vectors/v0 -type f -print0 | sort -z | xargs -0 sha256sum > /tmp/totipo-rc1-vectors-before.sha256
# The preceding snapshot was taken before edits; the following after edits.
find vectors/v0 -type f -print0 | sort -z | xargs -0 sha256sum > /tmp/totipo-rc1-vectors-after.sha256
cmp /tmp/totipo-rc1-vectors-before.sha256 /tmp/totipo-rc1-vectors-after.sha256
sha256sum -c /tmp/totipo-rc1-vectors-before.sha256
cmp requirements/v0-rc1.manifest.sha256 vectors/v0/manifest.sha256
cmp /tmp/totipo-rc1-case-ids-before.txt /tmp/totipo-rc1-case-ids-after.txt
sha256sum spec/totipo-vault-format-v0.md vectors/v0/manifest.sha256 requirements/v0-rc1.json requirements/v0-rc1.manifest.sha256
git diff --check -- . ':!spec/totipo-vault-format-v0.md'
```

The temporary case-ID lists were extracted from baseline/final CLI output with
`awk '/^PASS v0\// {print $2}'`. A temporary Go comparison also asserted exact
Section 1–73 equality and that Section 74 differed only by the stated sentence.

## Files changed

Modified:

- `.gitattributes`
- `.github/workflows/conformance.yml`
- `Makefile`
- `README.md`
- `conformance/README.md`
- `conformance/cmd/totipo-conformance/main.go`
- `conformance/reference/README.md`
- `review/README.md`
- `spec/totipo-vault-format-v0.md`

Added:

- `requirements/README.md`
- `requirements/v0-rc1.json`
- `requirements/v0-rc1.manifest.sha256`
- `conformance/internal/requirements/profile.go`
- `conformance/internal/requirements/profile_test.go`
- `conformance/internal/requirements/artifacts.go`
- `conformance/internal/requirements/artifacts_test.go`
- `conformance/internal/requirements/generate.go`
- `conformance/cmd/totipo-conformance/main_test.go`
- `conformance/cmd/totipo-requirements/main.go`
- `conformance/cmd/totipo-requirements/main_test.go`
- `review/releases/v0-rc1/V0_RC1_REQUIREMENTS_REPORT.md`

Removed by promotion: the incoming untracked
`spec/totipo-vault-format-v0-r37.md` duplicate. Its exact contents now occupy the
canonical path. The incoming untracked instructions file is not an agent-created
artifact and was not modified. No review/provenance inventory or vector file
changed. No optional evidence JSON or redundant release-preparation note was
added.

At the maintainer's request, the release-preparation documents were relocated:

- `V0_RC1_AGENT_INSTRUCTIONS.md` → `review/releases/v0-rc1/V0_RC1_AGENT_INSTRUCTIONS.md`
- `V0_RC1_REQUIREMENTS_REPORT.md` → `review/releases/v0-rc1/V0_RC1_REQUIREMENTS_REPORT.md`

Neither document is listed in a checksum inventory, so no inventory update was
needed. Instruction contents remain byte-for-byte unchanged, including their
historical root-path examples. Report links were adjusted for its new location,
and the review index now links to both documents. Commands and file lists in
this report remain relative to the repository root.

## Remaining human release actions

1. Review the working tree, including the frozen-manifest arrangement and exact IDs.
2. Commit the accepted changes.
3. Obtain green remote CI on that release commit, including portable platforms,
   Go 1.23/stable and the separate Linux integration/race jobs.
4. Optionally review/update GitHub repository metadata.
5. Create/tag `v0-rc1` and publish its GitHub Release.

No tag/release is claimed to exist. Portable conformance and Linux reference
evidence do not establish production readiness, physical power-loss correctness,
cross-platform storage conformance or an external audit.
