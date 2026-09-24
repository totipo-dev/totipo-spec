# Totipo v0 conformance vectors

The protocol source of truth is [the r36 specification](../../spec/totipo-vault-format-v0.md). These JSON files are language-neutral test artifacts, not wire encodings or a replacement specification. New `spec-derived` cases require review before they can be described as reviewed publication evidence.

Each `.json` file contains `{ "schema": 1, "cases": [...] }`. Every case has:

- `id`: unique stable `v0/category/name`; array position has no meaning.
- `kind`: the operation below.
- `description`: the assertion being exercised.
- `provenance`: `source`, `status`, and `locator` within the source.
- `input` and `expected`: objects with string-valued members. Integers are canonical decimal strings, preserving arbitrary-size time values across languages. `expected.disposition` is required.
- `scenario`: present only for symbolic `model` cases, as described below.

Unknown members/kinds/schema versions, duplicate JSON keys, nulls, missing required members, duplicate IDs, and malformed hex are rejected. Object key order is immaterial. All byte fields end in `_hex` and contain lowercase, even-length hex without prefix or whitespace. Empty hex encodes empty bytes. The password operation's `value` also uses this rule when `encoding` is `utf8-hex`.

Provenance status vocabulary:

| Status | Meaning |
|---|---|
| `reviewed-pinned-vector` | Expectations copied from the named historical fixture or literal assertion; not recomputed by the implementation under test. |
| `published-standard` | RFC 6238 Appendix B values, with Appendix A's algorithm-specific secret lengths. Six/seven-digit expectations are the corresponding suffixes of the published eight-digit decimal values. |
| `spec-derived` | New manually specified boundary, mutation, dispatch, or symbolic expectations from the cited r36 sections; not externally reviewed. |

Operations and fields:

| Kind | Input members | Expected members beyond `disposition` |
|---|---|---|
| `tlv` | `bytes_hex`, `form` (`signed` or `unsigned`) | `unsigned_hex` required for accepted signed forms |
| `envelope` | `bytes_hex` | On success: `semantic_hex`, `semantic_disposition` |
| `file_length` | `length` | None; tests only the pre-AEAD length gate |
| `dispatch` | `bytes_hex` | None; input is assumed authenticated, unframed, identity-checked |
| `pipeline` | `root_hex`, `object_id_hex`, `file_hex` | `stages`: comma-separated attempted stages in order |
| `object_crypto` | `root_hex`, `signing_seed_hex`, `unsigned_hex`, `object_type` (`TOKEN_UPDATE` or `DEVICE_UPDATE`) | Every intermediate listed below |
| `bootstrap` | `root_hex`, `password_utf8_hex`, `argon2_salt_hex`, `wrap_nonce_hex` | `k_wrap_hex`, `header_aad_hex`, `wrapped_root_hex`, `wrap_tag_hex`, `vault_file_hex` |
| `bootstrap_check` | `file_hex`, `password_utf8_hex` | `kdf_calls`; optional `root_hex` |
| `password` | `encoding` (`utf8-hex` or `codepoints`), `value` | None; `codepoints` is space-separated base-16 signed integers, allowing invalid source characters to be tested without replacement |
| `totp` | `secret_hex`, `algorithm`, `digits`, `period`, `seconds` | `code` required on success |
| `model` | Empty object; see `scenario` | `VERIFIED` means the entire symbolic expected view matched |

Object-crypto expectations: `public_key_hex`, `prk_hex`, `k_id_hex`, `k_object_root_hex`, `k_signature_context_hex`, `signature_domain_hex`, `signature_input_hex`, `signature_hex`, `signed_plaintext_hex`, `object_id_hex`, `object_key_hex`, `nonce_hex`, `aad_hex`, `semantic_length_hex`, `envelope_plaintext_hex`, `ciphertext_hex`, `tag_hex`, `file_hex`. The complete file is an exact concatenation of the reviewed ciphertext and tag. Domain and semantic-length fields are literal/slice normalization, not new cryptographic expectations.

TOTP algorithms are `1` = SHA-1, `2` = SHA-256, `3` = SHA-512. Digits are 6/7/8. Periods use the complete unsigned 32-bit range except zero. `seconds` is an arbitrary-precision signed decimal string so rejection of negative time and counter overflow is testable without host-language narrowing.

Disposition vocabulary:

| Disposition | Meaning |
|---|---|
| `FILE_LENGTH_VALID` / `FILE_LENGTH_REJECTED` | Exact file size gate only |
| `AEAD_REJECTED` | Authentication/decryption failed |
| `ENVELOPE_VALID` / `ENVELOPE_REJECTED` | Authenticated plaintext framing only |
| `OBJECT_ID_REJECTED` | Recomputed identity does not match the supplied filename ID |
| `INVALID_PREFIX` | Complete canonical common prefix cannot be read |
| `AUTHENTICATED_UNSUPPORTED_FUTURE` | Nonzero version after a valid common prefix; tail is opaque |
| `STRUCTURALLY_VALID` / `STRUCTURALLY_INVALID` | Immutable v0 grammar result; no signature or history validity implied |
| `MAGIC_REJECTED` / `BOOTSTRAP_VERSION_REJECTED` | Bootstrap pre-KDF rejection |
| `PASSWORD_VALID` / `PASSWORD_REJECTED` | Strict UTF-8/source-character/byte-length domain result |
| `VERIFIED` | Operation's exact known-answer or full symbolic-output comparison succeeded; not a claim of complete protocol validity |
| `INVALID` | Invalid TOTP input, or failed deterministic symbolic semantic rule |
| `PENDING` / `FULLY_VALID` | Symbolic model dependency/history result, assuming completed identity/structure/signature checks |

Pipeline trace vocabulary is `FILE_LENGTH,AEAD,ENVELOPE,OBJECT_ID,DISPATCH`. A trace contains exactly the stages attempted, including the failing stage. It proves that later stages were not entered. Envelope lengths zero and one are accepted framing with rejected semantic grammar.

A `scenario` contains `updates` and `expected`. Each update has symbolic `id`, `token`, `parents` (array of symbolic IDs), and `fields` (map from `STATUS`, `ISSUER`, `ACCOUNT`, `CREDENTIAL` to complete symbolic values). STATUS uses `LIVE`/`TOMBSTONE`. These are post-authentication model inputs, not canonical TLV or cryptographic IDs. Expected output has `validation` by ID, `tokens` by token, and `resolutions`. Each token contains `heads`, typed `fields` head sets, and `conflicts`; each conflict explains `status`, `credential`, and all witnessing `transitions`. Set arrays are sorted lexicographically by symbolic ID; duplicates are not a second member. Credential strings are opaque labels, not actual secrets.

`manifest.sha256` covers every normative vector artifact: all `.json` files recursively below this directory. Markdown files are documentation/status, not vector expectations, and are excluded. Paths are relative to this directory, use `/`, and are lexicographically ordered. Lines are lowercase SHA-256, two spaces, path, newline. Missing, extra, unreadable, symlinked, or stale artifacts fail loading. Normal tests verify the manifest and never rewrite it.

Maintenance, from the repository root:

```sh
go run ./tools/normalize.go --write-vectors
```

The command refuses to run without the explicit flag and checks the pinned source hashes first. It copies reviewed values, creates documented spec-derived inputs/expectations, and writes the manifest. It never imports the Go conformance implementation. Review every resulting diff; running this command does not confer review status. CI never runs it. Historical sources remain unmodified.

## Phase 2 extensions

Existing fixtures and expectations are unchanged. New provenance statuses are `reviewed-pinned` (alias for historical pinned inputs) and `spec-derived-reviewed` (explicit review recorded in the cited review document). A passing implementation result alone does not confer review status. The Phase 2 reviews are author reviews with separate calculations, not external human audit claims. [The case index](../../review/phase2/CASE_INDEX.md) lists every new fixture and its class.

New kinds:

| Kind | Payload | Expectations |
|---|---|---|
| `ed25519` | `input.public_key_hex`, `message_hex`, `signature_hex` | `disposition`: ACCEPT/REJECT; `reason`: ACCEPT or the reviewed first rejection predicate |
| `memory` | `input.binding`, top-level `trace` | Each event's `expected.result` and optional exact checkpoint facts; outer disposition VERIFIED |
| `presentation` | Empty input; top-level `presentation.updates` and `presentation.expected` | Full validation/head/name view; outer disposition VERIFIED |
| `use_profile` | Boolean strings for degraded/incomplete/lifecycle/credential completeness/metadata conflicts; comma-separated status and credential value labels | ALLOW/BLOCK, `active`, `recovery_visible`, `metadata_conflict` |
| `blocked` | `input.requirement` | BLOCKED and `reason`; produces no passing case |

Ed25519 reason vocabulary: `A_LENGTH`, `SIGNATURE_LENGTH`, `A_DECODE`, `R_DECODE`, `A_CANONICAL`, `R_CANONICAL`, `A_SMALL_ORDER`, `R_SMALL_ORDER`, `A_SUBGROUP`, `R_SUBGROUP`, `S_RANGE`, `EQUATION`, `ACCEPT`. These are test diagnostics, not wire error codes.

Token scenarios optionally include signer labels (no causal meaning), `unavailable` IDs and `observed` frontier IDs. Unavailable inputs are excluded; an explicit observation frontier selects its available causal closure. Missing parents remain missing. Optional `coverage` maps accepted resolution IDs to all incorporated historical credential revision IDs. Both token evaluators must agree with this map. Absent optional filters mean all supplied updates are available/observed.

A memory event is `{ "op": "...", "args": { ... }, "expected": { "result": "...", ... } }`. Scope labels are `token:T` or `device:K`; object IDs are symbolic names without colons or commas. Parent/head lists are comma-separated IDs. `persist` selects exactly one staged record by `key`; it never commits a batch implicitly. `crash` drops staged records and confirmations but retains durable evidence and the simulated synchronized namespace.

Memory operations: `valid`, `lose`, `path_failure`, `restore`, `observe_pending`, `observe_future`, `abandon`, `acknowledge`, `bypass_future`, `remember`, `persist`, `compact_head`, `discard_pending`, `terminal`, `future_handoff`, `clear_future`, `resource_limit`, `context`, `confirm`, `publish`, `begin`, `install`, `existing_vault`, `establish`, `crash`, `migrate`. `confirm`/`publish` default their desired status to LIVE when omitted; explicit differing statuses stale confirmation.

Checkpoint queries include `binding`, `establishment`, `installed`, `open_allowed`, `staged_count`; exact-ID facts `pending:SCOPE:ID`, `abandoned:SCOPE:ID`, `acknowledged:SCOPE:ID`, `future:ID`, `bypass:ID`, `published:ID`; scope facts `writer:SCOPE`, `complete:SCOPE`, `degraded:SCOPE`, `heads:SCOPE`, `head_count:SCOPE`, `remembered:SCOPE`, `capacity:SCOPE`; and migration facts `migration:FIELD`. Booleans are `true`/`false`, sets are sorted comma-separated IDs, and omitted query assertions impose no constraint at that checkpoint. Expected event results distinguish awaiting durability, durable completion, rejections, compaction and linearization; their exact vocabulary is fixed by the trace files and documented operation implementations.

Phase 2 maintenance is separate from the original converter:

```sh
go run ./review/ed25519/review.go --write-vectors
go run ./tools/phase2/*.go --write-vectors
go run ./tools/manifest/main.go --write-manifest
```

These are explicit maintenance commands, not normal test steps. The last command hashes existing JSON without rewriting any expectation. Run the Ed25519 reviewer without the write flag for a read-only mathematical check. New review artifacts have a separate inventory; the original historical review inventory remains unchanged.

The current reviewed memory-result vocabulary is: `AWAITING_DURABILITY`, `BINDING_MISMATCH`, `BINDING_NOT_DURABLE`, `CAS_CONFLICT`, `COMPACTED`, `CONFIRMATION_STALE`, `CONFIRMED`, `DEGRADED_KNOWN_HISTORY_LOSS`, `DESTINATION_EXISTS`, `DURABLE`, `HANDOFF_REJECTED`, `INSTALLED`, `LINEARIZED`, `MIGRATED_COPY`, `NOT_IDENTITY_GROUNDED`, `OBSERVED`, `REPLACEMENT_NOT_DURABLE`, `RESTARTED`, `UNAVAILABLE`, `UNKNOWN_FUTURE`, `UNKNOWN_PENDING`, and `WRITER_BLOCKED`. These distinguish completed persistence from provisional staging and policy rejection; none is a wire field. Additional API input-error outcomes are implementation diagnostics, not passing normative trace expectations.
