# TOTP Vault v0 r31 — canonical TLV corpus report

This companion corpus freezes the non-cryptographic byte-level cases for the r31 semantic-object registry. It covers canonical semantic TLV bytes, canonical unsigned forms, deterministic envelope-plaintext framing, and the pre-AEAD 2048-byte object-file length gate.

It intentionally does **not** claim that the 64-byte `FF01 SIGNATURE` placeholders are valid Ed25519 signatures. Ed25519, object-ID/HMAC, HKDF, AES-GCM, and history-dependent semantic validity belong to later vector layers.

## Corpus summary

- Total cases: **76**
- Structurally/framing valid cases: **22**
- Required-rejection cases: **54**
- Semantic TLV cases: **64**
- Envelope-plaintext cases: **9**
- Pre-AEAD object-file length cases: **3**

Frozen census assertions reproduced by the generator:

- typical creation-shaped signed `TOKEN_UPDATE`: **245 bytes**
- maximum valid `TOKEN_UPDATE`: **1987 bytes**
- maximum valid `DEVICE_UPDATE`: **1532 bytes**
- generic semantic envelope capacity: **2030 bytes**
- encryption plaintext: **2032 bytes**
- final semantic object file after 16-byte GCM tag: **2048 bytes**

## r31 §74 boundary coverage

The corpus explicitly covers:

- TLV framing: fixed `u16be` tag/length, wrong fixed widths, duplicate/unknown tags, numeric-order violations, top-level trailing bytes;
- context parents: accepted counts 0, 1, and 32; rejected 33; count/repetition mismatch; 31-byte parent; unsorted and duplicate raw IDs;
- human strings: accepted 0 and 256 encoded UTF-8 bytes and rejected 257/malformed UTF-8 independently for `ISSUER`, `ACCOUNT`, and `DISPLAY_NAME`;
- credential bounds: accepted secret lengths 1 and 128; rejected 0 and 129;
- enums: all three algorithm values, all three digit values, both status values, both object types; unknown enum rejection;
- nested credential: exact tag sequence/widths, missing/duplicate/unknown/out-of-order nested tags, and nested trailing-byte rejection;
- signature structure: exactly 64 bytes, `FF01` terminal in signed form, absent from unsigned form;
- signed/unsigned relation: unsigned bytes equal signed bytes with the complete final `FF01` TLV removed, with no reserialization changes;
- envelope plaintext: semantic lengths 0 and 1 succeed framing but fail semantic parsing, 1987 succeeds as the largest valid r31 object, 2030 succeeds as generic framing, 2031 rejects before extraction, and non-zero padding rejects;
- pre-AEAD file gate: 2048 bytes passes the length gate; 2047 and 2049 reject.

## Reproduction

Run:

```text
python3 totp-vault-v0-r31-tlv-corpus.py
```

The generator validates every case against its expected accept/reject disposition before rewriting the JSON corpus. Rejection *codes* emitted by this generator are diagnostic and are not proposed as consensus-visible error codes; the normative property of negative vectors is rejection.

## Artifact hashes

- `totp-vault-v0-r31-tlv-corpus.json` SHA-256: `c6c6639413f605ce18ecdb42d6e3d64b4ccd1709148eae43646dbcadf009b48c`
- `totp-vault-v0-r31-tlv-corpus.py` SHA-256: `18e7d28851e176cf1aeebe5e8a5bd4524a64c281fb298b69f39d568146e3777a`

## Result

This pass found **no contradiction requiring an r31 registry change**. The next useful layer is an independently written parser/validator consuming this frozen JSON without importing or sharing the generator implementation. After that passes, the same representative objects can be promoted into the cryptographic vector layer: exact unsigned bytes → signature input → signature → signed plaintext → object ID → padded envelope → AES-GCM ciphertext/tag.
