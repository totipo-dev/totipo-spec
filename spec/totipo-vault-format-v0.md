# Totipo Vault Format v0

**Status:** Byte-level-freeze candidate, revision 36  
**Scope:** State model, synchronization semantics, cryptographic construction, bootstrap semantics, lifecycle-conflict semantics, presentation history, and recovery behavior.  
**Still to freeze:** publication of the current lifecycle-oracle artifact, completion/publication of the mandatory cryptographic/TOTP/transition/recovery conformance corpus, and independent consumption of the semantic-object cryptographic vectors before interoperability is claimed.

**Project name:** Totipo. Frozen v0 protocol literals are not branding strings: the `TOTP-VAULT` bootstrap magic and the `TOTP-Vault/v0/...` domain-separation family retain their exact spellings and bytes as specified below. A conforming v0 implementation MUST NOT substitute `Totipo` or any other project/product name into those frozen literals.

**Revision 36 changes:** adds explicit Section 49.1 cross-references to the two earlier presentation-history rules that permit abandonment of pending same-key `DEVICE_UPDATE` history. This is cross-reference clarification only. All r35 Totipo branding, semantic, operational-security, recovery, application-safety, wire-format, bootstrap, cryptographic, and consensus rules remain unchanged.

---

## 1. Goals

This document specifies version 0 of the Totipo encrypted, append-only, multi-device TOTP vault format intended for synchronization over an untrusted or partially trusted file-synchronization medium such as Syncthing, Dropbox, or Google Drive.

The principal design goals are:

- small and auditable format and implementation surface;
- confidentiality against the synchronization medium;
- immutable authenticated content storage;
- deterministic convergence between independently operating devices;
- explicit preservation of conflicts rather than silent winner selection;
- device provenance without a device-authorization system;
- efficient incremental synchronization;
- minimal externally visible semantic metadata.

v0 intentionally does not define:

- device authorization, enrollment, or revocation;
- global membership;
- protocol-level garbage collection;
- secure deletion;
- root-key rotation;
- traffic-analysis resistance;
- global coordination or consensus.

---

## 2. Object model

v0 defines two encrypted, content-addressed semantic object types:

```text
TOKEN_UPDATE
DEVICE_UPDATE
```

It additionally defines one non-content-addressed bootstrap record:

```text
VAULT
```

These are the complete v0 protocol object types. Current heads, field revisions, conflicts, token/device state, and aggregate vault state are derived from valid object history rather than represented by additional synchronized objects.

A device is identified by its Ed25519 public key.

A `TOKEN_UPDATE` is one signed semantic event concerning exactly one logical token. Its context parents identify the token versions whose complete derived states the writer explicitly incorporates. The update may assert one or more semantic token fields. Fields not asserted by the update are inherited from the derived context, including unresolved field conflicts.

A `DEVICE_UPDATE` is one signed semantic event concerning presentation state for one device identity. In v0 its only semantic field is `DISPLAY_NAME`. Device-presentation history is independent of token history and MUST NOT affect token-state validity, token causality, or lifecycle-conflict calculation.

For each token, current token-update heads, per-field revision heads, ordinary field conflicts, and lifecycle conflicts are derived from valid `TOKEN_UPDATE` history for that token. The aggregate observed vault state is the collection of those independently derived token states.

A client MAY maintain off-store aggregate indexes/frontiers or materialized derived states for efficiency. Such local derived state is not synchronized protocol state and MUST be rebuildable from authenticated objects.

---

## 3. Threat model

The synchronization medium MUST NOT be trusted for:

- confidentiality;
- integrity;
- freshness;
- ordering;
- completeness;
- availability.

An attacker controlling the synchronization medium may:

- read files;
- copy files;
- add files;
- replace files;
- replay old files;
- reorder files;
- withhold files;
- delete files.

Without the vault root key, such an attacker MUST NOT be able to:

- decrypt semantic vault contents;
- compute valid object identifiers for guessed plaintext;
- construct a new valid vault object.

Possession of the vault root key grants read and semantic write capability in v0. A party able to recover `K_root` from a usable `VAULT` bootstrap therefore possesses the vault's v0 write capability; the protocol defines no additional membership or enrollment authorization check.

Device signing keys provide provenance, not authorization.

A party possessing `K_root` MAY generate a fresh vault-specific Ed25519 device identity and author valid `TOKEN_UPDATE` history under that new identity. It MAY also author `DEVICE_UPDATE` presentation history for that new identity. Such history is valid because v0 treats possession of `K_root`, together with a signing key controlled by the writer, as sufficient to create new vault history.

Possession of `K_root` alone MUST NOT permit a party to forge a `TOKEN_UPDATE` or `DEVICE_UPDATE` attributed to an already existing device identity without that device's private signing key. A friendly `DISPLAY_NAME` is self-asserted presentation data and MUST NOT be treated as enrollment, authorization, or proof of human identity; two device keys may legitimately claim the same friendly name.

A compromised unlocked client is assumed capable of reading secrets, generating a fresh device identity, and creating valid vault state under that identity.

Device revocation is outside v0.

---

## 4. Storage model

A vault is stored conceptually as:

```text
vault

objects/
    <OBJECT_ID>
    <OBJECT_ID>
    <OBJECT_ID>
    ...
```

`vault` is the bootstrap record.

One configured v0 vault owns one storage directory or equivalent synchronization namespace. A single `objects/` namespace MUST NOT intentionally mix objects belonging to different `K_root` values; multiple vaults require separate configured locations/namespaces.

Every file under `objects/` is immutable.

The object namespace is flat.

The filename is the lowercase hexadecimal representation of the object's 32-byte `OBJECT_ID`.

Only filenames consisting of exactly 64 lowercase hexadecimal characters are object candidates. Temporary files, synchronization conflict files, and unrelated filenames MUST be ignored by object discovery.

No semantic information appears directly in the storage hierarchy.

v0 never intentionally removes a valid historical object from the synchronized object store.

Writers SHOULD publish immutable objects using a temporary file followed by an atomic rename to the final object filename where the underlying platform permits it.

---

## 5. Cryptographic suite

v0 fixes the following primitives.

| Purpose | Primitive |
|---|---|
| Password KDF | Argon2id v=0x13 |
| Root/subkey derivation | HKDF-SHA-256 |
| Private content addressing | HMAC-SHA-256 |
| Object encryption | AES-256-GCM with 16-byte tag |
| Root-key wrapping | AES-256-GCM with 16-byte tag |
| Device signatures | Ed25519 |
| Random values | platform CSPRNG |

The Argon2id parameters for bootstrap version 0 are fixed:

```text
memory       = 65536 KiB
iterations   = 3
parallelism  = 4
salt         = 16 bytes
output       = 32 bytes
secret K     = empty byte string
associated X = empty byte string
```

The Argon2 version is `0x13`.

The number of execution threads used to evaluate Argon2id is an implementation detail. Implementations may evaluate the four lanes serially or in parallel as long as they produce the specified Argon2id output.

---

## 6. Password bytes

The password input domain is exactly `0..1024` bytes of well-formed UTF-8. The empty UTF-8 byte string is format-valid.

When an implementation accepts a character/string password, it MUST encode that input to UTF-8 strictly: encoding errors, including unpaired surrogate code units in APIs that can represent them, MUST be rejected rather than replaced. When an implementation accepts pre-encoded password bytes, it MUST reject malformed UTF-8.

The format performs no Unicode normalization.

Applications MUST NOT silently normalize or otherwise transform the password before KDF processing.

This means visually identical non-ASCII passwords entered in different Unicode normalization forms may not be interchangeable across devices. Clients SHOULD warn users of this possibility when setting or changing a password.

The resulting exact UTF-8 bytes are supplied directly to Argon2id.

Inputs exceeding 1024 UTF-8 bytes MUST be rejected. Implementations MUST NOT silently truncate a password to fit the limit. An application MAY impose a stronger password-strength or non-empty-password policy when creating or rewrapping a vault, but a conforming reader MUST remain able to process every format-valid password byte string, including the empty string.

---

## 7. Vault root key

At vault creation:

```text
K_root = CSPRNG(32 bytes)
```

`K_root` is permanent for the lifetime of the vault.

The password is an unlock credential for `K_root`; it is not itself the vault encryption key.

v0 does not support in-place `K_root` rotation.

Changing `K_root` constitutes creation of a new cryptographic vault.

---

## 8. VAULT bootstrap

The bootstrap record contains sufficient cleartext information to derive the password wrapping key and recover `K_root`. Unlike semantic objects, `VAULT` is not TLV-encoded. Bootstrap version 0 has one fixed byte layout and no padding, reserved fields, variable-length fields, or alternate encodings.

The exact v0 record is:

```text
offset  size  field
0       10    MAGIC = ASCII("TOTP-VAULT")
10       1    BOOTSTRAP_VERSION = u8 0x00
11      16    ARGON2_SALT
27      12    WRAP_NONCE
39      32    WRAPPED_ROOT
71      16    WRAP_TAG
-------------------------------
              total = 87 bytes
```

Thus the exact magic bytes are:

```text
54 4f 54 50 2d 56 41 55 4c 54
```

and the complete authenticated clear bootstrap header is exactly the first 39 bytes:

```text
VAULT_HEADER =
    MAGIC[10]
    || BOOTSTRAP_VERSION[1]
    || ARGON2_SALT[16]
    || WRAP_NONCE[12]
```

Bootstrap version 0 implies all KDF and wrapping algorithms and parameters. No algorithm identifiers, KDF work-factor fields, lengths, flags, or other attacker-controlled parameters are serialized. In particular, no attacker-controlled Argon2 memory, iteration, or parallelism fields exist.

On initial vault creation and every password rewrap:

```text
ARGON2_SALT = CSPRNG(16)
WRAP_NONCE  = CSPRNG(12)
```

The wrapping key is:

```text
K_wrap =
    Argon2id(
        password_bytes,
        ARGON2_SALT,
        m = 65536 KiB,
        t = 3,
        p = 4,
        output = 32,
        secret K = empty,
        associated data X = empty
    )
```

`K_root` is wrapped using AES-256-GCM with a 16-byte tag and `WRAP_NONCE`:

```text
WRAPPED_ROOT, WRAP_TAG =
    AES-256-GCM(
        key        = K_wrap,
        nonce      = WRAP_NONCE,
        plaintext  = K_root,
        AAD        = VAULT_HEADER,
        tag_length = 16 bytes
    )
```

`WRAPPED_ROOT` is therefore exactly 32 ciphertext bytes. `WRAP_TAG` is the final 16 bytes of the 87-byte record. The complete canonical clear bootstrap header through and including `WRAP_NONCE` — exactly bytes `0..38` — is authenticated as AEAD associated data. No part of `WRAPPED_ROOT` or `WRAP_TAG` is included in the AAD.

Before invoking Argon2id, a v0 reader MUST have rejected any candidate that fails the exact 87-byte v0 length, exact 10-byte magic, supported `BOOTSTRAP_VERSION = 0x00`, or the strict password-input domain of Section 6: the password must encode without replacement as well-formed UTF-8 and the resulting byte string must be at most 1024 bytes. Pre-encoded password bytes MUST likewise be validated as well-formed UTF-8 before the KDF. These checks MUST NOT invoke the password KDF. Unsupported bootstrap versions MUST be rejected before executing a KDF.

After those pre-KDF checks, the reader extracts the fixed salt and nonce, derives `K_wrap` with the fixed Section 5 parameters, and authenticates/decrypts exactly the final 48 bytes as `WRAPPED_ROOT[32] || WRAP_TAG[16]` with `VAULT_HEADER` as AAD. Wrong passwords and authenticated-field/ciphertext/tag modifications fail the GCM unwrap; such failure yields no candidate `K_root`.

The `VAULT` record is an offline password verifier. Anyone possessing a copy may attempt password guesses locally; Argon2id and password strength are the v0 defenses against this attack.

### 8.1 Initial bootstrap publication

Initial creation of `VAULT` MUST use the initial-creation publication rules of Section 9.1 and the durable first-establishment rules of Section 10.1.

Ordinary initial creation is permitted only for a configured vault location that has not already been initialized locally:

- the configured canonical `VAULT` pathname MUST be absent;
- the client MUST NOT initialize over an existing canonical `VAULT`; if one exists, creation fails and the client must use an explicit open, recovery, migration, or reconfiguration workflow instead;
- if the canonical `VAULT` is absent but candidate protocol object files already exist under `objects/`, ordinary creation MUST stop rather than generate a new unrelated root over potentially recoverable history; an explicit recovery or reconfiguration workflow is required. Such candidate files MUST be treated as potentially recoverable or damaged vault material, not automatically as garbage merely because `VAULT` is absent. Files that do not match the exact object-candidate filename grammar remain unrelated to this refusal rule.

A client MUST NOT create the configured `VAULT` pathname by directly truncating or writing the final file.

Before initial publication, the completed temporary bootstrap MUST be locally validated to ensure:

- exact expected v0 length and framing;
- the configured password unwraps the newly generated `K_root`;
- the unwrapped value exactly equals the newly generated root intended for this vault;
- the local `VAULT_BINDING` derived from that root is the value durably recorded in the pending establishment state of Section 10.1.

The canonical `VAULT` MUST then be installed with no-replace semantics as specified in Section 9.1. If another local creator or pre-existing vault wins the pathname race, the new candidate MUST NOT replace it.

Vault creation MUST NOT be reported as successful, and the new vault MUST NOT be used as ordinarily established state, until the installed canonical bootstrap has been revalidated against the durable pending `VAULT_BINDING` and establishment has been durably completed under Section 10.1.

---

## 9. Password changes and bootstrap replacement

v0 permits changing the password by rewrapping the same `K_root`.

The operation is:

```text
unlock current K_root
generate new Argon2 salt
derive K_wrap from new password
generate new 12-byte wrapping nonce
wrap same K_root
replace VAULT
```

No content-addressed object changes.

Password rewrapping protects the current `VAULT` bootstrap with the new password.

It does not revoke an old password against an attacker who possesses a historical copy of the previous bootstrap. A retained old bootstrap plus the corresponding old password still recovers the same `K_root`.

Clients MUST NOT describe password rewrapping as secure revocation of a leaked historical password/bootstrap pair.

Password changes are not mergeable distributed operations.

Concurrent password changes on unsynchronized devices may produce conflicting bootstrap files. Resolving such conflicts is outside the immutable object-history protocol.

An established client MAY remember the last accepted bootstrap representation and warn about apparent same-root bootstrap rollback. Durable `VAULT_BINDING` pinning is a MUST because a different `K_root` is a different cryptographic vault. Detecting rollback among bootstraps that unwrap to the already pinned same `K_root` is only a MAY because such rollback does not substitute vault identity or reveal semantic content beyond what a retained historical bootstrap/password pair already permits; its remaining effects are primarily unlock-credential rollback, availability, and user-experience/audit concerns.

A fresh client cannot prove that the bootstrap supplied by an untrusted storage medium is the newest historical bootstrap.

Bootstrap replacement is additionally subject to the durable vault-binding requirement in Section 10.

### 9.1 Crash-safe VAULT publication

`VAULT` is the only intentionally mutable/bootstrap protocol pathname in v0.

Initial creation and later same-root replacement have different pathname-install semantics and MUST NOT be conflated.

For either operation, a writer first constructs the complete candidate bootstrap in a separate temporary file in the same filesystem/storage context, writes the complete candidate bytes, durably flushes the temporary file to the extent supported by the platform, and validates the complete temporary bootstrap before any canonical-path installation.

For **initial creation**:

1. the candidate MUST satisfy the initial validation requirements of Section 8.1;
2. the `PENDING` establishment record containing the intended `VAULT_BINDING` MUST already be durably persisted as required by Section 10.1;
3. the configured canonical `VAULT` pathname MUST still be absent immediately before installation;
4. the writer MUST install the completed candidate with an atomic **no-replace** operation, exclusive-create/link/install primitive, or equivalent local exclusion that guarantees an existing canonical `VAULT` is not replaced;
5. if the destination already exists, the writer MUST NOT replace it and MUST treat ordinary creation as unsuccessful; if a pending establishment record already exists, recovery may validate the existing canonical bootstrap against that record as described in Section 10.1;
6. containing-directory metadata MUST be durably flushed when the platform provides such a facility;
7. the installed canonical bootstrap MUST be reopened/validated and required to derive exactly the pending `VAULT_BINDING` before establishment is durably completed;
8. creation success MUST be reported only after Section 10.1 marks the configured vault durably `ESTABLISHED`.

If the local platform cannot provide a primitive or exclusion mechanism that guarantees initial installation will not overwrite an existing canonical pathname in the presence of local concurrent creators, the implementation MUST NOT claim crash-safe ordinary initial creation in that context. It must use a stronger local exclusion/creation mechanism or require an explicit recovery/reconfiguration workflow.

For **password rewrap or other same-root bootstrap replacement**:

1. the configured vault MUST already be durably `ESTABLISHED` under Section 10.1;
2. the candidate MUST unwrap with the new password to exactly the already-established unchanged `K_root` and derive the already pinned `VAULT_BINDING`;
3. the writer MAY atomically replace the existing canonical `VAULT` because this operation intentionally republishes another wrapping of the same established root;
4. containing-directory metadata MUST be durably flushed when the platform provides such a facility;
5. password-change/bootstrap-replacement success MUST be reported only after publication completes.

The configured `VAULT` pathname MUST NOT be created or updated by truncating/overwriting that final file in place.

If the underlying filesystem/synchronization layer cannot provide atomic replacement semantics for a same-root rewrap, the implementation MUST use the strongest available crash-safe publication mechanism and MUST treat interrupted or ambiguous publication as incomplete.

Temporary files and provider/synchronization conflict copies are not authoritative bootstraps merely because they exist.

Only the single locally configured canonical `VAULT` pathname is considered automatically as the bootstrap candidate.

Files such as:

```text
VAULT (conflicted copy ...)
vault.sync-conflict-...
vault.tmp
```

or equivalent provider-generated alternate names MUST NOT be interpreted automatically as additional or alternate bootstraps.

A client MAY surface such files as evidence of a synchronization conflict, but selecting/importing one requires an explicit recovery workflow.

Crash-safe publication protects local creation/replacement from torn writes and the no-replace rule protects against overwriting a locally existing bootstrap during ordinary creation. These local rules do not coordinate two disconnected endpoints that each initialize an apparently empty synchronization location before seeing the other endpoint's files; such independently created roots remain separate cryptographic vaults and require explicit recovery/reconfiguration when synchronization later exposes the collision.

Crash-safe publication does not make concurrent password changes mergeable, prevent provider-side historical versions, or establish freshness against rollback.

---

## 10. Durable vault binding and root-key compromise

`K_root` is the cryptographic identity of a v0 vault.

### 10.1 Durable local vault binding and first establishment

`VAULT_BINDING` is the durable local identity pin for a configured vault. v0 defines it exactly as:

```text
VAULT_BINDING =
    HMAC-SHA-256(
        K_root,
        ASCII("TOTP-Vault/v0/local-vault-binding")
    )
```

This exact construction and literal domain string are part of v0 local security-state semantics. Implementations MUST NOT substitute an implementation-defined root identifier while claiming the v0 durable-binding guarantees.

`VAULT_BINDING` is local-only state and is not synchronized as part of the vault format.

A configured vault has a local establishment state:

```text
UNESTABLISHED
PENDING(VAULT_BINDING)
ESTABLISHED(VAULT_BINDING)
```

Transitions of this establishment state for one configured vault MUST be locally serialized or implemented with compare-and-set/equivalent transactional preconditions. A stale concurrent initializer MUST NOT overwrite a `PENDING` or `ESTABLISHED` binding produced by another initialization attempt. The canonical-`VAULT` no-replace rule of Section 9.1 does not by itself satisfy this local establishment-state requirement.

The establishment state and binding MUST be durably associated with the configured vault identity/location. A client MUST NOT report first creation/open success, expose ordinary accepted vault state, associate/reuse a device signing key for ordinary use, or author `TOKEN_UPDATE`/`DEVICE_UPDATE` history until the configured vault is durably `ESTABLISHED`.

For **initial creation**, after generating and locally validating the candidate root/bootstrap but before installing the canonical `VAULT`, the client MUST durably persist `PENDING(VAULT_BINDING)` for the newly generated `K_root`. Initial publication then follows the no-replace procedure of Section 9.1. After installation, the client MUST reopen/validate the canonical `VAULT`, require that it unwraps to a root deriving exactly the pending binding, and only then durably transition the establishment record to `ESTABLISHED`.

For **first opening of an existing configured vault**, after the bootstrap has passed framing/KDF validation and the supplied password successfully unwraps a candidate `K_root`, the client MUST derive `VAULT_BINDING` and durably establish that binding before the vault is treated as successfully opened or its state is accepted for ordinary use. An implementation MAY represent this as an atomic durable `ESTABLISHED(VAULT_BINDING)` write, or as `PENDING` followed by `ESTABLISHED`; either way, no successful/open accepted state may precede durable binding persistence.

Crash recovery is normative:

- If a `PENDING` establishment record exists and the canonical `VAULT` is present, the client MUST validate the canonical bootstrap and require that it derives the exact pending binding before completing establishment. A differing root MUST stop automatic establishment and require explicit recovery/reconfiguration.
- If a `PENDING` establishment record exists and the canonical `VAULT` is absent, setup remains incomplete. The client MUST NOT silently begin first contact with a different root. It MAY resume the original creation only if it can reproduce/validate a candidate deriving the same pending binding; otherwise explicit user-directed abort/reconfiguration is required before a different root may be established.
- A crash after canonical initial publication but before the durable `ESTABLISHED` transition therefore leaves a recoverable incomplete establishment, not an unpinned first-contact state.
- A crash after durable `ESTABLISHED` persistence retains the normal root-substitution protection below.

The minimal `PENDING(VAULT_BINDING)` record need not itself contain `K_root` or password-derived secret material. To make an interrupted initial creation resumable, an implementation MAY additionally retain the candidate bootstrap bytes, candidate `K_root`, or equivalent local recovery material sufficient to continue the same establishment attempt. Any such secret-bearing retained material is local sensitive state and MUST receive the confidentiality protections required by Section 61. An implementation MAY instead retain only the binding and require explicit abort/reconfiguration when the original candidate cannot be reproduced; it MUST NOT silently substitute a different root.

On every subsequent successful bootstrap unwrap for an `ESTABLISHED` configured vault, the client MUST derive the candidate binding and compare it to the pinned local binding.

If the binding differs:

```text
candidate vault != established vault
```

The client MUST NOT silently switch roots, continue using the existing device signing key, or author into the substituted vault.

A different root requires an explicit separate-vault, migration, or reconfiguration workflow.

Root pinning does not solve:

- first-contact substitution before any `PENDING` or `ESTABLISHED` binding has ever been durably recorded;
- same-root rollback to an older bootstrap wrapping the same `K_root`;
- withholding of newer history.

---

### 10.2 Device-key binding

A locally stored device signing identity MUST be bound in local security state to the same established vault binding.

A client MUST NOT silently reuse a device signing key with a different `K_root`.

### 10.3 Root-key compromise

If `K_root` is compromised, re-encrypting the same data using a new root key cannot undo disclosure of previously readable TOTP seeds.

Recovery requires creation of a new vault.

Credentials whose secret material may have been exposed SHOULD be rotated with their issuers.

---

## 11. Root-key hierarchy

After recovering `K_root`:

```text
PRK =
    HKDF-Extract(
        salt = 32 zero bytes,
        IKM  = K_root
    )
```

Derive:

```text
K_id =
    HKDF-Expand(
        PRK,
        ASCII("TOTP-Vault/v0/object-id"),
        32
    )

K_object_root =
    HKDF-Expand(
        PRK,
        ASCII("TOTP-Vault/v0/object-key-root"),
        32
    )

K_signature_context =
    HKDF-Expand(
        PRK,
        ASCII("TOTP-Vault/v0/signature-context"),
        32
    )
```

These literal domain strings are part of v0.

Keys derived for different purposes MUST NOT be interchanged.

---

## 12. Canonical plaintext and common prefix

Every v0 semantic object has one and only one canonical plaintext TLV representation:

```text
P = CANONICAL_TLV(object)
```

Every semantic object that relies on the v0-compatible encrypted-object envelope and authenticated-future dispatch begins with the frozen common prefix containing:

```text
OBJECT_VERSION
OBJECT_TYPE
```

at fixed canonical positions defined by the primitive TLV framing and common-prefix rules of Section 69.1.

This prefix remains encrypted in storage, but after successful AEAD decryption allows an older implementation to distinguish a supported v0 object from authenticated data belonging to a later semantic version.

The plaintext contains all semantic information, including:

- token IDs;
- public keys;
- context-parent references;
- token field assertions;
- credential values;
- device names;
- signatures.

Canonicalization is security-critical.

Non-canonical representations MUST be rejected.

---

## 13. OBJECT_ID

For canonical plaintext `P`:

```text
OBJECT_ID =
    HMAC-SHA-256(
        K_id,
        P
    )
```

`OBJECT_ID` is 32 bytes.

It is:

- deterministic within one vault;
- private to that vault;
- content-addressed;
- unsuitable for plaintext guessing without `K_id`;
- different across independently keyed vaults even for identical plaintext.

---

## 14. Per-object encryption

For object identifier `ID`:

```text
K_object =
    HKDF-Expand(
        K_object_root,
        ASCII("TOTP-Vault/v0/object-key") || ID,
        32
    )
```

The nonce is:

```text
nonce = first 12 bytes of ID
```

The associated data is:

```text
AAD =
    ASCII("TOTP-Vault/v0/object")
    || ID
```

v0 encrypts every semantic object using one fixed **2048-byte** object envelope. v0 does not use multiple size buckets.

Let:

```text
P = canonical semantic plaintext
```

The semantic object identifier remains:

```text
OBJECT_ID = HMAC-SHA-256(K_id, P)
```

Before encryption, the writer constructs one deterministic padded encryption plaintext:

```text
ENCRYPTION_PLAINTEXT =
    SEMANTIC_LENGTH_U16BE
    || P
    || ZERO_PADDING
```

where:

- `SEMANTIC_LENGTH_U16BE` is exactly 2 bytes, an unsigned big-endian integer equal to the exact byte length of `P`;
- `ZERO_PADDING` consists only of zero bytes;
- `ENCRYPTION_PLAINTEXT` is exactly 2032 bytes, so the final ciphertext plus the 16-byte GCM tag is exactly 2048 bytes;
- the envelope can frame at most 2030 semantic bytes;
- Section 69's v0 field/count limits make the largest valid v0 `TOKEN_UPDATE` 1987 semantic bytes and the largest valid v0 `DEVICE_UPDATE` 1532 semantic bytes;
- the representation is canonical: for a given `P`, there is exactly one valid `ENCRYPTION_PLAINTEXT` and therefore one envelope length.

Oversized semantic plaintext is invalid. Writers MUST NOT truncate fields, omit required context/field semantics, select an alternate envelope size, or otherwise change semantics merely to fit the envelope.

Encryption is AES-256-GCM with a 16-byte authentication tag:

```text
ciphertext, tag =
    AES-256-GCM(
        K_object,
        nonce,
        ENCRYPTION_PLAINTEXT,
        AAD,
        tag_length = 16 bytes
    )
```

The object file contains:

```text
ciphertext || tag
```

and therefore every semantic object file has exactly:

```text
object_file_length = 2048 bytes
```

No nonce or clear semantic header is stored.

Padding is not part of semantic object identity and is not included in the signed canonical semantic plaintext. It is a deterministic encryption-envelope encoding derived from that plaintext.

A per-object key MUST be used only for the exact canonical padded encryption plaintext determined by the semantic plaintext from which the corresponding object identifier was derived.

The required creation order is:

```text
canonical semantic plaintext P
        ↓
OBJECT_ID = HMAC(K_id, P)
        ↓
canonical fixed-size padded envelope for P
        ↓
per-object key from OBJECT_ID
        ↓
nonce from OBJECT_ID
        ↓
AES-GCM encryption
```

No random padding is permitted.

---

## 15. Object reading

Given object filename `ID`, a reader first requires the semantic object file to be exactly 2048 bytes. Any other size MUST be rejected before AEAD processing.

For a length-valid object, the reader:

```text
derives K_object from ID
derives nonce from ID
authenticates/decrypts the fixed 2032-byte encryption plaintext
reads authenticated SEMANTIC_LENGTH_U16BE
requires SEMANTIC_LENGTH_U16BE to fit within the fixed envelope
extracts exactly SEMANTIC_LENGTH_U16BE bytes as semantic plaintext P
requires every remaining padding byte to be zero
recomputes HMAC-SHA-256(K_id, P)
requires recomputed ID == filename ID
parses and validates only the frozen common prefix of P
reads OBJECT_VERSION
if OBJECT_VERSION != 0:
    classify as AUTHENTICATED_UNSUPPORTED_FUTURE under Section 59
    do not parse OBJECT_TYPE semantics or require the remaining payload to be valid v0 TLV
if OBJECT_VERSION == 0:
    parse the complete canonical v0 TLV plaintext P
    validate the supported v0 semantic object
```

`SEMANTIC_LENGTH_U16BE` and padding are interpreted only after successful AEAD authentication. The reader MUST reject non-canonical envelope encodings, including non-zero padding or a semantic length inconsistent with the fixed envelope.

The common-prefix parse occurs only after successful AEAD authentication, canonical envelope validation, and object-ID recomputation. For `OBJECT_VERSION=0`, complete semantic TLV parsing likewise occurs only after those checks. For an authenticated unsupported version, bytes after the frozen common prefix are opaque to v0 and MUST NOT be rejected merely because they are not a valid v0 `OBJECT_TYPE` or v0 TLV tail.

An authenticated plaintext that does not contain the complete canonical common prefix cannot be classified by `OBJECT_VERSION`. Such a plaintext is invalid framing and MUST NOT be classified as `AUTHENTICATED_UNSUPPORTED_FUTURE`. The opaque-tail rule begins only after the complete common prefix has parsed canonically and names an unsupported version.

A missing, truncated, wrong-length, unreadable, authentication-failing, substituted, or object-ID-mismatching current pathname representation MUST be treated as unavailable or untrustworthy and MUST be reconsidered if the bytes later change. Such a pathname/read failure MUST NOT by itself establish identity-grounded terminal invalidity of an immutable object. Only invalidity established from authenticated, identity-matching object contents may be treated as terminal under Section 56.2.

---

## 16. Device identity and Ed25519 profile

A device identity is its Ed25519 public key:

```text
DEVICE_ID := DEVICE_PUBLIC_KEY
```

No secondary device identifier exists.

A device MUST generate a distinct Ed25519 signing key for each vault. Reusing one device key across vaults is non-conforming.

Device keys provide authorship/provenance only. They do not imply authorization.

v0 uses PureEd25519 as defined by RFC 8032. Ed25519ctx and Ed25519ph are not used.

Deterministic signing according to RFC 8032 is a writer-conformance requirement. A verifier cannot prove from a public key and accepted signature that the signer used the required deterministic nonce derivation.

Verification is part of v0 consensus validity and MUST use the following exact acceptance profile:

1. `DEVICE_PUBLIC_KEY` is exactly 32 bytes.
2. `SIGNATURE` is exactly 64 bytes and is parsed as `R_enc[32] || S_enc[32]`.
3. Decode `A` from `DEVICE_PUBLIC_KEY` and `R` from `R_enc` using Edwards25519 point decoding. Decoding failure is rejection.
4. Re-encode each decoded point using canonical Edwards25519 encoding and require byte-for-byte equality with its input. Non-canonical encodings are rejected.
5. Reject `A` or `R` when `[8]A` or `[8]R` is the identity point.
6. Require prime-subgroup membership:

```text
[L]A = identity
[L]R = identity
```

where `L` is the prime order of the Ed25519 base point.

Implementations MUST perform these as true full-group membership tests. An API that reduces the scalar modulo `L` before multiplication does not implement this check correctly: supplying scalar `L` is reduced to `0`, so the API computes `[0]P = identity` for every input point and the purported subgroup test passes vacuously.
7. Interpret `S_enc` as a little-endian integer and require `0 <= S < L`.
8. Compute:

```text
h = SHA-512(R_enc || DEVICE_PUBLIC_KEY || message)
k = little_endian_integer(h) mod L
```

9. Accept only if the cofactorless verification equation holds:

```text
[S]B = R + [k]A
```

where `B` is the Ed25519 base point.

Because accepted `A` and `R` are required to lie in the prime-order subgroup, multiplication by the cofactor `8` is invertible on the accepted subgroup. Therefore the cofactorless equation above and the corresponding cofactored verification equation are equivalent for points that have already passed the v0 subgroup checks. v0 nevertheless states the cofactorless equation normatively.

Because accepted `A` and `R` are required to lie in the prime-order subgroup, v0 deliberately excludes mixed-order points even if a more permissive Ed25519 verifier would accept them.

Conformance MUST be established against the published v0 acceptance/rejection vectors rather than inferred from a library method name such as `verify_strict`.

---

## 17. TOKEN_ID

Every logical token receives:

```text
TOKEN_ID = CSPRNG(32 bytes)
```

at creation.

A token ID:

- has no semantic meaning;
- is never intentionally reused;
- remains constant for the logical lifetime of that token;
- is shared by all field-revision history for that token.

If independent stale histories nevertheless create different roots using the same `TOKEN_ID`, those roots are handled as ordinary concurrent token history and their asserted fields conflict or compose according to the normal field rules.

---

## 18. TOKEN_UPDATE

A `TOKEN_UPDATE` is one immutable signed semantic event affecting exactly one logical token.

Conceptually:

```text
TOKEN_UPDATE {
    OBJECT_VERSION = 0
    OBJECT_TYPE = TOKEN_UPDATE

    DEVICE_PUBLIC_KEY[32]

    CONTEXT_PARENT_COUNT
    CONTEXT_PARENT_UPDATE_ID[...]

    TOKEN_ID[32]

    STATUS?       = LIVE | TOMBSTONE
    ISSUER?       = UTF8
    ACCOUNT?      = UTF8
    CREDENTIAL? {
        ALGORITHM
        DIGITS
        PERIOD
        SECRET_BYTES
    }

    SIGNATURE[64]
}
```

At least one semantic field assertion MUST be present.

Every context parent MUST reference a fully valid `TOKEN_UPDATE` with the same `TOKEN_ID` as the child update.

The context-parent list is a set, is canonically sorted by raw object ID, contains no duplicates, and MUST be an antichain under token-update context ancestry.

A `TOKEN_UPDATE` MAY be parentless only when its required token context is empty. In normal valid history this is token creation.

No sequence number, wall-clock timestamp, action discriminator, distinguished same-device predecessor, per-device sequence, serialized field-parent list, serialized conflict list, or serialized complete token-state snapshot exists.

`DEVICE_PUBLIC_KEY` identifies signer/provenance only. It creates no causal relation to other updates signed by the same key.

Context-parent edges induce token-local causal ordering. Two valid updates for different `TOKEN_ID`s are never causally related by v0. Two valid updates for the same token that are not in one another's context ancestry remain concurrent and have no protocol ordering merely because they were observed in some order or signed by the same device.

---

## 19. Semantic-object signatures and vault binding

Both semantic object types are directly signed by the device identity named in the object.

A `TOKEN_UPDATE` signature is PureEd25519 over:

```text
ASCII("TOTP-Vault/v0/token-update")
||
K_signature_context
||
canonical_unsigned_token_update
```

A `DEVICE_UPDATE` signature is PureEd25519 over:

```text
ASCII("TOTP-Vault/v0/device-update")
||
K_signature_context
||
canonical_unsigned_device_update
```

The corresponding canonical unsigned object is the canonical plaintext representation with the `SIGNATURE` tag entirely absent. The signature field is not represented as an empty or zero value while signing.

Construction order for either signed semantic object is:

```text
build canonical unsigned object
        ↓
construct object-specific signature input
        ↓
PureEd25519 deterministic signing
        ↓
insert signature
        ↓
canonical signed plaintext
        ↓
compute OBJECT_ID
        ↓
encrypt
```

The object-specific literal context binds the signature to this vault and semantic purpose.

A semantic object whose signature fails the exact v0 Ed25519 acceptance profile of Section 16 is invalid.

Two independently produced byte-identical updates under the same signing key and identical context collapse to the same content-addressed object. Distinct semantic events necessarily differ in signed content or context and therefore have distinct object IDs except with negligible cryptographic collision probability.

---

## 20. DEVICE_UPDATE

A `DEVICE_UPDATE` is an immutable signed semantic event for friendly presentation state of one device identity.

Conceptually:

```text
DEVICE_UPDATE {
    OBJECT_VERSION = 0
    OBJECT_TYPE = DEVICE_UPDATE

    DEVICE_PUBLIC_KEY[32]

    CONTEXT_PARENT_COUNT
    CONTEXT_PARENT_UPDATE_ID[...]

    DISPLAY_NAME

    SIGNATURE[64]
}
```

Every context parent MUST reference a fully valid `DEVICE_UPDATE` signed by the same `DEVICE_PUBLIC_KEY`.

The context-parent list is a set, is canonically sorted by raw object ID, contains no duplicates, and MUST be an antichain under device-update ancestry.

A parentless `DEVICE_UPDATE` is a root of that device's presentation history. Multiple roots or incomparable descendants signed by the same key are valid forked presentation history.

A conforming writer publishing a new `DEVICE_UPDATE` MUST name the complete currently known maximal `DEVICE_UPDATE` head set for that device identity as its context-parent set. Withholding by the synchronization medium may nevertheless cause later-discovered presentation forks.

Because `DISPLAY_NAME` is the only v0 device field, every `DEVICE_UPDATE` asserts a new display-name revision and thereby resolves every display-name alternative present in its listed context. Later hidden presentation history may create a new fork.

A device may have valid `TOKEN_UPDATE`s without any available `DEVICE_UPDATE`; in that case user interfaces use the device fingerprint or an equivalent unnamed-device presentation.

---

## 21. Device-update ancestry and observed name semantics

Define `device_update_ancestors(U)` recursively from `U.CONTEXT_PARENT_UPDATE_ID` references, excluding `U` itself.

For device updates `U1` and `U2` signed by the same device key, define:

```text
U1 <=d U2
```

when `U1 == U2` or `U1` is in `device_update_ancestors(U2)`.

A current device-update head is a fully valid `DEVICE_UPDATE` not strictly dominated under `<=d` by another fully valid update signed by the same key.

Normally one current update head exists. If multiple incomparable current update heads exist, their `DISPLAY_NAME` assertions are the current name alternatives. Identical displayed values MAY be coalesced for presentation, but the distinct signed update heads remain separate ancestry until a later `DEVICE_UPDATE` incorporates them.

Device-update ordering is presentation-only. No `DEVICE_UPDATE` may be used as a token-update context parent.

---

## 22. Base state and token-local context

For `TOKEN_UPDATE C` affecting token `T`, define:

```text
CONTEXT_PARENTS(C) =
    set of C.CONTEXT_PARENT_UPDATE_ID values
```

Every member concerns the same `TOKEN_ID = T`.

For a parentless update:

```text
BASE_STATE(C) = empty state for T
```

Otherwise:

```text
BASE_STATE(C) =
    JOIN(
        state(P)
        for P in CONTEXT_PARENTS(C)
    )
```

where every `state(P)` is the derived state of token `T` only.

The context-parent set is signed and therefore forms an authenticated statement of the token versions explicitly incorporated as the basis of `C`. It proves that the named histories are in `C`'s causal past. It does not prove that no additional history was visible to the writer or that the synchronization medium disclosed everything.

No global merge or globally advancing vault frontier exists. Each token evolves independently.

For writer conformance, a `TOKEN_UPDATE(T)` MUST use the complete accepted observed current token-update frontier for `T` as its context-parent set. This is a writer-conformance requirement rather than a retrospective validity predicate: later-discovered or withheld history does not retroactively invalidate an update that was valid in its signed context.

---

## 23. TOKEN_UPDATE semantic fields

The v0 token field set is exactly:

```text
STATUS
ISSUER
ACCOUNT
CREDENTIAL
```

`STATUS` is `LIVE` or `TOMBSTONE`.

`ISSUER` and `ACCOUNT` are complete UTF-8 field values subject to the string rules and limits of Section 70.

`CREDENTIAL` is one atomic semantic field containing:

```text
ALGORITHM
DIGITS
PERIOD
SECRET_BYTES
```

These credential subfields MUST evolve together as one value. They are not independently mergeable because silently combining a seed with independently authored OTP parameters could create a credential value no writer intended.

Presence of a token field in update `C` means that `C` authors a new revision of that field. Absence means that the complete possibly-conflicted field state is inherited from `BASE_STATE(C)`.

The protocol does not distinguish whether a field assertion was supplied by a human, application logic, import, or recovery workflow. Authorship means only that the signed update introduces that semantic assertion.

---

## 24. Implicit field parents and typed field ancestry

For field `F` asserted by update `C`, define the field revision identity conceptually as:

```text
FIELD_REVISION(C, F) = (OBJECT_ID(C), F)
```

Its field-parent set is not serialized. It is defined deterministically as:

```text
FIELD_PARENTS(C, F) =
    field_heads(BASE_STATE(C), F)
```

Thus an assertion of `F` replaces every currently observed alternative of `F` in the update's context. v0 has no ordinary partial field merge or generic serialized `SUPERSEDES` list. If the user is not ready to choose a value for a conflicted field, the update simply does not assert that field and its conflict is inherited unchanged.

Define typed field ancestry `<=F` by following these implicit field-parent relations. Strict ancestry `<F` means `A != B` and `A <=F B`.

For every token field, field ancestry implies token-update causal ancestry:

```text
A <F B  =>  update(A) <c update(B)
```

An update asserting several fields creates one new revision in each corresponding typed field lineage. The lineages remain independent even though the revisions share one signed author event.

---

## 25. Token-update ancestry and concurrency

Define `token_update_ancestors(C)` recursively from context-parent references only, excluding `C` itself.

Every token-update ancestor of `C` MUST concern the same `TOKEN_ID` as `C`.

For updates `A` and `B` concerning the same token, define:

```text
A <c B
```

when `A` is in `token_update_ancestors(B)`.

Define reflexive causal incorporation:

```text
A <=c B
```

when `A == B` or `A <c B`.

Two distinct updates for the same token are concurrent when neither is in the other's causal ancestry:

```text
A ||c B
```

Updates for different `TOKEN_ID`s have no v0 causal relation. Implementations MUST NOT infer ordering between them from device key, file arrival, wall clock, canonical sorting, or an off-store aggregate index.

---

## 26. Derived token state

`TOKEN_UPDATE` does not serialize complete token state.

For every fully valid update `C` concerning token `T`, define deterministic derived state:

```text
state(C)
```

containing one current field-head set for each v0 token field:

```text
TOKEN_ENTRY {
    TOKEN_ID = T
    STATUS_HEADS
    ISSUER_HEADS
    ACCOUNT_HEADS
    CREDENTIAL_HEADS
}
```

Head sets are derived values, not wire fields. A single `state(C)` may therefore carry unresolved field conflicts even when `C` is the sole current token-update/version head.

Once a token has been created, every required field-head set remains non-empty in every descendant state.

Implementations MAY cache `state(C)`, field ancestry, materialized field values, or persistent-map/index equivalents. Such caches are local derived state and MUST be semantically reproducible from authenticated update history.

---

## 27. Maximal field heads and JOIN

For one field lineage, a revision is maximal in a state when no other revision in that state strictly descends from it under that field's typed ancestry.

For field `F`:

```text
JOIN_FIELD_HEADS_F(A, B):
    U = union(A, B)
    remove every X in U for which
        there exists Y in U such that X <F Y
    return remaining members
```

For two valid states of the same token, `JOIN` applies this operation independently to every token field.

`JOIN` is idempotent, commutative, and associative for valid derived states of the same token.

There is no cross-token `JOIN` operation in the causal protocol. The observed vault state is the mapping from each observed `TOKEN_ID` to its independently derived token state.

---

## 28. Token creation

Let `C = TOKEN_UPDATE(T)` and `B = BASE_STATE(C)`.

If `T` is absent from `B`, the operation is token creation.

Creation MUST assert all four required fields:

```text
STATUS = LIVE
ISSUER
ACCOUNT
CREDENTIAL
```

`CONTEXT_PARENTS(C)` MUST be empty and every implicit field-parent set is therefore empty.

The token ID MUST NOT already exist in `B`.

If independent concurrent histories nevertheless create the same random `TOKEN_ID`, their roots are joined as ordinary concurrent field histories when later observed together. Honest collision is cryptographically negligible; clients SHOULD surface multiple independent creation roots as anomalous duplicate creation rather than implying that they necessarily represent intentional edits of one credential.

---

## 29. Applying TOKEN_UPDATE to base state

Let `C = TOKEN_UPDATE(T)` with base state `B`.

For each token field `F`:

```text
if C asserts F:
    field_heads(state(C), F) = { FIELD_REVISION(C,F) }
else:
    field_heads(state(C), F) = field_heads(B,F)
```

The replacement of all old heads for an asserted field is implicit because those heads are exactly `FIELD_PARENTS(C,F)` from Section 24.

Thus context incorporation collapses the token-update/version frontier but does not necessarily resolve inherited field conflicts. An update may become the sole current token-update head while `state(C)` still contains multiple heads for one or more fields it did not assert.

Lifecycle restrictions in Sections 32–41 override ordinary application rules where applicable.

---

## 30. Ordinary field conflicts and value ambiguity

A token field has an ordinary revision conflict when its derived current field-head count is greater than one.

Conflict is a property of revision lineage, not merely distinct displayed bytes. Several concurrent field heads may carry byte-identical values; implementations MAY present such a field as value-unambiguous while retaining the distinct revision heads for ancestry and future validation.

Distinct current values of the same field MUST be surfaced as an ordinary field-value conflict unless a stronger lifecycle rule applies.

No ordinary conflict object or serialized conflict list exists.

---

## 31. Resolving ordinary field conflicts

An ordinary field conflict is resolved only by asserting a new value for that field in an update whose context incorporates the conflicting current versions.

Because `FIELD_PARENTS(C,F)` is implicitly the complete current field-head set in `BASE_STATE(C)`, the new assertion resolves every alternative of `F` currently visible in that context.

v0 intentionally has no partial ordinary field merge. A client that is not prepared to select a value for all currently observed alternatives leaves `F` absent from the new update; the conflict is inherited unchanged.

A chosen value MAY equal one of the existing alternatives. The new revision is still a new signed semantic assertion recording the resolution over the incorporated context.

---

## 32. Lifecycle status transitions

A `STATUS` field revision `L` is a lifecycle transition when its output value differs from the value of at least one implicit `STATUS` field parent.

Thus `LIVE -> TOMBSTONE` is deletion and `TOMBSTONE -> LIVE` is restoration. If a status assertion has multiple parents, differing from any parent makes it a lifecycle transition.

Initial `LIVE` creation is not a lifecycle transition for conflict detection.

Concurrent `STATUS` revisions alone do not constitute a lifecycle conflict. If no concurrent `CREDENTIAL` revision produces a lifecycle witness, differing `LIVE`/`TOMBSTONE` status heads are an ordinary field conflict and are resolved by an ordinary subsequent `STATUS` assertion over a context incorporating the conflicting status alternatives. The lifecycle-sensitive `STATUS` resolution and explicit-confirmation procedure exists specifically for status decisions that race with credential evolution.

---

## 33. Credential witness history

Lifecycle conflicts arise because token lifecycle (`STATUS`) and credential material (`CREDENTIAL`) evolve independently.

A lifecycle disagreement can remain relevant after the credential revision that first raced with a status transition ceases to be a current credential head. Therefore lifecycle analysis MUST consider historical credential revisions beneath current credential heads.

For token `T` in state `S`, define:

```text
CREDENTIAL_HISTORY(S,T)
```

as the union, over every current credential head `CH`, of `CH` and all credential-field ancestors of `CH`.

Missing required credential ancestry makes lifecycle analysis incomplete rather than proving absence of conflict.

---

## 34. Lifecycle-resolution candidate and coverage

Let `R = TOKEN_UPDATE(T)`. After envelope/object-identity, structural, signature, and context-parent validation have succeeded and every listed context parent is fully valid, derive:

```text
B = BASE_STATE(R)
PRE_R = LIFECYCLE_CONFLICTS(B,T)
```

using only already fully valid history and already fully valid lifecycle-resolution coverage.

`R` is a **lifecycle-resolution candidate** when:

- `PRE_R` is non-empty;
- `R` asserts exactly one token field, `STATUS`; and
- the asserted `STATUS` value is `LIVE` or `TOMBSTONE`.

Candidate classification occurs before final lifecycle-resolution validity is known. A candidate MUST NOT be treated as a fully valid update merely because it satisfies these preconditions.

No `LIFECYCLE_CONFIRMATION` wire field exists in v0.

For the sole purpose of validating lifecycle-resolution candidate `R`, define provisional candidate coverage:

```text
CANDIDATE_COVERS(R,W)
iff
    update(W) <c R
```

When computing the candidate's own post-state lifecycle conflicts, the newly authored `STATUS` revision of `R` is provisionally treated as covering exactly those historical credential revisions `W` satisfying `CANDIDATE_COVERS(R,W)`. Previously fully valid lifecycle-resolution revisions continue to contribute their ordinary coverage. Candidate coverage MUST NOT be used to validate any other object, MUST NOT enter accepted history, and disappears entirely if `R` fails its final validation checks.

After `R` passes the final lifecycle-resolution validation of Section 51, it becomes a fully valid **lifecycle-resolution update**. A fully valid lifecycle-resolution status revision `R` covers a historical credential revision `W` when:

```text
update(W) <c R
```

For current status head `H`, define:

```text
CONFIRMED(H,W)
```

when there exists a fully valid lifecycle-resolution status revision `R` such that:

```text
R <=STATUS H
and
update(W) <c R
```

Thus a fully valid resolution covers all credential history actually incorporated by its signed context. It does not cover a credential revision authored on a stale or hidden branch that the resolution did not incorporate. An ordinary non-`STATUS` update whose base has lifecycle conflicts may causally incorporate the same history, but that incorporation alone MUST NOT count as lifecycle resolution or make `CONFIRMED(H,W)` true.

---

## 35. Raw and unresolved lifecycle witnesses

For current status head `H` of token `T` in state `S`, and credential revision `W` in `CREDENTIAL_HISTORY(S,T)`, define:

```text
RAW_LIFECYCLE_WITNESS(H,W)
```

when there exists lifecycle-transition status revision `L` such that:

```text
L <=STATUS H
and
update(L) ||c update(W)
```

Define:

```text
UNRESOLVED_LIFECYCLE_WITNESS(H,W)
```

when `RAW_LIFECYCLE_WITNESS(H,W)` and not `CONFIRMED(H,W)`.

Historical credential witnesses MUST NOT be discarded merely because a causally later credential descendant is current. A valid history can contain an unresolved historical witness beneath a current credential revision that itself is no longer concurrent with the relevant status transition.

---

## 36. Maximal unresolved witness frontier and lifecycle-conflict set

For one current status head `H`, define `MAXIMAL_UNRESOLVED_WITNESSES(S,T,H)` as every unresolved witness `W` for which no other unresolved witness `W2` satisfies:

```text
W <CREDENTIAL W2
```

Define:

```text
LIFECYCLE_CONFLICTS(S,T) =
    {
        (H,W)
        for each current STATUS head H
        for each W in MAXIMAL_UNRESOLVED_WITNESSES(S,T,H)
    }
```

The maximal-witness rule prevents one stale credential chain from producing a conflict for every historical revision. Incomparable unresolved credential branches may contribute separate witnesses.

v0 does not require one wire resolution per `(H,W)` pair. The pair set is the explainable conflict evidence presented to the user; one confirmed status assertion over the complete current context resolves every lifecycle conflict visible in that context.

---

## 37. Lifecycle-sensitive writes

Let `C = TOKEN_UPDATE(T)`, `B = BASE_STATE(C)`, and:

```text
PRE = LIFECYCLE_CONFLICTS(B,T)
```

An outstanding lifecycle conflict is persistent derived evidence, not a distributed lock and not a token-wide writer freeze. A writer that has observed `PRE != empty` MAY continue to assert token fields that do not make a lifecycle decision, subject to the ordinary field and credential/status rules.

In particular:

- `ISSUER` and `ACCOUNT` assertions remain ordinary field writes;
- a `CREDENTIAL` assertion does not by itself resolve or confirm any lifecycle conflict and remains subject to Section 41;
- any update that asserts `STATUS` while `PRE` is non-empty MUST be routed to the lifecycle-resolution validation procedure of Section 51. It becomes a lifecycle-resolution candidate only after every candidate-classification precondition in Sections 34 and 51 succeeds. A `STATUS` assertion combined with another token field in this situation is invalid and receives no candidate-only provisional coverage. A conforming writer whose update becomes a candidate additionally satisfies Section 39.

An ordinary non-`STATUS` update over a lifecycle-conflicted context inherits the unresolved lifecycle evidence through its context and field ancestry. Such an update MUST NOT be interpreted as confirming or resolving the conflict merely because it causally incorporates the conflicting history.

This rule is necessarily local to the writer's accepted context: another writer that has not yet observed the lifecycle conflict may continue authoring under its own accepted context. When the histories later meet, the lifecycle oracle derives the conflict from the authenticated ancestry. Lifecycle conflicts for token `T` do not by themselves block `DEVICE_UPDATE` or authorship of another token `U`.

---

## 38. Lifecycle-resolution status assertion

Let `C = TOKEN_UPDATE(T)`, `B = BASE_STATE(C)`, and `PRE = LIFECYCLE_CONFLICTS(B,T)`.

When `PRE` is non-empty and `C` asserts `STATUS`, the validator MUST route `C` to Section 51. `C` becomes a lifecycle-resolution candidate only after the candidate-classification preconditions in Sections 34 and 51 have succeeded; routing alone does not grant candidate status or provisional coverage.

The deterministic object-validity requirements are:

1. `C` asserts exactly one token field: `STATUS`;
2. the asserted status is `LIVE` or `TOMBSTONE`; and
3. after constructing the candidate status revision and applying the provisional candidate coverage of Section 34, the candidate post-state lifecycle-conflict set is empty.

These validity predicates are determined solely from the signed object and authenticated validated history.

A conforming **writer** of such a resolution additionally MUST:

- use its complete accepted observed current token-update frontier as the context-parent set, as required by Section 22; and
- obtain explicit user confirmation satisfying Section 39 immediately before publication.

Those are writer-conformance requirements, not retrospective consensus-validity predicates. A remote validator cannot prove what additional history was visible to the writer or whether a human actually supplied confirmation. Failure to prove either fact MUST NOT by itself make an otherwise deterministically valid signed historical object invalid. Later discovery of omitted or withheld history may instead create a new ordinary or lifecycle conflict.

An update with `PRE` non-empty that does not assert `STATUS` is not a lifecycle-resolution candidate and is validated by the ordinary field rules plus Section 41. It does not clear the lifecycle conflict merely by incorporating it into context.

The status assertion implicitly replaces all current status heads in `B`. A fully valid resolution therefore reconciles every currently visible status branch and every lifecycle witness incorporated by its signed context in one lifecycle decision.

A resolution MAY assert the same scalar status value already present on one or all current status heads. The new revision is still semantically meaningful because it records a fresh status decision over a broader incorporated context.

Later-discovered hidden or stale credential/status history that was not incorporated by `C` may create a new lifecycle conflict. Such later conflict does not retroactively invalidate `C`.

---

## 39. Explicit user confirmation and confirmation freshness

Before publishing a lifecycle-resolution update, a conforming client MUST obtain explicit user confirmation of the resulting token `STATUS` given all lifecycle conflicts currently visible in the accepted context.

The human confirmation is client-side security state and is not separately represented on wire.

The pending confirmation MUST be bound to at least:

```text
TOKEN_ID
desired resulting STATUS
current accepted TOKEN_UPDATE context frontier
current token field-head state
current lifecycle-conflict set
current locally abandoned pending TOKEN_UPDATE set for T
current durable authenticated-future observation set for the vault
current durable future-version bypass set for the vault
```

Immediately before constructing and publishing the resolution update, the client MUST recompute those inputs. Any relevant token-context, conflict, pending-recovery, or future-version-context change makes the confirmation stale and requires explicit confirmation again.

Future-version context freshness is event-sensitive, not merely final-set-sensitive. Once the durable future-observation or future-bypass context changes after confirmation, that confirmation remains stale even if later processing returns the sets to equal values. Implementations MAY use a monotonically advancing local future-context generation/epoch.

Token-local accepted context does not require an analogous event epoch: withdrawal of previously accepted token history triggers known-history regression and blocks authorship while the context is not reconstructible; once the same accepted context is reconstructible, no token history has been silently crossed.

An unpublished lifecycle confirmation is session-local and MUST NOT survive client restart. After restart a lifecycle-resolution update requires a new explicit user confirmation even if every reconstructed input equals its pre-restart value.

After durable publication, a client SHOULD revalidate the exact published resolution and recompute current lifecycle conflicts before presenting resolution as complete. Newly arrived history may immediately reveal a new conflict without invalidating the already-valid resolution.

---

## 40. Deletion and restoration

When no lifecycle conflict is outstanding, deletion and restoration are ordinary `STATUS` assertions. When lifecycle conflicts are outstanding, any update asserting `STATUS` is routed to lifecycle-resolution validation under Sections 38 and 51; only a supported status-only assertion satisfying the candidate preconditions becomes a lifecycle-resolution operation subject to Section 39:

```text
LIVE -> TOMBSTONE
TOMBSTONE -> LIVE
```

Deleted tokens remain in derived token state. Deletion does not erase issuer/account/credential history or old seed material.

Clients MUST NOT describe token deletion as cryptographic erasure. User interfaces SHOULD normally suppress unambiguously tombstoned tokens from active token views.

---

## 41. Credential writes and tombstones

`CREDENTIAL` may be asserted without `STATUS` only when every current status head in the base state has value `LIVE`. This rule applies whether or not lifecycle conflicts are currently outstanding.

A `CREDENTIAL` assertion does not by itself resolve or confirm a lifecycle conflict. Historical credential witnesses remain part of the lifecycle oracle even when a later credential revision descends from them.

If any current status head has value `TOMBSTONE`, an update asserting `CREDENTIAL` MUST also assert:

```text
STATUS = LIVE
```

which resolves all currently observed status alternatives as part of that same author event when no lifecycle conflict is outstanding.

If lifecycle conflicts are outstanding, the required `STATUS = LIVE` assertion would itself be a lifecycle-resolution assertion under Sections 38–39, which MUST be status-only. Therefore a credential write from a base state containing any `TOMBSTONE` status head cannot proceed until the lifecycle decision has first been resolved.

An update MUST NOT simultaneously assert `STATUS = TOMBSTONE` and a new `CREDENTIAL`.

---

## 42. Observed token-update heads

For token `T`, a fully valid `TOKEN_UPDATE C` is causally subsumed when another fully valid `TOKEN_UPDATE H` concerning `T` satisfies `C <c H`.

The observed current token-update/version heads are the fully valid updates not causally subsumed by any other currently available fully valid update for `T`.

This frontier is token-local. It is not serialized as a checkpoint and does not imply that each field has only one current revision head.

---

## 43. Observed token and vault state

For token `T`, define:

```text
OBSERVED_TOKEN_STATE(T) =
    JOIN(
        state(H)
        for every observed current TOKEN_UPDATE head H of T
    )
```

The aggregate vault view is the mapping from each observed `TOKEN_ID` to its independently derived observed token state.

Observed ordinary field conflicts and lifecycle conflicts are derived from this state plus authenticated token-update/field ancestry.

A client MAY cache per-update materialized states and an aggregate per-token frontier off-store. Such caches are disposable derived state and create no synchronized global frontier.

---

## 44. Observed device presentation state

For each device public key, derive current device-update heads using Section 21.

Normally one current `DEVICE_UPDATE` head remains and its `DISPLAY_NAME` is the current friendly name. Multiple incomparable current heads represent presentation alternatives; identical displayed names MAY be coalesced in UI while the signed heads remain distinct.

If no valid `DEVICE_UPDATE` is available for a device key observed in token updates, the device is unnamed for presentation purposes.

Conflict/audit UIs SHOULD identify authors cryptographically by public-key identity or stable fingerprint and MAY render the current observed friendly name. v0 does not define a cryptographically authenticated friendly name at the time of an earlier token event.

---

## 45. Writer token-context attestation

For `TOKEN_UPDATE(T)`, a conforming writer MUST first derive the complete accepted observed token state on which the user or application is acting.

The writer MUST set:

```text
CONTEXT_PARENTS(C) = OBSERVED_TOKEN_UPDATE_HEADS(T)
```

and canonically sort the raw object IDs.

The writer MUST verify that `BASE_STATE(C)` equals the accepted observed token state used to make the operation. Failure indicates incomplete/inconsistent local validation and MUST block authorship.

The signed context frontier is therefore an authenticated attestation of the complete observed token version frontier incorporated by the operation. It does not prove completeness against withholding.

A writer MUST NOT author from resource-limited or otherwise incompletely validated v0 state, token-local known-history regression, or un-abandoned signature-verified pending history that can affect `T`. Section 56.1 may bypass only the corresponding durable pending-history block.

The sole additional completeness exception remains Section 59.1: if the only known incompleteness is authenticated future-version observations and every such durable observation is explicitly bypassed, v0 may author from the understood v0 subset while continuing to disclose aggregate incompleteness.

---

## 46. Signer independence and token-local concurrency

Device identity does not create semantic ordering.

Two `TOKEN_UPDATE`s signed by the same key but concerning different tokens are unrelated. Concurrent same-token updates signed by the same key are treated exactly like concurrency between different signers.

A conforming client MUST serialize its own local writes to the same token against one accepted context frontier so it does not intentionally create avoidable concurrent branches. Writes to different tokens need not be serialized merely because they use the same signing key.

---

## 47. Device-key restore

Possession or restoration of a device private key creates no token-semantic predecessor requirement.

A restored key may sign `TOKEN_UPDATE(T)` whenever the client has a complete accepted context for `T` and all ordinary writer rules are satisfied.

Before changing device presentation, a restored client MUST derive the currently known `DEVICE_UPDATE` head set and use all maximal same-key heads as the new update's context parents. Pending same-key presentation history blocks a new `DEVICE_UPDATE` until it resolves or the client explicitly abandons that presentation branch using the durable presentation-recovery workflow of Section 49.1.

---

## 48. Validation of context parents

For every `TOKEN_UPDATE C` concerning token `T`, a validator MUST:

```text
validate every listed context parent
require every context parent is a TOKEN_UPDATE
require every context parent concerns TOKEN_ID T
require the context-parent set duplicate-free
require canonical raw-ID ordering
require no listed context parent is a strict context ancestor of another
```

A zero-length context-parent set is permitted; ordinary token-creation rules determine whether such an update is valid.

Missing context parents leave the dependent update `PENDING`. They do not authorize treating visible descendants as roots or rewriting ancestry.

No validity rule requires a predecessor signed by the same device key.

---

## 49. Validation of DEVICE_UPDATE

A validator of `DEVICE_UPDATE U` MUST:

```text
verify the signature using Sections 16 and 19
validate every listed DEVICE_UPDATE context parent
require every context parent is signed by U.DEVICE_PUBLIC_KEY
require the set duplicate-free and canonically ordered
require no context parent is a strict device-update ancestor of another
validate DISPLAY_NAME framing, UTF-8, and bounds
```

Multiple roots or later-discovered concurrent presentation heads are valid forked presentation history, not automatic invalidity.

Writer conformance requires the context-parent set to equal all currently known maximal same-key device-update heads. Pending same-key presentation ancestry blocks another presentation write until it resolves or is explicitly abandoned under Section 49.1.

`DEVICE_UPDATE` validation never derives or changes token state.

### 49.1 Pending device-presentation recovery

A signature-verified `DEVICE_UPDATE U` that is `PENDING` only because referenced presentation dependencies are unavailable is presentation-only security evidence. It never blocks `TOKEN_UPDATE` authorship or changes token validity, but by default it blocks a new `DEVICE_UPDATE` under the same `DEVICE_PUBLIC_KEY` because the complete same-key presentation frontier cannot yet be reconstructed.

Once such a pending same-key update is observed, a client that may later author presentation history under that key MUST durably remember at least `(OBJECT_ID(U), DEVICE_PUBLIC_KEY)`, bound to the established `VAULT_BINDING`, before it treats processing of that observation as complete or permits a later presentation write that would otherwise forget the branch after restart.

A full read/write client MAY provide an explicit user-directed abandonment workflow for one or more exact pending `DEVICE_UPDATE` object IDs. An abandonment decision MUST be exact-ID, device-key-bound, vault-bound, and durable before a new presentation write relies on it. It changes only the local writer-completeness policy for `DEVICE_UPDATE`; it MUST NOT classify the abandoned object as invalid, erase the pending observation, affect token semantics, or suppress the branch if it later becomes `FULLY_VALID`.

If an abandoned presentation update later becomes `FULLY_VALID`, it MUST re-enter observed device-presentation history normally and may create a presentation fork with updates authored during the abandonment. Pending presentation evidence and its abandonment record may be compacted only after either (a) a durably remembered accepted `DEVICE_UPDATE` head descends from the formerly pending update under `<=d`, or (b) the exact object is established identity-grounded terminally invalid under Section 56.2. Mutable-path failure or disappearance is insufficient.

---

## 50. Validation of ordinary TOKEN_UPDATE

Let `C = TOKEN_UPDATE(T)` and `B = BASE_STATE(C)`.

A validator MUST:

```text
validate context parents
verify the update signature
require exactly one TOKEN_ID
require at least one token field assertion
validate every asserted field structurally
construct implicit field-parent sets from B
compute typed field ancestry and state(C)
compute PRE = LIFECYCLE_CONFLICTS(B,T)
```

If `PRE` is non-empty and `C` asserts `STATUS`, Section 51 applies instead.

Otherwise the validator applies token-creation or ordinary field rules and the credential/tombstone restrictions of Section 41. In particular, a non-`STATUS` update may remain valid with `PRE` non-empty; it does not thereby confirm or clear the lifecycle conflict.

No serialized claimed token state, field-parent list, supersedes list, or conflict list exists to compare against. Validity follows from the signed context, asserted field values, and deterministic state-transition rules.

Writer-conformance omission of already-observed concurrent history does not retroactively make an otherwise self-consistent update invalid; later joining that history surfaces the appropriate ordinary or lifecycle conflict.

---

## 51. Validation of lifecycle-resolution TOKEN_UPDATE

Lifecycle-resolution validation is deliberately staged so that candidate coverage does not depend on already knowing the candidate is valid.

For a prospective resolution `C`, a validator MUST perform the following in order:

```text
validate envelope/object identity, structure, signature, and context parents
require every context parent fully valid
compute B = BASE_STATE(C)
compute PRE = LIFECYCLE_CONFLICTS(B,T) using only fully valid prior history
require PRE non-empty
require C asserts exactly STATUS
require STATUS is LIVE or TOMBSTONE
classify C as a lifecycle-resolution candidate
construct the candidate STATUS revision with all current STATUS heads in B as implicit parents
compute candidate state(C)
compute POST* using ordinary fully valid resolution coverage plus Section 34 candidate-only provisional coverage for C
require POST* is empty
only then classify C as FULLY_VALID lifecycle-resolution history
```

Candidate-only provisional coverage exists solely during `C`'s own final lifecycle check. If any final check fails, `C` is invalid and contributes no lifecycle coverage to any state or later object.

A validator can prove that the signed status assertion incorporates the authenticated context named by the object and satisfies the deterministic resolution transition. It cannot prove that the writer used every history actually visible to it or that an actual human supplied the explicit confirmation required of conforming writers by Sections 22, 38, and 39.

---

## 52. Field-lineage validity

For every asserted token field `F`, the validator MUST treat the complete current `F` head set in `BASE_STATE(C)` as the implicit parent set of `FIELD_REVISION(C,F)`. There is no alternate wire declaration to trust or reconcile.

A field revision whose implicit ancestry cannot be reconstructed because required context/dependencies are missing remains incomplete/PENDING through its containing update.

Typed field ancestry is derived only from validated context and assertions; file-arrival order, signer identity, and value equality MUST NOT create or remove ancestry edges.

---

## 53. Context and field conflicts are distinct

A token may have one current token-update/version head while one or more fields inside `state(H)` remain conflicted. Conversely, several current token-update heads may compose to one unambiguous value for a field because their field revisions affect disjoint fields or converge by ancestry.

Validators and UIs MUST NOT infer field-conflict absence merely from token-update frontier width, nor infer value ambiguity merely from the number of field revision heads when all current heads carry an equal field value.

---

## 54. Semantic publication consequences

An update is minted only when it introduces at least one new semantic field assertion. v0 has no graph-only merge/adoption update whose sole purpose is to collapse a token frontier.

An unrelated assertion may nevertheless collapse the token-update/version frontier by incorporating all current context parents while inherited field conflicts remain unresolved inside its derived state.

This rule keeps synchronized history semantic: context describes what was incorporated, while asserted fields describe the new facts introduced by the event.

---

## 55. Cycles and defensive traversal

Because object IDs cryptographically commit to plaintext that contains their references, deliberately constructing a true reference cycle would require solving a cryptographic fixed-point problem and is not an expected live protocol state.

Implementations SHOULD nevertheless use cycle-safe DAG traversal for causal and typed field graphs.

A traversal MUST distinguish:

```text
VISITING  = node currently on the active traversal path
VISITED   = node already fully processed
```

Encountering a `VISITED` node is normal in a DAG with shared ancestors and MUST reuse the prior result.

Encountering a `VISITING` node is a back-edge/cycle and MUST abort that dependency traversal safely.

Cycle checks are defensive parser hygiene rather than a primary protocol threat.

---

## 56. Validation stages and object arrival order


Objects may arrive in any order.

v0 distinguishes:

```text
ENVELOPE_AUTHENTICATED
AUTHENTICATED_UNSUPPORTED_FUTURE
STRUCTURALLY_VALID
SIGNATURE_VERIFIED
PENDING
FULLY_VALID
INVALID
```

| Stage | Meaning |
|---|---|
| `ENVELOPE_AUTHENTICATED` | AEAD and object-ID verification succeeded for this vault |
| `AUTHENTICATED_UNSUPPORTED_FUTURE` | envelope authentication succeeded and the authenticated common prefix names an unsupported `OBJECT_VERSION`; Section 59 governs the durable observation and writer block/bypass |
| `STRUCTURALLY_VALID` | canonical plaintext parses as a supported semantic object with valid local field structure |
| `SIGNATURE_VERIFIED` | the Section 16/19 Ed25519 profile verifies |
| `PENDING` | all checks applicable so far succeed, but full validation is blocked only by unavailable referenced dependencies |
| `FULLY_VALID` | all applicable dependencies, ancestry, lifecycle, and state-transition checks validate |
| `INVALID` | an applicable required check failed rather than merely being unavailable |

A supported-version `TOKEN_UPDATE` normally progresses:

```text
ENVELOPE_AUTHENTICATED
    -> STRUCTURALLY_VALID
    -> SIGNATURE_VERIFIED
    -> PENDING or FULLY_VALID
```

An authenticated unsupported semantic version instead terminates v0 semantic parsing at `AUTHENTICATED_UNSUPPORTED_FUTURE` as specified in Section 59; `INVALID` is the terminal disposition when an applicable supported-v0 check fails.

A `DEVICE_UPDATE` follows the same stages, with dependencies limited to its own listed device-update context parents.

A bad signature produces `INVALID`.

These stage/disposition labels describe a client's current validation observation of the bytes and dependencies presently available for an object identity. Mutable pathname state may later disappear, become unreadable, or expose different bytes. Such a later observation does not erase durable security evidence already established for a previously authenticated pending or future-version object; Sections 56.2, 59, and 63 govern that durable evidence separately.

A signature-verified pending `TOKEN_UPDATE(T)` indicates that authenticated history may still affect token `T`. By default, a conforming client MUST NOT author another `TOKEN_UPDATE(T)` while such pending history remains unresolved, regardless of which device key signed the pending object. This block is token-local and does not by itself prevent writing unrelated token `U`.

The default pending-history block may be overridden only by the explicit user-directed local recovery procedure in Section 56.1.

Once a client reaches `SIGNATURE_VERIFIED` and classifies `TOKEN_UPDATE(T)` as `PENDING` because referenced dependencies are unavailable, the fact that this authenticated pending update exists is durable security evidence. Before the client may regard processing of that observation as complete, accept a token frontier that excludes it, or author `T`, it MUST durably record at least the pending update's `OBJECT_ID` and `TOKEN_ID`, bound to the established `VAULT_BINDING`. Disappearance of the object file, restart, or loss/rebuild of the disposable validation cache MUST NOT silently erase this pending-history block. The durable evidence MUST NOT be erased merely because validation later reaches `FULLY_VALID` or because the current storage representation becomes unreadable or invalid-looking. It remains security evidence until one of the durable replacement conditions in Section 63 is satisfied. Explicit abandonment under Section 56.1 changes whether the evidence blocks authorship; it does not erase the fact that the authenticated pending object was previously observed.

A pending or missing `DEVICE_UPDATE` affects only presentation completeness and does not block token authorship.

AEAD authentication plus a claimed public-key field is not device provenance.

Objects may arrive before dependencies, for example:

```text
C19 TOKEN_UPDATE(T) arrives before one context parent
C19 arrives before a required context dependency
C19 arrives before an update needed to reconstruct inherited field ancestry
U7 DEVICE_UPDATE arrives before U6
```

Pending objects MUST NOT affect validated token state or validated presentation state respectively.

Missing dependencies indicate incomplete synchronization, not necessarily corruption.

There are no independently stored token-field objects: field revisions exist only as semantic roles of their signed containing `TOKEN_UPDATE`.

Durable pending evidence is intentionally conservative and is not protocol-bounded: a party able to author authenticated history can create many distinct updates whose dependencies remain unavailable, causing local pending-evidence growth. Section 56.1 permits one explicit recovery action to acknowledge multiple pending object IDs for the same token, but batching the recovery decision does not erase the underlying durable evidence. Read-only clients need not expose abandonment controls; if their retained state may later be reused for writing, the writer-side durable-evidence and rescan requirements still apply before authorship.

### 56.1 Explicit recovery from indefinitely pending token history

The default pending-history block protects a writer from authoring token `T` while authenticated history that may affect `T` is not yet reconstructible. Because an untrusted synchronization medium may withhold dependencies indefinitely, and because a holder of `K_root` can create a signature-valid update referring to unavailable history, this block MUST NOT create an unrecoverable token-level liveness failure.

A full read/write v0 client SHOULD therefore provide an explicit user-directed recovery workflow that abandons waiting for one or more `TOKEN_UPDATE(T)` object IDs for which the client holds durable evidence that they were previously `SIGNATURE_VERIFIED` and `PENDING` because referenced dependencies were unavailable. A constrained or deliberately read-only implementation MAY omit this workflow, but when the condition is encountered it MUST clearly surface/document that the affected token cannot become writable in that client without another capable client, explicit migration/reconfiguration, or equivalent supported recovery. The object bytes need not still be present at recovery time; durable pending evidence exists specifically so disappearance cannot either erase the block or make explicit recovery impossible.

This recovery decision is local security/recovery state. It is not a synchronized semantic object, does not change the validation status of the pending update, does not declare the update invalid, and does not erase or rewrite authenticated history.

For each configured vault, the client MUST bind the locally abandoned pending-update set to the established `VAULT_BINDING` and `TOKEN_ID`, and MUST persist that recovery decision durably before allowing authorship that relies on the abandonment. An abandonment decision that permits a write MUST survive restart until the underlying pending security evidence has been durably handed off to an accepted frontier, the exact object has been established terminally invalid under Section 56.2, or a broader explicit recovery/migration supersedes the local state.

While a listed update remains `PENDING`, it no longer participates in the pending-history writer block for that client's accepted frontier of `T`. The client MAY then author `TOKEN_UPDATE(T)` from the fully validated accepted token frontier, provided all other writer requirements are satisfied.

The recovery workflow MUST:

- require explicit user direction; it MUST NOT occur automatically because of age, timeout, retry count, or resource pressure;
- identify the affected token and the pending update or updates being abandoned from the writer-completeness wait;
- visibly warn that authenticated but incomplete token history is being bypassed and may later become valid;
- operate only on exact object IDs for which durable security memory records a prior `SIGNATURE_VERIFIED` and `PENDING` observation due to unavailable referenced dependencies, and for which no durable replacement condition from Section 63 has yet completed; the synchronized object MAY currently be absent or unreadable;
- MUST NOT be used to bypass `DEGRADED_KNOWN_HISTORY_LOSS`, missing history that was previously accepted as fully valid, or resource-limited/incomplete v0 validation; pending-history recovery does not itself bypass authenticated future-version data, which requires its own independent durable Section 59.1 recovery decision;
- leave newly discovered pending updates blocked unless they are separately included in an explicit recovery decision.

If an abandoned pending update later becomes `FULLY_VALID`, the local abandonment ceases to suppress it automatically. The update MUST enter observed token history normally and may create ordinary or lifecycle conflicts with history authored during recovery. The durable pending/abandonment evidence is retained until the accepted-frontier handoff in Section 63 durably covers that update.

If the update is established to be terminally invalid by the identity-grounded rule of Section 56.2, it no longer requires abandonment because that exact immutable object ID has been disproven as valid history. A missing, truncated, substituted, unreadable, wrong-length, authentication-failing, or object-ID-mismatching current file is not such terminal invalidity and MUST NOT clear the remembered evidence.

A client MAY let one explicit recovery action select multiple currently pending updates for the same token. This is a usability batching mechanism only; the persisted recovery state remains the specific set of pending object IDs acknowledged by the user.

A token being authored under this recovery mode MUST initially remain visibly marked as having bypassed unresolved authenticated history. For an abandoned update that remains unresolved indefinitely, the user MAY explicitly acknowledge and finalize the local recovery warning for that exact pending object ID. This acknowledgement is local, vault-bound, durable recovery state. It MAY allow the client to stop presenting a persistent alarm for that already-acknowledged bypass, but it MUST NOT erase the durable pending evidence, erase the abandonment/bypass decision, declare the update invalid, or prevent the update from entering observed history if it later becomes fully valid. Audit/recovery UI MUST remain able to disclose that unresolved authenticated history was bypassed. Newly discovered or newly abandoned pending objects require their own warning/acknowledgement context.

If every abandoned pending update is later durably incorporated into an accepted frontier or established terminally invalid under Section 56.2, the corresponding recovery-warning acknowledgement state MAY be compacted after that stronger durable evidence is in place. Losing the synchronized file or merely removing an object ID from a UI list never clears the underlying security evidence.

### 56.2 Identity-grounded terminal invalidity

Durable evidence for a previously authenticated pending object MUST NOT be cleared merely because the bytes currently found at its pathname are missing, truncated, wrong-length, substituted, unreadable, fail AEAD authentication, or recompute to a different `OBJECT_ID`. Such observations show only that the current storage representation is unavailable or untrustworthy; they do not disprove the immutable object that was previously authenticated.

For purposes of clearing durable pending evidence, terminal invalidity must be **identity-grounded**. The client must have cryptographic evidence tied to the exact remembered object identity sufficient to establish that the object cannot become valid history. This requires either:

- the exact object bytes successfully authenticate under the remembered filename/object ID, recompute to that same `OBJECT_ID`, and then fail a deterministic required structural, signature, ancestry, lifecycle, or state-transition validity rule once all dependencies needed for that determination are available; or
- a referenced dependency is itself established terminally invalid using the same identity-grounded rule, making the remembered pending object deterministically invalid.

A transient read failure, missing dependency, authentication failure of newly observed bytes, resource exhaustion, or unavailable ancestry is never identity-grounded terminal invalidity. Those cases leave the remembered pending evidence unresolved.

Implementations MAY cache more detail to explain terminal invalidity, but the decision to clear durable pending evidence MUST be reproducible from authenticated, identity-matching object evidence rather than from mutable pathname state alone.

---

## 57. Missing ancestry is unknown, never rewritten

An implementation may be unable to establish an update's complete causal or typed field ancestry because referenced objects are missing or pending.

In that situation, ancestry-dependent facts are unknown/incomplete until the required objects are available.

Missing ancestry MUST NOT be interpreted as evidence that missing ancestors did not exist, that a visible node is a new root, or that two histories are concurrent.

v0 has no shallow-history, graft, replacement-parent, or history-rewrite semantics.

Implementations MUST NOT locally declare:

```text
"treat this update/revision as a root"
"pretend this parent edge does not exist"
"replace this parent with another parent"
```

in order to make incomplete history validate or merge.

Local caches, snapshots, or indexes are permitted only when semantically equivalent to the authenticated object graph and rebuildable from it.

---

## 58. Invalid and unrelated files

The object directory is controlled by an untrusted medium.

A file may:

- fail AEAD authentication;
- have the wrong object ID;
- be malformed;
- be unrelated to any valid device history;
- be temporarily incomplete while synchronization is in progress.

Such files MUST NOT automatically invalidate otherwise independent valid vault history.

A valid update depending upon an invalid or unavailable object cannot itself become fully valid.

---

## 59. Unsupported authenticated objects and immutable v0 grammar


A candidate pathname representation that fails envelope authentication does not reach `ENVELOPE_AUTHENTICATED` and cannot be classified as authenticated future data. For current processing it is unavailable or untrustworthy as specified in Section 15. Such an authentication failure alone is not identity-grounded terminal invalidity under Section 56.2 and MUST NOT clear durable evidence for a previously authenticated object identity.

After successful envelope authentication, the common semantic prefix is interpreted before version-specific grammar. Handling then depends first on `OBJECT_VERSION`:

- If `OBJECT_VERSION != 0`, the authenticated object is `AUTHENTICATED_UNSUPPORTED_FUTURE`. It may have been created by a newer implementation. A v0 implementation MAY continue displaying state it can validate, MUST clearly indicate that unsupported authenticated future data exists, and MUST NOT claim that its derived `OBSERVED_STATE` or `OBSERVED_DEVICE_STATE` is complete. By default it MUST NOT create new `TOKEN_UPDATE` or `DEVICE_UPDATE` history while any authenticated future-version observation remains unprocessed and un-bypassed. The only v0-local exception is the explicit durable future-version recovery procedure of Section 59.1.
- If `OBJECT_VERSION == 0`, the immutable v0 grammar applies. An unknown `OBJECT_TYPE`, unknown field/tag, unknown enum value, or other representation not accepted by the frozen v0 grammar is `INVALID`, not authenticated-future data. Such an otherwise unrelated invalid object does not globally block writing.

Version precedence is therefore normative: an unsupported semantic version is classified as authenticated future data before interpreting that version's object type; an unknown object type inside supported version 0 is invalid v0 data.

Once an established client authenticates an unsupported future-version object, that observation becomes durable vault-bound security evidence. Before processing of that observation may be considered complete, the client MUST durably retain at least the object's `OBJECT_ID` and authenticated `OBJECT_VERSION`, bound to the established `VAULT_BINDING`. Disappearance of the file, restart, rollback of the synchronized store, or loss/rebuild of disposable caches MUST NOT clear the durable future-version observation and, absent a durable Section 59.1 bypass for that exact observation, MUST NOT clear its default v0 write block.

A v0 implementation MUST NOT clear such future-version evidence merely because the object is no longer present. Granting a Section 59.1 bypass changes whether that exact durable observation blocks v0 writing; it does not erase or reinterpret the future object. The evidence may be cleared or superseded only by an implementation that understands the authenticated semantic version and, under the rules of that version, durably establishes a compatible state that explicitly permits such clearing; or by an explicit migration/reconfiguration to a different cryptographic vault.

### 59.1 Explicit local bypass of authenticated future-version history

The default future-version block is deliberately fail-closed because a v0 implementation cannot interpret the future object's semantics. However, making that block permanently non-bypassable would allow one authenticated future-version object — including a stray experimental object created by a `K_root` holder — to impose an unrecoverable write denial on every established v0-only client. v0 therefore permits an explicit local recovery decision analogous to pending-history recovery.

A full read/write v0 client SHOULD provide a user-directed workflow that bypasses the v0 writer block for one or more exact durable `AUTHENTICATED_UNSUPPORTED_FUTURE` object observations. A constrained or deliberately read-only implementation MAY omit this workflow, but when authenticated unsupported future data is encountered it MUST clearly surface/document that v0 writing remains blocked in that client unless a compatible client, explicit migration/reconfiguration, or another supported recovery path resolves the condition. The bypass:

- MUST identify the exact `OBJECT_ID` and authenticated `OBJECT_VERSION` being acknowledged;
- MUST be bound to the established `VAULT_BINDING` and persisted durably before any v0 authorship relies on it;
- MUST warn that the v0 client cannot interpret that history, that its view may be incomplete, and that v0 history authored during the bypass may later conflict with or be superseded by semantics understood by a compatible future implementation;
- MUST NOT declare the future object invalid, erase its durable observation, infer its token/device scope, or claim that the v0 aggregate view is complete;
- applies only to the exact future observations acknowledged by the user; any newly authenticated future-version object without a durable bypass re-establishes the global v0 write block.

When every durable future-version observation is either understood by compatible processing or covered by a durable explicit bypass, the v0 client MAY author `TOKEN_UPDATE` and `DEVICE_UPDATE` history under the narrow incompleteness exception defined by Section 45. All other ordinary v0 writer rules remain in force: future-version bypass does not relax resource-limit, pending-history, known-regression, accepted-frontier, lifecycle-confirmation, or publication requirements. This is a liveness recovery choice, not a statement that the future history is semantically irrelevant. Because v0 token updates cannot context-reference an unsupported future-version object as a valid v0 parent, continuation under bypass necessarily proceeds from the understood v0 subset.

A durable future-version bypass MUST survive restart while it is relied upon. If a compatible implementation later understands a bypassed future object, that object's actual semantics take effect normally; the bypass MUST NOT suppress resulting history, conflicts, migration requirements, or other compatible-version processing. Future evidence and bypass state may be compacted only under the universal durable-replacement rule of Section 63.

The accepted grammar and semantic interpretation of `OBJECT_VERSION=0` are immutable.

A later specification MUST NOT retrospectively declare a new field, enum value, object type, or changed interpretation to be valid inside `OBJECT_VERSION=0`.

Any semantic extension requires a new semantic object version.

Future semantic objects intended to trigger v0 authenticated-unknown handling MUST remain decryptable/authenticatable using the v0 cryptographic object envelope and MUST fit within the v0 pre-authentication framing limits.

A future format requiring a different envelope, different object-key derivation, or objects exceeding v0 framing limits requires an explicit bootstrap/format migration policy and cannot rely on old clients detecting it through the encrypted common prefix.

---

## 60. Fresh bootstrap scan and resource limits


A new client has no local object index.

Because filenames reveal no semantic type beyond being object candidates, a fresh client may need to inspect every candidate object: authenticate/decrypt, classify, verify signatures, resolve dependencies, validate updates, derive token-local causal and typed field graphs, derive per-token update frontiers/state, derive device-update presentation, and evaluate conflicts.

Object discovery requires inspection of each candidate object, but total validation complexity is not merely `O(number of objects)`.

Total work depends on total bytes, graph structure, per-token revision depth and width, lifecycle-conflict analysis, presentation-history structure, and number of observed tokens/frontier heads.

An implementation MAY enforce local CPU, memory, wall-clock, or work budgets for safety and usability.

Exhausting such a budget means:

```text
validation incomplete / resource-limited
```

It MUST NOT mean:

```text
unprocessed object = invalid
no conflict exists
partial view = complete
```

A resource-limited client MUST visibly indicate incomplete validation. It MUST NOT author a token whose accepted observed state may be incomplete, and MUST NOT claim the aggregate vault view is complete.

Serialized-object limits bound individual semantic objects and parent lists; they do not bound the total number of historical object files in storage.

---

## 61. Incremental synchronization and local cache


An established client SHOULD maintain a local cache containing known object IDs, validated object bytes or trustworthy replacement-detection metadata, token-local context-parent relationships, device-update context-parent relationships, typed token-field ancestry relationships, derived per-update token state, per-token current update heads, device-update heads, observed device presentation state, lifecycle-conflict witnesses, and detailed pending-dependency traversal state.

The detailed dependency graph and decrypted/parser state for a pending object may be disposable cache. However, the security fact that a specific signature-verified pending `TOKEN_UPDATE` exists for a specific `TOKEN_ID` is not disposable once learned; its minimum durable representation is governed by Sections 56 and 63.

A client MAY additionally maintain an aggregate off-store frontier/index spanning all tokens for efficient change detection and UI reconstruction. Such an index is derived local state, is not synchronized protocol state, and MUST be rebuildable from authenticated objects.

Normally only newly appearing files need decryption and processing, but an implementation MUST also detect when bytes at an already-known object filename have changed, disappeared, or become unreadable.

The derived cache is logically disposable, but that does not imply that persisting decrypted cache contents in plaintext is safe.

Implementations MUST define and enforce local confidentiality for:

- persisted decrypted `CREDENTIAL.SECRET_BYTES` material;
- `K_root`;
- device private signing keys;
- temporary plaintext files;
- backups;
- crash dumps and diagnostic logs.

Diagnostics MUST redact TOTP seed bytes (`CREDENTIAL.SECRET_BYTES`) by default.

A conforming client MAY choose not to persist decrypted secrets at all.

If the implementation retains a previously validated local immutable object copy, it MAY continue using that copy and SHOULD repair/republish the synchronized object when appropriate.

If it no longer retains trustworthy bytes for a required object, every dependent update for that token becomes unavailable for validation.

A cached authentication failure MUST be reconsidered if underlying file bytes later change.

Derived caches are not authoritative. Durable security memory from Sections 10 and 63 is intentionally not disposable in the same sense.

---

## 62. Freshness limitation


For token `T`, `OBSERVED_TOKEN_STATE(T)` means the join of all valid current token-update heads for `T` discoverable in the object set presently available to this client.

`OBSERVED_STATE` is the aggregate mapping of those per-token observed states.

Neither proves that no newer object exists elsewhere.

An untrusted synchronization medium may withhold newer token-update or device-update history.

A fresh client cannot distinguish a complete store from a valid older subset of the store.

A token update's context-parent set proves only that the named token versions/states were incorporated into that operation. It does not prove that the writer had globally complete knowledge.

The explicit confirmation rules similarly bind a lifecycle decision only to the accepted observed context available to the confirming client; later-discovered concurrent history may create new conflicts.

Observed friendly device names are likewise freshness-limited. A fresh client cannot prove that a visible `DEVICE_UPDATE` head is the newest historical update.

---

## 63. Durable security memory and remembered token frontiers


An established client MUST maintain durable security memory separate from its disposable derived cache. This security memory MUST be local state outside the hostile synchronized vault namespace, or be protected by an independent integrity/freshness mechanism that prevents the synchronization medium from rolling it back, replacing it, or deleting it without detection. v0's rollback and remembered-evidence guarantees assume the integrity and freshness of this local security memory relative to the synchronization-medium attacker.

Restoring an older local-device backup, deleting local security state, or otherwise resetting this memory can discard evidence of observations made after that snapshot. Unless an implementation has a separate trusted continuity mechanism proving otherwise, it MUST treat such a reset as loss of the prior local rollback guarantees and MUST NOT present the resulting state as though the former remembered-history guarantees were preserved.

That durable memory MUST include at least:

- the established `VAULT_BINDING`;
- for each `TOKEN_ID` whose semantic history the client has previously accepted, the last accepted current `TOKEN_UPDATE` head set for that token;
- for every signature-verified pending `TOKEN_UPDATE` whose unresolved existence has been observed, at least its `OBJECT_ID` and `TOKEN_ID`, until a durable replacement condition below is satisfied;
- every pending-history abandonment decision currently being relied upon to permit token authorship;
- every finalized/acknowledged pending-recovery warning state currently retained for audit/recovery presentation;
- for every authenticated unsupported-future-version object previously observed, at least its `OBJECT_ID` and authenticated `OBJECT_VERSION`, until compatible processing or migration/reconfiguration satisfies Section 59;
- every future-version bypass decision currently being relied upon to permit v0 authorship;
- for every signature-verified pending `DEVICE_UPDATE` whose same-key presentation-writer block may matter, at least its `OBJECT_ID` and `DEVICE_PUBLIC_KEY` until a durable replacement condition in Section 49.1 is satisfied;
- every pending-device-presentation abandonment decision currently being relied upon to permit `DEVICE_UPDATE` authorship.

For purposes of v0 security semantics, a newly reconstructed token frontier becomes **accepted** only after the corresponding remembered-head evidence has been durably persisted. Merely validating or provisionally displaying a newly discovered frontier from disposable memory does not make it accepted. A client MUST NOT use a newly discovered frontier as the base for token authorship until that acceptance persistence step has completed.

The durability requirement constrains ordering, not persistence granularity. An implementation MAY batch remembered-head advances for multiple tokens, device-presentation advances, pending/future evidence updates, or related recovery-policy records into one durability operation. The batch MAY be implemented as an atomic durable transaction, but a shared flush or other common durability barrier does **not** by itself make the individual record updates transactionally atomic. If the batching mechanism is non-atomic, every individual replacement/removal in the batch MUST still satisfy Section 63.1: prior durable evidence MUST remain in force until the evidence that supersedes it has itself become durable. No newly covered state may be treated as accepted, writer-eligible, successfully acknowledged, or safely compacted before its own required durable ordering condition is satisfied. A crash partway through a non-atomic batch may therefore leave a mixture of old and new records, but MUST leave conservative evidence for every replacement whose successor was not yet durable.

For token `T`, the remembered update-head set is an antichain of token-local causal heads that were current when last accepted. It is security evidence, not authoritative token state.

Durable authenticated-future evidence is vault-wide security evidence. By default it blocks v0 writing. A durable Section 59.1 bypass may change that blocking policy for the exact acknowledged observations, but neither ordinary v0 rescanning, file absence, cache rebuild, nor the bypass itself erases the underlying future-version evidence.

Durable pending evidence is likewise security evidence, not token state. It records that the client previously authenticated a signed update for `T` whose full validity could not yet be determined because dependencies were unavailable. The underlying object bytes and detailed missing-dependency graph need not be retained locally if the synchronized object later disappears, but the remembered `(OBJECT_ID, TOKEN_ID)` security fact MUST survive.

### 63.1 Universal durable-replacement ordering

All durable security-memory compaction follows one rule: **a durable security item MUST NOT be removed, replaced, or compacted until the evidence intended to supersede it has itself been durably persisted.** If the two changes cannot be made atomically, the stronger/new evidence MUST become durable first. A crash between those steps therefore leaves at least the older conservative evidence in place.

This rule applies at least as follows:

- a remembered token head `X` may be dropped only after a descendant/current accepted head `H` with `X <=c H` has been durably persisted in `REMEMBERED_TOKEN_HEADS(T)`;
- a remembered `DEVICE_UPDATE` head `U` may be dropped only after a durably remembered replacement update `U2` with `U <=d U2` exists;
- pending-update evidence may be dropped only after the accepted-frontier or identity-grounded-terminal-invalidity handoff below is durable;
- authenticated future-version evidence is not removed merely because a bypass exists; it may be cleared only after compatible-version processing or migration/reconfiguration has durably superseded the observation according to Section 59;
- recovery/bypass policy records that are still relied upon for authorship or audit presentation MUST remain durable until a stronger state makes them unnecessary.

For pending token history, one of the following MUST hold before its durable pending record is removed:

1. **Accepted-frontier handoff.** The update has become `FULLY_VALID`, and the client has durably persisted an accepted current token-head set containing some head `H` such that the formerly pending update `P` satisfies `P <=c H`. The accepted frontier therefore durably remembers the valid history. Frontier persistence MUST occur before, or atomically with, removal of the pending record. Merely reaching `FULLY_VALID` in volatile memory is insufficient.
2. **Identity-grounded terminal invalidity.** The update has been established terminally invalid under Section 56.2. Mutable-path failures such as disappearance, truncation, unreadability, authentication failure, or substitution are insufficient.

For pending device-presentation history, the analogous handoff rule from Section 49.1 applies: a pending `DEVICE_UPDATE P` may be removed only after a durably remembered accepted presentation head `H` satisfies `P <=d H`, or after identity-grounded terminal invalidity. Presentation abandonment alone does not erase the observation.

Explicit abandonment does not itself erase the pending security fact. It durably changes the local policy from blocking to bypassed for that exact object ID while the evidence remains unresolved. If the abandoned object later becomes fully valid, accepted-frontier handoff is still required before the durable record can be removed.

Let:

```text
REMEMBERED_TOKEN_HEADS(T)
```

be the previously persisted accepted token-update head set and:

```text
ACCEPTED_TOKEN_HEADS(T)
```

be a newly accepted current token-update head set.

A remembered head `X` is safely superseded when some `H` in `ACCEPTED_TOKEN_HEADS(T)` satisfies:

```text
X <=c H
```

When accepting the new token state, the client MAY replace remembered heads reflexively incorporated by new accepted heads only under Section 63.1: the replacement accepted-head evidence MUST already be durable, or be persisted atomically with removal of the older remembered head. It MUST NOT discard remembered semantic-history evidence merely because an update ceases to be a current head.

Because device friendly names may be used in conflict/audit presentation, an established client SHOULD also retain the last accepted current `DEVICE_UPDATE` head set for each known device identity. This is presentation rollback evidence, not token-state authority.

Remembered device-update heads may be compacted when a newly accepted update descends from them under `<=d`, but only after the descendant presentation-head evidence has become durable, or atomically with that durable replacement, as required by Section 63.1.

Durable semantic-head evidence grows approximately with:

```text
number of token IDs
+
number of unresolved concurrent token-update heads
```

Presentation-head evidence grows with device identities and unresolved device-update forks.

v0 has no token-history compaction, device revocation, or identity retirement, so these sets are not protocol-bounded by a constant over the lifetime of a vault.

---

## 64. Known-history regression and token-local degraded mode


For token `T`, let:

```text
CURRENT_RECONSTRUCTIBLE_TOKEN_HEADS(T)
```

be the set of currently available, fully valid current `TOKEN_UPDATE` heads for `T` reconstructible from synchronized objects plus retained validated local copies.

A remembered semantic update `X` remains represented iff:

```text
exists H in CURRENT_RECONSTRUCTIBLE_TOKEN_HEADS(T)
such that
    X <=c H
```

Known semantic-history regression exists for token `T` iff some `X` in `REMEMBERED_TOKEN_HEADS(T)` is not represented by any current reconstructible token head.

This is a closure-membership test, not a current-head-ID equality test.

If `CURRENT_RECONSTRUCTIBLE_TOKEN_HEADS(T)` is empty while `REMEMBERED_TOKEN_HEADS(T)` is non-empty, semantic-history regression exists for that token.

If known semantic-history regression exists, the client MUST place that token in:

```text
DEGRADED_KNOWN_HISTORY_LOSS
```

For that token:

- the client MUST visibly report known rollback/incomplete semantic-history status;
- it MUST NOT present a regressed available-state calculation as an ordinary complete current token state;
- it MUST NOT author new `TOKEN_UPDATE(T)` history based on that regressed state;
- it MAY permit read-only inspection under an explicit degraded warning;
- it MUST NOT silently resurrect a previously deleted token or older seed as current;
- restoring writability of the original `TOKEN_ID` requires restoration/republishing of enough previously accepted validated history to make every remembered head represented again; v0 defines no destructive same-vault reset that simply clears remembered accepted heads or reclassifies known-history loss as pending history.

Known history loss for token `T` does not by itself invalidate or block mutation of unrelated token `U` whose history is complete and non-regressed. The aggregate vault UI MUST nevertheless make any degraded token state visible and MUST NOT claim the entire vault is fully healthy while such degradation exists.

`DEVICE_UPDATE` rollback is evaluated independently when remembered presentation heads are available. If a remembered update head is no longer represented by any current update head under `<=d`, the client MUST visibly flag device-presentation history as regressed or incomplete and SHOULD prefer fingerprint-first identity presentation. Presentation-history regression alone MUST NOT change token-state validity.

This concerns history the client previously knew and accepted. It does not solve withholding of history a client has never observed.

Unrelated invalid garbage files do not by themselves trigger degraded mode.

An explicit recovery workflow for known-history loss MUST NOT clear or rewrite `REMEMBERED_TOKEN_HEADS(T)`, invent or remove authenticated parent edges, reclassify previously accepted missing history as merely pending, or omit fully valid current heads to make an operation encodable. If the missing accepted history cannot be restored, a client MAY let the user select recoverable values and materialize them as a **new logical token with a fresh `TOKEN_ID`**, or migrate selected values into a new vault under Section 65.1. Such materialization starts new history and MUST NOT be presented as restoration or continuation of the degraded original token; the original `TOKEN_ID` remains degraded unless its remembered history becomes reconstructible again.

---

## 65. Garbage collection, erasure, growth, capacity, and migration

v0 defines no garbage collection of valid historical objects.

Historical token updates and device updates are retained indefinitely.

Deleted tokens remain derivable from history and are not securely erased.

Unlike revision 9, an update does not serialize the complete token state, so protocol storage does not grow as `O(number of updates × token-state size)` merely because every update repeats the vault snapshot.

Storage growth is primarily proportional to the number and size of actual signed operations plus retained historical graph structure.

There is no serialized lifetime-token-entry limit imposed by a repeated `TOKEN_STATE` field.

Finite per-object limits still apply. In particular, final v0 limits on token/device context-parent counts, strings, credential-secret length, and other semantic fields may make a pathological operation impossible to encode if its required context frontier exceeds those bounds.

State or conflict history MUST NOT be truncated to force such an operation to fit.

A client that cannot encode a required operation within v0 limits is capacity-blocked for that operation and MUST use an explicit recovery/migration workflow rather than silently omitting required context or semantic assertions.

### 65.1 v0 migration procedure

Migration creates a new cryptographic vault and makes no continuity or retirement claim with respect to the old vault.

A migration client:

1. derives the best available `OBSERVED_STATE`;
2. requires the user to resolve/select any conflicts for credentials to migrate;
3. selects desired live logical tokens;
4. creates a new `VAULT` with a fresh `K_root`;
5. creates fresh device signing identities for clients joining the new vault;
6. assigns fresh `TOKEN_ID`s;
7. materializes each selected token as a new root `TOKEN_UPDATE` asserting `STATUS=LIVE`, `ISSUER`, `ACCOUNT`, and `CREDENTIAL`;
8. does not copy deleted tokens, tombstones, historical field revisions, lifecycle-resolution history, or old update history.

Conflict selection performed for migration is local migration input, not a semantic update to the source vault. A client MUST NOT require the selected values to be published first as a source-vault conflict resolution. In particular, when the source operation is capacity-blocked because its required frontier exceeds the v0 parent limit, migration may still select from the best available validated source state and create fresh destination history without first encoding the impossible source update.

Migration starts a new history; it is not v0 garbage collection.

v0 contains no old-vault retirement marker. Devices still configured for the old vault may continue creating valid old-vault history after migration.

A migration UI MUST direct the user to move/reconfigure every participating client onto the new vault and stop using/remove the old vault configuration from clients and synchronization endpoints.

Migration is not erasure. Old objects, deleted seeds, old bootstrap copies, and provider/version history persist until removed, and storage providers/backups may retain them even afterward.

Readers MUST reject any semantic object file whose size is not exactly 2048 bytes before AEAD processing.

---

## 66. Privacy properties and object-encryption rationale

Before unlocking, a storage observer directly sees bootstrap size/replacement activity, the number of opaque object files, object creation timing, total vault size, and equality of identical object IDs within that vault.

Semantic values such as device names, token IDs, issuer/account strings, credential secrets, context references, field assertions, signatures, and the unpadded semantic length of an object are not stored in plaintext.

Every encrypted semantic object file is exactly 2048 bytes. A storage observer therefore cannot distinguish `TOKEN_UPDATE` from `DEVICE_UPDATE`, field-assertion patterns, or simple from moderately complex valid updates by exact semantic-object file length.

This deliberately trades storage and synchronization efficiency for reduced size leakage. Even a small semantic object that would otherwise encode to only a few hundred bytes occupies 2048 bytes in the synchronized object store. For scale, 1,000 semantic objects occupy approximately 2 MiB before filesystem, synchronization-provider, or version-history overhead.

Fixed-size padding does not provide traffic-analysis resistance. Object counts, creation timing, equality/repetition, bootstrap activity, and total storage growth remain visible and may permit statistical inference of vault activity and approximate history size.

### 66.1 Deterministic AES-GCM rationale

v0 intentionally derives:

```text
OBJECT_ID = HMAC(K_id, canonical_plaintext)
K_object = HKDF(..., full OBJECT_ID)
nonce = first 12 bytes of OBJECT_ID
```

The construction relies on the invariant that one derived `K_object` encrypts only the one exact canonical padded encryption plaintext deterministically associated with the semantic plaintext that produced its full `OBJECT_ID`.

Distinct canonical semantic plaintexts are expected to produce distinct full object IDs and therefore distinct per-object keys.

A collision in the 96-bit nonce prefix across two distinct object IDs does not by itself constitute GCM nonce reuse because the encryption keys differ.

Repeating an identical canonical semantic object repeats the same ID, canonical padded envelope, key, nonce, AAD, and ciphertext, which is consistent with content addressing.

Random, alternative-size, or otherwise non-canonical padding is forbidden because it would permit different encryption plaintexts under the same per-object key and nonce for one semantic object ID.

The post-decryption keyed-ID recomputation is mandatory.

Implementations MUST preserve the exact plaintext → ID → key → nonce → encryption order.

---

## 67. Security meaning of signatures, context parents, and auditability

A valid semantic-object signature proves only that the named device private key signed the exact vault-bound object bytes accepted by v0. It does not prove device authorization, human identity, freshness, or global completeness.

For `TOKEN_UPDATE`, context parents prove which token versions/states the author event explicitly incorporated. Context edges induce token-local causal ordering, but they cannot prove that the synchronization medium did not withhold another concurrent update.

Per-field ancestry is derived from the signed context plus the fields asserted by each update. If update `C` asserts field `F`, every current `F` head in `BASE_STATE(C)` is implicitly superseded by the new field revision. If `F` is absent, its complete possibly-conflicted state is inherited.

The combination provides explainable provenance: a client can identify the update that authored each current field revision and can reconstruct the context in which a later field assertion resolved previous alternatives. Implementations MAY cache this derived provenance; it need not be redundantly serialized in every update.

`DEVICE_UPDATE` uses the same context concept for presentation history. Because v0 has only one device field (`DISPLAY_NAME`), each device update asserts a new name revision over its complete observed device-update context.

No signature or context mechanism establishes authorization membership. Possession of `K_root` plus a controlled signing key remains sufficient to create new valid vault history under a new device identity.

For lifecycle conflicts, diagnostic/audit tooling SHOULD be able to reconstruct an explainable witness containing at least:

```text
TOKEN_ID
current STATUS head H
current CREDENTIAL head(s)
lifecycle transition L
selected/maximal unresolved credential witness W
fully valid resolution coverage, if any
signing device public keys / fingerprints
analysis state/context
```

Diagnostics MUST redact TOTP seed bytes (`CREDENTIAL.SECRET_BYTES`) by default.

Human-facing strings such as `DISPLAY_NAME`, `ISSUER`, and `ACCOUNT` are untrusted presentation data. UIs and diagnostic tools MUST render them safely and MUST NOT allow control characters, bidirectional text behavior, terminal escapes, markup interpretation, or similar presentation effects to obscure cryptographic identity, lifecycle-conflict evidence, or security-relevant warnings.

---

## 68. TOTP semantics

`CREDENTIAL` contains the complete credential material needed to derive a TOTP value.

v0 fixes:

```text
T0 = 0
PERIOD is measured in integer seconds
PERIOD > 0
T = floor(Unix_time_seconds / PERIOD)
```

`Unix_time_seconds` MUST be non-negative.

The computed counter `T` MUST fit in `0 .. 2^64-1`. Negative time or overflow is an error; implementations MUST NOT wrap or narrow it.

`T` is encoded as an unsigned 8-byte big-endian HOTP counter.

TOTP uses RFC 6238 with the selected HMAC algorithm and RFC 4226 dynamic truncation, then decimal reduction and left zero-padding.

The v0 semantic algorithm set and wire values are exactly:

```text
ALGORITHM = 0x01  HMAC-SHA-1
ALGORITHM = 0x02  HMAC-SHA-256
ALGORITHM = 0x03  HMAC-SHA-512
```

No other OTP MAC algorithm value is valid in v0.

`DIGITS` is encoded directly as one unsigned byte and MUST be exactly `0x06`, `0x07`, or `0x08`.

`PERIOD` is encoded as an unsigned 32-bit big-endian integer and MUST be in `1 .. 2^32-1` seconds. Format validity accepts that entire range. Clients SHOULD warn when importing or configuring an unusually large `PERIOD`, because such a value may indicate misconfiguration or poor interoperability. v0 does not freeze a warning threshold, and an in-range `PERIOD` MUST NOT be classified as format-invalid solely because it is unusually large.

`SECRET_BYTES` MUST contain `1 .. 128` bytes for format validity. The bytes are represented exactly; an importer MUST NOT truncate a longer secret to force it into v0.

Format validity and secure secret generation are distinct. v0 may import non-empty legacy seeds shorter than modern generation recommendations, but clients SHOULD warn on unusually weak imports. Newly generated secrets MUST contain at least 20 CSPRNG-generated bytes (160 bits). They MAY be longer up to the 128-byte format maximum.

v0 represents only the RFC 6238-style profile above. The following are not representable as valid v0 credentials:

- HOTP/event-counter credentials whose moving factor is externally stored state rather than Unix time;
- TOTP credentials with `T0 != 0`;
- non-decimal or alphanumeric OTP output formats, including Steam-style 5-character codes;
- decimal OTP credentials requiring a digit count other than 6, 7, or 8.

An importer encountering such a credential MUST reject it as unsupported for v0 or preserve it outside the v0 token format; it MUST NOT silently coerce it into a different v0 TOTP credential.

The byte-level registry and limits for these fields are frozen by Sections 6, 14, 68, 69, and 70.

Actual OTP-generation vectors for every supported algorithm/digit combination and time-boundary behavior are mandatory publication artifacts.

### 68.1 Ordinary credential-consumption safety profile

This subsection governs **application use/presentation**, not consensus object validity. It does not add wire fields or change the derived-state rules.

A client MUST NOT present a credential as an ordinary, unqualified **current active token** for automatic OTP use or export when the token is in `DEGRADED_KNOWN_HISTORY_LOSS`, when validation of the token state is resource-limited/incomplete, when the current `STATUS` value is ambiguous or not unambiguously `LIVE`, when an unresolved lifecycle conflict exists, or when more than one distinct complete current `CREDENTIAL` value remains. Concurrent revision heads carrying byte-identical field values remain distinct lineage but MAY be coalesced as one value for this consumption decision, consistent with Section 30.

Ordinary `ISSUER` or `ACCOUNT` conflicts do not by themselves prohibit OTP generation when the status and credential requirements above are satisfied, but those metadata conflicts MUST remain visible and MUST NOT be silently collapsed into false provenance.

A client MAY provide explicit historical, degraded, recovery, or conflict-inspection interfaces that reveal otherwise non-current credential alternatives or derived OTPs, but such use MUST be clearly distinguished from ordinary current-token use. A lifecycle-conflicted or degraded token MUST remain discoverable through conflict/recovery UI even when every currently displayed status alternative is `TOMBSTONE`; suppression from the ordinary active-token list MUST NOT make the recovery condition undiscoverable. The same safety distinction applies to export, autofill/integration, APIs, and other application interfaces, not only the main token screen.

---

## 69. Canonical TLV registry

v0 uses a canonical fixed-width TLV encoding. This section is the normative byte-level registry for semantic plaintext `P`.

### 69.1 Primitive TLV framing

Every TLV is encoded as:

```text
TAG:u16be || LENGTH:u16be || VALUE[LENGTH]
```

`TAG` and `LENGTH` are unsigned 16-bit big-endian integers. No indefinite, varint, short-form, or alternate-length representation exists. A parser MUST bounds-check the complete `4 + LENGTH` bytes before reading or allocating the value and MUST reject trailing bytes.

Unless a grammar below explicitly permits repetition, tags in one TLV sequence MUST be strictly increasing numerically and MUST appear at most once. The only repeated top-level v0 tag is `CONTEXT_PARENT_UPDATE_ID = 0x0005`; its repetition rules are fixed below. Nested `CREDENTIAL` tags never repeat.

All fixed-width integer values use the exact widths listed below and unsigned big-endian encoding. A value with a different length is invalid even if its numeric value could otherwise be represented.

The first two TLVs of every semantic object using the v0-compatible cryptographic envelope are the common prefix:

```text
0x0001 OBJECT_VERSION
0x0002 OBJECT_TYPE
```

For `OBJECT_VERSION = 0`, the complete immutable grammar below applies. After envelope authentication, a v0 reader encountering another authenticated `OBJECT_VERSION` follows Section 59 rather than applying the rest of the v0 type grammar. Future versions that rely on that handling MUST preserve the common-prefix framing.

### 69.2 Top-level tag registry

| Tag | Name | Exact value encoding | v0 occurrence |
|---|---|---|---|
| `0x0001` | `OBJECT_VERSION` | `u8`; v0 value `0x00` | exactly once, first |
| `0x0002` | `OBJECT_TYPE` | `u8`; `0x01=TOKEN_UPDATE`, `0x02=DEVICE_UPDATE` | exactly once, second |
| `0x0003` | `DEVICE_PUBLIC_KEY` | exactly 32 raw bytes | exactly once |
| `0x0004` | `CONTEXT_PARENT_COUNT` | `u16be`; `0..32` | exactly once |
| `0x0005` | `CONTEXT_PARENT_UPDATE_ID` | exactly 32 raw object-ID bytes | repeated exactly `CONTEXT_PARENT_COUNT` times |
| `0x0101` | `TOKEN_ID` | exactly 32 raw bytes | exactly once in `TOKEN_UPDATE` |
| `0x0102` | `STATUS` | `u8`; `0x01=LIVE`, `0x02=TOMBSTONE` | optional in `TOKEN_UPDATE` |
| `0x0103` | `ISSUER` | UTF-8, `0..256` bytes | optional in `TOKEN_UPDATE` |
| `0x0104` | `ACCOUNT` | UTF-8, `0..256` bytes | optional in `TOKEN_UPDATE` |
| `0x0105` | `CREDENTIAL` | nested canonical TLV sequence from Section 69.4 | optional in `TOKEN_UPDATE` |
| `0x0201` | `DISPLAY_NAME` | UTF-8, `0..256` bytes | exactly once in `DEVICE_UPDATE` |
| `0xFF01` | `SIGNATURE` | exactly 64 raw Ed25519 signature bytes | exactly once and last in signed semantic plaintext; absent from canonical unsigned form |

The `0x00xx` range is the common semantic-object namespace, `0x01xx` is the `TOKEN_UPDATE` namespace, `0x02xx` is the `DEVICE_UPDATE` namespace, and `0x03xx` is the nested `CREDENTIAL` namespace. The `0xFFxx` range is reserved for terminal cryptographic/authentication fields. In v0 only `0xFF01 SIGNATURE` is assigned in that range; every other `0xFFxx` tag is unassigned and invalid. All other unassigned/unknown tags are likewise invalid when `OBJECT_VERSION=0`.

Because canonical ordering is numeric and `SIGNATURE = 0xFF01`, the signature is the final TLV of every signed semantic object without requiring a special ordering exception. The canonical unsigned form is obtained only by omitting the complete `0xFF01` TLV; no zero-length or zero-filled placeholder is inserted.

### 69.3 Context-parent encoding

`CONTEXT_PARENT_COUNT` is always present, including when zero. It MUST be `0..32`.

Exactly that many immediately following `0x0005` TLVs MUST occur before the next greater tag. When the count is zero, no `0x0005` TLV occurs. Every parent value is exactly 32 bytes.

Parent values MUST be strictly increasing under unsigned lexicographic comparison of the 32 raw bytes. This simultaneously freezes canonical set ordering and rejects duplicates. The semantic antichain/type/same-token or same-device requirements remain those of Sections 18, 20, 48, and 49.

The explicit count plus repeated ordinary TLVs is intentional. Packing all parent IDs into a special vector would save only `4N-4` bytes for `N>0` (124 bytes at the v0 maximum of 32 parents) while introducing a second list framing rule. The uniform repeated-TLV form remains within the 2048-byte envelope at every v0 maximum.

### 69.4 Nested CREDENTIAL registry

The value bytes of top-level `CREDENTIAL = 0x0105` are themselves one complete canonical TLV sequence with exactly these four tags, each exactly once and in numeric order:

| Tag | Name | Exact value encoding |
|---|---|---|
| `0x0301` | `ALGORITHM` | `u8`; `0x01=HMAC-SHA-1`, `0x02=HMAC-SHA-256`, `0x03=HMAC-SHA-512` |
| `0x0302` | `DIGITS` | `u8`; exactly `0x06`, `0x07`, or `0x08` |
| `0x0303` | `PERIOD` | `u32be`; `1..2^32-1` |
| `0x0304` | `SECRET_BYTES` | raw bytes; `1..128` bytes |

The nested value contains no trailing bytes and no additional tags. Its total encoded length is therefore `23..150` bytes, and the enclosing `0x0105` TLV occupies `27..154` bytes.

### 69.5 Object grammars

For signed `TOKEN_UPDATE`, the only valid top-level tag sequence is:

```text
0001 OBJECT_VERSION
0002 OBJECT_TYPE = 01
0003 DEVICE_PUBLIC_KEY
0004 CONTEXT_PARENT_COUNT
0005 CONTEXT_PARENT_UPDATE_ID   repeated count times
0101 TOKEN_ID
0102 STATUS?                    optional
0103 ISSUER?                    optional
0104 ACCOUNT?                   optional
0105 CREDENTIAL?                optional
FF01 SIGNATURE
```

At least one of `STATUS`, `ISSUER`, `ACCOUNT`, or `CREDENTIAL` MUST be present. Token-creation and lifecycle rules may require a stricter combination.

For signed `DEVICE_UPDATE`, the only valid top-level tag sequence is:

```text
0001 OBJECT_VERSION
0002 OBJECT_TYPE = 02
0003 DEVICE_PUBLIC_KEY
0004 CONTEXT_PARENT_COUNT
0005 CONTEXT_PARENT_UPDATE_ID   repeated count times
0201 DISPLAY_NAME
FF01 SIGNATURE
```

The canonical unsigned form of either object is the corresponding sequence with only the `FF01 SIGNATURE` TLV removed. All other bytes, ordering, counts, and values are unchanged.

A parser MUST reject missing required tags, forbidden type-specific tags, duplicates other than the exact parent repetition, count/repetition mismatch, non-canonical parent order, wrong fixed widths, invalid UTF-8, out-of-range values, unknown tags/enums, non-canonical nested credential bytes, or trailing bytes.

### 69.6 Envelope sizing under the frozen registry

With a 4-byte TLV header, 32-byte parent IDs cost 36 bytes each on wire. Representative signed semantic-plaintext sizes are:

| Object | Parents | Representative values | Semantic bytes |
|---|---:|---|---:|
| `DEVICE_UPDATE` root | 0 | 12-byte name | 136 |
| `DEVICE_UPDATE` rename | 2 | 64-byte name | 260 |
| `TOKEN_UPDATE` creation | 0 | 12-byte issuer, 18-byte account, 20-byte secret | 245 |
| issuer-only edit | 1 | 12-byte issuer | 208 |
| credential-only rotation | 1 | 32-byte secret | 250 |
| status-only resolution | 2 | — | 233 |
| status-only resolution | 8 | — | 449 |
| status-only resolution | 16 | — | 737 |
| status-only resolution | 32 | — | 1313 |
| all token fields | 8 | 128-byte issuer/account, 64-byte secret | 803 |
| maximum valid `TOKEN_UPDATE` | 32 | 256-byte issuer, 256-byte account, 128-byte secret | **1987** |
| maximum valid `DEVICE_UPDATE` | 32 | 256-byte name | **1532** |

For a candidate object-file size `B` with a 16-byte GCM tag and the 2-byte `SEMANTIC_LENGTH_U16BE`, semantic capacity is `B - 18`: 494 bytes at 512, 1006 at 1024, 2030 at 2048, and 4078 at 4096.

Under the frozen field maxima, a 1024-byte envelope can carry at most four parents in an update that simultaneously asserts every token field at its maximum size. A 2048-byte envelope carries the full 32-parent maximum even in that worst case, leaving 43 semantic bytes of headroom. Therefore v0 uses one fixed 2048-byte envelope. A 4096-byte envelope would double synchronized ciphertext payload without enabling any additional valid v0 object under the frozen registry, while 512/1024 would require materially smaller semantic or frontier limits.

Variable envelope buckets are not used. Consequently object-file length reveals no size class among valid v0 semantic objects, and the exact semantic plaintext still determines exactly one zero-padded encryption plaintext before deterministic per-object encryption.

---

## 70. Strings

Human-facing strings are UTF-8.

The protocol performs no Unicode normalization.

String limits are measured in encoded UTF-8 bytes.

`ISSUER`, `ACCOUNT`, and `DISPLAY_NAME` are each limited to at most 256 encoded UTF-8 bytes. The empty UTF-8 string is representable; semantic/UI policy may choose not to offer an empty value, but a parser MUST NOT invent a non-empty requirement absent another semantic rule.

Semantic strings exist only inside encrypted object plaintext.

---

## 71. Transition and publication summary

A normal token write:

```text
validate/accept complete current token context
        ↓
choose one or more semantic field assertions
        ↓
construct TOKEN_UPDATE whose CONTEXT_PARENTS equal the accepted token-update frontier
        ↓
sign, identify, encrypt, publish durably
        ↓
derive field revisions: each asserted field supersedes all of its current heads in BASE_STATE
        ↓
persist resulting accepted token frontier before acknowledging success
```

Fields absent from the update inherit their complete derived state, including conflicts.

Lifecycle conflicts do not freeze unrelated token fields. `ISSUER` and `ACCOUNT` remain ordinary writes; `CREDENTIAL` remains subject to Section 41 and does not itself resolve lifecycle conflicts. An update asserting `STATUS` over a lifecycle-conflicted context is routed to Section 51; only a valid status-only lifecycle-resolution candidate is subject to the explicit-confirmation writer rule of Section 39.

A normal device-presentation write similarly incorporates all current same-key `DEVICE_UPDATE` heads as context and asserts a new `DISPLAY_NAME`.

There is no semantic per-device token-update sequence and no serialized global vault frontier. A client may maintain off-store derived indexes/frontiers for efficiency only.

### 71.1 Crash-safe writer publication

A writer MUST serialize its own local `TOKEN_UPDATE` creation for the same token so two local operations do not intentionally race from the same accepted token context.

Writes to different tokens need not be serialized merely because they use the same signing key.

A writer MUST separately serialize local `DEVICE_UPDATE` creation per signing identity so two local presentation changes do not intentionally race from the same accepted device context.

Every local semantic publication MUST have a well-defined **publication linearization point** at which all writer gates relevant to that operation are rechecked against stable local security state. An implementation MUST either serialize acceptance of newly validated/pending history and authenticated future-version observations with the final gate check and publication, or use generations/compare-and-set/equivalent synchronization that aborts the operation if any relevant gate input changes before the object becomes published. For `TOKEN_UPDATE(T)`, this includes at least the accepted token frontier, known-history-regression state, applicable pending-history evidence/abandonment state, resource-completeness state, and vault-wide future-observation/bypass state. A lifecycle-resolution write additionally includes every Section 39 confirmation-freshness input. For `DEVICE_UPDATE`, the same rule applies to the accepted same-key presentation frontier, pending-presentation evidence/abandonment state, and vault-wide future-version writer gate.

A relevant observation accepted before the linearization point MUST be incorporated or cause the operation to abort/retry. An observation that becomes accepted only after the linearization point may create a later fork/conflict/incompleteness condition but does not retroactively invalidate an otherwise valid published object.

For immutable objects, writers MUST use crash-safe publication appropriate to the storage platform. The complete object bytes MUST be durable before publication is reported successful. TOKEN_UPDATE success is additionally gated by durable remembered-frontier acceptance under Section 63.

Atomic rename/temp-file publication SHOULD be used where available.

Concurrent token histories and device-name forks are graph properties, not normal concurrency-control strategies for one local writer instance.

### 71.2 Filesystem safety

Implementations SHOULD use bounded reads from stable file handles and MUST NOT trust attacker-preservable size/timestamp metadata alone as proof that an already-known file is unchanged.

Detecting same-metadata replacement may require periodic/event-triggered re-reading and authentication/hash verification of fixed-size object files when trustworthy retained local evidence is unavailable. Implementations should account for this cost explicitly.

Where the configured vault namespace may contain attacker-controlled filesystem entries, implementations MUST use a path/object-type handling policy that prevents protocol reads or publication from following untrusted symlinks, operating on special files, escaping the configured namespace through path traversal/rebinding, or causing unintended blocking or side effects; an equivalently strong sandbox/isolation guarantee is acceptable. Protocol object candidates SHOULD be restricted to ordinary files matching the exact object filename grammar. Temporary files used for immutable-object publication or `VAULT` creation/replacement MUST be created with exclusive safe-creation semantics and appropriate no-follow/path-component protections, or an equivalent platform mechanism, so an attacker-controlled pre-existing entry cannot redirect or overwrite the temporary write.

### 71.3 Bootstrap publication

Creation or replacement of the mutable `VAULT` bootstrap is not covered by immutable-object publication rules alone.

Every initial `VAULT` creation and later `VAULT` replacement MUST follow Section 9.1. Initial creation uses no-replace installation; same-root password rewrap uses authenticated replacement. The configured `VAULT` pathname MUST NOT be created or rewritten by in-place truncation/overwrite.

---

## 72. Core invariants

```text
A configured storage namespace contains exactly one v0 vault bootstrap identity.

K_root is the cryptographic identity of the vault.

Initial establishment is not successful until VAULT_BINDING is durable.

Every TOKEN_UPDATE and DEVICE_UPDATE file is exactly 2048 bytes
using one canonical deterministic zero-padded authenticated envelope.

OBJECT_ID is the keyed hash of exact canonical signed plaintext.

Every TOKEN_UPDATE concerns exactly one TOKEN_ID.

TOKEN_UPDATE context parents concern the same TOKEN_ID and induce token-local causal ancestry.

A conforming token writer uses the complete accepted observed TOKEN_UPDATE frontier as context.

Token semantic fields are STATUS, ISSUER, ACCOUNT, and atomic CREDENTIAL.

Presence of a field in TOKEN_UPDATE creates one new typed field revision.
Its implicit parents are all current heads of that field in BASE_STATE.
No serialized ordinary field-parent or supersedes list exists.

An absent field inherits its complete possibly-conflicted state from BASE_STATE.

Concurrent assertions of different fields compose.
Concurrent assertions of the same field remain as multiple current field revisions.

An ordinary field assertion resolves all currently observed alternatives of that field.
v0 has no partial ordinary field merge and no graph-only merge/adoption update.

A single current TOKEN_UPDATE/version head may carry unresolved field conflicts in its derived state.

STATUS lifecycle transitions concurrent with CREDENTIAL revisions are never silently resolved.
Historical credential witnesses remain relevant beneath current credential heads.

Lifecycle conflicts are persistent derived evidence, not a token-wide freeze.
Unrelated non-STATUS assertions may continue under ordinary rules without resolving the lifecycle conflict.
A STATUS assertion over a lifecycle-conflicted context is a user-confirmed STATUS-only lifecycle resolution over the complete current context.
No LIFECYCLE_CONFIRMATION wire field exists.
A valid lifecycle resolution clears all lifecycle conflicts visible in its own context.
Later hidden/stale history may create a new conflict.

CREDENTIAL may be written from a tombstoned status state only together with STATUS=LIVE.
STATUS=TOMBSTONE and CREDENTIAL MUST NOT be asserted by the same update.

DEVICE_UPDATE context is independent of token context and affects presentation only.

Device signatures prove provenance, not authorization.
Possession of K_root permits creation of a fresh device identity and valid new history.

Known accepted token-history regression is tracked per TOKEN_ID and blocks that token.
Pending authenticated token history is durable security evidence until a defined durable handoff or terminal-invalidity condition.
Authenticated unsupported-future observations remain durable, vault-bound security evidence. An observation without compatible processing or an exact durable Section 59.1 bypass blocks v0 authorship; a bypass changes the writer blocking policy but does not erase the observation or establish completeness.

Durable evidence replacement obeys Section 63.1: replacement evidence becomes durable before non-atomic removal of older evidence.

No v0 mechanism provides secure deletion, root-key rotation, device revocation, global consensus, or protocol-level garbage collection.
```

---

## 73. Explicit v0 exclusions

v0 intentionally does not provide:

- device authorization, enrollment, revocation, or membership consensus;
- a semantic per-device token-update chain;
- a serialized global vault frontier/checkpoint;
- graph-only adoption/merge updates;
- independently stored token-field objects;
- serialized ordinary field-parent/supersedes lists;
- partial ordinary field conflict resolution;
- a wire lifecycle-confirmation marker;
- secure deletion or cryptographic erasure of historical token material;
- root-key rotation inside the same cryptographic vault identity;
- protocol-level garbage collection;
- proof of freshness/completeness against a withholding synchronization medium;
- HOTP, nonzero-T0 TOTP, Steam-style/alphanumeric OTPs, or digit counts outside the v0 TOTP profile.

These exclusions are semantic design choices, not missing parser features. Later extensions that change semantic object grammar require a new semantic object version or explicit vault migration as applicable.

---

## 74. Required security, semantic, and conformance evidence

Before byte-level v0 freeze, executable transition/conformance evidence MUST cover at least:

```text
field-model semantics:
    token creation asserts STATUS=LIVE, ISSUER, ACCOUNT, CREDENTIAL
    issuer edit only
    account edit only
    credential rotation only
    combined disjoint-field edit
    concurrent issuer + account edits compose
    concurrent issuer + credential edits compose
    concurrent same-field edits remain conflicted
    equal-valued concurrent field revisions retain distinct lineage but may be value-coalesced in UI
    unrelated field assertion carries an existing conflict unchanged
    one later assertion of the conflicted field resolves all currently observed alternatives
    no partial ordinary field merge
    one current TOKEN_UPDATE head may carry unresolved field conflicts
    context frontier collapses version heads without silently resolving absent fields
    field ancestry implies context/causal ancestry
    JOIN is associative, commutative, idempotent

lifecycle semantics:
    delete vs concurrent credential rotation creates lifecycle conflict
    restoration vs concurrent credential rotation creates lifecycle conflict when appropriate
    restore+credential rotation in one causally informed update is valid when PRE is empty
    CREDENTIAL write from TOMBSTONE requires STATUS=LIVE in same update
    STATUS=TOMBSTONE + CREDENTIAL in same update is invalid
    lifecycle conflict + issuer edit -> issuer edit valid and lifecycle conflict remains
    lifecycle conflict + account edit -> account edit valid and lifecycle conflict remains
    lifecycle conflict + credential-only write with all current STATUS heads LIVE -> write may be valid and historical lifecycle conflict remains
    lifecycle conflict + credential write when any current STATUS head is TOMBSTONE -> blocked until status-only lifecycle resolution because CREDENTIAL would require STATUS=LIVE
    lifecycle conflict + STATUS assertion -> must use user-confirmed status-only lifecycle resolution
    user-confirmed STATUS-only resolution clears every lifecycle conflict visible in its context
    multiple visible credential witnesses are resolved by one status decision over the complete context
    multiple visible status branches are reconciled by the resolution status assertion
    hidden credential/status branch discovered later creates a new lifecycle conflict
    stale credential descendant after resolution creates a new conflict
    historical non-head credential witness remains detectable beneath a current credential descendant
    current-credential-head-only lifecycle oracle is rejected by counterexample
    no LIFECYCLE_CONFIRMATION wire field is required to reconstruct resolution coverage
    confirmation becomes stale on relevant token/pending/future context change
    future-context change-and-return still stales confirmation
    unpublished confirmation does not survive restart
    post-publication self-check distinguishes valid resolution from newly arrived conflict
    lifecycle-resolution candidate receives provisional coverage only for its own POST* validation
    correctly constructed lifecycle-resolution candidate -> POST* is empty
    mutation/corrupted-evaluator test producing nonempty POST* -> implementation invariant failure and provisional coverage does not enter history
    prospective resolution that fails lifecycle-resolution candidate classification -> receives no candidate-only provisional coverage
    PRE empty -> not a lifecycle-resolution candidate; validate under ordinary update rules and it may be FULLY_VALID
    required context dependency unavailable -> PENDING, not INVALID, and no candidate-only provisional coverage
    PRE non-empty with STATUS plus another asserted token field -> INVALID as a lifecycle resolution and no candidate-only provisional coverage
    PRE empty with STATUS plus another asserted token field -> ordinary-update rules apply and it may be FULLY_VALID
    invalid signature, invalid STATUS, invalid validated context dependency, or structural invalidity -> INVALID and no candidate-only provisional coverage
    deterministic validator result does not depend on proving human confirmation
    complete observed frontier and fresh human confirmation are writer-conformance requirements, not remote validity predicates
    later discovery of writer-omitted/withheld history may create a new conflict without retroactively invalidating the resolution

audit/presentation safety:
    DISPLAY_NAME/ISSUER/ACCOUNT containing control characters, bidi controls, terminal escapes, or markup cannot obscure fingerprints, lifecycle evidence, or security warnings
    lifecycle-conflict audit output reconstructs status head, lifecycle transition, maximal credential witness, resolution coverage, and signer fingerprint
    diagnostic output redacts CREDENTIAL.SECRET_BYTES by default

device presentation:
    parentless DEVICE_UPDATE
    concurrent same-key names form presentation fork
    later DEVICE_UPDATE uses all current same-key context heads and resolves visible name alternatives
    token updates may exist without DEVICE_UPDATE
    presentation history never changes token validity

context/writer rules:
    TOKEN_UPDATE context parents equal the accepted observed token-update frontier for writer conformance
    cross-token context parent is invalid
    redundant ancestor context parent is invalid
    omitted observed remote head is writer nonconformance but later discovery does not retroactively invalidate the signed update
    same signer creates no token causality by itself
    different tokens remain causally independent

pending/future/recovery:
    signature-verified pending TOKEN_UPDATE blocks only its token by default
    abandonment is exact-ID, vault/token-bound, durable before authorship, and does not erase evidence
    abandoned pending update later becoming FULLY_VALID rejoins history normally
    known-history regression cannot be bypassed by pending recovery
    authenticated future object creates durable vault-wide evidence and default write block
    exact durable future bypass permits only the understood v0 subset and does not claim completeness
    newly observed un-bypassed future object blocks again

first-establishment/bootstrap:
    exact v0 VAULT length is 87 bytes and exact magic is ASCII("TOTP-VAULT")
    BOOTSTRAP_VERSION=0x00 occupies byte 10; unsupported version is rejected before KDF
    VAULT_HEADER is exactly bytes 0..38 and is the complete AES-GCM AAD
    wrong total length, wrong magic, unsupported version, password >1024 UTF-8 bytes, malformed pre-encoded UTF-8, and character-to-UTF-8 encoding failure are rejected without invoking Argon2id
    empty password and exactly 1024 well-formed UTF-8 bytes are format-valid password inputs
    fixed Argon2id v=0x13/m=65536 KiB/t=3/p=4/output=32 with empty K/X reproduces the published K_wrap vector
    correct password unwraps the published K_root; wrong password fails authenticated unwrap
    mutation of ARGON2_SALT, WRAP_NONCE, WRAPPED_ROOT, or WRAP_TAG causes unwrap failure
    same-root password rewrap with fresh salt/nonce produces a distinct valid 87-byte VAULT while recovering the identical K_root
    initial creation uses no-replace canonical VAULT installation
    local racing creators have at most one winner
    PENDING(VAULT_BINDING) is durable before creation publication
    first open is not successful before durable binding
    restart from pending establishment cannot silently switch roots

security-memory durability:
    local TOKEN_UPDATE publication durable before remembered accepted frontier
    newly validated frontier not accepted/writer-eligible before remembered heads are durable
    pending-to-accepted-frontier handoff retains old evidence until replacement is durable
    non-atomic shared flush does not imply transaction atomicity
    batched persistence may share a durability barrier while each replacement obeys Section 63.1
    authenticated future evidence survives disappearance/restart/cache rebuild and is not erased by bypass
    rollback/reset of local durable security memory is not misrepresented as preserving prior rollback guarantees
    pending DEVICE_UPDATE evidence and exact abandonment survive restart while relied upon
    abandoned pending DEVICE_UPDATE later becoming FULLY_VALID re-enters presentation history

operational safety:
    lifecycle-conflicted STATUS+another-field is routed to Section 51, rejected before candidate coverage, and never treated as a candidate
    establishment-state racing initializers cannot overwrite another PENDING/ESTABLISHED binding
    writer gate change before publication linearization aborts/retries publication
    relevant observation after publication linearization may create later conflict without retroactive invalidation
    known-history recovery cannot clear remembered heads, invent parents, or reclassify accepted missing history as pending
    migration from a >32-head/capacity-blocked source may select values without first publishing an impossible source-vault resolution
    hostile symlink/special-file/temp-path cases cannot redirect protocol I/O or cause unsafe side effects
    ordinary OTP use/export is blocked for degraded, incomplete, lifecycle-conflicted, status-ambiguous, or credential-value-ambiguous token state

object framing/crypto:
    exact 2048-byte semantic object files
    fixed authenticated SEMANTIC_LENGTH_U16BE + zero padding
    canonical TLV framing/order positive and negative vectors cover wrong fixed widths, duplicate or unknown tags, out-of-order tags, and trailing bytes
    context-parent vectors cover counts 0/1/32, reject count 33, reject count/repetition mismatch, and reject duplicate or non-increasing raw parent IDs
    UTF-8 vectors cover 0/256-byte accepted values, 257-byte rejection, and malformed UTF-8 rejection for every human-facing string field
    SECRET_BYTES vectors cover lengths 1/128 accepted and 0/129 rejected
    nested CREDENTIAL vectors cover exact tag order/widths, every valid enum value, unknown enum/tag rejection, and nested trailing-byte rejection
    envelope-boundary vectors cover authenticated SEMANTIC_LENGTH_U16BE values 0 and 1 as framing-success/semantic-parse-failure cases, maximum valid v0 semantic length 1987, generic envelope maximum 2030, and 2031 rejection before semantic extraction
    nonzero padding rejected
    wrong 2048-byte object-file length rejected before AEAD processing
    recomputed OBJECT_ID must match filename
    strict Ed25519 positive/negative corpus including mixed-order rejection
    HKDF/AES-GCM known-answer composition checks
    RFC 6238 vectors for SHA-1/SHA-256/SHA-512 and 6/7/8 digits
```

Randomized semantic testing SHOULD generate valid token histories with stale contexts and compare at least two independently implemented lifecycle calculations. Deliberately incorrect variants MUST include current-credential-head-only witness scanning, future-context-insensitive confirmation freshness, early pending-evidence removal, and shared-flush-as-atomic assumptions so the corpus demonstrates sensitivity to the failures it is intended to catch.

Crypto vectors MUST include canonical bytes and all intermediate object-crypto values. `TOKEN_UPDATE` and `DEVICE_UPDATE` vectors MUST include unsigned canonical bytes, signature input/context, public key, signature, final signed plaintext, object ID, envelope plaintext, ciphertext, and tag.

Lifecycle-conflict reference-oracle outputs MUST include explainable conflict tuples and the witness data necessary to reproduce raw-witness selection, maximality, and fully valid resolution coverage rather than only booleans.

---

## 75. Remaining work before byte-level v0 freeze


Revision 36 is the Totipo byte-level-freeze candidate. It preserves every r34 wire byte and consensus accepted-history rule while incorporating the final review's cross-section consistency and operational-security clarifications: candidate routing, future-evidence durability, bootstrap password prechecks, known-history recovery endpoints, credential-consumption safety, hostile-filesystem handling, establishment/writer concurrency, local-security-memory assumptions, normative `VAULT_BINDING`, and pending device-presentation recovery. The independent Java-to-Python semantic-object TLV/envelope cross-consumption check and independent Java consumption of the r32 bootstrap Argon2id/AES-GCM vectors have completed without contradiction.

The remaining work is conformance-evidence and publication work rather than byte-format design:

1. implement/publish the current literal derived-state and lifecycle-conflict oracle from this document;
2. run all retained pre-r28 semantic/transition corpora against the r36 model and classify every changed outcome;
3. resolve any contradiction discovered by those regression passes;
4. retain executable tests freezing the r36 context/field writer rule, lifecycle-resolution validation procedure, confirmation freshness procedure, durable recovery ordering, publication linearization, and presentation-recovery rules;
5. complete/publish the mandatory object-crypto, Ed25519, TOTP, transition, degradation, rollback, presentation-history, resource-limit, bootstrap-publication, and password-input conformance vectors;
6. retain at least two independent semantic-object TLV/envelope parsers/validators against the frozen corpus;
7. independently consume the pinned semantic-object cryptographic vectors in another implementation before claiming full wire/crypto interoperability.

The canonical TLV/envelope boundary corpus, Java-to-Python semantic-object parser cross-check, r32 bootstrap vectors, and independent Java bootstrap-vector consumption are completed inputs to this remaining work and need not be redesigned absent a discovered contradiction.

Changes discovered during that work SHOULD be treated as specification bugs rather than opportunities to redesign the state model unless an actual contradiction is found.
