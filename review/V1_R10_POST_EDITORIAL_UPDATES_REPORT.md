# V1 r10 post-editorial updates report

## Changes

Starting commit: `65445082e3b7fc966f7566ee863858e60f5e30e8` on `main`.

Added OBJECT_VERSION allocation governance, replaced the v0 compatibility claim
with an informative historical note, and removed the entire normative migration
procedure. Top-level numbering remains contiguous:

| Previous section | Current section |
|---|---|
| 54 Migration from v0 | Removed |
| 55 Core invariants | 54 |
| 56 Required conformance evidence for v1 | 55 |
| 57 Open work after r10 | 56 |
| 58 v0 concepts intentionally absent from v1 | 57 |
| 59 Revision history | 58 |

Exact files changed:

- `spec/totipo-vault-format-v1.md`
- `tools/check_spec.py`
- `conformance/cmd/totipo-vector-gen/totp.go`
- `vectors/manifest.json`
- `requirements/v1-pre-rc.json`
- `review/V1_R10_POST_EDITORIAL_UPDATES_REPORT.md`

The only Go change updates the TOTP generator's section-reference metadata.
The checker now verifies the new numbering, allocation concepts, historical note,
and absence of the removed migration section. The moving profile pins the updated
spec and manifest; schema and case pins are unchanged.

The completed task to replace v0 field/lifecycle cases was removed: the current
corpus already contains v1 whole-state sequential, conflicting-concurrent, and
equal-concurrent graph cases. The misleading v0 crypto/bootstrap/TOTP porting
line now refers to remaining v1 coverage, without claiming complete release-candidate
evidence. Other open-work items were retained: representative vectors do not by
themselves establish exhaustive coverage or independent platform verification.

The repository cross-reference search found old section citations in the two r8
review reports; these are historical and remain unchanged. The supplied root-level
`totipo-vault-format-v1-r10.md` is a historical source input, not the live spec.
No current spec prose references required correction beyond headings. Archived v0,
revision-history prose, and prior reports remain unchanged.

## Protocol impact

- Protocol revision remains r10.
- Wire format changed: NO.
- Runtime semantics changed: NO.
- Case count or IDs changed: NO.
- Case expected results changed: NO.
- Object, signature, or crypto bytes changed: NO.

No fixture regeneration, Nix/CI changes, release-candidate freeze, tag, or release.
Future opaque-version routing and future-family compatibility projections retain
their existing behavior. Synthetic future-version fixtures allocate no interoperable
version value.

## OBJECT_VERSION governance

Final wording in §42.1:

> `OBJECT_VERSION` values are allocated by the published Totipo specification process for semantic grammars within an envelope family. r10 assigns `0x01`. All other values are unassigned by r10. Conforming implementations MUST NOT independently assign an unassigned value for interoperable/shared-vault use without a published Totipo specification allocating that value. r10 defines no private-use or experimental `OBJECT_VERSION` range.

This assigns only `0x01`, leaves every other value unassigned, requires published
specification allocation for interoperable use, and defines no private-use range.
The field remains `u8`.

## v0 historical status

Final top-level note:

> **Historical note:** Earlier Totipo design work used a v0 draft. It was never released as an implemented/deployed protocol and is not a supported predecessor of v1. v1 defines no migration protocol from v0; historical v0 draft artifacts are outside the v1 protocol.

v0 is design ancestry, not a supported deployed predecessor; v1 specifies no v0
migration protocol.

## Metadata regression

Machine comparison against the pre-edit baseline checked every case's SHA-256,
complete parsed JSON, protocol fields excluding `spec_sections`, ID set, and every
manifest field other than the permitted section metadata:

| Measurement | Result |
|---|---:|
| Total cases | 77 |
| Case file hashes unchanged | 77 |
| Case file hashes changed only by section metadata | 0 |
| Protocol-bearing payload changes | 0 |
| Expected-result changes | 0 |
| Case IDs added | 0 |
| Case IDs removed | 0 |

Changed case-file list: empty. Section references reside in the manifest, not the
case files. Only these three manifest entries changed from `['40', '56']` to
`['40', '55']`; their case hashes are unchanged:

- `v1.totp.rfc6238-sha1.001`
- `v1.totp.rfc6238-sha256.001`
- `v1.totp.rfc6238-sha512.001`

SHA-256 values:

| Artifact | Before | After |
|---|---|---|
| spec | `056303e547b33912967f612e07dd5e1a3f4fd3bc599599348504f60c1b9943a2` | `8c93d6e05b0e19682d7875e011aba4961558ff7c606919006b98e3df30219b52` |
| manifest | `bcc237b47c0ed9ad039659dffde84499778a8551a03529c9af4ed0e36eb88422` | `031456a8b35633ab86c9a7c04c6f54ba00697b7577de9f43379ce4932c6bca38` |
| profile | `7b204b2a8c3d382983dc41083361a6c9ba9bfabf1b15ae9d0e5777abf59874b2` | `d7b68ed799000a8c6449d3f114b8b1e62237db04b77400bf0a3c5e1123361b03` |

Session evidence is stored outside the repository in
`/tmp/totipo-r10-post-editorial-baseline.json`,
`/tmp/totipo-r10-post-editorial-compare.py`, and
`/tmp/totipo-r10-post-editorial-comparison.json`.

## Validation

Baseline `make check`, `make conformance`, and `make verify`: PASS, 77/77.
After edits:

- `python3 tools/check_spec.py`: PASS.
- `go -C conformance test ./...`: PASS.
- `go -C conformance vet ./...`: PASS.
- `make check`: PASS, including four Python tests and all Go tests.
- `make conformance`: PASS, 77/77.
- `make verify`: PASS, manifest/profile/schema and all 77 case files.
- `make race`: PASS.
- `make fuzz`: PASS, all three targets at the existing 10-second duration.
- Machine payload/hash comparison: PASS.
- `git diff --check`: PASS.

Go used the repository-local writable build cache. Nix is unavailable in this
session and was not modified. Remote CI was not verified for these uncommitted
changes. Final status contains the six intended changed/new files and the
pre-existing untracked execution instruction file. Nothing was staged or committed;
the changes are ready for the user's manual commit.

## Remaining work

r10 is now governance-complete and free of a fictitious deployed-v0 migration
obligation. No wire or semantic behavior changed. Stop editing `totipo-spec` and
move to the first live Totipo implementation as the independent vector consumer.
The remaining release-candidate evidence and external review obligations still apply.

Optional future editorial observation: §57 could eventually move to historical
design documentation. It remains intact as requested and is not a blocker.
