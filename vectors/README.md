# v1/r16 conformance vectors

The [manifest](manifest.json) lists the complete current corpus, with exact paths,
SHA-256 hashes, kinds, expected outcomes, and specification sections. Every case
is required by the [moving pre-RC profile](../requirements/v1-pre-rc.json).

Run `make conformance` to execute all cases or `make verify` to validate schemas,
hashes, physical file coverage, and profile pins. Neither command writes fixtures.
The [case contract](FORMAT.md) describes byte encoding and model boundaries.

Coverage includes canonical TOKEN grammar, client metadata, the maximum 1005-byte
TOKEN, deterministic 1024-byte encryption, authenticated invalid length/padding and keyed-ID
mismatch rejection, root wrapping and fingerprint stability,
RFC TOTP answers, SCC/current heads, missing ancestry, fixed folds, discovery
diagnostics, publication, and VAULT workflows. Abstract graph cycles and same-ID
failure cases do not pretend to be constructible cryptographic collisions.

The [rewrite report](../review/V1_R16_REWRITE_REPORT.md) classifies every old case.
This is moving pre-RC evidence, not a release, frozen RC profile, independent
security audit, or proof of production platform durability.
