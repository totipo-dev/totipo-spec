# V1 r13 hardening report

## Baseline

- r12 commit: `f0135637a9d8bff5df09d98dd8e2359f5b541375`.
- Actual starting corpus: **85 cases**. All cases present at task start form the
  regression baseline; no r11 integration or fixed target count was used.
- Baseline `make check`, `make conformance`, and `make verify`: PASS (85 cases).
- Baseline evidence captured in `/tmp/totipo-r12-baseline.json`: commit, spec,
  manifest and profile SHA-256, every case ID/path/hash, and total count.
- Work started in a clean worktree at the r12 commit because the original checkout
  contained a user change to `flake.nix`. That change and the supplied untracked
  artifacts are outside this integration. The execution instructions are not
  staged or committed.

| Baseline file | SHA-256 |
| --- | --- |
| `spec/totipo-vault-format-v1.md` | `2c601766845315f5a8c2122c80eb76bf882a9bcdc59edfedd3b86ce6b74fc7f6` |
| `vectors/manifest.json` | `7bbc4e3b645246f0bfa51648097d4954f263a5aa9cb11b2d09b36575d1ae3c2e` |
| `requirements/v1-pre-rc.json` | `6a2426557f0c6395ec4291ef9c03acd32cf9beefbdbdf427d81c9c3d6827abb3` |

## r13 delta

The live specification is byte-identical to the supplied
`totipo-vault-format-v1-r13.md`, SHA-256
`513426313c7850b998239ed19a3d7399b32de203994931ae2d873516717ae228`.
It was copied directly, not reconstructed from the review.

- Exact OPAQUE_UNSCOPED object retention: `LearnOpaque` authenticates the envelope,
  padding and keyed object ID, then stores OBJECT_ID as the record map key and a
  copied 1024-byte encrypted object. Storage observations carry those original
  bytes across the authentication boundary. Failed exact-byte persistence sets
  `PersistenceBlocked`; authoritative and candidate gates close.
- Retained-byte compatible reprocessing: `ReprocessOpaque` opens the retained
  encrypted object and feeds its authenticated semantic bytes to a compatible
  classifier. No synchronized copy is needed. Concrete retained records cannot
  use the older digest-only symbolic reclassification API. Successful durable
  supported/routable reclassification removes the active unscoped block; failed
  persistence leaves it active. Exact bytes may remain as a recovery copy.
- Reset historical-frontier consequence: the replacement epoch uses only new
  scan evidence. In `a <- b <- c`, dropping missing `b` during reset removes the
  learned path that suppressed `a`; `a` and `c` can become conflicting current
  heads. No ancestry is invented, timestamps are not changed, and reset does not
  make `a` semantically newer. The spec requires this warning before confirmation;
  reference model comments/case notes state the UI precondition.
- Late provenance recomputation: `graph.Provenance` tracks immutable known TOKEN
  evidence and recomputes on matching DEVICE/restored/local public-key arrival.
  Both UNRESOLVED -> VERIFIED and UNRESOLVED -> REJECTED are exercised with the
  existing fixed child signature and an in-memory invalid-signature variant.
  TOKEN_VALUE, validity, ancestry, heads, and credential authority are invariant.
  Attribution can also update for retained evidence whose value is unavailable,
  without restoring availability or authority.
- Success-gate wording alignment: existing r12 publication behavior already
  accepts DEVICE -> TOKEN -> success and TOKEN -> DEVICE -> success, and rejects
  TOKEN -> success before a matching durable DEVICE. No new ordering restriction.
- Resource-complete definition: every candidate in a fixed scan snapshot must
  reach terminal classification. Unclassified observations, exhausted resources,
  and unavailable required bytes prevent baseline completion; invalid storage is
  terminal. Replacement discovery/persistence failures preserve continuity unknown.
  `BaselineLearnOpaque` applies exact-byte retention to the replacement epoch.
- Editorial/fold cleanup: copied verbatim from the supplied artifact. The
  four-parent fold is illustrative and state-size-dependent; capacity formulas
  are unchanged. Structural checks cover the new concepts and retained history.

The older symbolic graph fixtures remain abstract topology evidence. They are
not presented as proof of concrete object retention. The new cases and boundary
tests supply that missing evidence. The compatible classifier is a controlled
interpretation of one existing synthetic fixture, not a new published grammar or
OBJECT_VERSION allocation. No production persistence or UI is implemented here.

## Existing r12 coverage reused

| Requirement / behavior | Existing coverage and delta decision |
| --- | --- |
| First DEVICE success gate, both allowed orders, premature success | `v1.provenance.initial-device-before-token.001` already proves all three; no new ordering case. |
| Sticky unscoped disappearance and compatible reclassification lifecycle | `v1.future.unscoped-sticky.001`; reused, but its symbolic digest does not prove exact bytes or retained-byte classifier input, so two concrete r13 cases were needed. |
| Reset with offending object absent/present | `v1.future.unscoped-rebaseline-absent.001`, `v1.future.unscoped-rebaseline-present.001`; reused. Missing-ancestry frontier changes need a new case. |
| General persistence blocks candidate use | `v1.candidate.persistence-block.001`; reused. Exact-object persistence failure is included in the new retention case, with no redundant standalone case. |
| Incomplete discovery allows selected safe candidate use | `v1.candidate.discovery-incomplete.001`; reused. It does not prove terminal classification for a replacement baseline, so a new baseline case was needed. |
| DEVICE convergence across rejected/unresolved heads | `v1.device.rename-incorporates-rejected-head.001`; no change. |
| Remote known-ID corruption versus local security-memory corruption | `v1.graph.known-id-corrupt-bytes-retain-node.001`, `v1.graph.local-security-memory-corruption.001`; no change. |
| Vault-bound signature context | `v1.provenance.signature-context-cross-vault.001`; no change. |
| Static TOKEN provenance outcomes | `v1.provenance.token-verified.001`, `v1.provenance.token-rejected.001`, `v1.provenance.token-unresolved.001`; reused. None proves late-key recomputation, so a new transition case was needed. |
| Fold/capacity behavior | Existing size/fold cases remain unchanged. New wording is checked structurally; no redundant byte case. |

## New cases

Exactly five cases were added, with independently authored expectations:

1. `v1.future.unscoped-retains-exact-object.001`
2. `v1.future.unscoped-reprocess-retained.001`
3. `v1.provenance.late-device-reclassify.001`
4. `v1.graph.rebaseline-history-reappears.001`
5. `v1.graph.rebaseline-resource-incomplete.001`

The case schema, consumer, moving manifest/profile, generator carry-forward list,
and format documentation cover these payloads. The generator retains reviewed
r13 workflow bytes and validates resolved references before reporting success.
It was not run to regenerate existing crypto fixtures during this integration.

## Regression evidence

All 85 pre-existing r12 case hashes and complete manifest entries were compared
with the captured baseline. Every comparison passed.

| Regression check | Changed |
| --- | ---: |
| Old case files / SHA-256 | 0 |
| Old IDs / paths / manifest entries | 0 |
| Old expected results | 0 |
| Old protocol-bearing payloads | 0 |
| Old crypto intermediates / encrypted objects | 0 |

Byte-identical files establish the unchanged encoded semantics, object bytes,
crypto intermediates and expected results. All old cases also execute unchanged.
No crypto, object, TLV, TOTP, Nix dependency, or historical report files changed.
No ECDSA nonce implementation, signature regeneration, or extra crypto library
was introduced. No RC profile, tag, or release was created.

## Protocol impact

- wire format changed: **NO**
- crypto construction changed: **NO**
- storage-family/routing changed: **NO**
- local durable-state requirement changed: **YES**
- provenance transition semantics changed/clarified: **YES**

## Validation

- Final count: **85 + 5 = 90 cases**.
- `python3 tools/check_spec.py`: PASS.
- `go -C conformance test ./...`: PASS.
- `go -C conformance vet ./...`: PASS.
- `make check`, `make conformance`, `make verify`: PASS, all 90 cases.
- `make race`: PASS; affected graph/vector packages were rechecked after final
  recovery boundary tests and provenance evidence validation.
- `make fuzz`: PASS for FuzzDispatch, FuzzOpen and FuzzArrivalAndDisappearance,
  each with a 10-second budget and two workers.
- Added boundary tests cover unauthenticated input, corrupted retained bytes,
  failed classifiers, prevention of digest-only bypass, incomplete resets,
  replacement exact-object persistence failure, unrelated key arrival, immutable
  provenance evidence, unavailable TOKEN values, and deliberately wrong case
  expectations/references.
- `git diff --check`, formatting, exact supplied-spec comparison, and regression
  hash/entry audit: PASS.
- Local environment: Go 1.26.7, Python 3.14.7, Linux amd64.
- Nix: unavailable on PATH; `nix develop --command make check` could not run.
  Nix files were not modified by the integration.
- CI: the existing workflow covers Linux/macOS/Windows and Go 1.23/stable. This
  local integration has not been pushed, so no new remote CI result is claimed.

## Remaining work

- Independent live implementation consuming these vectors.
- Production persistence/filesystem/crash/freshness validation, including exact
  opaque-byte storage and destructive-reset warnings before confirmation.
- Native platform cross-verification (Java/JCA, Android Keystore, Apple CryptoKit).
- External specification/security review before RC.
- Push the integration and obtain green cross-platform CI. Then stop modifying
  `totipo-spec` unless the independent live implementation exposes a concrete issue.

There are no unresolved r13 integration ambiguities or local validation blockers.
Nix and remote CI remain unverified as described above. r13 is integrated as a
delta from r12 with no wire-format regression.
