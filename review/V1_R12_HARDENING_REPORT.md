# V1 r12 hardening report

## Baseline

- r11 commit: `08d7f7ad46f53956c1554233a3bb68f17753eaf8` (`v11 tracability`).
- Working branch: `main`, the current v1 design branch.
- The user had already applied the supplied r12 artifact to the live spec and
  supplied `review/V1_R11_SPEC_REVIEW.md`. The spec was used exactly as supplied;
  its SHA-256 was recorded before integration and verified unchanged afterward.
- An unrelated pre-existing `flake.nix` modification was preserved and excluded
  from the integration commit. The local execution instructions were not staged.
- To validate a clean r11 tree without disturbing those changes, baseline checks
  ran in a detached checkout at `/tmp/totipo-r11-baseline-08d7f7a`.
- Baseline `make check`, `make conformance`, and `make verify`: PASS, 77/77.
- `/tmp/totipo-r12-baseline.json` records the r11 spec, manifest, profile, and all
  77 case-file SHA-256 values; it is local evidence outside the repository.

## Spec changes

The supplied r12 text defines active opaque-unscoped evidence and explicit
continuity reset/re-baseline, all-supported-head DEVICE convergence, remote
versus local corruption, safe trusted ECDSA signing, initial DEVICE publication,
and signature-context vault binding. It also clarifies empty-password creation,
direct-HMAC local vault binding, new TOTP secrets, and version governance.
The r11 full review is preserved verbatim as historical motivation.

## Protocol impact

- Wire format changed: NO.
- Crypto construction changed: NO.
- Routing/namespace changed: NO.
- Capacity changed: NO.
- TOTP algorithm changed: NO.
- Runtime/state-machine requirements changed: YES.

No codec, crypto construction, TOTP implementation, or signer dependency changed.
The case schema additions describe test workflows and introduce no wire fields.

## State model

Opaque-unscoped evidence remains sticky after disappearance. Compatible
reclassification requires the exact authenticated semantic digest and successful
persistence. Explicit confirmed reset enters continuity unknown, retains the old
epoch while a replacement scan is built, and completes only after a complete,
consistent, persisted baseline. Absent old evidence is not copied; present
unscoped evidence is rediscovered. Failed/incomplete replacement scans cannot
restore operations or replace the old epoch.

Supported rejected/unresolved DEVICE heads supply no friendly names but remain
causal. Rename derives parents from the entire supported current frontier; opaque
heads block it. Verified names can remain usable alongside inert names.

Remote absence, unreadable bytes, wrong size, AEAD failure, padding failure, and
keyed-ID failure retain durable nodes and parents. Values become unavailable
unless an exact trusted copy exists. Local security-memory corruption blocks
ordinary and candidate use; existing cycle/same-ID inconsistency cases still pass.

A focused single-vault/local-key workflow checks assertion validity, provenance,
identity/key matching and DEVICE durability before reporting first TOKEN success.
TOKEN-before-DEVICE publication can occur, but success cannot be reported until
both are durable. A DEVICE-first sequence succeeds after TOKEN publication.

Cross-vault provenance evidence reuses the fixed P-256 test key, canonical
unsigned TOKEN, and existing signature from `v1.crypto.token-root.001`. It verifies
under context A and is rejected under a distinct fixed context B, without signing.

## New cases

- `v1.future.unscoped-sticky.001`
- `v1.future.unscoped-rebaseline-absent.001`
- `v1.future.unscoped-rebaseline-present.001`
- `v1.device.rename-incorporates-rejected-head.001`
- `v1.graph.known-id-corrupt-bytes-retain-node.001`
- `v1.graph.local-security-memory-corruption.001`
- `v1.provenance.initial-device-before-token.001`
- `v1.provenance.signature-context-cross-vault.001`

The DEVICE case includes both REJECTED and UNRESOLVED heads, verifies their
ancestry after rename, and checks the opaque-head block. The sticky case also
exercises compatible reclassification. Unit tests cover incomplete, failed
persistence, and inconsistent replacement scans, plus unsafe reclassification.

## Regression evidence

- Baseline existing cases: 77.
- Existing case hashes changed: 0.
- Existing expected results changed: 0.
- Existing protocol-bearing payloads changed: 0.
- Existing crypto/object bytes changed: 0.
- New cases: 8.
- Final corpus: 77 + 8 = 85.

All original manifest entries and per-case profile pins remain unchanged.
Full-file hash equality proves that old expected results and byte fixtures are
unchanged. The generator was tested against an isolated copy under `/tmp` and
preserved all 85 case-file bytes; the working corpus was not regenerated.

| Artifact | r11 SHA-256 | r12 SHA-256 |
|---|---|---|
| `spec/totipo-vault-format-v1.md` | `63a61ede5263223754fb4d94409756c5c9512bda7bedee082ae500a02f6466e4` | `2c601766845315f5a8c2122c80eb76bf882a9bcdc59edfedd3b86ce6b74fc7f6` |
| `vectors/manifest.json` | `b14dabdb47afdb4863dd26ed27330168178db5e80c283fa21642cdfc1e24a951` | `7bbc4e3b645246f0bfa51648097d4954f263a5aa9cb11b2d09b36575d1ae3c2e` |
| `requirements/v1-pre-rc.json` | `9db43bc647befb1826a70fa322f61ac84382eb22c3c402b9b1b91ce64cfeca04` | `6a2426557f0c6395ec4291ef9c03acd32cf9beefbdbdf427d81c9c3d6827abb3` |

## Validation

- `python3 tools/check_spec.py`: PASS, r12 structure and semantic-presence checks.
- `go -C conformance test ./...`: PASS.
- `go -C conformance vet ./...`: PASS.
- `make check`: PASS.
- `make conformance`: PASS, 85/85.
- `make verify`: PASS, schemas, hashes and moving profile for 85 cases.
- `make race`: PASS.
- `make fuzz`: PASS, bounded 10-second runs for FuzzDispatch, FuzzOpen, and
  FuzzArrivalAndDisappearance, each with two workers.
- `git diff --check`: PASS.
- Stale revision search: only historical r11 references remain.
- Nix: unavailable on PATH; no Nix configuration was changed by this integration.
- CI: no result for this integration; no push was performed.

## Exact files changed

- `CONTRIBUTING.md`
- `Makefile`
- `README.md`
- `conformance/README.md`
- `conformance/cmd/totipo-conformance/main.go`
- `conformance/cmd/totipo-vector-gen/hardening.go`
- `conformance/cmd/totipo-vector-gen/main.go`
- `conformance/internal/graph/graph.go`
- `conformance/internal/graph/graph_test.go`
- `conformance/internal/graph/publication.go`
- `conformance/internal/vectors/hardening.go`
- `conformance/internal/vectors/runner.go`
- `conformance/internal/vectors/types.go`
- `requirements/README.md`
- `requirements/v1-pre-rc.json`
- `review/V1_R11_SPEC_REVIEW.md`
- `review/V1_R12_HARDENING_REPORT.md`
- `spec/totipo-vault-format-v1.md`
- `tools/check_spec.py`
- `vectors/FORMAT.md`
- `vectors/README.md`
- `vectors/case.schema.json`
- `vectors/cases/device/v1.device.rename-incorporates-rejected-head.001.json`
- `vectors/cases/future/v1.future.unscoped-rebaseline-absent.001.json`
- `vectors/cases/future/v1.future.unscoped-rebaseline-present.001.json`
- `vectors/cases/future/v1.future.unscoped-sticky.001.json`
- `vectors/cases/graph/v1.graph.known-id-corrupt-bytes-retain-node.001.json`
- `vectors/cases/graph/v1.graph.local-security-memory-corruption.001.json`
- `vectors/cases/provenance/v1.provenance.initial-device-before-token.001.json`
- `vectors/cases/provenance/v1.provenance.signature-context-cross-vault.001.json`
- `vectors/manifest.json`

## Remaining work

Independent live implementation, native platform P-256 verification, production
crash/persistence/freshness/filesystem behavior, and external review remain
required before an RC freeze. This in-memory model does not claim production
durability or a complete writer implementation. There are no integration blockers.
No tag, release, or frozen `v1-rc1` profile was created.

r12 hardening is integrated with no wire-format regression. After commit/push
and green CI, stop changing `totipo-spec` unless the independent live
implementation exposes a concrete issue.
