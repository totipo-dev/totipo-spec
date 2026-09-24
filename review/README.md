# Review inputs

These files are historical/review provenance inputs for building the Totipo v0 conformance suite.

- `source-vectors/`: reviewed or pinned input vectors/checksums to normalize into `vectors/v0/`.
- `source-code/`: historical/reference implementations and generators. They are not the normative protocol definition and should not be ported line-for-line.
- `reports/`: review and interoperability evidence.

The normative specification is `../spec/totipo-vault-format-v0.md`.

Historical filenames using `totp-vault` are intentionally preserved for provenance. Newly created public artifacts should use the Totipo name, while frozen protocol byte strings such as `TOTP-VAULT` and `TOTP-Vault/v0/...` remain unchanged.
