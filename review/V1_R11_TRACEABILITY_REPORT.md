# V1 r11 traceability report

## Purpose

r11 restores revision-history traceability for governance/historical changes made
after the original r10 entry. Baseline: `e98dcc4cba2a97dc824f70128a24e9c9b20c3596`
on `main`, with a clean working tree and passing baseline check/conformance/verify
(77/77 cases). Baseline hashes, including all 77 case files, were captured outside
the repository at `/tmp/totipo-r11-baseline.json`.

## Changes recorded

- cooperative rolling-upgrade limitation;
- OBJECT_VERSION allocation governance;
- v0 historical/non-deployed status;
- removal of normative v0 migration section;
- resulting section renumbering.

These body changes already existed at the baseline. This cut changes revision
metadata/history, not their requirements. The original r10 summary and entire
r10-and-earlier revision history remain byte-for-byte unchanged. Historical
reports and the root r10 source artifact remain unchanged. References dating
the storage-family work to r10 remain historical references.

The open-work heading now names r11. Only the completed canonical TOKEN/DEVICE
fixture-generation item was removed (with list renumbering): the existing
`v1.crypto.token-root.001`, `v1.crypto.token-child.001`,
`v1.crypto.device-root.001`, `v1.provenance.der-short-valid.001`, and
`v1.provenance.der-max-valid.001` cases provide the evidence. Other items retain
remaining coverage, independent/platform verification, and live implementation
work; passing the current corpus is not evidence that all such work is complete.

## Protocol impact

- wire format changed: NO
- runtime semantics changed: NO
- crypto changed: NO
- storage-family rules changed: NO
- vector expected results changed: NO
- case count changed: NO

Go tooling changes only identify, produce, or require r11 repository metadata.
No fixture generator was run. No tag, release, or frozen `v1-rc1` profile was made.

## Regression evidence

- old spec revision: r10; new spec revision: r11
- case count: 77 -> 77; required case IDs: 77
- case IDs added/removed/changed: 0/0/0
- expected results changed: 0
- protocol-bearing vector payloads changed: 0
- object/crypto bytes changed: 0
- case-file hashes changed: 0 of 77
- manifest case entries unchanged in full, including expectations and hashes
- moving profile required case IDs/hashes unchanged

Byte-identical case files preserve semantic bytes, signature inputs, DER
signatures, OBJECT_ID, HKDF outputs, AES-GCM ciphertext/tags, 1024-byte objects,
TOTP inputs/outputs, graph expectations, candidate-use expectations, and
storage-family expectations.

## Files changed

- `CONTRIBUTING.md`
- `Makefile`
- `README.md`
- `conformance/README.md`
- `conformance/cmd/totipo-conformance/main.go`
- `conformance/cmd/totipo-vector-gen/main.go`
- `conformance/internal/vectors/runner.go`
- `requirements/README.md`
- `requirements/v1-pre-rc.json`
- `spec/totipo-vault-format-v1.md`
- `tools/check_spec.py`
- `vectors/README.md`
- `vectors/manifest.json`
- `review/V1_R11_TRACEABILITY_REPORT.md` (new report)

All existing changed files except the spec and structural checker change solely
current revision labels/checks or checksum metadata. The spec additionally adds
revision history/summary and removes the evidenced completed open-work item;
the checker additionally requires the r11 summary and history heading. Neither
changes the protocol contract. The report is new traceability evidence.

## SHA-256

`spec/totipo-vault-format-v1.md`

- r10: `8c93d6e05b0e19682d7875e011aba4961558ff7c606919006b98e3df30219b52`
- r11: `63a61ede5263223754fb4d94409756c5c9512bda7bedee082ae500a02f6466e4`

`vectors/manifest.json`

- r10: `031456a8b35633ab86c9a7c04c6f54ba00697b7577de9f43379ce4932c6bca38`
- r11: `b14dabdb47afdb4863dd26ed27330168178db5e80c283fa21642cdfc1e24a951`

`requirements/v1-pre-rc.json`

- r10: `d7b68ed799000a8c6449d3f114b8b1e62237db04b77400bf0a3c5e1123361b03`
- r11: `9db43bc647befb1826a70fa322f61ac84382eb22c3c402b9b1b91ce64cfeca04`

## Validation

- python3 tools/check_spec.py: PASS
- go -C conformance test ./...: PASS
- go -C conformance vet ./...: PASS
- make check: PASS
- make conformance: PASS (77/77)
- make verify: PASS (77 case files and moving profile)
- make race: PASS
- make fuzz: PASS (10-second targets: FuzzDispatch, FuzzOpen, FuzzArrivalAndDisappearance)
- git diff --check: PASS
- Nix: unavailable in this environment; no Nix files changed
- CI: no r11 CI result available; this local cut has not been pushed

## Ambiguity

No protocol ambiguity found. Remaining pre-RC coverage and platform work was not
claimed complete merely because the current 77-case corpus passes.

## Next step

Stop editing totipo-spec and begin the first independent live Totipo implementation.

v1/r11 restores revision traceability without changing the protocol contract.
The spec repository is ready to stop changing; proceed to the first independent
live implementation.
