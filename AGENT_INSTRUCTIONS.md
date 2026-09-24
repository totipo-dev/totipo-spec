# Totipo v0 Conformance Suite — Agent Instructions

You are working in the Totipo specification repository.

The current normative specification is:

```text
spec/totipo-vault-format-v0.md
```

It corresponds to Totipo Vault Format v0 revision 36.

The repository may also contain historical/review material under `review/`. Those files are provenance, test evidence, source vectors, or historical implementations. They are **not** normative unless the current spec explicitly says so.

## Core rules

1. Treat `spec/totipo-vault-format-v0.md` as the protocol source of truth.
2. Do **not** redesign the wire format or semantics while building the conformance suite.
3. Do **not** rename frozen protocol byte strings merely because the product is now named Totipo.
   - `TOTP-VAULT` remains the bootstrap magic.
   - `TOTP-Vault/v0/...` domain strings remain byte-for-byte frozen protocol constants.
4. If a reviewed vector disagrees with your implementation, fail the test. Do not silently regenerate expectations.
5. If two normative parts of the spec genuinely contradict each other, document the contradiction and leave the affected test failing or blocked rather than silently choosing an interpretation.
6. Keep the distinction between:
   - normative specification;
   - language-neutral conformance vectors;
   - Go conformance/reference implementation;
   - non-normative property/fuzz tests.
7. The Go implementation is **not** the definition of the protocol.

---

# 0. Expected starting repository state

The seed material should already have been extracted into the repository. The useful inputs are expected at roughly:

```text
spec/
  totipo-vault-format-v0.md

review/
  README.md
  source-vectors/
  source-code/
  reports/
```

Historical source-vector filenames may still begin with `totp-vault-...`. Preserve those historical filenames under `review/` for provenance.

New public artifacts should use Totipo naming.

Before doing any implementation work:

- verify `spec/totipo-vault-format-v0.md` exists;
- inventory everything under `review/source-vectors`, `review/source-code`, and `review/reports`;
- do not modify the original review/source artifacts in place.

If the seed ZIP was extracted into a nested directory such as `totipo-conformance-seed/`, merge its `spec/` and `review/` directories into the repository root before proceeding.

---

# 1. Goal

Build the first official Totipo v0 language-neutral conformance corpus and a Go implementation that consumes it independently.

The main milestone is:

> A Go implementation, written from the current Totipo v0 spec, consumes externally pinned reviewed vectors and reproduces exact canonical bytes and cryptographic intermediates without using those expected values as implementation logic.

The suite should eventually cover:

- canonical TLV grammar;
- deterministic semantic-object envelope;
- semantic-object key hierarchy and crypto;
- bootstrap record and Argon2id/AES-GCM wrapping;
- common-prefix future-version dispatch;
- TOTP behavior;
- Totipo's strict Ed25519 acceptance profile;
- semantic/lifecycle reference model;
- pending/future/rollback/recovery state transitions;
- publication and durability invariants;
- fuzz/property testing.

---

# 2. Create the repository layout

Create as needed:

```text
vectors/
  v0/
    README.md
    tlv/
    envelope/
    object-crypto/
    bootstrap/
    dispatch/
    totp/
    ed25519/
    transitions/
    lifecycle/
    recovery/
    manifest.sha256

conformance/
  go.mod
  README.md
  cmd/
    totipo-conformance/
      main.go
  internal/
    corpus/
    tlv/
    envelope/
    objectcrypto/
    bootstrap/
    dispatch/
    totp/
    ed25519profile/
    model/

fuzz/
  README.md
```

Do not create a second copy of the specification.

Do not move or rewrite historical `review/` material except where a separate normalized copy is deliberately created under `vectors/v0/`.

---

# 3. Define the language-neutral vector format

Use JSON for public conformance vectors.

JSON is only the test-artifact format; it is not part of the Totipo wire protocol.

Use a stable case structure where practical:

```json
{
  "id": "v0/tlv/token-create",
  "description": "...",
  "provenance": {
    "source": "review/source-vectors/...",
    "status": "reviewed-pinned-vector"
  },
  "input": {},
  "expected": {}
}
```

Rules:

- every case has a unique stable `id`;
- IDs must not depend on array ordering;
- hex byte strings are lowercase;
- no `0x` prefix;
- no whitespace in hex;
- even number of hex characters;
- do not depend on JSON object key order;
- explicitly document any enum vocabulary used by vector files.

For negative cases, record expected processing stage/disposition rather than only `valid: false`.

Use a vocabulary that maps cleanly onto the spec. Candidate labels include:

```text
FILE_LENGTH_REJECTED
AEAD_REJECTED
ENVELOPE_REJECTED
OBJECT_ID_REJECTED
INVALID_PREFIX
AUTHENTICATED_UNSUPPORTED_FUTURE
STRUCTURALLY_INVALID
SIGNATURE_INVALID
PENDING
FULLY_VALID
INVALID
```

Do not invent stage names that blur distinctions made by the spec.

Document the final schema/vocabulary in:

```text
vectors/v0/README.md
```

---

# 4. Normalize existing reviewed artifacts

Search `review/` for the already produced reviewed inputs. Expected examples include equivalents of:

```text
totp-vault-v0-r31-tlv-corpus.json
r31-object-vectors.txt
totp-vault-v0-r32-bootstrap-vectors.txt
```

and their reports/source implementations.

Do **not** regenerate reviewed expected values merely because conversion is easier.

Convert reviewed material into the new `vectors/v0/` schema while preserving exact expected bytes.

Every converted vector should record provenance back to the original review artifact.

Example:

```json
"provenance": {
  "source": "review/source-vectors/r31-object-vectors.txt",
  "status": "reviewed-pinned-vector"
}
```

Generated fuzz output and current Go results are never automatically promoted to reviewed expected data.

---

# 5. TLV corpus requirements

At minimum normalize/create reviewed cases for the canonical v0 TLV grammar.

## Positive cases

Include:

- TOKEN_UPDATE root/create;
- DEVICE_UPDATE root;
- ordinary token edit;
- credential update;
- lifecycle STATUS-only update;
- parent counts 0, 1, 2, and 32;
- maximum TOKEN_UPDATE;
- maximum DEVICE_UPDATE;
- valid empty ISSUER / ACCOUNT / DISPLAY_NAME;
- valid 256-byte human-facing strings;
- SECRET_BYTES length 1;
- SECRET_BYTES length 128;
- every valid STATUS value;
- every valid ALGORITHM value;
- DIGITS 6, 7, and 8;
- PERIOD boundary values already required by the reviewed corpus.

## Negative cases

Include:

- incomplete TLV header;
- declared TLV length larger than remaining bytes;
- trailing bytes;
- wrong fixed-width integer/key/signature values;
- unknown v0 tag;
- unknown v0 enum;
- duplicate singleton tag;
- out-of-order tags;
- misplaced/non-terminal SIGNATURE;
- missing required field;
- TOKEN_UPDATE with no semantic field assertion;
- forbidden type-specific tag;
- parent count mismatch;
- parent count 33;
- parent IDs of wrong size, including 31 and 33 bytes;
- duplicate parent IDs;
- non-increasing parent IDs;
- malformed UTF-8;
- 257-byte human-facing string;
- malformed nested CREDENTIAL;
- duplicate nested tag;
- out-of-order nested tag;
- unknown nested tag;
- nested trailing bytes;
- SECRET_BYTES length 0;
- SECRET_BYTES length 129;
- PERIOD zero.

The parser must reject noncanonical input, not repair or normalize it.

---

# 6. Envelope corpus requirements

Totipo v0 semantic-object framing is:

```text
SEMANTIC_LENGTH_U16BE[2]
|| P
|| ZERO_PADDING
```

with:

```text
authenticated encryption plaintext = 2032 bytes
semantic capacity                  = 2030 bytes
AES-GCM tag                        = 16 bytes
stored semantic object             = 2048 bytes
```

Include at minimum:

```text
semantic length 0
semantic length 1
maximum actual TOKEN_UPDATE length
maximum actual DEVICE_UPDATE length
generic semantic length 2030
semantic length 2031 -> reject
nonzero padding -> reject
wrong semantic-object file sizes -> reject before AEAD
```

Record the expected stage at which each failure occurs.

Do not parse semantic TLV before authenticated envelope validation and OBJECT_ID verification have succeeded.

---

# 7. Object-crypto vectors

For at least one TOKEN_UPDATE and one DEVICE_UPDATE, publish/normalize exact pinned values for:

```text
K_root
PRK
K_id
K_object_root
K_signature_context

canonical_unsigned
signature domain/context input
signature_input
device public key
signature

canonical_signed
OBJECT_ID
K_object
nonce
AAD

SEMANTIC_LENGTH_U16BE
envelope_plaintext
ciphertext
GCM tag
complete 2048-byte semantic-object file
```

Prefer the existing reviewed/pinned artifact under `review/source-vectors/` as provenance.

The Go tests must reproduce every intermediate byte exactly.

A decrypt(encrypt(x)) round trip alone is insufficient.

---

# 8. Bootstrap vectors

Normalize both reviewed bootstrap cases:

```text
INITIAL
REWRAP
```

Pin at least:

```text
password UTF-8 bytes
ARGON2_SALT
WRAP_NONCE
K_wrap
K_root
VAULT_HEADER
WRAPPED_ROOT
WRAP_TAG
complete 87-byte VAULT record
```

Totipo v0 bootstrap layout is fixed and non-TLV:

```text
offset 0   size 10  MAGIC = ASCII("TOTP-VAULT")
offset 10  size 1   BOOTSTRAP_VERSION = 0x00
offset 11  size 16  ARGON2_SALT
offset 27  size 12  WRAP_NONCE
offset 39  size 32  WRAPPED_ROOT
offset 71  size 16  WRAP_TAG

total = 87 bytes
AAD   = bytes 0..38
```

Do not rename `TOTP-VAULT`.

Include negative cases for:

- wrong total lengths;
- wrong magic;
- unsupported bootstrap version;
- password >1024 encoded UTF-8 bytes;
- malformed pre-encoded UTF-8;
- character-to-UTF-8 encoding failure;
- empty password accepted;
- exactly 1024 bytes of well-formed UTF-8 accepted;
- wrong password;
- mutation of salt;
- mutation of nonce;
- mutation of wrapped ciphertext;
- mutation of authentication tag;
- wrong AAD boundary where reviewed evidence exists.

Framing/version/password-domain failures must be rejected before Argon2id.

---

# 9. Common-prefix and future-version dispatch corpus

Create/normalize cases proving the exact dispatch ordering.

Requirements:

1. semantic-object file framing/AEAD/envelope/OBJECT_ID checks happen before semantic dispatch;
2. an incomplete or malformed canonical common prefix is invalid framing;
3. a malformed prefix cannot become AUTHENTICATED_UNSUPPORTED_FUTURE;
4. OBJECT_VERSION=0 enters the full immutable v0 grammar;
5. OBJECT_VERSION!=0 becomes AUTHENTICATED_UNSUPPORTED_FUTURE after a valid common prefix;
6. after nonzero version dispatch, the remaining authenticated payload is opaque to v0;
7. unknown OBJECT_TYPE under version 0 is INVALID;
8. unknown future OBJECT_TYPE is irrelevant to a v0 reader after nonzero-version dispatch.

Do not accidentally parse an opaque future tail using the v0 grammar.

---

# 10. Manifest

Generate:

```text
vectors/v0/manifest.sha256
```

The manifest must include every committed normative vector file beneath `vectors/v0/`, excluding the manifest itself.

Rules:

- SHA-256;
- lexicographic path order;
- stable formatting;
- normal test runs verify but never rewrite it.

If vector-writing tooling exists, it must require an explicit maintenance command such as:

```text
--write-vectors
```

CI must never regenerate expected vectors automatically.

---

# 11. Go module and dependencies

Implement the official conformance runner under `conformance/`.

Prefer the Go standard library.

Initially acceptable external dependency:

```text
golang.org/x/crypto/argon2
```

Pin dependency versions.

Do not add general serialization, assertion, test-framework, or crypto dependencies without a concrete reason documented in `conformance/README.md`.

The following must pass:

```bash
go fmt ./...
go vet ./...
go test ./...
```

If practical, also run with the race detector for stateful/model tests:

```bash
go test -race ./...
```

---

# 12. Corpus loader

Implement:

```text
conformance/internal/corpus
```

Responsibilities:

- recursively load JSON vector files;
- validate artifact schema;
- decode strict lowercase hex;
- reject malformed artifacts;
- expose stable case IDs;
- detect duplicate IDs;
- verify `vectors/v0/manifest.sha256`;
- report provenance in failures where useful.

Tests must fail for:

- duplicate IDs;
- malformed hex;
- malformed JSON/schema;
- unreadable vector files;
- stale manifest.

The corpus loader should not implement Totipo protocol semantics beyond validating the artifact format itself.

---

# 13. Independent Go TLV parser

Implement the parser independently from the historical Python/Java review code.

Do not port those implementations line-by-line.

Use the current spec as the grammar source.

Implement at least:

```text
TAG:u16be
LENGTH:u16be
bounded value extraction
canonical numeric tag ordering
special repeated 0x0005 parent rule
fixed-width validation
strict well-formed UTF-8
v0 object-type-specific allowed tags
nested CREDENTIAL grammar
SIGNATURE = 0xFF01 terminal rule
parent count validation
strict parent raw-ID ordering
all frozen field/count limits
```

Never silently canonicalize malformed bytes.

For positive signed semantic objects:

- parse the signed bytes;
- independently remove the complete final `FF01 SIGNATURE` TLV;
- assert byte equality with the pinned `canonical_unsigned` expected value.

The implementation should report enough stage information for negative corpus assertions.

---

# 14. Envelope implementation

Implement deterministic framing/unframing.

Unframing requirements:

1. input authenticated plaintext is exactly 2032 bytes;
2. first two bytes decode `SEMANTIC_LENGTH_U16BE`;
3. semantic length must be <=2030;
4. extract exactly that many bytes;
5. all remaining bytes must be zero.

Test all envelope corpus cases.

Do not invoke semantic TLV parsing before envelope validation succeeds.

---

# 15. Key hierarchy and semantic-object crypto

Implement the frozen Totipo v0 key hierarchy exactly.

Do not substitute an API with merely similar semantics without verifying exact inputs/outputs.

Implement:

```text
K_root -> PRK
PRK -> K_id
PRK -> K_object_root
PRK -> K_signature_context

OBJECT_ID = HMAC-SHA-256(K_id, canonical signed plaintext)

K_object = HKDF-Expand(
    K_object_root,
    ASCII(frozen object-key domain) || OBJECT_ID,
    32
)

nonce = first 12 bytes of OBJECT_ID

AAD = ASCII(frozen object domain) || OBJECT_ID
```

The literal domain strings must come from the current spec and remain byte-for-byte unchanged.

AES-256-GCM encrypts the exact deterministic 2032-byte envelope.

For every pinned object-crypto fixture, require exact equality for every published intermediate value.

Also test stage ordering using hooks/counters if necessary:

```text
wrong file length -> reject before AEAD
AEAD failure -> no envelope/TLV parse
bad semantic length/padding -> no TLV parse
OBJECT_ID mismatch -> no TLV parse
```

---

# 16. Bootstrap codec

Implement a fixed-record bootstrap codec from the current spec.

Password processing requirements:

- strict well-formed UTF-8;
- no Unicode normalization;
- no replacement on encoding error;
- encoded byte length 0..1024 inclusive;
- malformed pre-encoded UTF-8 rejected;
- invalid source character encoding rejected;
- pre-KDF failures must not invoke Argon2id.

Use the exact frozen Argon2id parameters and AES-GCM AAD boundary.

Consume the external INITIAL and REWRAP vectors and require exact equality for:

```text
K_wrap
VAULT_HEADER
WRAPPED_ROOT
WRAP_TAG
complete VAULT bytes
unwrapped K_root
```

Add test hooks/counters proving pre-KDF rejection ordering.

---

# 17. TOTP implementation and vectors

Implement exactly the Totipo v0 TOTP profile:

```text
T0 = 0
PERIOD = uint32 in 1..2^32-1
DIGITS = 6, 7, or 8
ALGORITHM = SHA-1, SHA-256, or SHA-512 only
counter = floor(nonnegative Unix seconds / PERIOD)
counter encoded unsigned 8-byte big-endian
counter must fit 0..2^64-1
RFC 4226 dynamic truncation
decimal reduction
left-zero padding
```

Test at minimum:

- every algorithm;
- every supported digit count;
- period 1;
- period 30;
- period 2^31;
- period 2^32-1;
- period boundaries/start/end behavior where reviewed vectors exist;
- negative time rejection;
- zero period rejection;
- period >2^32-1 rejection;
- counter 2^64-1 accepted;
- counter overflow rejected.

Consume existing reviewed expected TOTP values where available instead of replacing them with Go-generated expectations.

---

# 18. Strict Ed25519 profile

This area requires special care.

Do **not** assume Go's `crypto/ed25519.Verify` proves Totipo conformance.

Totipo's acceptance profile includes explicit point/canonical/subgroup requirements beyond merely calling a generic verifier.

Tasks:

1. locate/import the reviewed positive/negative Ed25519 corpus if available;
2. run Go's standard verifier against it for diagnostic comparison;
3. document exactly which cases pass/fail;
4. implement `ed25519profile.Verify` that satisfies the full Totipo acceptance corpus;
5. do not weaken the spec to fit the library.

If a lower-level elliptic-curve primitive is required for canonical/subgroup checks, add the smallest reasonable dependency and document why.

Do not claim strict Ed25519 conformance until the complete corpus passes, including mixed-order rejection cases.

If the reviewed strict-Ed25519 corpus is not present in `review/`, mark that as an explicit blocked deliverable rather than inventing a substitute and calling it official.

---

# 19. CLI conformance runner

Implement:

```bash
go run ./conformance/cmd/totipo-conformance ./vectors/v0
```

Expected style:

```text
PASS v0/tlv/token-create
PASS v0/bootstrap/initial
PASS v0/object-crypto/token-root
FAIL v0/ed25519/mixed-order-07
     expected: INVALID
     actual:   FULLY_VALID
```

Requirements:

- stable, concise output;
- nonzero exit status on failure;
- optional case-prefix filter, e.g.:

```bash
... --filter v0/bootstrap/
```

Do not include vector regeneration in the normal runner.

---

# 20. Prove cross-implementation independence

Add tests/documentation demonstrating that Go is consuming externally pinned expectations.

At minimum prove:

- Go parses historical external canonical TLV fixtures;
- Go derives the expected unsigned representation from external signed plaintext;
- Go reproduces external object-crypto intermediates exactly;
- Go consumes the external INITIAL/REWRAP bootstrap fixtures exactly;
- expected outputs are not generated by the Go code under test during normal runs.

Document this in:

```text
conformance/README.md
```

---

# 21. Literal semantic/lifecycle reference model

After deterministic byte/crypto consumption is working, implement a deliberately simple reference model under:

```text
conformance/internal/model
```

This model must follow the **current r36 spec**, not the historical lifecycle oracle source verbatim.

Historical files such as `LifecycleOracle.java`, if present, are methodology/reference evidence only unless they exactly match r36 semantics.

The current reference model should prioritize transparency over performance:

- explicit sets/maps;
- direct ancestry traversal;
- no hidden caches required for correctness;
- deterministic outputs;
- explainable conflict witnesses.

It should eventually produce/validate:

- token-update causal frontier;
- typed field heads;
- JOIN behavior;
- lifecycle transitions;
- historical credential witness selection;
- maximal unresolved witnesses;
- lifecycle-resolution provisional coverage;
- lifecycle-resolution final coverage;
- conflict tuples/witness explanations;
- writer eligibility gates where modeled.

Do not optimize before the literal model is demonstrably correct.

---

# 22. Transition/recovery vectors

Create language-neutral fixtures for stateful behavior after the byte/crypto layers are stable.

Useful fixture shape:

```json
{
  "id": "v0/recovery/pending-abandonment-reentry",
  "initial": {},
  "events": [],
  "expected": {
    "validation": {},
    "token_heads": [],
    "field_heads": {},
    "lifecycle_conflicts": [],
    "durable_security_evidence": {},
    "writer_allowed": false
  }
}
```

Cover the mandatory cases from Section 74, including at least:

- ordinary field composition/conflict behavior;
- lifecycle witness persistence;
- status-only lifecycle resolution;
- stale/hidden witness reappearance;
- pending TOKEN_UPDATE durable evidence;
- pending abandonment and later re-entry;
- unsupported-future durable evidence and bypass;
- known-history regression;
- remembered-head durable replacement ordering;
- pending DEVICE_UPDATE abandonment/re-entry;
- establishment-state race behavior;
- publication linearization/freshness gates;
- capacity-blocked >32-head operations;
- migration selection from capacity-blocked source state.

Do not infer missing behavior from old review code if r36 specifies something different.

---

# 23. Fuzzing

After deterministic conformance passes, add Go fuzz tests.

At minimum:

```text
FuzzTLVFraming
FuzzV0Grammar
FuzzCredentialGrammar
FuzzEnvelopeUnframe
FuzzCommonPrefixDispatch
```

Seed fuzzers from committed positive and negative vectors.

Properties:

- parser never panics;
- invalid bytes are never silently canonicalized;
- accepted v0 bytes reserialize identically if an encoder exists;
- envelope unframing never exposes bytes outside the authenticated declared semantic range;
- future-version opaque tails are never parsed as v0 grammar;
- fuzzing must not rewrite normative vectors.

A fuzz discovery must be minimized and reviewed before being promoted into `vectors/v0/`.

---

# 24. CI expectations

Add CI after the base suite works.

At minimum run:

```bash
go fmt -w/check as appropriate
go vet ./...
go test ./...
```

Prefer testing on:

- Linux;
- macOS;
- Windows;
- at least two supported Go versions where practical.

Do not let CI fetch or regenerate normative vector expectations from an implementation under test.

---

# 25. Deliverables

Do not stop at scaffolding.

Deliver as much of the following as the available reviewed inputs permit:

1. `vectors/v0/README.md` with schema and disposition vocabulary;
2. normalized TLV vectors;
3. normalized envelope vectors;
4. normalized object-crypto vectors;
5. normalized bootstrap vectors;
6. common-prefix dispatch vectors;
7. TOTP vectors from reviewed inputs;
8. Ed25519 corpus import/status;
9. `vectors/v0/manifest.sha256`;
10. Go corpus loader;
11. independent Go TLV parser;
12. envelope implementation;
13. key hierarchy/object-crypto reproduction;
14. bootstrap codec and vector consumption;
15. TOTP implementation and vector consumption;
16. strict-Ed25519 status/implementation to the extent reviewed corpus permits;
17. CLI conformance runner;
18. deterministic Go tests;
19. initial fuzz targets;
20. `conformance/README.md`;
21. current-r36 semantic/lifecycle reference model, if time permits after byte/crypto conformance is complete;
22. transition/recovery fixtures where source evidence is available;
23. final implementation report.

---

# 26. Final report

Produce a report containing at least:

```text
- vector files created
- case counts by category
- PASS / FAIL / BLOCKED counts
- commands used to run the suite
- Go version(s)
- external dependencies and exact versions
- whether every externally pinned object-crypto intermediate reproduced exactly
- whether both bootstrap vectors reproduced exactly
- TOTP vector status
- strict Ed25519 profile status
- lifecycle/reference-model status
- fuzz targets added
- any reviewed artifact that could not be normalized
- any vector whose provenance could not be established
- any spec ambiguity or contradiction found
- any historical corpus that was referenced by the spec/reviews but not present in the repository
```

Use `BLOCKED` rather than fabricating expected results when required reviewed source material is absent.

---

# 27. Contradiction handling

Do not modify the Totipo spec just to make the suite pass.

If you find a genuine normative contradiction, record:

```text
section A
section B
minimal reproducer
why both requirements cannot be satisfied
current test result
proposed resolution
```

Then stop only the affected portion and continue independent work where possible.

Do not silently resolve the contradiction in code.

---

# 28. Priority order

Work in this order unless a hard dependency requires otherwise:

1. inventory reviewed inputs;
2. vector schema + normalized corpus;
3. manifest;
4. Go corpus loader;
5. TLV parser;
6. envelope parser;
7. object-crypto reproduction;
8. bootstrap reproduction;
9. common-prefix dispatch;
10. TOTP;
11. strict Ed25519 corpus/status;
12. CLI runner;
13. independence documentation;
14. literal current-r36 semantic/lifecycle model;
15. transition/recovery vectors;
16. fuzzing;
17. CI;
18. final report.

The first major checkpoint is complete when:

> `go test ./...` independently consumes the reviewed TLV, object-crypto, bootstrap, dispatch, and available TOTP/Ed25519 vectors and reproduces every pinned expected byte/result without regenerating expectations.
