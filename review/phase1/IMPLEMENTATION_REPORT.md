# Totipo v0 conformance implementation report

The first byte/crypto checkpoint is implemented against `spec/totipo-vault-format-v0.md` revision 36. The Go consumer reproduces the externally pinned TLV, object-crypto, bootstrap and available TOTP expectations without regenerating them during tests. This is an initial conformance suite, not completion of every Section 74 publication obligation.

## Results

**Available corpus: 270 PASS, 0 FAIL.** There are no silently skipped corpus cases. One required deliverable is **BLOCKED** by missing reviewed input: strict Ed25519 acceptance/rejection conformance. This is a deliverable count, not a fabricated negative-vector count. Additional stateful work remains explicitly incomplete below.

| Category | Cases | Result |
|---|---:|---|
| TLV | 65 | PASS |
| Envelope/file-length gates | 13 | PASS |
| Object crypto | 2 | PASS: every pinned intermediate matches |
| Bootstrap/password | 37 | PASS: INITIAL and REWRAP match exactly |
| Common-prefix dispatch and encrypted processing order | 56 | PASS |
| TOTP | 82 | PASS |
| Initial symbolic token/lifecycle model | 15 | PASS |
| Strict Ed25519 acceptance corpus | Not supplied | BLOCKED |

Provenance totals: 104 reviewed-pinned cases, 54 published-standard cases, 112 spec-derived cases. Spec-derived cases have not been represented as externally reviewed. The 104 reviewed cases comprise 76 original TLV/envelope/file-size cases, two object-crypto fixtures, two bootstrap fixtures, and 24 TOTP period-boundary assertions copied from the historical Java review's literal expected values.

All 17 historical review files match `review-inventory.sha256`. The normative specification and historical inputs were not edited. No expected value was changed to accommodate a Go result. A converter bug in interpreting the source enum `valid_signed` was corrected during development; the independent source-comparison test verifies the retained source dispositions.

## Artifacts

The 11 JSON vector files are:

```text
vectors/v0/bootstrap/cases.json
vectors/v0/bootstrap/supplemental.json
vectors/v0/dispatch/supplemental.json
vectors/v0/envelope/cases.json
vectors/v0/envelope/supplemental.json
vectors/v0/lifecycle/cases.json
vectors/v0/object-crypto/cases.json
vectors/v0/tlv/cases.json
vectors/v0/tlv/supplemental.json
vectors/v0/totp/cases.json
vectors/v0/totp/supplemental.json
```

`vectors/v0/README.md` documents the schema, enum vocabulary, symbolic-model extension, provenance, and exact manifest rules. `vectors/v0/manifest.sha256` covers every JSON vector file in lexicographic path order. Both the Go loader and an independent `sha256sum -c` check passed. Documentation/status Markdown is not a normative vector file.

`conformance/` contains the corpus loader, independent TLV and credential grammar, envelope codec, key hierarchy and object crypto, bootstrap codec, dispatch, TOTP, initial token model, CLI, deterministic tests, and fuzz targets. `tools/normalize.go --write-vectors` is an explicit maintenance converter with pinned source-hash checks; it does not import the implementation under test. No normal runner or CI path rewrites expectations.

## Validation and environment

Local Go toolchain: **go1.26.7 linux/amd64**. Declared module language floor: Go 1.23. Other operating systems/toolchains are configured in CI but have not been executed locally.

External dependencies:

- `golang.org/x/crypto v0.36.0`: Argon2id.
- `golang.org/x/sys v0.31.0`: indirect CPU/platform support used by x/crypto.

Commands from the repository root:

```sh
gofmt -l conformance tools/normalize.go
go -C conformance fmt ./...
go -C conformance vet ./...
go -C conformance test -count=1 ./...
go run ./conformance/cmd/totipo-conformance ./vectors/v0
go run ./conformance/cmd/totipo-conformance ./vectors/v0 --filter v0/object-crypto/
sha256sum -c conformance/review-inventory.sha256
(cd vectors/v0 && sha256sum -c manifest.sha256)
go -C conformance test -race ./...
```

Formatting, vet, deterministic tests, CLI execution/filtering, original-input hashes, and the vector manifest pass. The execution environment used `GOMODCACHE=/tmp/totipo-gomod` and `GOCACHE=/tmp/totipo-gocache` to keep caches in writable storage.

Race-detector status: **PASS** with GCC 15.3.0 and Go 1.26.7. After the shell was refreshed to expose the compiler, `go -C conformance test -race ./...` completed successfully across all packages. The refreshed dependency cache required downloading the same pinned module versions. Earlier missing-compiler and restricted-network setup failures were resolved; no race assertion failed.

All five fuzz targets passed short local smoke runs, in addition to their deterministic corpus seeds:

| Target | Smoke executions |
|---|---:|
| `FuzzTLVFraming` | 241,739 |
| `FuzzV0Grammar` | 124,978 |
| `FuzzCredentialGrammar` | 75,927 |
| `FuzzEnvelopeUnframe` | 1,982 |
| `FuzzCommonPrefixDispatch` | 79,670 |

Example fuzz command: `go -C conformance test ./internal/runner -run '^$' -fuzz '^FuzzV0Grammar$' -fuzztime=2s -parallel=2`. These short runs are smoke evidence, not exhaustive fuzzing. Discoveries remain in the non-normative Go fuzz cache and do not rewrite vectors.

CI checks formatting, vet, deterministic tests/fuzz seeds and the CLI on Linux/macOS/Windows, with Go 1.26.x and `stable`; Linux also runs the race detector. CI does not download or regenerate expected vector data. The configured toolchain entries may coincide while 1.26 is stable.

## Cryptographic and semantic evidence

Both TOKEN_UPDATE and DEVICE_UPDATE reproduce every reviewed intermediate exactly: PRK, K_id, K_object_root, K_signature_context, public key, signature input, deterministic signature, canonical signed bytes, OBJECT_ID, K_object, nonce, AAD, padded envelope, ciphertext and tag. The corpus additionally pins the exact signature domain, length prefix and complete 2048-byte file by literal copying/concatenation. Independent signed-to-unsigned parsing and externally stored file decryption also pass.

INITIAL and REWRAP reproduce K_wrap, all header/salt/nonce offsets, wrapped root, tag and complete 87-byte record; both externally stored records unwrap to the same reviewed root. Additional tests check every unsupported bootstrap version, strict UTF-8/source-character failure, empty/exactly-1024-byte actual wrapping, no Unicode normalization, wrong password, tampering, and incorrect AAD boundaries. Preflight rejection counters prove Argon2id is not called for framing/version/password-domain failures.

TOTP covers all supported algorithms and digit counts, periods 1/30/2^31/(2^32-1), boundary behavior, negative time, invalid periods, and counter overflow. Exact values come from the historical counter assertions and RFC 6238 Appendix B (using Appendix A's algorithm-specific secret lengths). The unsigned maximum counter is additionally tested for acceptance at several periods; that test asserts range behavior, not an invented reviewed OTP expectation.

Go's standard Ed25519 verifier accepts both pinned genuine signatures. There are zero supplied reviewed negative Ed25519 cases to compare. No strict profile is claimed or exposed through a misleading generic-verifier wrapper. The required canonical/subgroup/mixed-order acceptance corpus and strict implementation remain BLOCKED pending that corpus, as required by the agent instructions.

The literal token model implements causal and typed field frontiers, JOIN, historical credential ancestry, maximal unresolved lifecycle witnesses with transition explanations, candidate-only provisional coverage, final resolution coverage, ordinary tombstone/credential restrictions, pending dependencies, invalid dependencies, redundant context rejection and cycle rejection. Its tests include arrival-order permutations, JOIN laws, field-ancestry implications, and injected nonempty candidate POST that must not contribute accepted coverage. Model inputs assume completed identity/structure/signature checks; they do not establish end-to-end signature or writer conformance.

## Outstanding deliverables and missing inputs

The following remain **incomplete implementation/review work**, rather than passing tests or newly discovered spec contradictions:

- Complete reviewed r36 semantic/lifecycle corpus and independent differential oracle evidence beyond the 15 initial spec-derived scenarios.
- DEVICE_UPDATE semantic/presentation history validation.
- Durable pending TOKEN_UPDATE/DEVICE_UPDATE evidence, exact-ID abandonment and later re-entry, future evidence/bypass, and known-history regression.
- Remembered-frontier replacement ordering, non-atomic durability batches, establishment-state races and restart behavior.
- Writer confirmation freshness, publication linearization/freshness, actual filesystem/power-loss fault injection and hostile filesystem cases.
- Capacity-blocked >32-head operations and migration selection from capacity-blocked source state.

The seed contains `LifecycleOracle.java` and review reports, but no complete reviewed language-neutral r36 lifecycle/recovery/transition fixture corpus. Historical implementation code was not promoted to normative expected behavior. Status files under `vectors/v0/transitions/` and `vectors/v0/recovery/` record these omissions.

Referenced historical material absent from this repository includes `R17Review.java` (including `strict` and primitive/Ed25519 evidence), `R20Review`, `R26Review`, `R26Oracle`, `BootstrapV0.java`, `review/run.sh`, and `review/run-bootstrap.sh`. The historical checksum file also refers to an r32 spec and a differently named bootstrap report not supplied under those exact paths. Existing r32 review prose and the independent bootstrap vectors were preserved as supplied. Old reports' aggregate pass counts were not treated as locally reproduced results.

Every supplied standalone vector fixture was normalized. No supplied vector was rejected for unknown provenance. Historical reports and implementations are inventoried provenance/methodology, not themselves complete vector artifacts. Newly authored cases explicitly say `spec-derived`; external RFC cases identify their source URL and derivation of reduced digit counts.

No genuine contradiction was found in the implemented r36 byte/crypto or token-model portions. This is not an audit claim about the unimplemented portions. No spec change was made, and no contradiction was silently resolved to obtain passing tests.
