# r31 Java → Python cross-implementation byte check

## Result

PASS: 12 cross-implementation checks across the Java-pinned TOKEN_UPDATE and DEVICE_UPDATE fixtures.

The Java `R31Review --write-vectors` path generated `r31-object-vectors.txt`. The independent Python r31 TLV/envelope implementation then consumed those Java-emitted bytes directly. The Python implementation did not regenerate the fixture objects.

For each object type, the cross-check established:

1. Java `signed_plaintext` is accepted by the Python canonical signed-v0 grammar.
2. Java `unsigned` is accepted by the Python canonical unsigned-v0 grammar.
3. Java `unsigned` is byte-for-byte equal to `signed_plaintext` with exactly the complete final `FF01` signature TLV (4-byte TLV header + 64-byte signature value) omitted.
4. Java `envelope_plaintext` is accepted by the Python envelope validator and extracts exactly the Java `signed_plaintext`; all remaining padding is zero.
5. Java fixture framing relationships match r31: 32-byte object ID, nonce = first 12 bytes of object ID, AAD = `ASCII("TOTP-Vault/v0/object") || object_id`, 2032-byte ciphertext, 16-byte tag, 2048-byte stored object.
6. Java `signature_input` begins with the correct object-type domain literal and ends with the exact canonical unsigned bytes; the intervening vault signature-context key is 32 bytes.

## Scope

This is deliberately a byte-grammar/interoperability check. The Python side does not independently verify Ed25519, HMAC object IDs, HKDF, or AES-GCM. Those values remain pinned by the Java executable review and existing primitive known-answer tests.

The result closes the previously identified self-generated-corpus gap for the TLV/envelope layer: Java-generated object fixtures are accepted by an implementation that shares neither Java framing nor Java registry-validation code.

## Conclusion

No r31 registry or envelope contradiction was found. The semantic-object TLV/envelope layer now has cross-implementation consumption evidence sufficient to proceed to the VAULT bootstrap byte layout.
