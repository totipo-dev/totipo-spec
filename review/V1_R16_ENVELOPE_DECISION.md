# Totipo v1/r16 fixed envelope — fifth design checkpoint

This is a non-normative design checkpoint. **v1/r15 remains normative until
the r16 rewrite**. The first four checkpoints remain unchanged historical
records, read in chronological order: simplification, state/policy,
storage/identity, and storage/operation decisions. This checkpoint resolves
the last substantial protocol-design question left open before normative r16.
Exact same-OBJECT_ID failure wording still needs final normative phrasing;
its high-level defensive integrity-failure behavior is already settled and
does not require another design checkpoint or a recovery state machine.

## Settled decision

Retain the fixed **1024-byte encrypted TOKEN object envelope** for r16:

```text
object file size                 1024 bytes
AES-GCM tag                        16 bytes
encrypted plaintext              1008 bytes
authenticated semantic length       2 bytes
maximum semantic plaintext       1006 bytes
```

The exact envelope format otherwise remains the existing v1 construction,
including authenticated semantic length and canonical zero padding. Removal
of obsolete semantic fields requires grammar and explanatory wording changes,
not a redesign of keyed OBJECT_ID, HKDF object keys, or AES-GCM.

## Rationale

The fixed representation gives every v1 TOKEN one uniform physical shape.
It simplifies bounded reading, storage validation, publication, memory
allocation, test fixtures, cross-platform implementations, and interoperability
reasoning. It also incidentally hides exact semantic plaintext length. This
is a secondary benefit, not the primary security rationale. Variable-size
authenticated encryption would be technically viable; its length leakage is
not considered a major security defect. The choice is not primarily a
filesystem-block-size optimization.

The settled r16 semantic grammar makes 1024 essentially the minimum practical
fixed envelope with the agreed limits:

```text
worst-case legal non-parent fields       861 semantic bytes
four explicit parents: 4 * 36            144 bytes
maximum legal TOKEN                    1005 semantic bytes
semantic capacity                      1006 bytes
spare worst-case semantic capacity         1 byte
```

Every legal TOKEN always fits `MAX_PARENTS = 4`, even with every permitted
field at its maximum. Shrinking the fixed envelope would require reopening
already-settled field limits or the four-parent invariant. Growing it provides
no currently justified semantic benefit. Bucketed sizes would add
canonicalization and envelope complexity without enough benefit.

r16 no longer reserves speculative in-family semantic-version headroom.
Future families define their own compatibility relationship.

## Preservation and phase boundary

This checkpoint is added before normative edits. The previous four checkpoint
hashes are recorded in the rewrite report's baseline and are verified unchanged
at this boundary. `git diff --check` passes; no tracked normative, corpus, or
tooling file has changed. Normative r16 rewriting follows directly, without a
commit. This document remains the historical decision record after that rewrite.
