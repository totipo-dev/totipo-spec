# Repository cleanup review

This change is repository cleanup only. No protocol text, frozen literal, vector byte, Go implementation/test code, dependency version, or Go minimum version changed. No commit, push, tag, release, visibility change, or repository-setting change was made.

## Files added

- `README.md`: project overview, authoritative paths, evidence boundaries, minimum Go version, and current test commands.
- `CONTRIBUTING.md`: protocol/vector change discipline and verification commands.
- `SECURITY.md`: private-reporting guidance with an explicit unresolved contact TODO.
- `LICENSE-TODO.md`: records that no project license exists and requires maintainer selection; Apache-2.0 is only a candidate, not adopted.
- `.github/dependabot.yml`: weekly Go module and GitHub Actions update proposals; no automatic merging.
- `review/process/README.md`: index and historical-context note for archived instructions and the seed index.
- `REPO_CLEANUP_REPORT.md`: this review summary.
- `review/source-code/BootstrapV0.java`: added by the maintainer during cleanup, retained without edits and linked from the current review index. This is not agent-authored code or newly executed conformance evidence.

A new current index was also written at `review/README.md`, replacing the original index after its byte-preserving archive move below. Other updated documentation is `conformance/README.md` and `review/phase3/README.md`.

## Files moved

| Old path | New path | Content/inventory treatment |
|---|---|---|
| `AGENT_INSTRUCTIONS.md` | `review/process/AGENT_INSTRUCTIONS.md` | Byte-identical; not inventoried |
| `PHASE2_AGENT_INSTRUCTIONS.md` | `review/process/PHASE2_AGENT_INSTRUCTIONS.md` | Byte-identical; not inventoried |
| `PHASE3_AGENT_INSTRUCTIONS.md` | `review/process/PHASE3_AGENT_INSTRUCTIONS.md` | Byte-identical; not inventoried |
| `conformance/IMPLEMENTATION_REPORT.md` | `review/phase1/IMPLEMENTATION_REPORT.md` | Byte-identical historical Phase 1 checkpoint; not inventoried |
| `PHASE2_IMPLEMENTATION_REPORT.md` | `review/phase2/IMPLEMENTATION_REPORT.md` | Only Markdown link destinations adjusted for the new directory; not inventoried |
| `PHASE3_IMPLEMENTATION_REPORT.md` | `review/phase3/IMPLEMENTATION_REPORT.md` | Only Markdown link destinations adjusted for the new directory; not inventoried |
| `review/README.md` | `review/process/SEED_REVIEW_README.md` | Byte-identical; inventory path updated, hash unchanged |
| `fuzz/README.md` | `conformance/FUZZING.md` | Byte-identical; not inventoried; no tools/tests depend on the old directory |

References and inventory membership were searched before the moves. Current documentation links point to the new locations. Archived process instructions retain their original requested output names and layout examples; `review/process/README.md` explicitly identifies those as historical context. Phase reports retain their original execution and CI observations rather than being rewritten as current results.

**The only inventory edit is one path in `conformance/review-inventory.sha256`:**

```text
review/README.md -> review/process/SEED_REVIEW_README.md
```

Its SHA-256 remains `e6b14493eec2f101c34e4b82bdb8fe56db1ae42a1e3ca4c1b993bab9b91bc261`. All 17 original seed files remain verifiable with their original hashes. No inventory hash was recomputed to accommodate edited evidence. The Phase 2 inventory and Phase 3 source inventory are byte-identical to their pre-cleanup versions.

## Files deleted and material retained

No file content was discarded. The empty top-level `fuzz/` directory was removed after its sole README moved. Paths shown as deleted by unstaged Git status correspond to the moves listed above, not discarded reports.

The inspection found no duplicate report contents or committed scratch, coverage, profile, fuzz-cache, test-binary, or OS-junk files requiring deletion. The actively used, ignored `.direnv/` and `.phase3-test-tmp/` directories were retained. Historical vector files, old implementations, checksum lists, and `review/phase3/local-results.json` remain because they are provenance or referenced execution evidence. Necessary Go workspace and Nix configuration remain at the root.

`.gitignore` retains the local integration scratch/log and direnv ignores and adds `coverage.out`, `*.prof`, `*.test`, `*.test.exe`, `.DS_Store`, `/result`, and `/result-*`. No normative or review evidence is ignored.

## CI and dependency updates

The existing action majors were resolved from the official upstream tag APIs, without guessing hashes or changing major versions:

| Action | Before | Immutable pin | Release |
|---|---|---|---|
| `actions/checkout` | `@v4` | `11d5960a326750d5838078e36cf38b85af677262` | [v4.4.0](https://github.com/actions/checkout/releases/tag/v4.4.0) |
| `actions/setup-go` | `@v5` | `40f1582b2485089dde7abd97c1529aa768e1baff` | [v5.6.0](https://github.com/actions/setup-go/releases/tag/v5.6.0) |
| `actions/upload-artifact` | `@v4` | `ea165f8d65b6e75b540449e92b4886f43607fa02` | [v4.6.2](https://github.com/actions/upload-artifact/releases/tag/v4.6.2) |

Sources: official [checkout tags](https://api.github.com/repos/actions/checkout/tags?per_page=100), [setup-go tags](https://api.github.com/repos/actions/setup-go/tags?per_page=100), and [upload-artifact tags](https://api.github.com/repos/actions/upload-artifact/tags?per_page=100). Each selected release and its major alias resolved to the same commit. Human-readable release comments accompany every workflow pin.

- Portable Linux/macOS/Windows matrix: `1.26.x, stable` becomes **`1.23.x, stable`**.
- Linux integration matrix: **`1.23.x, stable`**.
- Linux race detector: **`stable`**.
- Linux inventory verification additionally checks `review/phase3/source.sha256`.
- Portable tests remain separate from Linux-only storage/process tests. No non-Linux storage claim was added, and CI never regenerates vectors.
- Weekly Dependabot updates cover `/conformance` Go modules and repository GitHub Actions. No extra action, dependency, or automatic merge policy was added.

`conformance/go.mod` and `go.work` still declare `go 1.23.0`. The root and conformance READMEs agree. The latest Go 1.23 patch returned by the official Go release metadata was **1.23.12**, and the full suite was tested with it. The metadata's current stable release was **1.27.1**, so the two CI entries are distinct. The installed local compiler remains 1.26.7; local results below do not claim execution of stable 1.27.1. No minimum-version contradiction was found. Historical reports' 1.26.x CI descriptions are preserved as checkpoint history, not current support policy.

## Integrity and test results

| Check | Before | After |
|---|---|---|
| Spec SHA-256 | `90bb937fc376517b16cee48b07c64c32df1e687fa32354a19ea1b851bab3eff7` | Identical |
| Vector manifest SHA-256 | `bf6a3a670ddf1acab773e44fe563f2e3d9392299fb9229bddb64042194e474a0` | Identical |
| Vector manifest verification | PASS, all 16 JSON files | PASS, all 16 JSON files |
| Entire `vectors/v0/` tree | Hashes saved for all 23 files, including documentation and manifest | All 23 byte-identical |
| Historical inventory | PASS, 17 files | PASS, same hashes; seed-index path move only |
| Phase 2 inventory | PASS, 4 files | PASS, unchanged |
| Phase 3 source inventory | PASS, 9 files | PASS, unchanged |
| Full tests, Go 1.23.12 | Additional minimum-version check | PASS |
| Normative CLI, Go 1.23.12 | Additional minimum-version check | **388 PASS / 0 FAIL / 0 BLOCKED** |
| Full tests, local Go 1.26.7 | Required cleanup check | PASS |
| Race tests, local Go 1.26.7 | Required cleanup check | PASS |
| Normative CLI, local Go 1.26.7 | Required cleanup check | **388 PASS / 0 FAIL / 0 BLOCKED** |
| Formatting and vet | Go 1.23.12 and Go 1.26.7 | PASS |

Linux integration runs use the documented repository-local `TMPDIR` on ext4. The tests require Unix sockets; the final test commands ran with permission outside the restrictive sandbox rather than skipping the hostile-socket test. No expected value was changed to recover a test.

The link audit found no newly broken relative Markdown links. It preserves three pre-existing historical links in hashed reports: `R31Review.java`, `r31-object-vectors.txt`, and `BootstrapV0.java`. The current review index documents them and links to all three current counterparts. Historical authoring-environment `*SHA256SUMS.txt` paths remain preserved evidence, not rewritten current-layout inventories. Mathematical `[L](B+T)` notation was excluded from pathname checking. No protocol-text link change was needed.

The maintainer supplied `review/source-code/BootstrapV0.java` during cleanup. Its received SHA-256 is `42ec7605b2ef4f61b1f536b14391147a502499e94fcf967853fd2c39a3cabee4`. No existing inventory records a hash for that Java source; the r32 report's displayed hash belongs to the bootstrap vector text, not the Java file. The source was not edited, compiled, executed, or silently added to an older review inventory. Its Bouncy Castle imports did not introduce a new Go dependency or build step. The Phase 1 report's historical absence statement is retained as a statement about that checkpoint.

## Commands run

Commands were run from the repository root unless a directory is stated. Verification never ran a vector-writing maintenance command.

Integrity before and after:

```sh
sha256sum spec/totipo-vault-format-v0.md vectors/v0/manifest.sha256
(cd vectors/v0 && sha256sum -c manifest.sha256)
sha256sum -c conformance/review-inventory.sha256
sha256sum -c review/phase2/inventory.sha256
sha256sum -c review/phase3/source.sha256
```

Additional whole-tree preservation checks:

```sh
sha256sum spec/totipo-vault-format-v0.md vectors/v0/manifest.sha256 > /tmp/totipo-cleanup-before.sha256
rg --files vectors/v0 | sort | xargs sha256sum > /tmp/totipo-cleanup-all-vectors.sha256
sha256sum -c /tmp/totipo-cleanup-before.sha256
sha256sum -c /tmp/totipo-cleanup-all-vectors.sha256
```

Full tests and CLI, with these environment settings:

```sh
mkdir -m 700 -p .phase3-test-tmp
export TMPDIR=/home/niki/Sources/totipo-spec/.phase3-test-tmp
export GOMODCACHE=/tmp/totipo-gomod
export GOCACHE=/tmp/totipo-gocache
GOTOOLCHAIN=go1.23.12 go -C conformance test -count=1 -timeout=180s ./... > /tmp/totipo-cleanup-go123-tests.txt
GOTOOLCHAIN=go1.23.12 go run ./conformance/cmd/totipo-conformance ./vectors/v0 > /tmp/totipo-cleanup-go123-cli.txt
go -C conformance test -count=1 -timeout=180s ./... > /tmp/totipo-cleanup-tests.txt
go -C conformance test -race -count=1 -timeout=300s ./... > /tmp/totipo-cleanup-race.txt
go run ./conformance/cmd/totipo-conformance ./vectors/v0 > /tmp/totipo-cleanup-cli.txt
```

Formatting, static checks, and review:

```sh
gofmt -l conformance tools review/ed25519/review.go
/tmp/totipo-gomod/golang.org/toolchain@v0.0.1-go1.23.12.linux-amd64/bin/gofmt -l conformance tools review/ed25519/review.go
GOTOOLCHAIN=go1.23.12 GOMODCACHE=/tmp/totipo-gomod GOCACHE=/tmp/totipo-gocache go -C conformance vet ./...
GOMODCACHE=/tmp/totipo-gomod GOCACHE=/tmp/totipo-gocache go -C conformance vet ./...
GOCACHE=/tmp/totipo-gocache go run /tmp/totipo-cleanup-links.go
git diff --check
```

The temporary link-audit helper scans relative Markdown targets, excludes inline/code-block notation, and explicitly reports the three preserved historical exceptions. Searches of Markdown/YAML/Go and inventory files checked old report/instruction/fuzz paths before and after moves. Duplicate-report and tracked-debris searches found nothing eligible for deletion. The direct upstream `curl` tag/release requests and Go toolchain download needed network permission; the Go download completed with `GOTOOLCHAIN=go1.23.12 ... go version`.

## Remaining TODOs

- Maintainer license selection and review of retained third-party rights; no license installed.
- Private security-reporting address or advisory instructions; no contact invented.
- Human review/commit of these cleanup changes, then a fresh remote CI run for the revised minimum/stable matrix and pinned actions. No remote workflow was triggered during this task.
- Release/status work, any r37 bookkeeping, requirement-set versioning, `v0-rc1` tags, and GitHub Releases remain a separate follow-up. None was performed here.
- Existing evidence limits remain: no external audit, actual power-loss proof, production-readiness claim, or macOS/Windows storage-adapter conformance.
