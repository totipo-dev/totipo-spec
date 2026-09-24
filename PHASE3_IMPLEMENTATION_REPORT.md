# Phase 3 implementation report

The Linux reference implementation now joins frozen protocol bytes, strict crypto validation, the existing semantic models, durable local security memory, filesystem publication, restart reconstruction, and cross-process locking. Local tests pass on Linux/ext4. This is reference/conformance evidence, not production readiness, power-loss proof, cross-platform durability, or full Totipo client conformance.

The implementation is under [`conformance/reference/`](conformance/reference/README.md). Machine-readable local results and source hashes are under [`review/phase3/`](review/phase3/README.md).

## Baseline and frozen artifacts

Before implementation, the normal suite, race suite, CLI, vector manifest, historical inventory, and Phase 2 inventory passed. Baseline: **388 PASS / 0 FAIL / 0 BLOCKED**. A temporary copy of the baseline manifest was used again after implementation to verify all 16 JSON files remained byte-identical. No normative vectors or expected results were regenerated.

The final CLI remains **388 PASS / 0 FAIL / 0 BLOCKED**. Platform tests are counted separately. Both existing review inventories are unchanged and pass. The r36 specification is unchanged, and no normative contradiction was found.

| Frozen artifact | SHA-256 |
|---|---|
| `vectors/v0/manifest.sha256` | `bf6a3a670ddf1acab773e44fe563f2e3d9392299fb9229bddb64042194e474a0` |
| `spec/totipo-vault-format-v0.md` | `90bb937fc376517b16cee48b07c64c32df1e687fa32354a19ea1b851bab3eff7` |

## Executed results

Environment: **Go 1.26.7, Linux 6.18.44, amd64, GCC 15.3.0**. The final tests used repository-local temporary directories on **ext4**, not `/tmp`'s tmpfs. The Unix-socket test required an unsandboxed execution because the sandbox denied socket setup; it was executed successfully rather than skipped.

- **28 substantive integration test groups, 98 leaf test cases, zero failures.** The helper-only entry and parent groups are excluded from leaf counts.
- **64 subprocess crash/restart cases across 52 distinct named crash points.** Each calls `os.Exit(86)` without deferred cleanup, then opens fresh clients/adapters from persisted files.
- Full suite: PASS. Integration portion: **40.523 seconds**.
- Full race-detector suite: PASS. Integration portion: **82.964 seconds**.
- Formatting, vet, normative CLI, both original review inventories, and original vector hashes: PASS.
- Portable packages cross-compiled for `darwin/amd64` and `windows/amd64`. They were **not executed** on those operating systems; Linux-only adapters/tests are intentionally absent there.

The complete passed-case list, exact commands, crash-point names, and environment are in [`review/phase3/local-results.json`](review/phase3/local-results.json). [`review/phase3/source.sha256`](review/phase3/source.sha256) identifies the tested integration Go source.

## Reference-client coverage

The ingest pipeline uses actual 2048-byte encrypted files: safe bounded read; length gate; AEAD; deterministic envelope; OBJECT_ID recomputation; common-prefix/version dispatch; canonical v0 parsing; strict signature verification; dependency/semantic validation; and durable observation/frontier/handoff persistence. Token results must match both the Phase 2 primary model and independent reference oracle. DEVICE_UPDATE uses the existing presentation model. Future tails never enter the v0 grammar.

Authorship prepares from the persisted accepted frontier, checks writer eligibility and confirmation freshness, constructs canonical unsigned bytes, signs, identifies, envelopes, encrypts, publishes without replacement, reopens through the ingest pipeline, and persists acceptance before returning success. The 32-parent limit is checked without truncating heads. `AssessUse` reuses the Phase 2 ordinary OTP/export policy separately from consensus validity.

Restart discards the whole client and derived state. The implementation re-reads synchronized files and durable local evidence. Optional retained encrypted immutable copies are independently revalidated and can reconstruct history missing from synchronization. Every scan rereads current bytes, including same-size/same-mtime replacements.

| Integration category | Demonstrated behavior |
|---|---|
| `integration/object-ingest` | Valid encrypted objects accepted; invalid strict signatures gain no authority; dependency/type boundaries use existing models |
| `integration/pending` | Child-first arrival, durable token-local block, disappearance/restart, wrong-length/AEAD-invalid replacement, reappearance, abandonment, acknowledgement, accepted/terminal handoff |
| `integration/future` | Authenticated nonzero-version opaque tail, malformed-prefix rejection, durable observation, disappearance/restart, exact bypass, continued visible incompleteness |
| `integration/restart` | Same token state after reconstruction, missing-history degradation, retained validated copies, remembered-head preservation, database rollback limitation |
| `integration/presentation` | Same-key pending recovery, fork reconstruction, later-valid re-entry, durable presentation handoff, rollback warning, token independence |
| `integration/object-publish` | Complete durable immutable files, identical existing bytes, conflicting bytes, no overwrite, success after accepted-frontier persistence |
| `integration/vault-create` | Candidate validation, PENDING-before-install, no-replace, matching canonical recovery, differing-root refusal, refusal over existing object candidates |
| `integration/vault-rewrap` | Complete temporary bootstrap, unchanged root/binding, atomic canonical replacement, old/new-password behavior across interruption |
| `integration/filesystem` | Actual symlinks, FIFO, directory, Unix socket, hostile temp entries, parent replacement, name/bound checks, preserved outside targets |
| `integration/multiprocess` | Establishment CAS, immutable publication races, same-token writer serialization, publication/observation lock ordering |
| `integration/migration` | 33 real heads retained and capacity-blocked; new root/device/token and complete parentless creation; unchanged source |
| `integration/application` | Lifecycle confirmation required and freshness checked; ordinary use blocked by conflict/incompleteness; resource budget fails closed |

## Persistence and crash matrix

The local adapter is a small **fsync-backed JSON snapshot**, not a database dependency. It is outside the hostile synchronized namespace in a caller-provided private 0700 directory. One stable `flock` inode serializes all cooperating processes for the configured vault. Each transaction reloads state while locked.

A commit writes an exclusive 0600 temporary file, fsyncs its complete bytes, renames the snapshot, and fsyncs the local directory. This is one atomic snapshot replacement. Reopening a visible snapshot finishes the directory durability barrier before using it. The 16 MiB snapshot bound rejects capacity exhaustion before modifying canonical state.

Persisted records cover location, exact r36 binding, establishment, generation, token/presentation heads, pending facts, abandonment, warning acknowledgement, future versions, bypass, and terminal-invalidity handoffs. No root, password, plaintext credential, or private signing key is placed in this database.

| Crash family | Cases | Durable invariant and recovery expectation |
|---|---:|---|
| Initial establishment | 19 | No ordinary open without a matching canonical bootstrap and durable binding; pre-install PENDING is retained; a different root cannot finish it |
| Object publication/acceptance | 14 | Pre-install crash leaves old history; installed object may precede acknowledgement; old remembered evidence persists until replacement transaction; restart revalidates and accepts |
| Same-root rewrap | 7 | Canonical bootstrap is complete old or complete new bytes; binding is unchanged |
| Pending observation/accepted/terminal handoff | 11 | Durable pending fact survives disappearance; accepted successor is durable before pending removal; terminal fact and removal are one transaction |
| Future observation/recovery decisions | 3 | Observation persists independently of bypass; durable abandonment/bypass survives restart and does not erase evidence |
| Atomic local snapshot replacement | 6 | Complete old or new pending+future batch; unrelated remembered history remains intact |
| Presentation frontier handoff | 4 | Same ordering as token pending handoff, with separate presentation scope |
| **Total** | **64** | **52 distinct operation/stage hook names** |

Tests assert the resulting durable records, reconstructed state, and applicable writer block/recovery outcome. The JSON snapshot intentionally offers transaction semantics; the Phase 2 non-atomic model remains separate evidence. Crash leftovers are ignored noncanonical files/directories, never authoritative bootstraps.

These tests kill application processes while the kernel remains alive. A rename visible after death but before directory fsync is not presented as proof of power-loss persistence. Physical disk failure, lost write caches, filesystem corruption, and power-cut experiments were not performed.

## Filesystem evidence

The Linux adapter uses component-by-component `openat(O_DIRECTORY|O_NOFOLLOW)`, pinned directory descriptors, `O_PATH|O_NOFOLLOW` plus regular-file `fstat`, and bounded reads by reopening the pinned inode through `/proc/self/fd`. Special files are rejected before an ordinary file open, avoiding FIFO waits/device-open side effects. Actual device nodes were not created in this test environment.

Temporary candidates use random exclusive 0700 staging directories and `O_CREAT|O_EXCL|O_NOFOLLOW` 0600 files. Candidate bytes are read back against the validated input. Immutable objects and initial VAULT use **`renameat2(RENAME_NOREPLACE)`**; same-root rewrap uses **`renameat2(..., 0)`**. File and final-directory fsyncs precede success. Existing identical immutable files are read, compared, and fsynced; different or unsafe entries are not overwritten. Final bytes are checked again.

Filesystem assumptions, primitive references, API preconditions, and trust boundaries are documented in [`conformance/reference/README.md`](conformance/reference/README.md). No unsafe fallback is offered when Linux/procfs/filesystem facilities are unavailable. Network/sync-provider APIs and filesystem behavior beyond the tested local ext4 environment remain unsupported.

The existing pinned **`golang.org/x/sys v0.31.0`** is now a direct dependency for these syscalls. Its version did not change. No new crypto or database dependency was introduced.

## Concurrency and freshness

- **3 two-process establishment races:** both helpers inspect absent VAULT before a release barrier. Exactly one configured binding wins; the other cannot overwrite it. Reopening verifies the winner's canonical root against durable state.
- **3 two-process raw publication races:** identical object bytes, differing bytes at one final name, and independent names. Identical/independent publication succeeds; conflicting bytes produce one loser without corruption or replacement.
- **1 two-process same-token writer race:** operations serialize or stale-retry; the resulting token history has one head rather than an avoidable local fork.
- **1 two-process observation/publication race:** observation acceptance cannot cross the writer's held lock; bytes arriving afterward can become durable future evidence and block subsequent writes without invalidating the published update.
- **4 pre-publication freshness changes:** accepted head, pending evidence, future observation/bypass change, and restarted session all reject the prepared lifecycle operation without publishing it. A further after-publication case preserves validity while surfacing a later field conflict. A successful explicitly confirmed lifecycle resolution is also tested.

The exact linearization point is the successful final gate check entering immutable installation under the same cross-process lock used by observation acceptance and recovery. The confirmation binds desired fields/status, accepted token/field heads, conflicts, pending policy, and future context; durable generation changes and session identity prevent stale reuse. Global serialization is conservative and may cause retries for unrelated changes.

## End-to-end recovery, migration, and randomness

Pending observations survive lost files and process restart. Invalid-looking current representations never erase authenticated evidence. Reappearing valid bytes are reconsidered; accepted-frontier handoff becomes durable before evidence removal. Known-history regression retains missing heads, blocks the affected scope, and can recover from restored bytes or retained validated copies. Presentation history has separate heads/recovery and cannot change token consensus validity.

The database rollback test explicitly restores an older snapshot after persisting a future observation. The later observation is lost. There is **no claim of protection against local database backup rollback** and no automatic detection of an externally restored old snapshot. The trust assumption is intact local memory relative to the synchronization attacker.

Migration uses a source with 33 real concurrent token heads. No source resolution is published and no head is omitted. Locally selected complete creation values are written under a fresh destination root, fresh signing key, and fresh TOKEN_ID; the source object list stays unchanged. Production/reference creation paths directly use `crypto/rand` for root, token ID, Ed25519 key generation, salt, wrap nonce, and generated 20-byte credential secrets. Tests check identities/structure and the API paths; they do not claim a statistical randomness test. No deterministic vector generator is used by these creation paths.

## CI and remaining limits

Per the user's clarification, the repository is private and remote CI is not configured yet. **No remote workflow was triggered, repository published, or CI result claimed.** The local workflow file now separates:

1. Linux/macOS/Windows portable semantic and crypto tests for Go 1.26.x and stable;
2. Linux filesystem/process integration jobs with JSON artifacts for both versions;
3. Linux race-detector coverage.

Remote execution is pending. macOS/Windows filesystem adapters are unsupported, not passing or silently substituted with the abstract simulator. Cross-compilation is reported only as compilation evidence.

Other limits: no actual power-loss proof, network-filesystem/provider guarantees, external audit, protected persistent device-key implementation, credential UI, or production resource/performance hardening. Pending setup without a canonical bootstrap retains only its binding; it requires explicit recovery/reconfiguration if the original candidate is lost. An optional retained object cache must itself be outside the hostile namespace. Private staging/local state assume no malicious same-UID/root process. Process-crash coverage is finite, not every possible instruction-level interleaving.

These limits do not change the 388 normative-vector result. Linux integration is evidenced separately; full production/client/platform conformance is not claimed.

## Reproduction

From the repository root:

```sh
gofmt -l conformance tools
go -C conformance vet ./...
mkdir -m 700 -p .phase3-test-tmp
TMPDIR="$PWD/.phase3-test-tmp" go -C conformance test -count=1 -timeout=180s ./...
TMPDIR="$PWD/.phase3-test-tmp" go -C conformance test -race -count=1 -timeout=300s ./...
go run ./conformance/cmd/totipo-conformance ./vectors/v0
(cd vectors/v0 && sha256sum -c manifest.sha256)
sha256sum -c conformance/review-inventory.sha256
sha256sum -c review/phase2/inventory.sha256
sha256sum -c review/phase3/source.sha256
GOOS=darwin GOARCH=amd64 go -C conformance build ./...
GOOS=windows GOARCH=amd64 go -C conformance build ./...
```

This environment additionally used `GOMODCACHE=/tmp/totipo-gomod` and `GOCACHE=/tmp/totipo-gocache`. Full local JSON output was collected with `go test -json`; the extracted non-normative results are committed-ready files under `review/phase3/`. No commits or pushes were made by the agent.
