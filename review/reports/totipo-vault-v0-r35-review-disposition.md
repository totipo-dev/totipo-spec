# Totipo Vault Format v0 r35 — review disposition

r35 incorporates the attached r34 review and the project rename to **Totipo**. It intentionally changes no semantic-object wire bytes, bootstrap bytes, TLV registry values, envelope sizes, object-crypto construction, or frozen cryptographic domain strings.

## Branding

The specification title/project name is now **Totipo Vault Format v0**. The already-frozen protocol constants remain byte-for-byte unchanged:

- `ASCII("TOTP-VAULT")`
- every `ASCII("TOTP-Vault/v0/...")` domain separator

Those strings are protocol constants, not branding strings. Replacing them with `Totipo` would define different bytes and is explicitly non-conforming for v0.

## Review findings

### A. Lifecycle-resolution candidate terminology — accepted

Sections 37, 38, 40, and 71 now distinguish **routing to lifecycle-resolution validation** from **candidate classification**. An update with `PRE != empty` and any `STATUS` assertion is routed to Section 51, but becomes a candidate only after the exact status-only/supported-status preconditions succeed. `STATUS` plus another token field is invalid in that route and receives no provisional candidate coverage.

### B. Durable future evidence summary — accepted

Section 72 now matches Sections 59.1 and 63: bypass changes the v0 writer-blocking policy for the exact observation but never erases authenticated future evidence or establishes completeness.

### C. Password preprocessing in bootstrap gate — accepted

Section 8 now explicitly includes strict well-formed UTF-8 validation/encoding and the 1024-byte bound in the pre-Argon2 checks, matching Sections 6 and 74.

### D. Recovery endpoints and capacity-blocked migration — accepted

Section 64 now states that writability of the original `TOKEN_ID` can be restored only by reconstructing its remembered accepted history. Recovery cannot clear remembered heads, invent/remove parent edges, or downgrade known history loss to pending history. Recoverable values may instead be materialized into a fresh `TOKEN_ID` or migrated to a new vault, with no continuity claim.

Section 65.1 now makes conflict selection for migration local migration input. A capacity-blocked source vault does not need to publish an impossible old-vault resolution before migration can proceed.

### E. Credential consumption profile — accepted as application semantics

New Section 68.1 is explicitly application-facing, not consensus validity. Ordinary current-token OTP use/export is blocked for degraded or incompletely validated state, ambiguous/non-LIVE status, unresolved lifecycle conflict, or multiple distinct credential values. Equal-valued revision heads may still coalesce for value use while retaining lineage. Metadata conflict alone does not block OTP generation, but remains visible. Historical/degraded inspection must not masquerade as normal current use.

### F. Filesystem safety — accepted

Section 71.2 now requires safe handling when the sync namespace can contain hostile filesystem entries: no unsafe symlink/special-file/path rebinding behavior, and exclusive/no-follow-equivalent temporary-file creation for protocol publication.

### G. Establishment and writer isolation — accepted

Section 10.1 now requires serialized/CAS-equivalent configured-vault establishment transitions, so racing initializers cannot overwrite another pending/established binding.

Section 71.1 now defines a local publication linearization point and requires relevant writer gates to be stable or generation/CAS checked through publication. Observations accepted before that point must be incorporated or abort the write; observations accepted later may produce a later fork/conflict without retroactive invalidation.

### H. Local security-memory assumptions and VAULT_BINDING — accepted

Section 63 now explicitly requires durable security memory to live outside the hostile sync namespace or have an independent integrity/freshness mechanism. Restoring/resetting older local state is acknowledged as loss of later remembered evidence unless a separate continuity mechanism exists.

The previously recommended `VAULT_BINDING` formula is now the exact mandatory v0 construction. This changes local conformance requirements, but no synchronized bytes.

### I. Pending DEVICE_UPDATE recovery — accepted

New Section 49.1 defines presentation-only durable pending evidence and exact abandonment. The decision is vault/key/object-bound, survives restart when relied upon, never affects token semantics, and cannot suppress a branch that later validates. Section 63/63.1 now carries the corresponding durable handoff rules.

## Reviewer lemmas

The field-ancestry equivalence and lifecycle-resolution POST* observations remain useful review lemmas/test-oracle guidance. r35 does not add them as new wire or consensus rules because the existing normative definitions already determine those results.

## Remaining release evidence

The review's release-evidence concerns remain in scope of Section 75: current lifecycle-oracle publication, complete pinned conformance artifacts, crash/restart/concurrent-publication testing, and independent consumption of the semantic-object cryptographic vectors before a full interoperability claim.

## Verification performed while cutting r35

A mechanical comparison confirmed that:

- every frozen `TOTP-VAULT` / `TOTP-Vault/v0/...` literal occurs unchanged;
- Sections 69.1–69.6 (the complete TLV registry and sizing rules) are byte-for-byte unchanged from r34;
- Section 14 (object envelope/encryption construction) is byte-for-byte unchanged from r34.
