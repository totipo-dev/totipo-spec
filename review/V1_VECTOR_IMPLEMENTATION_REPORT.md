# V1 vector implementation report

Date: 2026-09-25. Protocol: Totipo v1, revision r9. Status: moving pre-RC evidence.

At the original implementation handoff, the Go reference consumer passed all
**68 manifest cases**, and `make check` passed.
All 62 originally planned IDs were present, together with three bootstrap cases
and three additional graph cases. At that handoff, the specification was
byte-for-byte identical to the supplied seed. No release, tag, or frozen RC
profile was created.

## Post-implementation acceptance update

The implementation was subsequently committed as
`8042729fe6b1c00fc57e284000f39c060d76d773` (`v1 work`), which is the cleanup
baseline for both local `main` and `origin/main`. The pushed
[Totipo v1 conformance run](https://github.com/totipo-dev/totipo-spec/actions/runs/36197834195)
completed successfully for that exact commit; this was verified through the
GitHub Actions API during cleanup. Current main contains the accepted Go/vector
implementation.

The GitHub refs API also confirms remote `archive/v0` at
`5157f14e1af9d28b6f15db20a51cd384f97c8a37`, matching the local and remote-tracking
archive refs. Remote tags/releases were not enumerated during cleanup; none were
modified. The earlier lack of SSH did not prevent these read-only HTTPS checks.

Final cleanup retires the completed plan and temporary bundle, removes the
transition inventory, makes three non-semantic rationale edits, and adds three
RFC 6238 cases (18 known-answer rows). The current corpus contains **71 cases**;
all 68 original case files remain byte-identical. See the
[final cleanup report](V1_FINAL_CLEANUP_REPORT.md) for current validation and scope.
The coverage and validation tables below describe the original 68-case handoff.

## Historical preservation and commits

The pre-reset v0 HEAD was:

```text
5157f14e1af9d28b6f15db20a51cd384f97c8a37
```

The local `archive/v0` branch points to that exact commit. It preserves the old
protocol, corpus, Go implementation, reviews, and build files without copying
them into current main. No existing branch was moved, no history rewritten, and
no tag/release changed. During original execution there were no tags in the local checkout. The attempted
`git fetch --tags --prune` could not complete because `ssh` is unavailable, so
remote tags/releases were not independently verified then. The archive was not
pushed by the agent during that session; its later remote presence is verified above.

The user explicitly authorized replacing the dirty working tree, superseding the
initial clean-tree stop, and then manually committed the staged historical reset:

```text
0fe389345ecf6f4cf1a043b369bfd7ed30ee8aa0 archive v0 and reset main to v1/r9
```

At the original agent handoff, only the reset was committed and the implementation
was left for the user's manual commit. The user subsequently committed it as
`8042729`; the reset boundary remains separate and was not squashed. The temporary
instruction/bundle inputs were included in that human commit and have now been
removed in cleanup after verifying their reset-only purpose.

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
stable. Linux adds race/fuzz checks. Those remote jobs had not yet executed at the original handoff; the later
successful run is recorded above. Original local results were Linux with Go 1.26.7.

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
No r9 normative wording was changed or proposed during implementation; the later
cleanup changes rationale wording only. Supported assertion validity,
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
- Broader adversarial/negative corpus coverage, native
  platform password/UTF-8 behavior, and Java/Android/Apple P-256 cross-verification.
- Independent consumption by the first real Totipo implementation and external
  protocol/security review before a frozen RC profile or release.

The recommended next step is to implement the live client's parser/envelope layer
against these fixed fixtures, consume the symbolic graph cases using its own
state engine, then add platform persistence/freshness evidence. Compare semantic
and signature-input bytes exactly, verify fixed signatures, and do not demand
fresh-signature byte equality. Keep current profile changes explicitly reviewed.

## Change history

Exact file history is maintained by Git: reset commit
`0fe389345ecf6f4cf1a043b369bfd7ed30ee8aa0` and accepted implementation commit
`8042729fe6b1c00fc57e284000f39c060d76d773`. The transition-only file inventory and
its large appendix were retired during final cleanup.
