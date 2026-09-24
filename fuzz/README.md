# Non-normative Go fuzz/property tests

The targets live in `conformance/internal/runner/fuzz_test.go` and seed from the committed positive and negative artifacts after schema/manifest verification:

- `FuzzTLVFraming`: bounded extraction and byte-identical reconstruction.
- `FuzzV0Grammar`: no panic, accepted-field reconstruction, exact signature-TLV omission.
- `FuzzCredentialGrammar`: no panic, accepted length/framing bounds.
- `FuzzEnvelopeUnframe`: exact declared range and canonical zero-padding reconstruction.
- `FuzzCommonPrefixDispatch`: no panic; arbitrary future tails stay opaque.

Example from the repository root:

```sh
cd conformance
go test ./internal/runner -run '^$' -fuzz '^FuzzV0Grammar$' -fuzztime=30s
```

`go test ./...` runs all deterministic fuzz seeds. Fuzz discoveries are non-normative and must be minimized and reviewed before adding public vectors. Fuzzing never rewrites `vectors/v0` or its manifest. The tests do not generate official expected bytes.
