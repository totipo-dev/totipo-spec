# Totipo v1/r16 normative rewrite report

The five design checkpoints have been applied to a coherent normative v1/r16
specification and current conformance corpus. Work is uncommitted and ready for
manual review. No commits, pushes, tags, releases, RC freezes, Java work, or changes
outside `totipo-spec` were performed.

## Baseline

Starting commit: `0a01fc1f0493ec1fefc0c4b89f8b22490d586852`. `git status --short` was empty (clean); no
user checkpoint exceptions were needed. The normative baseline was v1/r15.

Baseline `make check`, `make verify`, and `make conformance` all passed.
The baseline had **105 cases: 93 baseline and 12 conditional advisory-history**.
The manifest and all 105 physical case hashes were verified before editing.

| Baseline artifact | SHA-256 |
| --- | --- |
| `spec/totipo-vault-format-v1.md` | `f62022718044b172092ec321b48ddd14c83c7b791587050cc0ae4678cdefec35` |
| `vectors/manifest.json` | `2ddf067d119e3512fa033bd4736a1a9f64042985b8a23968ff89778eb4ce5512` |
| `requirements/v1-pre-rc.json` | `756cb4cf2ec642c597928bca896c895098f79d3870e1a493ae95e71e5e51ec81` |

Every baseline case ID and full SHA-256 appears in the migration table below and
in [machine-readable case audit](V1_R16_CASE_AUDIT.json). The read-only
[`tools/audit_r16.py`](../tools/audit_r16.py) reproduces that audit against the
starting commit and current worktree.

## Design inputs and preservation

Chronological order is authoritative; later explicit resolutions supersede earlier
open questions. The first four checkpoints were hashed before editing and verified
byte-identical at Phase 1 and final validation. The fifth was completed before
normative edits and has remained unchanged since then.

| Order | Checkpoint | SHA-256 |
| --- | --- | --- |
| 1 | [V1_R16_SIMPLIFICATION_DECISIONS.md](V1_R16_SIMPLIFICATION_DECISIONS.md) | `0e4811836c2a9f6fc0fc590093dee94554b68cb965b5f66639da882a47a52940` |
| 2 | [V1_R16_STATE_POLICY_DECISIONS.md](V1_R16_STATE_POLICY_DECISIONS.md) | `99dffeeda7e41287b994c58519b1d67d20373e0dffa44df671487ede4cfac077` |
| 3 | [V1_R16_STORAGE_IDENTITY_DECISIONS.md](V1_R16_STORAGE_IDENTITY_DECISIONS.md) | `7755c3d912b63a2c251b27bc550e6ee89386294072236e424274f76a647285a0` |
| 4 | [V1_R16_STORAGE_OPERATION_DECISIONS.md](V1_R16_STORAGE_OPERATION_DECISIONS.md) | `cba4b2cff04dbde5733b04df1488d83f1de0e2905fa8575aa84a264b61ec6ac8` |
| 5 | [V1_R16_ENVELOPE_DECISION.md](V1_R16_ENVELOPE_DECISION.md) | `d9a0d389921f87f2b166e1c2b17b0e4798cfe939cb3b2a86e41041b931a7487e` |

Phase 1 added only the non-normative envelope decision. `git diff --check` passed;
all baseline artifact hashes still matched, and `git diff --name-only` was empty,
confirming no tracked normative/corpus/tooling edits before Phase 2. The fifth
checkpoint records r15 as normative until the rewrite, the unchanged historical
status of earlier checkpoints, the settled envelope decision, and the remaining
editorial same-ID phrasing task. No separate checkpoint was needed for that wording.

The r15-and-earlier revision-history suffix in the specification is byte-identical
to baseline. All pre-existing review artifacts and historical audit tooling are
unchanged. No archived v0 or release/tag state was changed.

## Normative changes

- Replaced the 58-section r15 structure with 21 coherent sections, including preserved
  historical revision entries and a new r16 entry. Internal numeric references resolve.
- TOKEN is the sole semantic grammar in exact `objects-v1/`. Removed DEVICE objects,
  advertisement, rename/folding, device identity, key discovery, first-device setup,
  and first-TOKEN advertisement prerequisites.
- Removed P-256/ECDSA/DER signatures, provenance fields/status/reclassification,
  device private-key custody, local key binding, and signature-context derivation.
  The unlocked vault root is the authority to author valid state.
- Flattened complete TokenValue fields. Tombstones retain full credentials and may
  be inspected or used for TOTP. No protocol use-permission catalog or gate remains.
- Fixed `MAX_PARENTS = 4`, sorted duplicate-free explicit parent TLVs, and linear
  folds taking four originals then previous fold plus up to three originals. Every
  stage preserves the same operation TokenValue and exact client metadata presence/value.
- Replaced fatal cycles with mutual-reachability/SCC causal equivalence. Every member
  of each maximal group is a head. Equal values remain unambiguous; differing values
  are ordinary conflict. Descendants supersede the entire group. Each current or
  retained historical object exposes its own exact metadata; equal semantic values
  never synthesize, choose, or discard head metadata.
- Missing, invalid, or different-TOKEN_ID parents leave unresolved edges. Children
  remain valid; no skipped paths or inferred grandparents are allowed.
- History unavailability alone does not prohibit authorship when TOKEN_ID and the
  complete desired TokenValue are otherwise known or supplied. Disappearance of previously
  considered information does not stale an informed decision; new relevant
  alternatives must be disclosed before claiming resolution of all known alternatives.
- Removed exhaustive discovery, readiness/resource-completeness states, mandatory
  graph persistence, advisory-history capability, remembered-head warnings, local
  cache/store union, and confirmation-generation authorization machinery.
- Defined the configured durable store independently of optional synchronization.
  Observation yields validated objects, unresolved parents, and diagnostics.
- Specified complete-byte staging, reasonable crash safety, and file/namespace
  persistence intent for new immutable publication. Exact existing bytes succeed
  read-only without fresh persistence; non-exact targets stay untouched. Ambiguous
  outcomes report no success and introduce no pending protocol state.
- Canonical `vault` absence suffices for creation; orphan files do not block it.
  Creation has no mandatory local establishment step. Password change retains root
  and fingerprint, compares exact current bytes to the opened BASE immediately before
  replacement, rejects observed stale changes, and accepts the residual non-CAS race.
- Replaced mandatory local binding with non-secret optional VAULT_FINGERPRINT
  recognition. Mismatch means a different vault, not invalidity or lack of permission.
- Same-ID contradiction is a narrowly scoped integrity failure: do not choose a
  plaintext, exclude that identity from normal semantic evaluation, and leave its
  references unresolved. No persistent recovery architecture is introduced.

## Wire-format and naming changes

r16 intentionally changes TOKEN semantic plaintext and consequently TOKEN wire
bytes. Removed fields: OBJECT_TYPE, OBJECT_VERSION, AUTHOR_DEVICE_ID, AUTHOR_TIME,
SIGNATURE, and the nested CREDENTIAL framing. DEVICE grammar is entirely removed.
Added fields: optional CLIENT_NAME and CLIENT_TIME. ALGORITHM, DIGITS, PERIOD,
and SECRET_BYTES are now root-level fields.

Canonical tags are consecutive `0x0001..0x000c`, in this order: TOKEN_ID,
PARENT_COUNT, PARENT_ID, STATUS, ISSUER, ACCOUNT, ALGORITHM, DIGITS, PERIOD,
SECRET_BYTES, CLIENT_NAME, CLIENT_TIME. Parent count is u16be; explicit IDs remain
32 bytes. Metadata absence differs from present empty/zero and has no value,
causality, freshness, conflict-priority, authorization, or TOTP meaning. Canonical
width/UTF-8 checks still apply to metadata encodings.

The exact fingerprint domain is ASCII `totipo/v1/vault-fingerprint`, with no
terminating NUL, under HMAC-SHA-256 keyed directly by K_root.

The canonical bootstrap pathname remains exact lowercase ASCII **`vault`**, as in
r15. Conceptual VAULT names the representation, not an alternate pathname. No
uppercase or case-insensitive alias/fallback is allowed. The initial rewrite's
uppercase pathname interpretation was corrected before adversarial review; no
bootstrap encoding, wrapping input, or fingerprint derivation changed.

The outer object cryptographic construction is unchanged: keyed OBJECT_ID over P,
HKDF-SHA-256 per-object key, ID-derived nonce, fixed AAD, AES-256-GCM, authenticated
u16 semantic length, and canonical zero padding. Root wrapping and password encoding
remain unchanged. Existing ASCII/empty/Unicode bootstrap password, salt, nonce,
wrap key, header, and wrapped record bytes were compared against baseline and match.

## Crypto and envelope impact

Signature/provenance crypto and its key branch are deleted. Object addressing,
HKDF, and AES-GCM have not been redesigned. Deterministic IDs, derived object keys,
ciphertexts, tags, and 1024-byte TOKEN fixtures were regenerated because their
semantic plaintext changed, not because the envelope algorithm changed.

The envelope remains exactly 1024 bytes: 1008 encrypted plaintext plus 16-byte tag,
with two authenticated length bytes and 1006-byte semantic capacity. Worst-case
non-parent fields are 861 bytes; four parents cost 144; maximum TOKEN is 1005,
leaving one spare semantic byte. Uniform storage shape, bounded implementations,
canonicality, and cross-platform consistency are primary. Exact-length hiding is
secondary; variable-size authenticated encryption is technically viable. No
filesystem-block-size security rationale or speculative version headroom is used.

## Corpus changes

Old count: **105**. New count: **90**.

By ID: **80 removed**, **65 added**, **25 retained IDs**.
Of retained IDs, three RFC TOTP files are byte-identical (manifest section metadata
changes), eight TOKEN cases are wire-regenerated, and fourteen are simplified
bootstrap/graph/storage equivalents. Nineteen retired IDs have explicit replacement
IDs; sixty-one cases are removed because their protocol concepts disappeared.

| Classification of every baseline case | Count |
| --- | --- |
| retain unchanged | 0 |
| retain but metadata/section refs change | 3 |
| regenerate for changed r16 wire bytes | 8 |
| replace with new r16 equivalent | 33 |
| remove because concept is deleted | 61 |

“Retain unchanged” counts the full case contract including section metadata; none
qualify because sections were renumbered. The three byte-identical files are listed
explicitly below. Retained bootstrap case records change only binding-to-fingerprint
metadata; their wrapping bytes stay exact.

New coverage includes metadata absent/empty/max and absent/zero/max-u64, zero through
four parents, fifth-parent rejection independent of value size, maximum envelope,
malformed/unknown grammar, complete tombstones, fixed folds, equal/conflicting/self/
late cycles, missing parents, same-ID defensive exclusion, identical collapse,
diagnostic-only incomplete observation, fingerprint/rewrap, orphan-safe creation,
exact-existing and non-exact publication, ambiguous failure, and exact replacement
comparison. Integration tests connect real encrypted folds to graph evaluation and
compute the RFC answer from complete live and tombstoned values. Parent-unavailable
publication uses a real child fixture whose parent is not supplied to the backend.

Graph expectations are authored explicitly from the specification; cycles and
same-ID contradictions are abstract validated-fact fixtures, not fabricated crypto
collisions. Storage workflow booleans describe backend outcomes and cannot prove
actual staging/flush syscall behavior. Descriptive UI/confirmation obligations remain
normative; no replacement permission-state simulator was invented merely for tests.

### Removed IDs

```text
v1.routing.token-v1.001
v1.routing.device-v1.001
v1.encoding.device-root.001
v1.device.explicit-id.001
v1.routing.token-future-opaque.001
v1.routing.device-future-opaque.001
v1.crypto.future-token-opaque.001
v1.crypto.future-device-opaque.001
v1.routing.unknown-type-unscoped.001
v1.routing.future-token-malformed-prefix.001
v1.routing.future-device-malformed-prefix.001
v1.device.id-mismatch.001
v1.encoding.author-time-zero.001
v1.encoding.author-time-u64max.001
v1.timestamp.zero.001
v1.timestamp.normal.001
v1.timestamp.i64max.001
v1.timestamp.u64max.001
v1.size.token-max-5-fold.001
v1.size.device-max-14.001
v1.size.device-max-15-fold.001
v1.size.short-der-no-extra-parent.001
v1.crypto.device-root.001
v1.provenance.token-verified.001
v1.provenance.token-rejected.001
v1.provenance.token-unresolved.001
v1.provenance.der-short-valid.001
v1.provenance.der-max-valid.001
v1.graph.remembered-head-absent.001
v1.graph.cycle-integrity-failure.001
v1.graph.global-object-id-conflict.001
v1.future.scoped-token.001
v1.future.concurrent-supported-opaque.001
v1.future.unrelated-token.001
v1.future.supported-descendant.001
v1.future.version-does-not-order.001
v1.future.device-presentation.001
v1.future.unscoped-warning.001
v1.future.unscoped-candidate-use.001
v1.future.disappearance-removes-current.001
v1.candidate.conflict-a.001
v1.candidate.current-peer-missing.001
v1.candidate.historical-current-missing.001
v1.candidate.opaque-current.001
v1.candidate.discovery-incomplete.001
v1.candidate.cache-write-warning.001
v1.candidate.history-memory-lost.001
v1.timestamp.equal-value-different-times.001
v1.timestamp.fold-common-time.001
v1.graph.reappearance-mismatch.001
v1.storage.wrong-size-not-opaque.001
v1.future.family-compat-shadow.001
v1.future.family-no-shadow.001
v1.device.rename-incorporates-rejected-head.001
v1.graph.known-id-corrupt-bytes-warning.001
v1.graph.history-cache-corrupt.001
v1.provenance.initial-device-before-token.001
v1.provenance.signature-context-cross-vault.001
v1.provenance.late-device-reclassify.001
v1.snapshot.late-arrival-before-publication.001
v1.snapshot.hidden-history-after-publication.001
v1.snapshot.unrelated-change.001
v1.history.memory-lost.001
v1.history.regression-and-return.001
v1.snapshot.incomplete-usable.001
v1.publication.ambiguous-retry-sibling.001
v1.confirmation.visible-conflict.001
v1.confirmation.unrelated-change.001
v1.confirmation.relevant-change.001
v1.confirmation.relevant-information-return.001
v1.snapshot.opaque-token-plan-blocked.001
v1.publication.cache-failure-after-success.001
v1.bootstrap.local-binding-establishment.001
v1.history.unscoped-disappearance-warning.001
v1.history.opaque-disappearance-warning.001
v1.history.current-peer-missing-warning.001
v1.history.current-child-missing-warning.001
v1.graph.corrupt-current-bytes-excluded.001
v1.history.cache-failure-after-publication-warning.001
v1.snapshot.no-advisory-history.001
```

### Regenerated retained IDs

```text
v1.encoding.token-root.001
v1.encoding.parent-order.001
v1.encoding.parent-count-mismatch.001
v1.encoding.duplicate-nonrepeatable.001
v1.encoding.utf8-boundary.001
v1.size.token-max-4.001
v1.crypto.token-root.001
v1.crypto.token-child.001
```

### Byte-identical case files

```text
v1.totp.rfc6238-sha1.001
v1.totp.rfc6238-sha256.001
v1.totp.rfc6238-sha512.001
```

### Every baseline case: exact hash and disposition

The linked JSON audit adds exact new hashes, reasons, and replacement-ID lists.

| Baseline ID | Baseline SHA-256 | Classification and replacement |
| --- | --- | --- |
| `v1.routing.token-v1.001` | `7c9cfa6f6a045f4b3cde16fd7e5463848bfa992f04604c5bd8ee2b9cfbf66db7` | replace with new r16 equivalent; `v1.encoding.token-root.001` |
| `v1.routing.device-v1.001` | `7582fd0ff94b78986c610fe1d55654497f1eada8da35080ec5fec0f27c62624d` | remove because concept is deleted; — |
| `v1.encoding.token-root.001` | `55fca1e27555d8e98f1eff9235a17912ff5ec528d1cad8c819ef52b04084ae51` | regenerate for changed r16 wire bytes; `v1.encoding.token-root.001` |
| `v1.encoding.device-root.001` | `4f6f19d87c0543c734235079019033ee6f94abaf91e7b8d7756c0cc4087df765` | remove because concept is deleted; — |
| `v1.device.explicit-id.001` | `2c3986efdcb032d5cf4317339eeb7c9a44946de4ae1c6fd10e0aa9114330eb87` | remove because concept is deleted; — |
| `v1.routing.token-future-opaque.001` | `2dba7e6f240bba4dfc68343cf08be5abd9da1b4f5947d513acbbfd11c02a57b2` | remove because concept is deleted; — |
| `v1.routing.device-future-opaque.001` | `ae923d48d64f87a0d600bfa3b0ddaf30558d155cdaf2f748486a9ca8f0ccb0c3` | remove because concept is deleted; — |
| `v1.crypto.future-token-opaque.001` | `ad004790fb056dcc426f059bbb22ebbbde99f1af0a6624fad088dd0c0072c39e` | remove because concept is deleted; — |
| `v1.crypto.future-device-opaque.001` | `f3cc040f8cf16d6319143a8bc679b4f1e2418931ab4cad464b1ac5c21c649d39` | remove because concept is deleted; — |
| `v1.routing.unknown-type-unscoped.001` | `35640f317dcbae50f81e491ea6831d2a038ce93580ccfab9992dbc7388b6b857` | remove because concept is deleted; — |
| `v1.routing.future-token-malformed-prefix.001` | `8f6a31ffd162a1e7be4d817d090003604a8ae5a19e9c333cb8bab57b25477af4` | remove because concept is deleted; — |
| `v1.routing.future-device-malformed-prefix.001` | `525111d12685afc0189fabe16afa66dbb841eb1020518d388092b9819511a376` | remove because concept is deleted; — |
| `v1.encoding.parent-order.001` | `1ef03d4d80d388bbb7a45402516df0a6a52f9fe1770cdeec5c0dfed3fc821748` | regenerate for changed r16 wire bytes; `v1.encoding.parent-order.001` |
| `v1.encoding.parent-count-mismatch.001` | `ff91d61efa0a877e96c4c7fb6dcb44f4824e4b3f6dd92370be6889f90ebc8f06` | regenerate for changed r16 wire bytes; `v1.encoding.parent-count-mismatch.001` |
| `v1.encoding.duplicate-nonrepeatable.001` | `1a85b74b4c5cf818b48f9139cedd6763f3bf1f02afad955f3a1b7660a89ab947` | regenerate for changed r16 wire bytes; `v1.encoding.duplicate-nonrepeatable.001` |
| `v1.device.id-mismatch.001` | `e4f8e5e1e91f09fd0314520541bb29d0666bb92222a3555b069f52c45582605b` | remove because concept is deleted; — |
| `v1.encoding.author-time-zero.001` | `5fe8833c172304b87324d5258d5f82f9ce9895876d564a397e53d678d1329c5d` | replace with new r16 equivalent; `v1.metadata.client-time-zero.001` |
| `v1.encoding.author-time-u64max.001` | `b5396ee0553c7354b11a02f54979e5fa7c9439fab6441fceaba4d58ca6282d5b` | replace with new r16 equivalent; `v1.metadata.client-time-u64max.001` |
| `v1.timestamp.zero.001` | `13c04284b1aaa30c9a6a8bc217332a41536a071bc213db76df323667b0d5066e` | replace with new r16 equivalent; `v1.metadata.client-time-zero.001` |
| `v1.timestamp.normal.001` | `7bbddb0d34b3d53af6c8891214006ef926b71fdfdfdc157aeeb7ca6dd83901c7` | replace with new r16 equivalent; `v1.metadata.client-time-normal.001` |
| `v1.timestamp.i64max.001` | `fa5d971499e70ef2f003de09b10f964dbbe19ceb85ea5c6e20bea89e4f1e7bd8` | replace with new r16 equivalent; `v1.metadata.client-time-u64max.001` |
| `v1.timestamp.u64max.001` | `cbc283e9f73d01c3d4ec7a5330b7a5478a3fc988638f0350715b229bca023e84` | replace with new r16 equivalent; `v1.metadata.client-time-u64max.001` |
| `v1.encoding.utf8-boundary.001` | `06aa463ca48e73f4cd41d99e7d547910b037c4d927ab12f2ccf0ad05d794b5db` | regenerate for changed r16 wire bytes; `v1.encoding.utf8-boundary.001` |
| `v1.size.token-max-4.001` | `6e7951f71ad002d2570303a1a9faabff9e038f26fa69740cc05669d3ec808c8d` | regenerate for changed r16 wire bytes; `v1.size.token-max-4.001` |
| `v1.size.token-max-5-fold.001` | `d7037b17fa7faf2bb21a99f8cf5c0f48450a9375cac711de2039cfff6cff9a02` | replace with new r16 equivalent; `v1.encoding.five-parents.001`, `v1.fold.width-5.001` |
| `v1.size.device-max-14.001` | `50c8171786c0fe57eb521c31918b73f2d19daac9fe7903e6ee13de29921406d7` | remove because concept is deleted; — |
| `v1.size.device-max-15-fold.001` | `cf59636fbbf88438c02e7e02b8f08a69e8d9ad8c0cf876ef0afe2328d85a7696` | remove because concept is deleted; — |
| `v1.size.short-der-no-extra-parent.001` | `8b46b395387179ee4810cda17642a01cd1fcdb849179406a71b0be3abea13fc4` | remove because concept is deleted; — |
| `v1.crypto.token-root.001` | `ae82fd22120a02f8adb3f00496ed0841f97402c9325f7482fe2404c390571bff` | regenerate for changed r16 wire bytes; `v1.crypto.token-root.001` |
| `v1.crypto.token-child.001` | `e7cc354fa64d73dedfa0ae31fc848a50339992d00a2c25a5794b58532752f97e` | regenerate for changed r16 wire bytes; `v1.crypto.token-child.001` |
| `v1.crypto.device-root.001` | `0f45cf193ff8a04e5c20dbb39bf569d58a0a1dc3ce44071e203bc6a4d024ec92` | remove because concept is deleted; — |
| `v1.provenance.token-verified.001` | `7e9448c3c173dc4fadc36551796b0e4d3b674bcdaac18bda8c577473f36e0f30` | remove because concept is deleted; — |
| `v1.provenance.token-rejected.001` | `b7b662b4c798873a69bea1257e27d8c16b97cdad64ea135720a442e39665248e` | remove because concept is deleted; — |
| `v1.provenance.token-unresolved.001` | `ad4d175619bb4e3bae32e106742711b127d9946990bcf3630a4bf7ca3a6d4639` | remove because concept is deleted; — |
| `v1.provenance.der-short-valid.001` | `cf0c80b95f72052c55d782bc6c0aab3483d321d96b951560f9c6153c07e8a1bb` | remove because concept is deleted; — |
| `v1.provenance.der-max-valid.001` | `0a0757e56cc9021873feac6426849bad8ba5f0cf0b2aa95d21f0ef66120ba3bd` | remove because concept is deleted; — |
| `v1.bootstrap.ascii.001` | `cd8e39898af81b186fdd5056f776e1f7bdebb97f2da8fd07b43d13d99fd380c8` | replace with new r16 equivalent; `v1.bootstrap.ascii.001` |
| `v1.bootstrap.empty.001` | `95da80febe764a28ee900b226f4799dd4c1aa6bb0056e254c12c63b00da9bbba` | replace with new r16 equivalent; `v1.bootstrap.empty.001` |
| `v1.bootstrap.unicode.001` | `c080d612e1916734cd2d960d7bc0c99779ddfea77ce0d8e8da66c4f9d41a1b8d` | replace with new r16 equivalent; `v1.bootstrap.unicode.001` |
| `v1.graph.sequential.001` | `47eaa8eb6aab126871294be1fa0dd8ac7d3dc8f298b36615da95524d74f119f5` | replace with new r16 equivalent; `v1.graph.sequential.001` |
| `v1.graph.equal-concurrent.001` | `4636e373e55d5c02e6064653786c5b5fd9eb037af34c02ecaa2862675626bd8e` | replace with new r16 equivalent; `v1.graph.equal-concurrent.001` |
| `v1.graph.conflicting-concurrent.001` | `510e1135ac39c1ad3755ab748352f34266941ae6d276e726a3aee018bd1b8423` | replace with new r16 equivalent; `v1.graph.conflicting-concurrent.001` |
| `v1.graph.remembered-head-absent.001` | `7b9d3c6300996e8fd6b7c3947d22470a39ee542a8f576140a5f45bab8ab9d6f6` | remove because concept is deleted; — |
| `v1.graph.intermediate-disappears.001` | `008410dd7316984925f2b8c08404dce3732c4b412731a65ab827bcb2ba84fd2a` | replace with new r16 equivalent; `v1.graph.intermediate-disappears.001` |
| `v1.graph.late-parent.001` | `8c55564e288d856a06657ff574db1a810b0c67a625ea58874e3ced12f2a0081f` | replace with new r16 equivalent; `v1.graph.late-parent.001` |
| `v1.graph.cycle-integrity-failure.001` | `cd548f4ac91f843024b96662df88d076b6a6c259e7530e4ef7518ea5c30bd4d6` | replace with new r16 equivalent; `v1.graph.cycle-equal.001`, `v1.graph.cycle-conflicting.001`, `v1.graph.late-cycle.001` |
| `v1.graph.global-object-id-conflict.001` | `47e4382be26aff7a44dcac00fd38c6629ac7d978d180f86ade7de698d1938b26` | replace with new r16 equivalent; `v1.graph.same-id-defensive.001` |
| `v1.future.scoped-token.001` | `30101f1aeebffe70b08847f2c14346b9ca7914857054542574ac5ad0ff14522c` | remove because concept is deleted; — |
| `v1.future.concurrent-supported-opaque.001` | `9f79f73a41ad7070c92fc0dfeb8960fbb1bb1d03862798253e6b334fb9087c15` | remove because concept is deleted; — |
| `v1.future.unrelated-token.001` | `146c624cecd2a3bdb8a1f622411795526232eebd5325353fed968dca68c23544` | remove because concept is deleted; — |
| `v1.future.supported-descendant.001` | `d491b15b4fd86fbe11eae1b51984d315d538957330a20432702432a577b7fb58` | remove because concept is deleted; — |
| `v1.future.version-does-not-order.001` | `302f3b6469438d62ef1c7ecbdd0370b12878b5e5def169f94ae834bc511af56b` | remove because concept is deleted; — |
| `v1.future.device-presentation.001` | `786b52408405d0e2bd4573ae5c456ed6b499dfb2e4a9911b22dac1e0d40b878e` | remove because concept is deleted; — |
| `v1.future.unscoped-warning.001` | `72b585f4c0f205c8cd6da77f866eb52ba51dd35ef97c7b258e8f0a29b112f99b` | remove because concept is deleted; — |
| `v1.future.unscoped-candidate-use.001` | `f44001e169ea612c041330af3ac8fa461619634e061e478ddf4841c8e934ea15` | remove because concept is deleted; — |
| `v1.future.disappearance-removes-current.001` | `ff726a2aa1ec06d8b33a5f93d44ee6d84fbbb8d68d4f3f4433face4096c243fd` | remove because concept is deleted; — |
| `v1.candidate.conflict-a.001` | `107209cdea134ca2bd39f2f6728e261a38fb1db688a8c2d52e0a6aed5ce0870b` | remove because concept is deleted; — |
| `v1.candidate.current-peer-missing.001` | `7b2645dfc4a0021305344b029b36a2ca3cdbbd618aa2b5e6bb73a21ed98ea5b6` | remove because concept is deleted; — |
| `v1.candidate.historical-current-missing.001` | `f0307cabe51d3a46c56a902a706a07138e8193593fc9b14007d8d3f41ac1c0e0` | remove because concept is deleted; — |
| `v1.candidate.opaque-current.001` | `7eedd8bee05df3e018bbfc11618eea014b8b6081c4c4de288d68a30aa7890ca5` | remove because concept is deleted; — |
| `v1.candidate.discovery-incomplete.001` | `2bb3d356410591e71d90d824ff5244559c8ec442f046288ea6f65711437836b6` | remove because concept is deleted; — |
| `v1.candidate.cache-write-warning.001` | `0d02deac2bf894b714d900156b1ee7e2f082d456325f2ef2d2622ec836928ab4` | remove because concept is deleted; — |
| `v1.candidate.history-memory-lost.001` | `d1fae5d8c793afc9a5c9b63560826d3fbde2d3702033d8669dc875d88164ec60` | remove because concept is deleted; — |
| `v1.timestamp.equal-value-different-times.001` | `948a9ede1ee4528823fad88db7ab6e39028c56e0738c386eb22aa56da794a162` | replace with new r16 equivalent; `v1.graph.equal-concurrent.001`, `v1.metadata.client-time-normal.001` |
| `v1.timestamp.fold-common-time.001` | `bce78510a62ea2fe22ad21ffc66242b61efacc8149d0bdc921d64bfa34dda11f` | replace with new r16 equivalent; `v1.fold.width-10.001` |
| `v1.graph.wrong-identity-parent.001` | `44240522fca46902c3b4067486202ad2b1fe029f01fecc9e2d9e8b0d6052c474` | replace with new r16 equivalent; `v1.graph.wrong-identity-parent.001` |
| `v1.graph.reappearance.001` | `5c5a02fefa67d990f5d37b34108615669227805d1d619c6ccd9be62c2a0c638d` | replace with new r16 equivalent; `v1.graph.reappearance.001` |
| `v1.graph.reappearance-mismatch.001` | `ae6954abf6c89659c477ae723c600b1d111a4a959abef02ea94f2df8eede9780` | replace with new r16 equivalent; `v1.graph.same-id-defensive.001` |
| `v1.totp.rfc6238-sha1.001` | `f3fed610df11e572f1c5c569e8822c5d1076df40197fadb559eca1a4aa0b6625` | retain but metadata/section refs change; `v1.totp.rfc6238-sha1.001` |
| `v1.totp.rfc6238-sha256.001` | `eb9baec09844984cb7d7072164beb4d794ab9fb997428ca06ffc92d68ea5f25b` | retain but metadata/section refs change; `v1.totp.rfc6238-sha256.001` |
| `v1.totp.rfc6238-sha512.001` | `c2787120d475f894371b490d37f5096fd255d6a36e385e20bf092bd7f148b07a` | retain but metadata/section refs change; `v1.totp.rfc6238-sha512.001` |
| `v1.storage.objects-v1.001` | `39bf5efa810f84f97dd3c9ccd9bd074882ebeb48cdcd7dd0ccf922994eeec1dc` | replace with new r16 equivalent; `v1.storage.objects-v1.001` |
| `v1.storage.nonobject-name-ignored.001` | `abf7064713ee1b3db17affac184eab934fa0c8309c61608a8e174717a351667c` | replace with new r16 equivalent; `v1.storage.nonobject-name-ignored.001` |
| `v1.storage.wrong-size-not-opaque.001` | `9e964551073cadd76960210616aed6dae53491de44036e738410a00a3e127949` | replace with new r16 equivalent; `v1.storage.wrong-size.001` |
| `v1.storage.unknown-sibling-ignored.001` | `7da08a2f739ecd0dd7f6cbcf0157b4e22975b3d711c04ddcc8954f69c7d2baed` | replace with new r16 equivalent; `v1.storage.unknown-sibling-ignored.001` |
| `v1.future.family-compat-shadow.001` | `b8b8525836462c4247ce8ce70832434f99d81a1140c2d4df9ba943245d611ae5` | remove because concept is deleted; — |
| `v1.future.family-no-shadow.001` | `bc002da706c4bfaabe1b5c951f06826d3851fda8cc8e6e64c0dff6328f488aa4` | remove because concept is deleted; — |
| `v1.device.rename-incorporates-rejected-head.001` | `7669f424f2e1a6926b6ad9820bf8c11c9d1a35cedc4cb8f482879de2fec724c6` | remove because concept is deleted; — |
| `v1.graph.known-id-corrupt-bytes-warning.001` | `65b1d90767551986b54f2ff225d22755da552b29f04ea03b03653e74e034549d` | remove because concept is deleted; — |
| `v1.graph.history-cache-corrupt.001` | `5a25f03ff4048eb40f15cc81084b9d2b0045588af0628b74b45332ad7a306cae` | remove because concept is deleted; — |
| `v1.provenance.initial-device-before-token.001` | `999029b0726fc113354d1232669f197de19bb026c537bbf53725df6f350f5800` | remove because concept is deleted; — |
| `v1.provenance.signature-context-cross-vault.001` | `df926d8526bd26b08ded0517fb19bb555aad80e1295daecc1fa4c4bce3d39f54` | remove because concept is deleted; — |
| `v1.provenance.late-device-reclassify.001` | `9bb0148a2066440188f82e286c1b97ce8b52e58154606b3059c4a32ff5428a99` | remove because concept is deleted; — |
| `v1.snapshot.late-arrival-before-publication.001` | `153d9a21c0dd25319469c5732e67e54f54a5edfcebf547e868cad576149b59f4` | replace with new r16 equivalent; `v1.graph.late-parent.001`, `v1.storage.missing-parent-publication.001` |
| `v1.snapshot.hidden-history-after-publication.001` | `a7aa78855f14dfea46c205cccc754016a89c31cfc1d0ae8efea2eb7858b3071b` | replace with new r16 equivalent; `v1.graph.conflicting-concurrent.001` |
| `v1.snapshot.unrelated-change.001` | `f522cccc761572430e477c6bef284b9282ec8a037be2c9fcfdfe6ab76d318bc3` | remove because concept is deleted; — |
| `v1.history.memory-lost.001` | `e285e7fd58f4ccde248c703e814ede87c2c77d2eb612a88770593c9ebf329ea2` | remove because concept is deleted; — |
| `v1.history.regression-and-return.001` | `f41819b688d7766bb0fb4f6844369b4e25f15edc1d9091119e31b20234ca83f9` | remove because concept is deleted; — |
| `v1.snapshot.incomplete-usable.001` | `054c8176eac043f8e3774cdff596b52d644509ce1196eb375a55e6546f5dc7b0` | replace with new r16 equivalent; `v1.storage.incomplete-diagnostics.001` |
| `v1.publication.ambiguous-retry-sibling.001` | `f143377eeffc18c246c3909818e78a7795c208ad5f82d0e5ab5742d479379790` | replace with new r16 equivalent; `v1.storage.ambiguous-publication.001`, `v1.storage.exact-existing-publication.001` |
| `v1.confirmation.visible-conflict.001` | `20161ba34287b0ff471de189fb51583ddb106a3f638e52b78518f13088aee2e0` | remove because concept is deleted; — |
| `v1.confirmation.unrelated-change.001` | `79bfc10e31e6352472c117941840648a6cf30fce521b2ede7c0b4ceb99a6a12a` | remove because concept is deleted; — |
| `v1.confirmation.relevant-change.001` | `a1cfb569aee89b00a8839824110372c9cfd1a98bc05c171868ff344d11865e53` | remove because concept is deleted; — |
| `v1.confirmation.relevant-information-return.001` | `383e064e6ed852731fed7d8c0555512e5afc189543eb65fe713f83a31b7530b1` | remove because concept is deleted; — |
| `v1.snapshot.opaque-token-plan-blocked.001` | `bb333b9f526a7aab172d1c9f8a89ae080a1096f85962519666e12abdf5f69c34` | remove because concept is deleted; — |
| `v1.publication.cache-failure-after-success.001` | `9020a225f052cb30e13e0c7f61a18623d7421aed97d7b47a610aab19c25be9ed` | remove because concept is deleted; — |
| `v1.storage.exact-existing-publication.001` | `0367706fd21d636af8405596637dfe55367f920987a74b6214c32dcf27136c8c` | replace with new r16 equivalent; `v1.storage.exact-existing-publication.001` |
| `v1.bootstrap.local-binding-establishment.001` | `92889fc703c94953b912f656ba28099fd541782a33138c894498a99bf9c6ce88` | remove because concept is deleted; — |
| `v1.history.unscoped-disappearance-warning.001` | `56e080c998529a92a4553fff6c71e7e104bbb7051453891d6a5c5b2424d13065` | remove because concept is deleted; — |
| `v1.history.opaque-disappearance-warning.001` | `386d3c07b004ff195c8ea67f17bdc0822f59ce000aac1ecc4f507bfce7becc43` | remove because concept is deleted; — |
| `v1.history.current-peer-missing-warning.001` | `3fcfa1eee7521708d6756dacb6eaed09da1abf4a37ed4302edd7458f8d4e1610` | remove because concept is deleted; — |
| `v1.history.current-child-missing-warning.001` | `df6591e9e331059a99bdf81be895264cd911f937eb0243edfc662ec26a37bddc` | remove because concept is deleted; — |
| `v1.graph.corrupt-current-bytes-excluded.001` | `f14a4d224b9476dac7250af041067c23d13f45c3d5f89eee4445e09da537f57c` | replace with new r16 equivalent; `v1.storage.failed-aead.001` |
| `v1.history.cache-failure-after-publication-warning.001` | `9c5dbc28fb3fe82991cfd377d89a6948d9979a8683916babec9901f2f989571d` | remove because concept is deleted; — |
| `v1.snapshot.no-advisory-history.001` | `c9ecf724ebb6def958069c72c3a16cdafb36d9d864767b42dda92c1d287772dc` | remove because concept is deleted; — |

### Added IDs and exact hashes

| New ID | SHA-256 |
| --- | --- |
| `v1.bootstrap.rewrap.001` | `88f337eadb06b8054dedcb2323da237feb932e725fae0ddc1488111474f4bcc3` |
| `v1.crypto.nonzero-padding.001` | `c88238486775495cb5b538eccbf5630b84a7db1b3dd36f08321f5e15b813eaa6` |
| `v1.crypto.object-id-mismatch.001` | `b95a6c07d355c79bb5bee38ae140b7ca877825276d2c198869f63a147ecfe199` |
| `v1.crypto.semantic-length-invalid.001` | `1dc9935f90536e7ae8e1766a1e641a4623c4b4fabc86eaebb0bad5f9d188ba67` |
| `v1.encoding.account-too-long.001` | `9224d0a78c7f76f47373fcc23d84f4cc2f93873e3fb6453ad54f999b4f944735` |
| `v1.encoding.algorithm-range.001` | `48b5d49a95f6e3467bae0d8feda41fb717ca3e849b04ca104ff51322ee726a0b` |
| `v1.encoding.client-name-invalid-utf8.001` | `8f3bd35ecc2987835873d574aa326d540a2a19ab887fcb81976d0286bb952120` |
| `v1.encoding.client-name-too-long.001` | `20de7ce90f93d2c91cc8dc8f6e4c3baf4555477ed39f8145059e2924c46a5841` |
| `v1.encoding.client-time-width.001` | `cfbb54b74b7633667ef626c7863532534e9810477135ad9b4889234b4f65d26a` |
| `v1.encoding.complete-tombstone.001` | `7bbc696d5e50c00a409053798455225d1a493e06ad013c167e7fcda2082ce6e8` |
| `v1.encoding.digits-range.001` | `67fccd5522def5ea2ddf7d9a9eaffe1379b92137b4268cbe313e3ab5b7b94db7` |
| `v1.encoding.five-parents.001` | `1f782aca234d9a818e1c21cd2d4ff0f69e7176009c9d1b81f234ba866bf32192` |
| `v1.encoding.identity-width.001` | `f076851ad93e54a74688b0ba5e980f06e776a29be2d6172f1dc684115d23798b` |
| `v1.encoding.issuer-invalid-utf8.001` | `c13e16a11ea064607ec62dd0751f6f7a36150cfd834ca12481ef7a450c9ec64a` |
| `v1.encoding.issuer-too-long.001` | `0952a938c4ecfbad7c19d401d883202c8aa88ab295c37445928f84d84a660d43` |
| `v1.encoding.metadata-order.001` | `11354f239c8c43c25559a0e424de4211228a3b3527cc3be44b27b95d200908a3` |
| `v1.encoding.missing-secret.001` | `34c62ab540acc5e5a6888c712e09c28bce9830bc0efdba93f900a9f9a9febfb9` |
| `v1.encoding.non-token-plaintext.001` | `2bb906d9a319d624390ca720a7eddfcaef3e059fc4ac75f37e488620a720ab92` |
| `v1.encoding.parent-duplicate.001` | `f25965beb62471d62221e681ccfc4dce42d77251304ebf030fa2a74576884eab` |
| `v1.encoding.parents-1.001` | `7afb5abbf174ffcdfc5826cc8ac16bff02021bdc01e9a5f120996d42e77d79d4` |
| `v1.encoding.parents-2.001` | `7c1c72d920686539f6ac658d82717867d5a2815a4215d601239acb1506f3f499` |
| `v1.encoding.parents-3.001` | `e10040b5244d8400409b1f73cd4addc3fff007340af3dcfb9afc505e29149c1c` |
| `v1.encoding.parents-4.001` | `53083ad189f2e2bf3dba06d1848b07d6df2b3fa275149cec84fc36a6bac373bc` |
| `v1.encoding.period-zero.001` | `b7e68c46cc3e6d7c7a4b1df8e30d132e013adbb8c2cb8ddddef9e0058f714b3b` |
| `v1.encoding.secret-empty.001` | `399c3582b3367d7a64a75d89791aaaee46b6f93357498214fe1a237f91048f59` |
| `v1.encoding.secret-too-long.001` | `34e6a98064cfdfe6e7d14a9601dcb758603b046da82f67a4a168f05d5f57985e` |
| `v1.encoding.status-range.001` | `6f7307cdbae6926ac01af694d1e301c41ce4110fe99c9189f8db8d228f2819ca` |
| `v1.encoding.trailing-byte.001` | `8a378a8559c437c0d06018e9d7deb29d2347a77b1fffe2d50f8af89d9f3ce90b` |
| `v1.encoding.unknown-field.001` | `60c23ba0ddcea4c11b2e4c0cb5f0fd21863c8cd1805800bbe3b310f90332e77b` |
| `v1.fold.width-0.001` | `7f798f7171bb65f13ee754550997c89d6b51ac71093a7d0c74a172ecdd7931a1` |
| `v1.fold.width-10.001` | `11020f4c7f5ffa2df316affa909d02886f4dda2eda634e06f2f9b40a5b975b39` |
| `v1.fold.width-11.001` | `089fbecf28003ff0b6877f9ab0dab2a65b33950a913400fef6bb9ac532def55d` |
| `v1.fold.width-4.001` | `997d8cfd3841cca4002f169a959527f313b4da47ee28c668bb467495bc370554` |
| `v1.fold.width-5.001` | `e687aa9ccf26d1921a1415c20dae1ac442aa7006768e51a965641ff32b5bb0ce` |
| `v1.fold.width-7.001` | `00ee671a1acb0c100603a31ad7b9e725005a9827d19b280ba05410e3376d1df9` |
| `v1.graph.cycle-conflicting.001` | `1776c9d41d22391a6ca8d7c55010fd00216c973b0330658b2ea9c364570069d6` |
| `v1.graph.cycle-equal.001` | `277ee0bddd8843b728286856face217ab4d4db5cc77fc4d8123472019768301f` |
| `v1.graph.identical-collapse.001` | `b40f5e20c5569c9b2dc4916851b29215245d2675bba40f18386ce420d124fc88` |
| `v1.graph.late-cycle.001` | `77c6adb091513d3b9df43679313c9100b34075eb3b1e35e2d415f282811e01ae` |
| `v1.graph.same-id-defensive.001` | `ec715df198f0097570b45f018f046783d99cba430cb68936de4683bd1f0a5748` |
| `v1.graph.self-cycle.001` | `17e496222dc3e427ad1b49278c4a5bc1ec3d7a0e9dc242a9974567e680c73ea2` |
| `v1.metadata.client-name-absent.001` | `fc94b1cf7dd64265833958fe777179eb81e14ce9f809d748e1b21b855e91001e` |
| `v1.metadata.client-name-empty.001` | `077b83cccbb1482c5d3d66bafad64a8a8dc392ef883744136ae8294edbb187c0` |
| `v1.metadata.client-name-max.001` | `8c4fd39f39b0e59945c7e21c1c554c3278e173455ac5ba7ef806abaa9aee434e` |
| `v1.metadata.client-time-absent.001` | `14b4221b6dc00de4bf283c5695c92b32054a27e5f9896135a92b7b544c82fb8f` |
| `v1.metadata.client-time-normal.001` | `33e216d4ed94b777f62f766b73610fa197ce4c94b74ebd6bc09d0e32be84439f` |
| `v1.metadata.client-time-u64max.001` | `a8b22f77f418c10951155019c24b38b6604c9a0900619cc6dd81048a5ab37b8d` |
| `v1.metadata.client-time-zero.001` | `2523e7b7961faf69333d40d3efaa4321c2c4ac03e2072f105aa16b8c3358d1cd` |
| `v1.storage.ambiguous-publication.001` | `66fedf9b84c502699a31a532072e0aad7240c335c45867ac1833521be4be3c00` |
| `v1.storage.failed-aead.001` | `a6edd275a551a2f27ca9e7d1e5028c82b0e8351cfb48960d7440a0302f955e1f` |
| `v1.storage.incomplete-diagnostics.001` | `0c56bcf2b0234e6c80df0abc286ff941013b14efc0a07b2278103815a75f913c` |
| `v1.storage.invalid-grammar.001` | `7b19437516c6933628d695e54621ef7780951326c80bcc532940da2d384ccf77` |
| `v1.storage.missing-parent-publication.001` | `ac8e852457bacef8b41eb07768a8014fda266c5d59cebcb6240d94bcb1031a6e` |
| `v1.storage.new-publication.001` | `7896eda8040208180421ea2de58e82445a44a840ba19c05cd590273575b7c009` |
| `v1.storage.non-exact-publication.001` | `62cec110c07346670fb6a154618c064b33efa5257f8e8dc3b330693b8f2ce886` |
| `v1.storage.unreadable-target.001` | `14237ba381a8ac809378252a9d02312cc1c59cec4c2e61e1002991f13c051057` |
| `v1.storage.wrong-size.001` | `0de739abc2ff67cb2b3ab52c70a4b75df9c6904f00c1045ec8b693d80164dd30` |
| `v1.vault.create-ambiguous.001` | `1966f2d7146152aa490d8cbc8d46aa10c23ab82e5ba8b3ab959af7f96878cd28` |
| `v1.vault.create-empty.001` | `ab4e133f97d71f1363c00be91214a101d2697cec4bce992f90126d148f009468` |
| `v1.vault.create-existing.001` | `174a961b5302fee0079baa047a0eafe77b0105589d114b1dcbd47d05696ee22f` |
| `v1.vault.create-orphans.001` | `46c0cddae7bc4d7a8e84d9d93e40a5be4ad31f448202b565363409e3cf76dc0b` |
| `v1.vault.replace-ambiguous.001` | `d961390934f63226ec0c7c0d4812136d73f0efcc7740612b00fd2557aabed08e` |
| `v1.vault.replace-equal.001` | `f164864d0e9ffe7d97e16398f50589d8db8d06a080cd2a544f6cf33da4e113ab` |
| `v1.vault.replace-stale.001` | `3bd9f3df4e5a255e49abb3ff6f2aa905c2a8b7b282801b540fbd32ef91e388b1` |
| `v1.vault.replace-unreadable.001` | `5fed129f3ac86916e254a6d47206559e89cc6e1fad4b4bd7076def04d42b338e` |

## Active implementation and tooling changes

Deleted graph authorship permission/generation, provenance discovery, advertisement
publication, and recovery layers (`authorship.go`, `provenance.go`,
`provenance_test.go`, `publication.go`, `recovery.go`). Replaced graph evaluation
with cycle-safe reachability and maximal causal groups; folds use fixed capacity.

Deleted vector applicability/capability, local binding, provenance hardening,
recovery, and old storage graph adapters (`applicability.go`, `local.go`,
`hardening.go`, `recovery.go`, `storage.go`). Rewrote case types/schema and consumer
to remove their dead fields and branches. Removed the CLI capability selector and
`make conformance-all`; all manifest cases are now moving-profile requirements.

Removed cryptographic signing, verification, signature-input/public-key helpers,
signature context, and binding derivation. Retained root/object crypto and TOTP.
Replaced the object routing/version/type parser, nested credential encoding,
signature reservation, and value-dependent fan-in with the exact flattened parser.

Added an explicit reviewed r16 generator, independent Python wire/HMAC/HKDF checks,
exhaustive small-graph oracle checks, actual encrypted fold tests, and workflow tests.
The generator shares Go AEAD/Argon2 primitives with the consumer; independent Python
checks cover canonical TLV fields, HMAC/HKDF intermediates, fingerprint, and maximum
size, not independent AES/Argon2. Generator reproducibility was verified; normal
checks do not generate cases. Existing RFC answer files remain independent inputs.

Updated README, CONTRIBUTING, conformance/vector/requirements documentation, Makefile,
structural checker, schema tests, schemas, manifest, moving profile, and revision labels.

## Open issues, interpretations, and limits

No unresolved design ambiguity blocks the rewrite. Routine finalized choices are
consecutive TLV tags, the checkpoint's proposed fingerprint domain, and local
exclusion for exceptional same-ID failure. Canonical bootstrap pathname is exact
lowercase `vault`. Metadata is optional for an authoring operation; every fold stage
preserves the operation's exact CLIENT_NAME and CLIENT_TIME presence/values as well
as the same complete desired TokenValue. All head objects retain exact metadata,
including in equal-valued concurrent heads and SCC groups, while metadata stays
outside TokenValue equality. No new protocol semantic choice was required by the
correction pass; the Go snapshot/projection helpers are implementation details.

The model provides no filesystem adapter and does not claim crash-proof publication,
atomic CAS, remote completeness, or production resource scalability. The full UI
policy for truthful descriptions/disclosure is application work. No Java API,
platform module, future-family migration, or garbage collection design is selected.

## Validation

Baseline: `git status --short` empty; `git rev-parse HEAD` as above; `make check`,
`make verify`, and `make conformance` passed. Phase 1 hash and tracked-diff checks
passed before normative editing.

During implementation, the first full Go corpus test found generator pointer reuse:
input descriptions aliased a later TOKEN while semantic bytes were already captured.
Copying each fixture input fixed the mismatch; all regenerated fixtures then passed.
An initial preservation/whitespace audit found one extra EOF newline after historical
revision text. It was removed, history was rechecked byte-identical, and exact spec
pins were regenerated. These were corrected before final validation, not waived.

Final command results:

| Command | Result |
| --- | --- |
| `python3 tools/check_spec.py` | PASS |
| `go -C conformance test ./...` | PASS |
| `go -C conformance vet ./...` | PASS |
| `make check` | PASS |
| `make conformance` | PASS |
| `make verify` | PASS |
| `make race` | PASS |
| `make fuzz` | PASS |
| `git diff --check` | PASS |
| `command -v nix` | UNAVAILABLE — Nix-backed check not run |

`make check` includes the Python schema/independent-wire tests, all Go packages,
full conformance run, and exact pin verification. `make fuzz` runs FuzzDispatch,
FuzzOpen, and FuzzArrivalAndDisappearance for 10 seconds each with parallelism 2.
`gofmt -l conformance` produced no paths. Targeted parser/graph/crypto/storage tests
also passed during development. Final preservation checks passed for all previous
checkpoints, historical revision text, unchanged bootstrap wrapping bytes, and all
three RFC case files. Explicit generator rerun produced identical corpus/manifest/
profile hashes. Every manifest case exists, every physical case is listed exactly
once, every hash matches, and the moving profile requires exactly all current cases.

Nix is not installed/on PATH in this environment; the normal
`nix develop --command make check` could not be run. No Nix success is claimed.

### Focused legacy-reference audit

Searched every requested term: DEVICE, DEVICE_ID, P-256, ECDSA, DER, SIGNATURE,
AUTHOR_DEVICE_ID, AUTHOR_TIME, PROVENANCE, OPAQUE_ROUTABLE, OPAQUE_UNSCOPED,
OBJECT_TYPE, OBJECT_VERSION, VAULT_BINDING, READY, PROCESSING_INCOMPLETE,
HISTORY_REGRESSION, HISTORY_MEMORY_LOST, candidate-use, and current r15 references.
Reviewed surviving matches:

- `spec/`: only historical revision records contain retired concepts; raw substring
  `DER` also matches the valid `VAULT_HEADER` name. Active grammar/requirements are clean.
- `tools/check_spec.py`: retired names are a rejection list, not supported semantics.
- `tools/audit_r15.py`: unchanged, explicitly historical r14-to-r15 audit, excluded
  from active checks and intended for its historical checkout.
- `tools/audit_r16.py` and this report/audit: baseline revision and migration evidence.
- Publication code/tests/schema/case: raw `READY` matches `ALREADY_PRESENT_EXACT`,
  the required idempotent publication result, not a readiness state.
- Existing review reports/checkpoints: historical text preserved unchanged.

No current-revision r15 constant remains in active r16 consumer/generator/docs/checks.
Structural checks reject active retired terms before the revision-history boundary.

## Focused correction pass before adversarial review

This pass started from the existing uncommitted r16 rewrite at HEAD
`0a01fc1f0493ec1fefc0c4b89f8b22490d586852`, with 87 manifest cases. `git status --short` confirmed
the expected modified/deleted r16 files and untracked r16 additions, not a clean
tree or a new commit. Baseline `make check`, `make verify`, and `make conformance`
all passed before correction edits. The exact starting status is recorded below.

The two substantive corrections are:

1. Restore exact lowercase ASCII `vault` as the sole canonical bootstrap path;
   conceptual VAULT and identifiers such as VAULT_HEADER/VAULT_FINGERPRINT remain.
2. Preserve exact per-object metadata on heads/history and one operation's exact
   metadata through all fold stages. Presence/absence, present-empty CLIENT_NAME,
   raw validated UTF-8, zero, and full unsigned CLIENT_TIME values are preserved.
   Metadata remains excluded from TokenValue equality and all causal/conflict/use
   calculations. No synthetic merged metadata or event identity is introduced.

The Go graph model now returns complete head records alongside head IDs, preserves
metadata in retained historical records, and uses detached copies to prevent caller
mutation. A validated-object conversion carries metadata into graph records.
`FoldObjects` snapshots one operation and constructs each stage changing only its
parents. Existing equal-concurrent and cycle fixtures now assert different exact
per-head metadata, and all fold cases carry one complete operation as input.

Integration tests open real encrypted three-stage folds for named/time-stamped,
absent, present-empty/zero, absent-name/max-u64-time, and exact decomposed-UTF-8/
absent-time operations. They verify exact decoded metadata at every stage, equal
semantic values across differently attributed heads, retained historical metadata,
and unchanged TOTP. SCC tests include absent versus present-empty/zero. Mutation
and negative consumer tests detect dropped metadata or accidental aliasing.

All five checkpoint hashes were recorded before editing and verified unchanged
afterward (the exact values are in the design-input table). Bootstrap fixtures,
including all 87-byte representations and fingerprints, are byte-identical to the
pre-correction rewrite; all other encrypted TOKEN byte fixtures are unchanged too.
The preserved r15-and-earlier history and RFC TOTP files still match baseline.
The corpus count stays 87: no IDs added or removed by this correction pass.
The historical common-time case is now classified as replaced by the r16
operation-metadata fold case, rather than as simply removed.

That correction pass's required checks passed; the validation table above records
the latest final run. Generator reproducibility and exact profile/case hashes were
checked again. No new semantic choice was needed. This earlier pass preceded the
focused adversarial fixes documented below. Java work has not begun.

### Correction search audit

Reviewed active matches for VAULT/vault, CLIENT_NAME/CLIENT_TIME, fold, metadata,
head, and TokenValue. The spec layout, path rule, creation/replacement path mentions,
and README now use `vault`. Uppercase occurrences in headers/crypto/fingerprint,
conceptual representation prose, and rejection checks are intentional. Abstract
workflow cases do not encode a bootstrap pathname; their documented target is
`vault`, so no bootstrap bytes change. Historical checkpoints/reports are unchanged.
Per-head metadata is retained in graph Node/HeadObjects and validated-object
projection. Fold operation metadata is copied exactly; no active per-stage optional
metadata or refreshed-time rule remains. Metadata is still outside ValueBytes and
the semantic tuple. Schema and fixtures encode optional fields without collapsing
empty names or zero times. All relevant documentation now agrees.

### Exact correction-pass starting status

```text
 M CONTRIBUTING.md
 M Makefile
 M README.md
 M conformance/README.md
 M conformance/cmd/totipo-conformance/main.go
 M conformance/internal/cryptov1/crypto.go
 M conformance/internal/cryptov1/crypto_test.go
 D conformance/internal/graph/authorship.go
 M conformance/internal/graph/graph.go
 M conformance/internal/graph/graph_test.go
 D conformance/internal/graph/provenance.go
 D conformance/internal/graph/provenance_test.go
 D conformance/internal/graph/publication.go
 D conformance/internal/graph/recovery.go
 M conformance/internal/object/object.go
 M conformance/internal/object/object_test.go
 M conformance/internal/storage/publication.go
 M conformance/internal/storage/storage.go
 M conformance/internal/storage/storage_test.go
 M conformance/internal/totp/totp.go
 D conformance/internal/vectors/applicability.go
 D conformance/internal/vectors/hardening.go
 D conformance/internal/vectors/local.go
 D conformance/internal/vectors/recovery.go
 M conformance/internal/vectors/runner.go
 M conformance/internal/vectors/runner_test.go
 D conformance/internal/vectors/storage.go
 M conformance/internal/vectors/types.go
 M requirements/README.md
 M requirements/v1-pre-rc.json
 M spec/totipo-vault-format-v1.md
 M tools/check_spec.py
 M tools/test_check_vectors.py
 M vectors/FORMAT.md
 M vectors/README.md
 M vectors/case.schema.json
 M vectors/cases/bootstrap/v1.bootstrap.ascii.001.json
 M vectors/cases/bootstrap/v1.bootstrap.empty.001.json
 D vectors/cases/bootstrap/v1.bootstrap.local-binding-establishment.001.json
 M vectors/cases/bootstrap/v1.bootstrap.unicode.001.json
 D vectors/cases/candidate/v1.candidate.cache-write-warning.001.json
 D vectors/cases/candidate/v1.candidate.conflict-a.001.json
 D vectors/cases/candidate/v1.candidate.current-peer-missing.001.json
 D vectors/cases/candidate/v1.candidate.discovery-incomplete.001.json
 D vectors/cases/candidate/v1.candidate.historical-current-missing.001.json
 D vectors/cases/candidate/v1.candidate.history-memory-lost.001.json
 D vectors/cases/candidate/v1.candidate.opaque-current.001.json
 D vectors/cases/confirmation/v1.confirmation.relevant-change.001.json
 D vectors/cases/confirmation/v1.confirmation.relevant-information-return.001.json
 D vectors/cases/confirmation/v1.confirmation.unrelated-change.001.json
 D vectors/cases/confirmation/v1.confirmation.visible-conflict.001.json
 D vectors/cases/crypto/v1.crypto.device-root.001.json
 D vectors/cases/crypto/v1.crypto.future-device-opaque.001.json
 D vectors/cases/crypto/v1.crypto.future-token-opaque.001.json
 M vectors/cases/crypto/v1.crypto.token-child.001.json
 M vectors/cases/crypto/v1.crypto.token-root.001.json
 D vectors/cases/device/v1.device.explicit-id.001.json
 D vectors/cases/device/v1.device.id-mismatch.001.json
 D vectors/cases/device/v1.device.rename-incorporates-rejected-head.001.json
 D vectors/cases/encoding/v1.encoding.author-time-u64max.001.json
 D vectors/cases/encoding/v1.encoding.author-time-zero.001.json
 D vectors/cases/encoding/v1.encoding.device-root.001.json
 M vectors/cases/encoding/v1.encoding.duplicate-nonrepeatable.001.json
 M vectors/cases/encoding/v1.encoding.parent-count-mismatch.001.json
 M vectors/cases/encoding/v1.encoding.parent-order.001.json
 M vectors/cases/encoding/v1.encoding.token-root.001.json
 M vectors/cases/encoding/v1.encoding.utf8-boundary.001.json
 D vectors/cases/future/v1.future.concurrent-supported-opaque.001.json
 D vectors/cases/future/v1.future.device-presentation.001.json
 D vectors/cases/future/v1.future.disappearance-removes-current.001.json
 D vectors/cases/future/v1.future.family-compat-shadow.001.json
 D vectors/cases/future/v1.future.family-no-shadow.001.json
 D vectors/cases/future/v1.future.scoped-token.001.json
 D vectors/cases/future/v1.future.supported-descendant.001.json
 D vectors/cases/future/v1.future.unrelated-token.001.json
 D vectors/cases/future/v1.future.unscoped-candidate-use.001.json
 D vectors/cases/future/v1.future.unscoped-warning.001.json
 D vectors/cases/future/v1.future.version-does-not-order.001.json
 M vectors/cases/graph/v1.graph.conflicting-concurrent.001.json
 D vectors/cases/graph/v1.graph.corrupt-current-bytes-excluded.001.json
 D vectors/cases/graph/v1.graph.cycle-integrity-failure.001.json
 M vectors/cases/graph/v1.graph.equal-concurrent.001.json
 D vectors/cases/graph/v1.graph.global-object-id-conflict.001.json
 D vectors/cases/graph/v1.graph.history-cache-corrupt.001.json
 M vectors/cases/graph/v1.graph.intermediate-disappears.001.json
 D vectors/cases/graph/v1.graph.known-id-corrupt-bytes-warning.001.json
 M vectors/cases/graph/v1.graph.late-parent.001.json
 D vectors/cases/graph/v1.graph.reappearance-mismatch.001.json
 M vectors/cases/graph/v1.graph.reappearance.001.json
 D vectors/cases/graph/v1.graph.remembered-head-absent.001.json
 M vectors/cases/graph/v1.graph.sequential.001.json
 M vectors/cases/graph/v1.graph.wrong-identity-parent.001.json
 D vectors/cases/history/v1.history.cache-failure-after-publication-warning.001.json
 D vectors/cases/history/v1.history.current-child-missing-warning.001.json
 D vectors/cases/history/v1.history.current-peer-missing-warning.001.json
 D vectors/cases/history/v1.history.memory-lost.001.json
 D vectors/cases/history/v1.history.opaque-disappearance-warning.001.json
 D vectors/cases/history/v1.history.regression-and-return.001.json
 D vectors/cases/history/v1.history.unscoped-disappearance-warning.001.json
 D vectors/cases/provenance/v1.provenance.der-max-valid.001.json
 D vectors/cases/provenance/v1.provenance.der-short-valid.001.json
 D vectors/cases/provenance/v1.provenance.initial-device-before-token.001.json
 D vectors/cases/provenance/v1.provenance.late-device-reclassify.001.json
 D vectors/cases/provenance/v1.provenance.signature-context-cross-vault.001.json
 D vectors/cases/provenance/v1.provenance.token-rejected.001.json
 D vectors/cases/provenance/v1.provenance.token-unresolved.001.json
 D vectors/cases/provenance/v1.provenance.token-verified.001.json
 D vectors/cases/publication/v1.publication.ambiguous-retry-sibling.001.json
 D vectors/cases/publication/v1.publication.cache-failure-after-success.001.json
 D vectors/cases/routing/v1.routing.device-future-opaque.001.json
 D vectors/cases/routing/v1.routing.device-v1.001.json
 D vectors/cases/routing/v1.routing.future-device-malformed-prefix.001.json
 D vectors/cases/routing/v1.routing.future-token-malformed-prefix.001.json
 D vectors/cases/routing/v1.routing.token-future-opaque.001.json
 D vectors/cases/routing/v1.routing.token-v1.001.json
 D vectors/cases/routing/v1.routing.unknown-type-unscoped.001.json
 D vectors/cases/size/v1.size.device-max-14.001.json
 D vectors/cases/size/v1.size.device-max-15-fold.001.json
 D vectors/cases/size/v1.size.short-der-no-extra-parent.001.json
 M vectors/cases/size/v1.size.token-max-4.001.json
 D vectors/cases/size/v1.size.token-max-5-fold.001.json
 D vectors/cases/snapshot/v1.snapshot.hidden-history-after-publication.001.json
 D vectors/cases/snapshot/v1.snapshot.incomplete-usable.001.json
 D vectors/cases/snapshot/v1.snapshot.late-arrival-before-publication.001.json
 D vectors/cases/snapshot/v1.snapshot.no-advisory-history.001.json
 D vectors/cases/snapshot/v1.snapshot.opaque-token-plan-blocked.001.json
 D vectors/cases/snapshot/v1.snapshot.unrelated-change.001.json
 M vectors/cases/storage/v1.storage.exact-existing-publication.001.json
 M vectors/cases/storage/v1.storage.nonobject-name-ignored.001.json
 M vectors/cases/storage/v1.storage.objects-v1.001.json
 M vectors/cases/storage/v1.storage.unknown-sibling-ignored.001.json
 D vectors/cases/storage/v1.storage.wrong-size-not-opaque.001.json
 D vectors/cases/timestamp/v1.timestamp.equal-value-different-times.001.json
 D vectors/cases/timestamp/v1.timestamp.fold-common-time.001.json
 D vectors/cases/timestamp/v1.timestamp.i64max.001.json
 D vectors/cases/timestamp/v1.timestamp.normal.001.json
 D vectors/cases/timestamp/v1.timestamp.u64max.001.json
 D vectors/cases/timestamp/v1.timestamp.zero.001.json
 M vectors/manifest.json
 M vectors/manifest.schema.json
?? conformance/cmd/generate-vectors/
?? conformance/internal/vectors/integration_test.go
?? review/V1_R16_CASE_AUDIT.json
?? review/V1_R16_ENVELOPE_DECISION.md
?? review/V1_R16_FINAL_STATUS.txt
?? review/V1_R16_REWRITE_REPORT.md
?? tools/audit_r16.py
?? tools/test_r16_vectors.py
?? vectors/cases/bootstrap/v1.bootstrap.rewrap.001.json
?? vectors/cases/encoding/v1.encoding.account-too-long.001.json
?? vectors/cases/encoding/v1.encoding.algorithm-range.001.json
?? vectors/cases/encoding/v1.encoding.client-name-invalid-utf8.001.json
?? vectors/cases/encoding/v1.encoding.client-name-too-long.001.json
?? vectors/cases/encoding/v1.encoding.client-time-width.001.json
?? vectors/cases/encoding/v1.encoding.complete-tombstone.001.json
?? vectors/cases/encoding/v1.encoding.digits-range.001.json
?? vectors/cases/encoding/v1.encoding.five-parents.001.json
?? vectors/cases/encoding/v1.encoding.identity-width.001.json
?? vectors/cases/encoding/v1.encoding.issuer-invalid-utf8.001.json
?? vectors/cases/encoding/v1.encoding.issuer-too-long.001.json
?? vectors/cases/encoding/v1.encoding.metadata-order.001.json
?? vectors/cases/encoding/v1.encoding.missing-secret.001.json
?? vectors/cases/encoding/v1.encoding.non-token-plaintext.001.json
?? vectors/cases/encoding/v1.encoding.parent-duplicate.001.json
?? vectors/cases/encoding/v1.encoding.parents-1.001.json
?? vectors/cases/encoding/v1.encoding.parents-2.001.json
?? vectors/cases/encoding/v1.encoding.parents-3.001.json
?? vectors/cases/encoding/v1.encoding.parents-4.001.json
?? vectors/cases/encoding/v1.encoding.period-zero.001.json
?? vectors/cases/encoding/v1.encoding.secret-empty.001.json
?? vectors/cases/encoding/v1.encoding.secret-too-long.001.json
?? vectors/cases/encoding/v1.encoding.status-range.001.json
?? vectors/cases/encoding/v1.encoding.trailing-byte.001.json
?? vectors/cases/encoding/v1.encoding.unknown-field.001.json
?? vectors/cases/fold/
?? vectors/cases/graph/v1.graph.cycle-conflicting.001.json
?? vectors/cases/graph/v1.graph.cycle-equal.001.json
?? vectors/cases/graph/v1.graph.identical-collapse.001.json
?? vectors/cases/graph/v1.graph.late-cycle.001.json
?? vectors/cases/graph/v1.graph.same-id-defensive.001.json
?? vectors/cases/graph/v1.graph.self-cycle.001.json
?? vectors/cases/metadata/
?? vectors/cases/storage/v1.storage.ambiguous-publication.001.json
?? vectors/cases/storage/v1.storage.failed-aead.001.json
?? vectors/cases/storage/v1.storage.incomplete-diagnostics.001.json
?? vectors/cases/storage/v1.storage.invalid-grammar.001.json
?? vectors/cases/storage/v1.storage.missing-parent-publication.001.json
?? vectors/cases/storage/v1.storage.new-publication.001.json
?? vectors/cases/storage/v1.storage.non-exact-publication.001.json
?? vectors/cases/storage/v1.storage.unreadable-target.001.json
?? vectors/cases/storage/v1.storage.wrong-size.001.json
?? vectors/cases/vault/
```

### Case hashes changed by the correction pass

| Case ID | Before correction SHA-256 | After that correction SHA-256 |
| --- | --- | --- |
| `v1.fold.width-0.001` | `f96a8711646b7f388847e06382a70797f699a01e0338fd06b47bbdbcd557a451` | `7f798f7171bb65f13ee754550997c89d6b51ac71093a7d0c74a172ecdd7931a1` |
| `v1.fold.width-10.001` | `900006f475b5f23fcfd192d40f0ea0fdfe0563e391cfaba08bff355b0f645978` | `11020f4c7f5ffa2df316affa909d02886f4dda2eda634e06f2f9b40a5b975b39` |
| `v1.fold.width-11.001` | `b3e1028ab3d3cca791175e362fe9caada21ff6d7f11526bda74742742a3bb744` | `089fbecf28003ff0b6877f9ab0dab2a65b33950a913400fef6bb9ac532def55d` |
| `v1.fold.width-4.001` | `2171ae57849c3326fa77b61793f4b23d379f2eef93b72bbbb633a7c4d28f474f` | `997d8cfd3841cca4002f169a959527f313b4da47ee28c668bb467495bc370554` |
| `v1.fold.width-5.001` | `bcb9e462421b204282e00cf7691e30666401f3a7e5695e680d7ee89c9d469aef` | `e687aa9ccf26d1921a1415c20dae1ac442aa7006768e51a965641ff32b5bb0ce` |
| `v1.fold.width-7.001` | `7b196ee5eb6021bf9559ac8ea27d57e13a1db895fcada779a266052bd8281d58` | `00ee671a1acb0c100603a31ad7b9e725005a9827d19b280ba05410e3376d1df9` |
| `v1.graph.conflicting-concurrent.001` | `8b3b2cfd277571e91c184ab416e2eaca31e6cd5ec7015c1793dc4c9a5ac99797` | `87c00c3c78f4fa42804e0f619e160323b2ded1546a22b2f3ad5e645f311453c7` |
| `v1.graph.cycle-conflicting.001` | `19e44fbabd178fe3038220e3655d988b5c882f9f71783a5733007fcde1f7ddf9` | `1776c9d41d22391a6ca8d7c55010fd00216c973b0330658b2ea9c364570069d6` |
| `v1.graph.cycle-equal.001` | `0886072230e36d6a485961791ea053e1448368341324af81d5a854f8456f8978` | `277ee0bddd8843b728286856face217ab4d4db5cc77fc4d8123472019768301f` |
| `v1.graph.equal-concurrent.001` | `5f71851ba43471353504f7405b70eb31deadfa1597ed02b879caeb2e71f099d9` | `63350e320521772b4fab67b649832bbe270231d0d5be0c02ad10881bc2ed3520` |
| `v1.graph.identical-collapse.001` | `54fdbf8fac38d943bfc2e6c2a2897bba5227743979c09ac94c6c3d649c5f65b3` | `b40f5e20c5569c9b2dc4916851b29215245d2675bba40f18386ce420d124fc88` |
| `v1.graph.intermediate-disappears.001` | `802bf46ef8fb1eb1567a47379aa263d927f617c8f533aa584c59b96cb6bacd3a` | `bfedd33b499ffa3bce0de33379eea1f6091c913ba8e920e1d1cb4905f762ab5a` |
| `v1.graph.late-cycle.001` | `b3454c1694b28fd0f8a028a44bdecd8cfc4707f20ebe7dc4a59e8572b40600b8` | `77c6adb091513d3b9df43679313c9100b34075eb3b1e35e2d415f282811e01ae` |
| `v1.graph.late-parent.001` | `f55909738ffc014d0251a68fd7d2ce8b455157e9c5aac3dad4aca21c321e7963` | `680855e3210518e5d1a2e3a40af576e7eb3137003faa9ff72761c8a50aa88a04` |
| `v1.graph.reappearance.001` | `f6d8175afd5e0299614b8abfbb022dd3ce6174c8a369b2453704c8ab8aa91f8e` | `0082ecf1462cce4f0bc8e409b1f5d80574379bd22f22fb5ab302c9a11e82a61a` |
| `v1.graph.same-id-defensive.001` | `b4acd44129c6037b81dc90ebd847e5af66814f5244cd69055fc521177c11ed3e` | `ec715df198f0097570b45f018f046783d99cba430cb68936de4683bd1f0a5748` |
| `v1.graph.self-cycle.001` | `4c247cd05dafc2797dc03bbe1bb62118bacc4e30b0502269470d75ce61d00f51` | `17e496222dc3e427ad1b49278c4a5bc1ec3d7a0e9dc242a9974567e680c73ea2` |
| `v1.graph.sequential.001` | `666138ee2e4ea7f83b34d5761f594748d60a67c3d9d48e9f1617bca20abe6748` | `1a96ef6b78156004c75d35fd72a66318c83de8259d4ad406787a5aac95c652d7` |
| `v1.graph.wrong-identity-parent.001` | `80bb6fbc69f9fa552cc147e6d98f047ed8d9979fe6b990a677b277a08aaf0f8f` | `890fb23c7048f6f21aa041d50d887305cc82c80b1fd4b87d05a383328259de63` |

## Final adversarial correction pass

This focused pre-commit pass began at unchanged HEAD `0a01fc1f0493ec1fefc0c4b89f8b22490d586852`
in the existing uncommitted r16 rewrite. Its exact starting `git status --short`
was identical to the correction-pass starting status reproduced above. Baseline
`make check`, `make verify`, and `make conformance` all passed with 87 cases.
The five checkpoint hashes were recorded before edits and verified unchanged
again afterward; their exact values remain in the design-input table.

### Four findings addressed

1. Canonical observed entry types: lowercase `vault` must be a regular file before
   bootstrap interpretation, and exact `objects-v1` must be a directory before
   traversal. Symlinks and other wrong types are not deliberately followed,
   interpreted, or traversed. Absence permits the existing creation/lazy-creation
   behavior; a wrong-type existing entry is not absence. Replacement must re-observe
   a regular canonical `vault` for exact BASE comparison. No stable inode, hostile
   same-UID race proof, native filesystem API, or vault-wide gate was introduced.
   The model rejects wrong types before invoking a read callback and tests every
   listed special-entry class, missing namespace, and alternate bootstrap names.
2. Metadata lifetime: exact metadata preservation applies for as long as an object
   is represented, returned, exposed, or retained. Once the store no longer supplies
   it, the protocol requires no persistent retention of that object, its metadata,
   or its graph facts. Disposable caches remain disposable. Per-head/SCC metadata
   and exact common operation metadata across fold stages remain intact.
3. Required authoring inputs: history unavailability alone is not a prohibition when
   TOKEN_ID and complete desired TokenValue are known or supplied. Known unavailable
   parent IDs remain usable; no history contents or ancestry may be invented.
   New-token creation still generates a CSPRNG TOKEN_ID, supplies/generates the full
   value, and uses no parents. This adds no permission/readiness model.
4. Three required post-AEAD negative fixtures: authenticated nonzero padding,
   filename-derived context whose ID differs from HMAC(K_id, P), and authenticated
   semantic length 1007. All contain otherwise canonical TOKEN P. The generator
   uses AES-GCM directly only for malformed fixture construction; strict production
   Seal/Open APIs and their validation order are unchanged. No HMAC collision is
   simulated. The consumer proves raw authentication succeeds, verifies the isolated
   defect, then requires strict rejection with no semantic state. A broken tag or
   an ordinary valid object mislabeled as defective fails fixture verification.

The validation boundaries remain separate: wrong physical size, failed AEAD,
authenticated invalid length, authenticated nonzero padding, authenticated keyed-ID
mismatch, authenticated invalid grammar, and supported valid TOKEN. Existing
wrong-size, failed-AEAD, and grammar cases were retained unchanged.

Reread storage Sections 2, 3, 7, 8, 14, and 18 together, and metadata/state Sections
12 and 15–17 together. Observed-type constraints do not strengthen the local threat
boundary; metadata lifetime does not introduce remembered history. Ambiguous
publication/replacement, exact-existing publication, SCCs, value equality, and fixed
fold rules are unchanged. No additional protocol-design ambiguity was exposed.

### Pass baseline pins

| Artifact | Before this pass SHA-256 |
| --- | --- |
| `spec/totipo-vault-format-v1.md` | `1720eca1e709f8dd554944c39e84b49dc5f00aae2b2f0ca7aab078f1de96d4ba` |
| `vectors/manifest.json` | `53c4378a8bc2f5c2bb11fb2ac4eff150badcf739460a55bdc38486f06ecac5f6` |
| `requirements/v1-pre-rc.json` | `338a45ca19f613cebf7f490884abbf7e91e6403b5fcc9faac40edcf1d1daea18` |

Case count: **87 → 90**. **3 added**, no removed IDs, **4 existing case files changed**.
The four changed cases are replacement workflows whose already-existing `kind`
field now explicitly records `regular`. All 83 other previous cases are byte-identical,
including every bootstrap wrapping/fingerprint fixture, prior encrypted TOKEN
fixture, and RFC TOTP file. No grammar, domain, key, or bootstrap bytes changed.

| Added required case | SHA-256 |
| --- | --- |
| `v1.crypto.nonzero-padding.001` | `c88238486775495cb5b538eccbf5630b84a7db1b3dd36f08321f5e15b813eaa6` |
| `v1.crypto.object-id-mismatch.001` | `b95a6c07d355c79bb5bee38ae140b7ca877825276d2c198869f63a147ecfe199` |
| `v1.crypto.semantic-length-invalid.001` | `1dc9935f90536e7ae8e1766a1e641a4623c4b4fabc86eaebb0bad5f9d188ba67` |

### Baseline case IDs/hashes and final disposition for this pass

| Case ID | Before pass SHA-256 | Final disposition/hash |
| --- | --- | --- |
| `v1.bootstrap.ascii.001` | `2fa8fbd10507a9b8043ee335ab500844fd1e4d323afbe747bb6396a6b9ae0873` | unchanged |
| `v1.bootstrap.empty.001` | `b1b253e82bf434d1de0cc307e2c4bcddee3fdf4abe989330f733b19945d5b3c2` | unchanged |
| `v1.bootstrap.rewrap.001` | `88f337eadb06b8054dedcb2323da237feb932e725fae0ddc1488111474f4bcc3` | unchanged |
| `v1.bootstrap.unicode.001` | `84114af5be3292bd38d02b73fe6a5fcd552bba7cb6076eb289a685ea788e07e6` | unchanged |
| `v1.crypto.token-child.001` | `39a570a31f2a103b1a2604dec62c2428a584696303fd912eaec8f8b34de429ea` | unchanged |
| `v1.crypto.token-root.001` | `94d20ffa3e84b09500fa67d491c7cd50ecd53fc961e420fa496c335dab1c14ef` | unchanged |
| `v1.encoding.account-too-long.001` | `9224d0a78c7f76f47373fcc23d84f4cc2f93873e3fb6453ad54f999b4f944735` | unchanged |
| `v1.encoding.algorithm-range.001` | `48b5d49a95f6e3467bae0d8feda41fb717ca3e849b04ca104ff51322ee726a0b` | unchanged |
| `v1.encoding.client-name-invalid-utf8.001` | `8f3bd35ecc2987835873d574aa326d540a2a19ab887fcb81976d0286bb952120` | unchanged |
| `v1.encoding.client-name-too-long.001` | `20de7ce90f93d2c91cc8dc8f6e4c3baf4555477ed39f8145059e2924c46a5841` | unchanged |
| `v1.encoding.client-time-width.001` | `cfbb54b74b7633667ef626c7863532534e9810477135ad9b4889234b4f65d26a` | unchanged |
| `v1.encoding.complete-tombstone.001` | `7bbc696d5e50c00a409053798455225d1a493e06ad013c167e7fcda2082ce6e8` | unchanged |
| `v1.encoding.digits-range.001` | `67fccd5522def5ea2ddf7d9a9eaffe1379b92137b4268cbe313e3ab5b7b94db7` | unchanged |
| `v1.encoding.duplicate-nonrepeatable.001` | `1f7390e00ed6831b39b48e524dbfe3a54fa2cdd1dffd168eae66ec910f7637db` | unchanged |
| `v1.encoding.five-parents.001` | `1f782aca234d9a818e1c21cd2d4ff0f69e7176009c9d1b81f234ba866bf32192` | unchanged |
| `v1.encoding.identity-width.001` | `f076851ad93e54a74688b0ba5e980f06e776a29be2d6172f1dc684115d23798b` | unchanged |
| `v1.encoding.issuer-invalid-utf8.001` | `c13e16a11ea064607ec62dd0751f6f7a36150cfd834ca12481ef7a450c9ec64a` | unchanged |
| `v1.encoding.issuer-too-long.001` | `0952a938c4ecfbad7c19d401d883202c8aa88ab295c37445928f84d84a660d43` | unchanged |
| `v1.encoding.metadata-order.001` | `11354f239c8c43c25559a0e424de4211228a3b3527cc3be44b27b95d200908a3` | unchanged |
| `v1.encoding.missing-secret.001` | `34c62ab540acc5e5a6888c712e09c28bce9830bc0efdba93f900a9f9a9febfb9` | unchanged |
| `v1.encoding.non-token-plaintext.001` | `2bb906d9a319d624390ca720a7eddfcaef3e059fc4ac75f37e488620a720ab92` | unchanged |
| `v1.encoding.parent-count-mismatch.001` | `f1f33dc9e5c5c84ab9eaa95d949d7a20ce69092f7b016d1a3c5b295f53aa9f61` | unchanged |
| `v1.encoding.parent-duplicate.001` | `f25965beb62471d62221e681ccfc4dce42d77251304ebf030fa2a74576884eab` | unchanged |
| `v1.encoding.parent-order.001` | `d32f11e252ac1aa39b2603dffa1ceeb6f60a6234bdf44e8dfc8b3c31632bf3b2` | unchanged |
| `v1.encoding.parents-1.001` | `7afb5abbf174ffcdfc5826cc8ac16bff02021bdc01e9a5f120996d42e77d79d4` | unchanged |
| `v1.encoding.parents-2.001` | `7c1c72d920686539f6ac658d82717867d5a2815a4215d601239acb1506f3f499` | unchanged |
| `v1.encoding.parents-3.001` | `e10040b5244d8400409b1f73cd4addc3fff007340af3dcfb9afc505e29149c1c` | unchanged |
| `v1.encoding.parents-4.001` | `53083ad189f2e2bf3dba06d1848b07d6df2b3fa275149cec84fc36a6bac373bc` | unchanged |
| `v1.encoding.period-zero.001` | `b7e68c46cc3e6d7c7a4b1df8e30d132e013adbb8c2cb8ddddef9e0058f714b3b` | unchanged |
| `v1.encoding.secret-empty.001` | `399c3582b3367d7a64a75d89791aaaee46b6f93357498214fe1a237f91048f59` | unchanged |
| `v1.encoding.secret-too-long.001` | `34e6a98064cfdfe6e7d14a9601dcb758603b046da82f67a4a168f05d5f57985e` | unchanged |
| `v1.encoding.status-range.001` | `6f7307cdbae6926ac01af694d1e301c41ce4110fe99c9189f8db8d228f2819ca` | unchanged |
| `v1.encoding.token-root.001` | `7cccb7b1eb52761fab8025abeefd3680b213ea1d3e56e8419f61bdbc2d5a0cd6` | unchanged |
| `v1.encoding.trailing-byte.001` | `8a378a8559c437c0d06018e9d7deb29d2347a77b1fffe2d50f8af89d9f3ce90b` | unchanged |
| `v1.encoding.unknown-field.001` | `60c23ba0ddcea4c11b2e4c0cb5f0fd21863c8cd1805800bbe3b310f90332e77b` | unchanged |
| `v1.encoding.utf8-boundary.001` | `66f2392965142444cd422dc167c2929230872fdab4a083065d5a04465fd711e5` | unchanged |
| `v1.fold.width-0.001` | `7f798f7171bb65f13ee754550997c89d6b51ac71093a7d0c74a172ecdd7931a1` | unchanged |
| `v1.fold.width-10.001` | `11020f4c7f5ffa2df316affa909d02886f4dda2eda634e06f2f9b40a5b975b39` | unchanged |
| `v1.fold.width-11.001` | `089fbecf28003ff0b6877f9ab0dab2a65b33950a913400fef6bb9ac532def55d` | unchanged |
| `v1.fold.width-4.001` | `997d8cfd3841cca4002f169a959527f313b4da47ee28c668bb467495bc370554` | unchanged |
| `v1.fold.width-5.001` | `e687aa9ccf26d1921a1415c20dae1ac442aa7006768e51a965641ff32b5bb0ce` | unchanged |
| `v1.fold.width-7.001` | `00ee671a1acb0c100603a31ad7b9e725005a9827d19b280ba05410e3376d1df9` | unchanged |
| `v1.graph.conflicting-concurrent.001` | `87c00c3c78f4fa42804e0f619e160323b2ded1546a22b2f3ad5e645f311453c7` | unchanged |
| `v1.graph.cycle-conflicting.001` | `1776c9d41d22391a6ca8d7c55010fd00216c973b0330658b2ea9c364570069d6` | unchanged |
| `v1.graph.cycle-equal.001` | `277ee0bddd8843b728286856face217ab4d4db5cc77fc4d8123472019768301f` | unchanged |
| `v1.graph.equal-concurrent.001` | `63350e320521772b4fab67b649832bbe270231d0d5be0c02ad10881bc2ed3520` | unchanged |
| `v1.graph.identical-collapse.001` | `b40f5e20c5569c9b2dc4916851b29215245d2675bba40f18386ce420d124fc88` | unchanged |
| `v1.graph.intermediate-disappears.001` | `bfedd33b499ffa3bce0de33379eea1f6091c913ba8e920e1d1cb4905f762ab5a` | unchanged |
| `v1.graph.late-cycle.001` | `77c6adb091513d3b9df43679313c9100b34075eb3b1e35e2d415f282811e01ae` | unchanged |
| `v1.graph.late-parent.001` | `680855e3210518e5d1a2e3a40af576e7eb3137003faa9ff72761c8a50aa88a04` | unchanged |
| `v1.graph.reappearance.001` | `0082ecf1462cce4f0bc8e409b1f5d80574379bd22f22fb5ab302c9a11e82a61a` | unchanged |
| `v1.graph.same-id-defensive.001` | `ec715df198f0097570b45f018f046783d99cba430cb68936de4683bd1f0a5748` | unchanged |
| `v1.graph.self-cycle.001` | `17e496222dc3e427ad1b49278c4a5bc1ec3d7a0e9dc242a9974567e680c73ea2` | unchanged |
| `v1.graph.sequential.001` | `1a96ef6b78156004c75d35fd72a66318c83de8259d4ad406787a5aac95c652d7` | unchanged |
| `v1.graph.wrong-identity-parent.001` | `890fb23c7048f6f21aa041d50d887305cc82c80b1fd4b87d05a383328259de63` | unchanged |
| `v1.metadata.client-name-absent.001` | `fc94b1cf7dd64265833958fe777179eb81e14ce9f809d748e1b21b855e91001e` | unchanged |
| `v1.metadata.client-name-empty.001` | `077b83cccbb1482c5d3d66bafad64a8a8dc392ef883744136ae8294edbb187c0` | unchanged |
| `v1.metadata.client-name-max.001` | `8c4fd39f39b0e59945c7e21c1c554c3278e173455ac5ba7ef806abaa9aee434e` | unchanged |
| `v1.metadata.client-time-absent.001` | `14b4221b6dc00de4bf283c5695c92b32054a27e5f9896135a92b7b544c82fb8f` | unchanged |
| `v1.metadata.client-time-normal.001` | `33e216d4ed94b777f62f766b73610fa197ce4c94b74ebd6bc09d0e32be84439f` | unchanged |
| `v1.metadata.client-time-u64max.001` | `a8b22f77f418c10951155019c24b38b6604c9a0900619cc6dd81048a5ab37b8d` | unchanged |
| `v1.metadata.client-time-zero.001` | `2523e7b7961faf69333d40d3efaa4321c2c4ac03e2072f105aa16b8c3358d1cd` | unchanged |
| `v1.size.token-max-4.001` | `8ae31b9221c91dcd9717a6f2533d041453fcc157ecf53e358981b784e735ddde` | unchanged |
| `v1.storage.ambiguous-publication.001` | `66fedf9b84c502699a31a532072e0aad7240c335c45867ac1833521be4be3c00` | unchanged |
| `v1.storage.exact-existing-publication.001` | `fe9b4514d76cc767f256b82e2d509f44b3725c23694c89bfecf6da858c6478af` | unchanged |
| `v1.storage.failed-aead.001` | `a6edd275a551a2f27ca9e7d1e5028c82b0e8351cfb48960d7440a0302f955e1f` | unchanged |
| `v1.storage.incomplete-diagnostics.001` | `0c56bcf2b0234e6c80df0abc286ff941013b14efc0a07b2278103815a75f913c` | unchanged |
| `v1.storage.invalid-grammar.001` | `7b19437516c6933628d695e54621ef7780951326c80bcc532940da2d384ccf77` | unchanged |
| `v1.storage.missing-parent-publication.001` | `ac8e852457bacef8b41eb07768a8014fda266c5d59cebcb6240d94bcb1031a6e` | unchanged |
| `v1.storage.new-publication.001` | `7896eda8040208180421ea2de58e82445a44a840ba19c05cd590273575b7c009` | unchanged |
| `v1.storage.non-exact-publication.001` | `62cec110c07346670fb6a154618c064b33efa5257f8e8dc3b330693b8f2ce886` | unchanged |
| `v1.storage.nonobject-name-ignored.001` | `fc5eb0214a1a22b044cc3fb2c7c92a1cfbe40ba48123d7f7d4e446f4a93df609` | unchanged |
| `v1.storage.objects-v1.001` | `acba56ab7b85e8251daf4d850767a731ee4670121a5023a099962b79f49db767` | unchanged |
| `v1.storage.unknown-sibling-ignored.001` | `2322277808aa3dd0828a36858d19e58d9e1dbf3631105efaeb1498c8ff49a255` | unchanged |
| `v1.storage.unreadable-target.001` | `14237ba381a8ac809378252a9d02312cc1c59cec4c2e61e1002991f13c051057` | unchanged |
| `v1.storage.wrong-size.001` | `0de739abc2ff67cb2b3ab52c70a4b75df9c6904f00c1045ec8b693d80164dd30` | unchanged |
| `v1.totp.rfc6238-sha1.001` | `f3fed610df11e572f1c5c569e8822c5d1076df40197fadb559eca1a4aa0b6625` | unchanged |
| `v1.totp.rfc6238-sha256.001` | `eb9baec09844984cb7d7072164beb4d794ab9fb997428ca06ffc92d68ea5f25b` | unchanged |
| `v1.totp.rfc6238-sha512.001` | `c2787120d475f894371b490d37f5096fd255d6a36e385e20bf092bd7f148b07a` | unchanged |
| `v1.vault.create-ambiguous.001` | `1966f2d7146152aa490d8cbc8d46aa10c23ab82e5ba8b3ab959af7f96878cd28` | unchanged |
| `v1.vault.create-empty.001` | `ab4e133f97d71f1363c00be91214a101d2697cec4bce992f90126d148f009468` | unchanged |
| `v1.vault.create-existing.001` | `174a961b5302fee0079baa047a0eafe77b0105589d114b1dcbd47d05696ee22f` | unchanged |
| `v1.vault.create-orphans.001` | `46c0cddae7bc4d7a8e84d9d93e40a5be4ad31f448202b565363409e3cf76dc0b` | unchanged |
| `v1.vault.replace-ambiguous.001` | `5797e05edb205de4005c5c72e649677152e01de56d0d308f45c28b0836eec1f4` | `d961390934f63226ec0c7c0d4812136d73f0efcc7740612b00fd2557aabed08e` |
| `v1.vault.replace-equal.001` | `6c13df8a4ae24ccdfe32a5c32106827d401e10ea3547d3b150d253fca354ddfc` | `f164864d0e9ffe7d97e16398f50589d8db8d06a080cd2a544f6cf33da4e113ab` |
| `v1.vault.replace-stale.001` | `95bdec40d40a401eea40e9d47dd01f2ccdc059052ae25f77056a6b0f64892fb5` | `3bd9f3df4e5a255e49abb3ff6f2aa905c2a8b7b282801b540fbd32ef91e388b1` |
| `v1.vault.replace-unreadable.001` | `810fe6dde9eb48f95f82ceb5cc4f418ed1d318df5e9acd232ef0833ba120dba7` | `5fed129f3ac86916e254a6d47206559e89cc6e1fad4b4bd7076def04d42b338e` |

### Validation and focused search

The final command/result table above records the complete post-fix suite: spec
checker, Go tests/vet, make check/conformance/verify/race/fuzz, and whitespace.
Generator reruns reproduced exact case/manifest/profile bytes. Every manifest
file exists exactly once, hashes match, and moving pins match the final artifacts.
All five checkpoints, historical review files, r15-and-earlier revision suffix,
bootstrap wrapping fixtures, and RFC answers were verified unchanged. Nix remains
unavailable; no Nix-backed pass is claimed.

An intermediate generator build failed because an import edit also inserted import
strings into case-kind assignments. Those stray strings were removed; generation
then succeeded. The accompanying stale-spec-pin failure was from that aborted
generation and resolved with the normal successful generator run. No failure was
waived in the final suite.

Reviewed active occurrences of vault, objects-v1, symlink, regular file, directory,
CLIENT_NAME, CLIENT_TIME, retain/retained, history, cache, author/authorship,
unavailable, padding, OBJECT_ID, semantic length, and AEAD. Canonical entry-type
checks are observed-type checks only. Metadata exactness is scoped to represented
objects; no permanent-memory implication remains. Authorship still requires supplied
inputs but missing history is not a gate. Successful raw AEAD in fixture verification
is never an accepting alternate parser: strict length/padding/ID validation follows.
Historical checkpoints and older reports retain their original wording unchanged.

The repository is left uncommitted for final pre-commit review. No commit, push,
tag, release, RC freeze, or Java work was performed.

## Final exact pins

| Artifact | SHA-256 |
| --- | --- |
| `spec/totipo-vault-format-v1.md` | `96f362d68251a801b2c433bd6435c0056d5b070bfe1258a12ae12bdf7bb7990e` |
| `vectors/manifest.json` | `d03821fa2a0c41145bd8b82c6ea46800ab1a847b50c92fb9996b3c585d6f0a33` |
| `vectors/manifest.schema.json` | `3b54847e7a21bacc7956406c0d79dc172a0fb6415e9eb9c0335a511c1987664a` |
| `vectors/case.schema.json` | `d38618f53dcf0a558c389831e8838a07248066beff392812f3d277cc974e25c9` |
| `requirements/v1-pre-rc.json` | `687faf7757de7c63da68c28889bf3d0976b1858f9647b67569e3d8dfafc59c56` |

## Diff summary and final status

The complete modified/added/deleted file list below includes untracked files expanded
individually. `M` modifies, `D` deletes, and `??` adds an untracked reviewable file.
No files are staged. Ignored `.direnv/` scratch validation logs are not deliverables.

```text
 M CONTRIBUTING.md
 M Makefile
 M README.md
 M conformance/README.md
 M conformance/cmd/totipo-conformance/main.go
 M conformance/internal/cryptov1/crypto.go
 M conformance/internal/cryptov1/crypto_test.go
 D conformance/internal/graph/authorship.go
 M conformance/internal/graph/graph.go
 M conformance/internal/graph/graph_test.go
 D conformance/internal/graph/provenance.go
 D conformance/internal/graph/provenance_test.go
 D conformance/internal/graph/publication.go
 D conformance/internal/graph/recovery.go
 M conformance/internal/object/object.go
 M conformance/internal/object/object_test.go
 M conformance/internal/storage/publication.go
 M conformance/internal/storage/storage.go
 M conformance/internal/storage/storage_test.go
 M conformance/internal/totp/totp.go
 D conformance/internal/vectors/applicability.go
 D conformance/internal/vectors/hardening.go
 D conformance/internal/vectors/local.go
 D conformance/internal/vectors/recovery.go
 M conformance/internal/vectors/runner.go
 M conformance/internal/vectors/runner_test.go
 D conformance/internal/vectors/storage.go
 M conformance/internal/vectors/types.go
 M requirements/README.md
 M requirements/v1-pre-rc.json
 M spec/totipo-vault-format-v1.md
 M tools/check_spec.py
 M tools/test_check_vectors.py
 M vectors/FORMAT.md
 M vectors/README.md
 M vectors/case.schema.json
 M vectors/cases/bootstrap/v1.bootstrap.ascii.001.json
 M vectors/cases/bootstrap/v1.bootstrap.empty.001.json
 D vectors/cases/bootstrap/v1.bootstrap.local-binding-establishment.001.json
 M vectors/cases/bootstrap/v1.bootstrap.unicode.001.json
 D vectors/cases/candidate/v1.candidate.cache-write-warning.001.json
 D vectors/cases/candidate/v1.candidate.conflict-a.001.json
 D vectors/cases/candidate/v1.candidate.current-peer-missing.001.json
 D vectors/cases/candidate/v1.candidate.discovery-incomplete.001.json
 D vectors/cases/candidate/v1.candidate.historical-current-missing.001.json
 D vectors/cases/candidate/v1.candidate.history-memory-lost.001.json
 D vectors/cases/candidate/v1.candidate.opaque-current.001.json
 D vectors/cases/confirmation/v1.confirmation.relevant-change.001.json
 D vectors/cases/confirmation/v1.confirmation.relevant-information-return.001.json
 D vectors/cases/confirmation/v1.confirmation.unrelated-change.001.json
 D vectors/cases/confirmation/v1.confirmation.visible-conflict.001.json
 D vectors/cases/crypto/v1.crypto.device-root.001.json
 D vectors/cases/crypto/v1.crypto.future-device-opaque.001.json
 D vectors/cases/crypto/v1.crypto.future-token-opaque.001.json
 M vectors/cases/crypto/v1.crypto.token-child.001.json
 M vectors/cases/crypto/v1.crypto.token-root.001.json
 D vectors/cases/device/v1.device.explicit-id.001.json
 D vectors/cases/device/v1.device.id-mismatch.001.json
 D vectors/cases/device/v1.device.rename-incorporates-rejected-head.001.json
 D vectors/cases/encoding/v1.encoding.author-time-u64max.001.json
 D vectors/cases/encoding/v1.encoding.author-time-zero.001.json
 D vectors/cases/encoding/v1.encoding.device-root.001.json
 M vectors/cases/encoding/v1.encoding.duplicate-nonrepeatable.001.json
 M vectors/cases/encoding/v1.encoding.parent-count-mismatch.001.json
 M vectors/cases/encoding/v1.encoding.parent-order.001.json
 M vectors/cases/encoding/v1.encoding.token-root.001.json
 M vectors/cases/encoding/v1.encoding.utf8-boundary.001.json
 D vectors/cases/future/v1.future.concurrent-supported-opaque.001.json
 D vectors/cases/future/v1.future.device-presentation.001.json
 D vectors/cases/future/v1.future.disappearance-removes-current.001.json
 D vectors/cases/future/v1.future.family-compat-shadow.001.json
 D vectors/cases/future/v1.future.family-no-shadow.001.json
 D vectors/cases/future/v1.future.scoped-token.001.json
 D vectors/cases/future/v1.future.supported-descendant.001.json
 D vectors/cases/future/v1.future.unrelated-token.001.json
 D vectors/cases/future/v1.future.unscoped-candidate-use.001.json
 D vectors/cases/future/v1.future.unscoped-warning.001.json
 D vectors/cases/future/v1.future.version-does-not-order.001.json
 M vectors/cases/graph/v1.graph.conflicting-concurrent.001.json
 D vectors/cases/graph/v1.graph.corrupt-current-bytes-excluded.001.json
 D vectors/cases/graph/v1.graph.cycle-integrity-failure.001.json
 M vectors/cases/graph/v1.graph.equal-concurrent.001.json
 D vectors/cases/graph/v1.graph.global-object-id-conflict.001.json
 D vectors/cases/graph/v1.graph.history-cache-corrupt.001.json
 M vectors/cases/graph/v1.graph.intermediate-disappears.001.json
 D vectors/cases/graph/v1.graph.known-id-corrupt-bytes-warning.001.json
 M vectors/cases/graph/v1.graph.late-parent.001.json
 D vectors/cases/graph/v1.graph.reappearance-mismatch.001.json
 M vectors/cases/graph/v1.graph.reappearance.001.json
 D vectors/cases/graph/v1.graph.remembered-head-absent.001.json
 M vectors/cases/graph/v1.graph.sequential.001.json
 M vectors/cases/graph/v1.graph.wrong-identity-parent.001.json
 D vectors/cases/history/v1.history.cache-failure-after-publication-warning.001.json
 D vectors/cases/history/v1.history.current-child-missing-warning.001.json
 D vectors/cases/history/v1.history.current-peer-missing-warning.001.json
 D vectors/cases/history/v1.history.memory-lost.001.json
 D vectors/cases/history/v1.history.opaque-disappearance-warning.001.json
 D vectors/cases/history/v1.history.regression-and-return.001.json
 D vectors/cases/history/v1.history.unscoped-disappearance-warning.001.json
 D vectors/cases/provenance/v1.provenance.der-max-valid.001.json
 D vectors/cases/provenance/v1.provenance.der-short-valid.001.json
 D vectors/cases/provenance/v1.provenance.initial-device-before-token.001.json
 D vectors/cases/provenance/v1.provenance.late-device-reclassify.001.json
 D vectors/cases/provenance/v1.provenance.signature-context-cross-vault.001.json
 D vectors/cases/provenance/v1.provenance.token-rejected.001.json
 D vectors/cases/provenance/v1.provenance.token-unresolved.001.json
 D vectors/cases/provenance/v1.provenance.token-verified.001.json
 D vectors/cases/publication/v1.publication.ambiguous-retry-sibling.001.json
 D vectors/cases/publication/v1.publication.cache-failure-after-success.001.json
 D vectors/cases/routing/v1.routing.device-future-opaque.001.json
 D vectors/cases/routing/v1.routing.device-v1.001.json
 D vectors/cases/routing/v1.routing.future-device-malformed-prefix.001.json
 D vectors/cases/routing/v1.routing.future-token-malformed-prefix.001.json
 D vectors/cases/routing/v1.routing.token-future-opaque.001.json
 D vectors/cases/routing/v1.routing.token-v1.001.json
 D vectors/cases/routing/v1.routing.unknown-type-unscoped.001.json
 D vectors/cases/size/v1.size.device-max-14.001.json
 D vectors/cases/size/v1.size.device-max-15-fold.001.json
 D vectors/cases/size/v1.size.short-der-no-extra-parent.001.json
 M vectors/cases/size/v1.size.token-max-4.001.json
 D vectors/cases/size/v1.size.token-max-5-fold.001.json
 D vectors/cases/snapshot/v1.snapshot.hidden-history-after-publication.001.json
 D vectors/cases/snapshot/v1.snapshot.incomplete-usable.001.json
 D vectors/cases/snapshot/v1.snapshot.late-arrival-before-publication.001.json
 D vectors/cases/snapshot/v1.snapshot.no-advisory-history.001.json
 D vectors/cases/snapshot/v1.snapshot.opaque-token-plan-blocked.001.json
 D vectors/cases/snapshot/v1.snapshot.unrelated-change.001.json
 M vectors/cases/storage/v1.storage.exact-existing-publication.001.json
 M vectors/cases/storage/v1.storage.nonobject-name-ignored.001.json
 M vectors/cases/storage/v1.storage.objects-v1.001.json
 M vectors/cases/storage/v1.storage.unknown-sibling-ignored.001.json
 D vectors/cases/storage/v1.storage.wrong-size-not-opaque.001.json
 D vectors/cases/timestamp/v1.timestamp.equal-value-different-times.001.json
 D vectors/cases/timestamp/v1.timestamp.fold-common-time.001.json
 D vectors/cases/timestamp/v1.timestamp.i64max.001.json
 D vectors/cases/timestamp/v1.timestamp.normal.001.json
 D vectors/cases/timestamp/v1.timestamp.u64max.001.json
 D vectors/cases/timestamp/v1.timestamp.zero.001.json
 M vectors/manifest.json
 M vectors/manifest.schema.json
?? conformance/cmd/generate-vectors/main.go
?? conformance/internal/vectors/integration_test.go
?? conformance/internal/vectors/post_aead.go
?? review/V1_R16_CASE_AUDIT.json
?? review/V1_R16_ENVELOPE_DECISION.md
?? review/V1_R16_FINAL_STATUS.txt
?? review/V1_R16_REWRITE_REPORT.md
?? tools/audit_r16.py
?? tools/test_r16_vectors.py
?? vectors/cases/bootstrap/v1.bootstrap.rewrap.001.json
?? vectors/cases/crypto/v1.crypto.nonzero-padding.001.json
?? vectors/cases/crypto/v1.crypto.object-id-mismatch.001.json
?? vectors/cases/crypto/v1.crypto.semantic-length-invalid.001.json
?? vectors/cases/encoding/v1.encoding.account-too-long.001.json
?? vectors/cases/encoding/v1.encoding.algorithm-range.001.json
?? vectors/cases/encoding/v1.encoding.client-name-invalid-utf8.001.json
?? vectors/cases/encoding/v1.encoding.client-name-too-long.001.json
?? vectors/cases/encoding/v1.encoding.client-time-width.001.json
?? vectors/cases/encoding/v1.encoding.complete-tombstone.001.json
?? vectors/cases/encoding/v1.encoding.digits-range.001.json
?? vectors/cases/encoding/v1.encoding.five-parents.001.json
?? vectors/cases/encoding/v1.encoding.identity-width.001.json
?? vectors/cases/encoding/v1.encoding.issuer-invalid-utf8.001.json
?? vectors/cases/encoding/v1.encoding.issuer-too-long.001.json
?? vectors/cases/encoding/v1.encoding.metadata-order.001.json
?? vectors/cases/encoding/v1.encoding.missing-secret.001.json
?? vectors/cases/encoding/v1.encoding.non-token-plaintext.001.json
?? vectors/cases/encoding/v1.encoding.parent-duplicate.001.json
?? vectors/cases/encoding/v1.encoding.parents-1.001.json
?? vectors/cases/encoding/v1.encoding.parents-2.001.json
?? vectors/cases/encoding/v1.encoding.parents-3.001.json
?? vectors/cases/encoding/v1.encoding.parents-4.001.json
?? vectors/cases/encoding/v1.encoding.period-zero.001.json
?? vectors/cases/encoding/v1.encoding.secret-empty.001.json
?? vectors/cases/encoding/v1.encoding.secret-too-long.001.json
?? vectors/cases/encoding/v1.encoding.status-range.001.json
?? vectors/cases/encoding/v1.encoding.trailing-byte.001.json
?? vectors/cases/encoding/v1.encoding.unknown-field.001.json
?? vectors/cases/fold/v1.fold.width-0.001.json
?? vectors/cases/fold/v1.fold.width-10.001.json
?? vectors/cases/fold/v1.fold.width-11.001.json
?? vectors/cases/fold/v1.fold.width-4.001.json
?? vectors/cases/fold/v1.fold.width-5.001.json
?? vectors/cases/fold/v1.fold.width-7.001.json
?? vectors/cases/graph/v1.graph.cycle-conflicting.001.json
?? vectors/cases/graph/v1.graph.cycle-equal.001.json
?? vectors/cases/graph/v1.graph.identical-collapse.001.json
?? vectors/cases/graph/v1.graph.late-cycle.001.json
?? vectors/cases/graph/v1.graph.same-id-defensive.001.json
?? vectors/cases/graph/v1.graph.self-cycle.001.json
?? vectors/cases/metadata/v1.metadata.client-name-absent.001.json
?? vectors/cases/metadata/v1.metadata.client-name-empty.001.json
?? vectors/cases/metadata/v1.metadata.client-name-max.001.json
?? vectors/cases/metadata/v1.metadata.client-time-absent.001.json
?? vectors/cases/metadata/v1.metadata.client-time-normal.001.json
?? vectors/cases/metadata/v1.metadata.client-time-u64max.001.json
?? vectors/cases/metadata/v1.metadata.client-time-zero.001.json
?? vectors/cases/storage/v1.storage.ambiguous-publication.001.json
?? vectors/cases/storage/v1.storage.failed-aead.001.json
?? vectors/cases/storage/v1.storage.incomplete-diagnostics.001.json
?? vectors/cases/storage/v1.storage.invalid-grammar.001.json
?? vectors/cases/storage/v1.storage.missing-parent-publication.001.json
?? vectors/cases/storage/v1.storage.new-publication.001.json
?? vectors/cases/storage/v1.storage.non-exact-publication.001.json
?? vectors/cases/storage/v1.storage.unreadable-target.001.json
?? vectors/cases/storage/v1.storage.wrong-size.001.json
?? vectors/cases/vault/v1.vault.create-ambiguous.001.json
?? vectors/cases/vault/v1.vault.create-empty.001.json
?? vectors/cases/vault/v1.vault.create-existing.001.json
?? vectors/cases/vault/v1.vault.create-orphans.001.json
?? vectors/cases/vault/v1.vault.replace-ambiguous.001.json
?? vectors/cases/vault/v1.vault.replace-equal.001.json
?? vectors/cases/vault/v1.vault.replace-stale.001.json
?? vectors/cases/vault/v1.vault.replace-unreadable.001.json
```

The exact final `git status --short` output is also returned as
[`V1_R16_FINAL_STATUS.txt`](V1_R16_FINAL_STATUS.txt).

## Remaining work

> repin/reconcile `totipo-java` against r16 and evaluate what existing implementation architecture/code should be kept, changed, simplified, or deleted.

That Java work was not started. Review this uncommitted r16 rewrite before committing;
no commit, push, tag, release, or RC-profile freeze has been performed.
