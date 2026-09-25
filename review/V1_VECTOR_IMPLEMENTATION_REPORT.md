# V1 vector implementation report

Date: 2026-09-25. Protocol: Totipo v1, revision r9. Status: moving pre-RC evidence.

The Go reference consumer passes all **68 manifest cases**. `make check` passes.
All 62 IDs in the supplied `vectors/CASE_PLAN.md` are present, together with three
bootstrap cases and three additional graph cases. The normative specification is
byte-for-byte identical to the supplied seed. No release, tag, or frozen RC
profile was created.

## Historical preservation and commits

The pre-reset v0 HEAD was:

```text
5157f14e1af9d28b6f15db20a51cd384f97c8a37
```

The local `archive/v0` branch points to that exact commit. It preserves the old
protocol, corpus, Go implementation, reviews, and build files without copying
them into current main. No existing branch was moved, no history rewritten, and
no tag/release changed. There were no tags in this local checkout. The attempted
`git fetch --tags --prune` could not complete because `ssh` is unavailable, so
remote tags/releases were not independently verified. The archive was not pushed.

The user explicitly authorized replacing the dirty working tree, superseding the
initial clean-tree stop, and then manually committed the staged historical reset:

```text
0fe389345ecf6f4cf1a043b369bfd7ed30ee8aa0 archive v0 and reset main to v1/r9
```

This is the only new commit. The subsequent Go, vector, profile, documentation,
and CI implementation remains uncommitted for the user to review and commit.
No further staging or commits were performed after the user chose manual commits.
The reset boundary remains separate and was not squashed.

`AGENT_INSTRUCTIONS.md`, `PROPOSED_TREE.md`, `SHA256SUMS.txt`, and `copy-to-repo/`
are user-supplied local inputs, intentionally excluded from implementation changes
and commits. A temporary copy of those inputs also exists at
`/tmp/totipo-v1-input/`; it is not a repository artifact.

## Current tree and implementation

Current main contains the v1/r9 specification, seed design reviews, portable JSON
cases, a moving requirements profile, the Go reference consumer, and current
checks. No `vectors/v0`, v0 requirements profile, v0 conformance target, or v0
protocol subtree remains in the tracked current tree. README and contribution
instructions now describe the current implementation and development commands.

The implementation separates:

- `internal/tlv`: bounds-checked u16be framing.
- `internal/object`: frozen TOKEN/DEVICE routing, v1 full encoders/parsers,
  explicit DEVICE identity equality, unsigned u64 timestamps, dispatch classes,
  whole-state value bytes, and deterministic 72-byte signature capacity planning.
- `internal/cryptov1`: root hierarchy, private content IDs, deterministic 1024-byte
  AES-GCM envelopes, zero padding, filename validation, P-256 provenance, Argon2id
  bootstrap wrap/unwrap, and local vault binding.
- `internal/graph`: persisted routing/evidence semantics, availability kept
  separate from topology, current heads, conflict/completeness, candidate and
  authoritative safety gates, DEVICE presentation, cycles and global-ID conflicts.
- `internal/vectors`: strict typed case IO, manifest/file checksums, moving-profile
  verification, diagnostic intermediate comparisons, and semantic case execution.
- `cmd/totipo-conformance`: manifest-driven verification and execution.
- `cmd/totipo-vector-gen`: deliberate fixture maintenance using the same consumer.

The Python tools check specification structure and the checked-in JSON schema
contracts using the standard library. The schema checker intentionally supports
only the vocabulary used by these contracts and rejects unsupported keywords.
It is not another protocol implementation.

## Corpus coverage

| Category | Cases |
|---|---:|
| routing | 7 |
| encoding | 8 |
| device | 2 |
| size | 5 |
| crypto | 5 |
| provenance | 5 |
| bootstrap | 3 |
| graph | 11 |
| future | 9 |
| candidate | 7 |
| timestamp | 6 |
| **Total** | **68** |

Kinds: **25 exact-byte**, **6 negative/parser**, **37 semantic/state** cases.
Every manifest entry has a real file, stable ID, expected result, specification
sections, and SHA-256. Cases are written before manifest entries. The moving
`requirements/v1-pre-rc.json` pins all 68 IDs/file hashes plus the spec, manifest,
and both schema hashes. No `v1-rc1.json` exists.

Every routing/encoding/identity parser fixture is authenticated through the same
v1 encrypted envelope, including the intentionally invalid supported grammars.
The future fixtures retain frozen routing with `OBJECT_VERSION=2` and opaque
bytes that fail the v1 tail grammar. Malformed future prefixes and unknown types
are opaque-unscoped; invalid current storage is tested separately.

Graph expectations cover scoped future TOKEN degradation; unaffected unrelated
tokens; explicit candidate availability and warnings; supported descendants of
opaque ancestors; version/time non-ordering; DEVICE-only presentation degradation;
unscoped authoritative blocking without candidate blocking; disappearing supported
and opaque bytes; conflicts; late/wrong-identity parents; persistence and continuity
gates; cycles; global-ID conflicts; and matching/mismatching reappearance.
Queries assert nonmutation. Graph expectations are authored independently of the
evaluator's output, using symbolic authenticated object IDs and complete values.

## Crypto fixture method

All keys, salts, nonces, passwords, and secrets in this corpus are public test
material. Full supported crypto fixtures use P-256 private scalar 1 and fixed DER
signature bytes, initially generated by Go's CSPRNG-backed ECDSA signer. Signatures
are verified locally, committed as fixtures, and reused on regeneration. The
consumer verifies signatures rather than expecting independent signers to reproduce
them. Fixtures expose unsigned semantic bytes, signature input, DER signature,
signed bytes, all root-derived keys, object ID/key, nonce, AAD, semantic length,
padding, ciphertext, GCM tag, and the exact 1024-byte object.

The offline 70/72-byte DER fixture selection exists only to test verifier length
coverage. It never controls a writer's parent set. The capacity rejection fixture
uses a single signature and an equivalent low-S representation over a 15-parent,
254-byte-display DEVICE: its actual bytes fit but its 1007-byte reserved size does
not. The writer helper plans with 72 bytes before signing, signs once, verifies
locally, and never retries to obtain a fitting signature.

Future objects expose their frozen routing inputs and opaque tails; no invented
v1 signature semantics are imposed on them. Bootstrap vectors include ASCII,
empty, and non-normalized Unicode password bytes with the fixed r9 Argon2id
parameters. No custom cipher or elliptic-curve implementation was introduced.

The generator and consumer share implementation ancestry. These results are one
implementation's evidence, not independent interoperability evidence.

## Validation results

| Check | Result |
|---|---|
| `python3 tools/check_spec.py` | PASS: Totipo v1/r9 structural spec checks |
| `go test ./...` in `conformance/`, writable `GOCACHE` | PASS |
| `make check` | PASS: spec, Python tests, Go tests, execution, schema/hash/profile verification |
| `make conformance` | PASS: 68/68 |
| `make verify` | PASS: schema contracts, manifest/files, moving profile |
| `make race` | PASS |
| `make fuzz` | PASS: three 10-second targets, two workers each |
| `go -C conformance vet ./...` | PASS |
| `git diff --check` and Go formatting | PASS |
| Seed/spec byte comparison | PASS: identical |
| Deliberate regeneration stability | PASS: all 72 JSON artifact hashes unchanged |

Recorded bounded fuzz runs executed approximately 1,020,495 semantic-parser inputs,
1,081,840 envelope inputs, and 115,347 graph arrival/disappearance inputs. These
are bounded smoke checks, not exhaustive evidence. Parser tests also directly
exercise authenticated bad lengths, padding, keyed-ID mismatch, AEAD tampering,
UTF-8/field/range errors, malformed provenance, unsigned timestamp extremes,
high-S/low-S signatures, canonical DER rejection, and short-signature planning.

CI now runs only current v1 work on Linux, macOS, and Windows with Go 1.23 and
stable. Linux adds race/fuzz checks. Those remote jobs have been configured but
were not executed in this session; local results are Linux with Go 1.26.7.

## Dependencies and preserved Nix setup

The original flake inputs, pinned `llm-agents`, development-shell model, Go tools,
compiler packages, `.envrc`, and `flake.lock` were preserved. The user's flake
change adds `python3` to both the development shell and jailed-agent package list
and forwards `GOPATH`/`GOBIN`. No input or lockfile change was needed.

Make selects a writable `GOCACHE` under the existing ignored `.direnv/` directory.
This is cache configuration, not a replacement dependency environment. The resumed
Nix-supplied environment successfully ran the documented Make commands. A separate
fresh `nix develop` evaluation could not be run because the `nix` command is not
exposed inside this jail.

The new module uses only Go standard-library crypto plus:

- `golang.org/x/crypto v0.36.0`: Argon2id and the official HKDF implementation for
  the supported Go 1.23 toolchain.
- `golang.org/x/sys v0.31.0`: indirect CPU/platform support for x/crypto.

These are the existing project's previously pinned official module versions,
reused with a new minimal module and verified module checksums. The old Edwards25519
dependency was removed with the old module. There is no CBOR/TLV framework, general
JSON-schema dependency, or second synthetic protocol implementation.

## Ambiguities, limitations, and next live-implementation work

No concrete contradiction requiring a wire-format or semantic change was found.
No r9 normative wording was changed or proposed. Supported assertion validity,
provenance validity, and opaque routing were implemented as separate checks,
including the explicit acceptance of structurally sized but off-curve DEVICE
public-key bytes as rejected provenance rather than invalid TOKEN authority.

This pass is reference/vector evidence, not a full application conformance claim.
The remaining work includes:

- A real durable storage implementation with fault/crash tests for VAULT pending
  establishment, root binding, atomic no-replace object publication, graph flush
  ordering, continuity baseline recovery, and persistence retry handling.
- Operation freshness epochs, confirmation invalidation, resource-complete scans,
  staged bounded-fold publication/interruption, and concurrent-client workflows.
  Capacity/fan-in is implemented; the graph fold-timestamp case supplies a shared
  timestamp and does not implement or validate an application writer transaction.
- Broader adversarial/negative corpus coverage, TOTP RFC 6238 vectors, native
  platform password/UTF-8 behavior, and Java/Android/Apple P-256 cross-verification.
- Independent consumption by the first real Totipo implementation and external
  protocol/security review before a frozen RC profile or release.

The recommended next step is to implement the live client's parser/envelope layer
against these fixed fixtures, consume the symbolic graph cases using its own
state engine, then add platform persistence/freshness evidence. Compare semantic
and signature-input bytes exactly, verify fixed signatures, and do not demand
fresh-signature byte equality. Keep current profile changes explicitly reviewed.

## Exact file changes

The following inventory compares the intended resulting repository tree with
`archive/v0` at the recorded pre-reset commit. It includes the already committed
reset and the uncommitted implementation, including the user's flake adjustment.
`A` means added, `D` removed, and `M` modified. Untracked instruction/bundle inputs,
Go caches, temporary files, and `.git` metadata are excluded. It is also available
as [`V1_VECTOR_FILE_CHANGES.tsv`](V1_VECTOR_FILE_CHANGES.tsv).

```text
M .github/workflows/conformance.yml
M .gitignore
M CONTRIBUTING.md
M Makefile
M README.md
D REPO_CLEANUP_REPORT.md
D conformance/FUZZING.md
M conformance/README.md
M conformance/cmd/totipo-conformance/main.go
D conformance/cmd/totipo-conformance/main_test.go
D conformance/cmd/totipo-requirements/main.go
D conformance/cmd/totipo-requirements/main_test.go
A conformance/cmd/totipo-vector-gen/graph.go
A conformance/cmd/totipo-vector-gen/main.go
M conformance/go.mod
M conformance/go.sum
D conformance/internal/bootstrap/bootstrap.go
D conformance/internal/bootstrap/bootstrap_test.go
D conformance/internal/corpus/corpus.go
D conformance/internal/corpus/corpus_test.go
D conformance/internal/corpus/trace.go
A conformance/internal/cryptov1/crypto.go
A conformance/internal/cryptov1/crypto_test.go
D conformance/internal/dispatch/dispatch.go
D conformance/internal/ed25519profile/doc.go
D conformance/internal/ed25519profile/verify.go
D conformance/internal/envelope/envelope.go
A conformance/internal/graph/graph.go
A conformance/internal/graph/graph_test.go
D conformance/internal/model/model.go
D conformance/internal/model/model_test.go
D conformance/internal/model/safety.go
A conformance/internal/object/object.go
A conformance/internal/object/object_test.go
D conformance/internal/objectcrypto/objectcrypto.go
D conformance/internal/presentation/presentation.go
D conformance/internal/publication/publication.go
D conformance/internal/publication/publication_test.go
D conformance/internal/referenceoracle/memory.go
D conformance/internal/referenceoracle/oracle.go
D conformance/internal/requirements/artifacts.go
D conformance/internal/requirements/artifacts_test.go
D conformance/internal/requirements/generate.go
D conformance/internal/requirements/profile.go
D conformance/internal/requirements/profile_test.go
D conformance/internal/runner/fuzz_test.go
D conformance/internal/runner/independence_test.go
D conformance/internal/runner/phase2_test.go
D conformance/internal/runner/runner.go
D conformance/internal/runner/runner_test.go
D conformance/internal/securitymemory/memory.go
D conformance/internal/securitymemory/memory_test.go
M conformance/internal/tlv/tlv.go
A conformance/internal/tlv/tlv_test.go
D conformance/internal/totp/totp.go
D conformance/internal/totp/totp_test.go
A conformance/internal/vectors/runner.go
A conformance/internal/vectors/runner_test.go
A conformance/internal/vectors/types.go
D conformance/reference/README.md
D conformance/reference/client/client.go
D conformance/reference/fault/fault.go
D conformance/reference/integration/crash_test.go
D conformance/reference/integration/filesystem_test.go
D conformance/reference/integration/integration_test.go
D conformance/reference/localstate/linux.go
D conformance/reference/localstate/state.go
D conformance/reference/storage/linux.go
D conformance/reference/storage/storage.go
D conformance/review-inventory.sha256
M flake.nix
D go.work.sum
M requirements/README.md
D requirements/v0-rc1.json
D requirements/v0-rc1.manifest.sha256
A requirements/v1-pre-rc.json
D review/README.md
A review/V1_R8_FINAL_CONSISTENCY_ADVERSARIAL_REVIEW.md
A review/V1_R8_SIMPLIFICATION_REVIEW.md
A review/V1_R9_FORWARD_COMPAT_REVIEW.md
A review/V1_VECTOR_FILE_CHANGES.tsv
A review/V1_VECTOR_IMPLEMENTATION_REPORT.md
D review/ed25519/REVIEW.md
D review/ed25519/review.go
D review/phase1/IMPLEMENTATION_REPORT.md
D review/phase2/CASE_INDEX.md
D review/phase2/IMPLEMENTATION_REPORT.md
D review/phase2/REVIEW.md
D review/phase2/inventory.sha256
D review/phase3/IMPLEMENTATION_REPORT.md
D review/phase3/README.md
D review/phase3/local-results.json
D review/phase3/source.sha256
D review/process/AGENT_INSTRUCTIONS.md
D review/process/PHASE2_AGENT_INSTRUCTIONS.md
D review/process/PHASE3_AGENT_INSTRUCTIONS.md
D review/process/README.md
D review/process/SEED_REVIEW_README.md
D review/releases/v0-rc1/V0_RC1_AGENT_INSTRUCTIONS.md
D review/releases/v0-rc1/V0_RC1_REQUIREMENTS_REPORT.md
D review/reports/r31-executable-review.md
D review/reports/r31-java-python-crosscheck-report.md
D review/reports/r32-bootstrap-review.md
D review/reports/totipo-vault-v0-r35-review-disposition.md
D review/reports/totp-vault-v0-r31-tlv-corpus-report.md
D review/source-code/BootstrapV0.java
D review/source-code/LifecycleOracle.java
D review/source-code/R31Review.java
D review/source-code/r31-java-vector-crosscheck.py
D review/source-code/totp-vault-v0-r31-tlv-corpus.py
D review/source-code/totp-vault-v0-r32-bootstrap-vectors.py
D review/source-vectors/r31-crosscheck-SHA256SUMS.txt
D review/source-vectors/r31-object-vectors.txt
D review/source-vectors/totp-vault-v0-r31-tlv-corpus-SHA256SUMS.txt
D review/source-vectors/totp-vault-v0-r31-tlv-corpus.json
D review/source-vectors/totp-vault-v0-r32-SHA256SUMS.txt
D review/source-vectors/totp-vault-v0-r32-bootstrap-vectors.txt
D spec/totipo-vault-format-v0.md
A spec/totipo-vault-format-v1.md
A tools/check_spec.py
A tools/check_vectors.py
D tools/manifest/main.go
D tools/normalize.go
D tools/phase2/main.go
D tools/phase2/memory.go
D tools/phase2/presentation.go
D tools/phase2/safety.go
A tools/test_check_vectors.py
A vectors/CASE_PLAN.md
A vectors/FORMAT.md
A vectors/README.md
A vectors/case.schema.json
A vectors/cases/bootstrap/v1.bootstrap.ascii.001.json
A vectors/cases/bootstrap/v1.bootstrap.empty.001.json
A vectors/cases/bootstrap/v1.bootstrap.unicode.001.json
A vectors/cases/candidate/v1.candidate.conflict-a.001.json
A vectors/cases/candidate/v1.candidate.continuity-block.001.json
A vectors/cases/candidate/v1.candidate.current-peer-missing.001.json
A vectors/cases/candidate/v1.candidate.discovery-incomplete.001.json
A vectors/cases/candidate/v1.candidate.historical-current-missing.001.json
A vectors/cases/candidate/v1.candidate.opaque-current.001.json
A vectors/cases/candidate/v1.candidate.persistence-block.001.json
A vectors/cases/crypto/v1.crypto.device-root.001.json
A vectors/cases/crypto/v1.crypto.future-device-opaque.001.json
A vectors/cases/crypto/v1.crypto.future-token-opaque.001.json
A vectors/cases/crypto/v1.crypto.token-child.001.json
A vectors/cases/crypto/v1.crypto.token-root.001.json
A vectors/cases/device/v1.device.explicit-id.001.json
A vectors/cases/device/v1.device.id-mismatch.001.json
A vectors/cases/encoding/v1.encoding.author-time-u64max.001.json
A vectors/cases/encoding/v1.encoding.author-time-zero.001.json
A vectors/cases/encoding/v1.encoding.device-root.001.json
A vectors/cases/encoding/v1.encoding.duplicate-nonrepeatable.001.json
A vectors/cases/encoding/v1.encoding.parent-count-mismatch.001.json
A vectors/cases/encoding/v1.encoding.parent-order.001.json
A vectors/cases/encoding/v1.encoding.token-root.001.json
A vectors/cases/encoding/v1.encoding.utf8-boundary.001.json
A vectors/cases/future/v1.future.concurrent-supported-opaque.001.json
A vectors/cases/future/v1.future.device-presentation.001.json
A vectors/cases/future/v1.future.disappearance-retains-routing.001.json
A vectors/cases/future/v1.future.scoped-token.001.json
A vectors/cases/future/v1.future.supported-descendant.001.json
A vectors/cases/future/v1.future.unrelated-token.001.json
A vectors/cases/future/v1.future.unscoped-authoritative-block.001.json
A vectors/cases/future/v1.future.unscoped-candidate-use.001.json
A vectors/cases/future/v1.future.version-does-not-order.001.json
A vectors/cases/graph/v1.graph.conflicting-concurrent.001.json
A vectors/cases/graph/v1.graph.cycle-integrity-failure.001.json
A vectors/cases/graph/v1.graph.equal-concurrent.001.json
A vectors/cases/graph/v1.graph.global-object-id-conflict.001.json
A vectors/cases/graph/v1.graph.intermediate-disappears.001.json
A vectors/cases/graph/v1.graph.late-parent.001.json
A vectors/cases/graph/v1.graph.missing-current-value.001.json
A vectors/cases/graph/v1.graph.reappearance-mismatch.001.json
A vectors/cases/graph/v1.graph.reappearance.001.json
A vectors/cases/graph/v1.graph.sequential.001.json
A vectors/cases/graph/v1.graph.wrong-identity-parent.001.json
A vectors/cases/provenance/v1.provenance.der-max-valid.001.json
A vectors/cases/provenance/v1.provenance.der-short-valid.001.json
A vectors/cases/provenance/v1.provenance.token-rejected.001.json
A vectors/cases/provenance/v1.provenance.token-unresolved.001.json
A vectors/cases/provenance/v1.provenance.token-verified.001.json
A vectors/cases/routing/v1.routing.device-future-opaque.001.json
A vectors/cases/routing/v1.routing.device-v1.001.json
A vectors/cases/routing/v1.routing.future-device-malformed-prefix.001.json
A vectors/cases/routing/v1.routing.future-token-malformed-prefix.001.json
A vectors/cases/routing/v1.routing.token-future-opaque.001.json
A vectors/cases/routing/v1.routing.token-v1.001.json
A vectors/cases/routing/v1.routing.unknown-type-unscoped.001.json
A vectors/cases/size/v1.size.device-max-14.001.json
A vectors/cases/size/v1.size.device-max-15-fold.001.json
A vectors/cases/size/v1.size.short-der-no-extra-parent.001.json
A vectors/cases/size/v1.size.token-max-4.001.json
A vectors/cases/size/v1.size.token-max-5-fold.001.json
A vectors/cases/timestamp/v1.timestamp.equal-value-different-times.001.json
A vectors/cases/timestamp/v1.timestamp.fold-common-time.001.json
A vectors/cases/timestamp/v1.timestamp.i64max.001.json
A vectors/cases/timestamp/v1.timestamp.normal.001.json
A vectors/cases/timestamp/v1.timestamp.u64max.001.json
A vectors/cases/timestamp/v1.timestamp.zero.001.json
A vectors/manifest.json
A vectors/manifest.schema.json
D vectors/v0/README.md
D vectors/v0/bootstrap/cases.json
D vectors/v0/bootstrap/supplemental.json
D vectors/v0/dispatch/supplemental.json
D vectors/v0/ed25519/README.md
D vectors/v0/ed25519/phase2.json
D vectors/v0/envelope/cases.json
D vectors/v0/envelope/supplemental.json
D vectors/v0/lifecycle/README.md
D vectors/v0/lifecycle/cases.json
D vectors/v0/manifest.sha256
D vectors/v0/object-crypto/cases.json
D vectors/v0/presentation/README.md
D vectors/v0/presentation/phase2.json
D vectors/v0/recovery/README.md
D vectors/v0/recovery/phase2.json
D vectors/v0/tlv/cases.json
D vectors/v0/tlv/supplemental.json
D vectors/v0/totp/cases.json
D vectors/v0/totp/supplemental.json
D vectors/v0/transitions/README.md
D vectors/v0/transitions/phase2.json
D vectors/v0/transitions/use-profile.json
```
