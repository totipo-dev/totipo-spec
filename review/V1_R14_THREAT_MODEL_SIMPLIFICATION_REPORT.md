# V1 r14 threat-model simplification report

## 1. Decision

Totipo treats synchronized data and provider history as hostile, while the local
OS/filesystem execution environment is part of the trusted computing base for
baseline conformance. Implementations must remain robust to ordinary synchronization
races and corrupted/malicious synchronized bytes, but baseline v1 does not require
immunity to an actively malicious same-privilege local process racing individual
filesystem syscalls. Stronger hostile-local-filesystem hardening remains permitted
as an implementation security feature, not required for protocol conformance.

Every candidate still earns protocol meaning only through the normal v1 structural,
keyed-identity, AEAD, semantic, and durable-state validation rules. Materialization
by a local synchronization client never makes synchronized state truthful.

## 2. Motivation

The r13 filesystem wording supported a defensible interpretation requiring hostile
same-UID syscall-race defenses. That interpretation drove implementations toward
complexity and audit costs disproportionate to the core protocol goal, while still
not establishing a globally atomic filesystem snapshot. r14 explicitly narrows the
baseline instead of describing this security tradeoff as editorial cleanup. It
neither rejects the prior work nor prohibits implementations from keeping it.

This revision uses only the current repository and the supplied design decision;
no newer external Totipo text or platform documentation was imported. Work stayed
in `totipo-spec`; no `totipo-java` work or new protocol behavior was implemented.

## 3. What remains hostile

Synchronized bytes, directory entries, provider history, ordering, completeness,
freshness, and availability remain untrusted. The storage attacker may read, copy,
add, replace, replay, reorder, withhold, and delete synchronized entries. Rollback,
replay, malformed files, corrupt ciphertext, and malicious bootstrap bytes remain
in scope, with unchanged cryptographic consequences and durable-history handling.

## 4. What is now trusted for baseline

The kernel/OS execution environment, filesystem implementation, mount/process
namespace, and processes acting with the application's local privileges belong to
the baseline trusted computing base. Baseline conformance assumes absence of an
actively malicious same-privilege adversary deliberately racing pathname targets,
file types, directory identities, or already-open inode contents between syscalls.
The synchronization client remains untrusted as a source of synchronized state.

## 5. Ordinary synchronization races

Appearance, disappearance, replacement, and read failure remain normal robustness
cases. Each discovery pass has a fixed accepted observation set: unavailable required
bytes make that pass incomplete; a later pass may observe the new state. An observed
pathname/type change requires re-evaluation or failure/retry. Torn or malformed bytes
undergo normal validation. Such observations cannot invent authenticated knowledge,
erase durable graph knowledge, or by themselves trigger `LOCAL_CONTINUITY_UNKNOWN`.
Authenticated same-ID contradictions and local durable-memory corruption retain
their existing continuity-failure behavior. Operation freshness remains required;
newly observed changes still stale/recompute operations under the existing rules.

## 6. Baseline filesystem obligations

| Obligation | Normative scope retained in r14 |
|---|---|
| Logical confinement | Exact `vault`, `objects-v1`, and 64 lowercase hex names; no untrusted traversal strings or sibling recursion |
| Candidate selection | Only observed regular-file direct children of exact `objects-v1/`; observed symlinks, directories, FIFOs, sockets, and devices are not objects |
| Family isolation | Unknown sibling names and unrelated/temp/conflict names have no semantic effect |
| Bounded processing | Bounded reads/allocations and parsing; hostile metadata cannot determine unchecked allocation or establish content identity |
| Validation | Normal structural, canonical, keyed OBJECT_ID, AEAD, semantic, and durable-state validation |
| Wrong-size evidence | Invalid current-family storage evidence, never authenticated opaque evidence |
| Discovery | Every accepted observation reaches terminal classification; required unavailable bytes/resource exhaustion prevent completeness |
| Immutable publication | Complete bytes, no overwrite, safe temporary creation against accidental collisions/clobbering |
| Crash correctness | Section 9.1 VAULT publication, required durability acknowledgement, no false success |
| Local custody | Existing permissions, secret-custody, and independently retained security-memory obligations remain |

Ordinary no-follow, exclusive-create, atomic-move/no-replace, and stable-handle
facilities remain recommended where readily available and useful. Baseline does
not require proof that they close every malicious local TOCTOU race.

## 7. Optional hardening

| Technique | r14 status |
|---|---|
| O_PATH/statx and repeated inode identity checks | Permitted implementation hardening; no baseline mandate |
| procfd exact-inode reopening | Permitted; no required portable equivalent |
| openat2 confinement | Permitted stronger local confinement |
| Hostile inode/path/type/rebinding race defenses | Valuable implementation evidence, outside baseline interoperability proof |
| SecureDirectoryStream or other stronger platform facilities | Permitted; no platform API prescription |
| Sandboxing and sandbox-escape defenses | Permitted additional endpoint protection |

No normative hostile-local-filesystem profile is introduced.

## 8. Crash durability

This revision does not relax full temporary writes, candidate file flush, candidate
validation, no-overwrite initial/immutable installation, atomic replacement where
available, containing-directory flush where supported, canonical reopen/authentication,
or the prohibition on in-place VAULT truncate/overwrite. Required durability must
be acknowledged before success. Ambiguous/interrupted publication remains incomplete
and non-success. Object publication still precedes durable graph insertion.
Section 9, including Section 9.1, is byte-identical to the starting specification.

## 9. Protocol-security invariance

There is no change to wire bytes, cryptographic domains/construction, OBJECT_ID,
v1 envelope, routing prefixes, object size, capacity formulas, TOKEN/DEVICE semantic
state, graph rules, readiness, candidate use, provenance, password semantics, or
TOTP algorithms. Bootstrap/root binding, rollback/history handling, opaque evidence,
canonical parsing, and secret-custody rules remain intact. Protocol version remains
1; BOOTSTRAP_VERSION, OBJECT_VERSION allocations, filenames, and namespaces do not
change. Semantic evaluator logic is unchanged.

## 10. Conformance impact

Existing storage cases remain valid with identical expectations. Section 55 adds
explicit baseline evidence for static special-entry exclusion, unavailable/changed
candidate handling, bounded reads, immutable no-overwrite publication, and no success
on ambiguous/interrupted publication. Syscall-race immunity is no longer baseline
evidence; stronger filesystem tests remain useful implementation security evidence.
The portable model does not prove live crash durability or local syscall atomicity.
No new vector or native filesystem-race simulator was added.

## 11. Vector impact

All 90 case JSON files are byte-identical to starting HEAD. All 90 required case IDs
and SHA-256 values are unchanged. No signatures, ciphertext, crypto/encoding fixtures,
storage expectations, or graph results were regenerated. The generator was not run.
The manifest differs only by its revision metadata; order, section references, and
per-case hashes are unchanged. Both schemas are byte-identical: neither enumerates
a revision requiring an edit. The requirements profile remains `moving-pre-rc`,
with only revision, spec hash, and manifest hash updated. No RC profile was created.

## 12. Implementation impact

Existing hardened Linux implementations remain conforming and may retain their
stronger defenses. Java, Android, and iOS implementations need not reproduce Linux
procfd/statx machinery solely for baseline v1 conformance. They still need ordinary
storage robustness, cryptographic validation, operation freshness, and platform
crash/durability evidence. No stronger implementation was removed or modified here.

## 13. Security tradeoff

An attacker who can execute as the Totipo user or compromise the trusted local
filesystem execution environment may race pathname operations, tamper synchronized
and local files, cause availability failures, potentially defeat filesystem-level
assumptions, and compromise file-backed local secret custody. Totipo's protocol
cryptography is not intended to turn a compromised local endpoint into a secure
execution environment. This limitation is distinct from malicious bytes delivered
through sync storage, which remain fully in scope.

This boundary does not make world-readable private keys acceptable, excuse careless
security-memory storage, permit password/key leakage, or remove permissions hygiene.

## 14. Verification

Starting local branch: `main`. Starting HEAD:
`618ced33487aebf768e908923abfd7ae34537a35`.
Before any modification, `git status --short` was empty; `make check` and
`make verify` both passed. The baseline header was Totipo Vault Format v1, design
revision 13 / r13. The profile identified `totipo-v1`, `moving-pre-rc`, r13, with
90 required cases and these baseline hashes:

| Baseline file | SHA-256 |
|---|---|
| Specification | `513426313c7850b998239ed19a3d7399b32de203994931ae2d873516717ae228` |
| Manifest | `da289e75ccb2a8d73080cb8cc6ebcc2a6d39623edf2bfcbccce84b6b67dba8bd` |

Final exact hashes:

| File | SHA-256 |
|---|---|
| `spec/totipo-vault-format-v1.md` | `a60cf63a742b4855b5742e5280ac15cdf6a08e60c81b1a6242b1dea10a1630fb` |
| `vectors/manifest.json` | `2776ed6a8c62609384f0f5e34747b4ce80359fb1abc956137cc64e1ad92def41` |
| `vectors/manifest.schema.json` | `59bdc9165b1c7e940031c447c2da640185774a6e4fd52fa0fee0b245061e0440` |
| `vectors/case.schema.json` | `79e33b68bf948dcc332075827b791c8eb8e54d9cb27634dd59c0c6ac1fd95fd5` |

| Command/check | Result |
|---|---|
| `make spec-check` | PASS: v1/r14 structural checks |
| `make test` | PASS: four Python tests and all Go packages |
| `make conformance` | PASS: all 90 v1/r14 cases |
| `make verify` | PASS: manifest and 90 case JSON schemas, manifest/profile/file hash validation |
| `make check` | PASS: full aggregate checks |
| `make race` | PASS: Go race tests |
| `git diff --check` | PASS |
| Direct byte comparisons against `git show HEAD:<path>` | PASS: all 90 cases and both schemas unchanged |
| Required-case array comparison | PASS: all 90 IDs/hashes and their order unchanged |
| Manifest byte comparison allowing only revision replacement | PASS |
| Specification section/history comparison | PASS: Sections 5–23, 25–32, 34–49, 51–53 and all history from r13 onward unchanged |

The broad revision and filesystem searches were manually reviewed. The final
`current.*r13|spec_revision.*r13|Revision: r13|revision 13|Open work after r13`
search had no matches before this report recorded that search expression.
The final hostile-namespace/same-privilege/TOCTOU search retained only the new
boundary, optional defenses, synchronized-state custody rules, and historical text;
no old mandatory rebinding/safe-open proof remains.

## 15. Diff summary

| Changed path | Change |
|---|---|
| `spec/totipo-vault-format-v1.md` | r14 identity, summary, threat boundary, Sections 4/24/33/50/54/55/56 clarifications, revision history |
| `vectors/FORMAT.md` | Portable storage-model/conformance boundary |
| `vectors/manifest.json` | Revision metadata only |
| `requirements/v1-pre-rc.json` | Revision and exact spec/manifest hash pins only |
| `README.md` | Current revision, boundary, next step, review link |
| `CONTRIBUTING.md` | Current normative revision link label |
| `Makefile` | Current revision in help text |
| `conformance/README.md` | Current revision and live-adapter obligations |
| `conformance/cmd/totipo-conformance/main.go` | Current revision in result label |
| `conformance/cmd/totipo-vector-gen/main.go` | Current manifest/profile revision constants; generator not executed |
| `conformance/internal/storage/storage.go` | Adapter obligation comment only |
| `conformance/internal/vectors/runner.go` | Active manifest/profile revision validation |
| `requirements/README.md` | Current profile revision and unchanged corpus description |
| `tools/check_spec.py` | Active revision/header checks and r14 history presence |
| `vectors/README.md` | Current corpus revision |
| `review/V1_R14_THREAT_MODEL_SIMPLIFICATION_REPORT.md` | This decision, audit, verification, and tradeoff report |

All changes are uncommitted. There are no changed case, schema, crypto fixture, or
historical report/snapshot files.

## 16. Issues/blockers

No blocking wording or verification issue remains. The stronger mandatory adapter
wording in `conformance/README.md` and the storage package comment was discovered
and corrected alongside Section 50. Secret-custody requirements referring to the
hostile synchronized namespace remain baseline obligations, not local TOCTOU proof.

Remaining r13 occurrences are intentionally historical: the r13 revision entry and
report/link; descriptions of r13 recovery/provenance additions in spec open work,
consumer/vector docs, and this report; the unchanged r13 recovery fixture-preservation
comment/error; the pre-r13 graph-fixture comment; and structural checks retaining
r13 history/concepts. Current manifest/profile validation expects r14.

Independent live implementation consumption, platform durability/interoperability,
and external security review remain pre-RC work under Section 56. No native safe-open
or hostile same-UID race proof is a baseline RC blocker. Repository verification does
not claim to supply missing production platform evidence.
