# Versioned conformance requirements

A **corpus** contains all currently committed normative vector cases. A
**requirements profile** freezes the exact portable case-ID set and artifact
hashes for a named release. [v0-rc1.json](v0-rc1.json) pins canonical revision 37
and exactly 388 IDs. It is release metadata prepared for review; it does not
assert that a Git tag or GitHub Release already exists.

From the repository root:

```sh
go run ./conformance/cmd/totipo-conformance --requirements requirements/v0-rc1.json
go run ./conformance/cmd/totipo-conformance --requirements requirements/v0-rc1.json --verify-only
go run ./conformance/cmd/totipo-conformance --requirements requirements/v0-rc1.json --filter v0/totp/
```

Equivalent Make targets are `make conformance-v0-rc1` and
`make verify-requirements`. Full execution requires every frozen ID to pass,
with zero failures and zero blocked cases. Filters are diagnostic **PARTIAL**
runs, even when they match the entire set; they never establish full profile
conformance. Integrity-only verification does not execute cases. Invalid
artifacts, missing/duplicate IDs, empty filter results, failures, blocked cases,
or unexpected totals return nonzero.

## Schema and derivation

Paths in a profile use repository-relative forward slashes; profiles reside
directly under `requirements/`. The caller's working directory does not change
artifact resolution. An optional positional corpus directory must satisfy the
same pinned-artifact and current-manifest checks.

The profile's `vectors.manifest_path` is
[`requirements/v0-rc1.manifest.sha256`](v0-rc1.manifest.sha256), an exact copy of
`vectors/v0/manifest.sha256` at profile creation. Its SHA-256 is
`bf6a3a670ddf1acab773e44fe563f2e3d9392299fb9229bddb64042194e474a0`.
Pinning the moving manifest itself forever would prevent corpus additions;
this retained copy resolves that conflict without weakening the original
artifact checks. Both the frozen manifest's exact hash and every file hash
listed in it are verified. The current corpus's own complete manifest is also
validated by the existing loader.

Future cases must be added in **new JSON files**, with the current manifest
updated for review. Existing pinned JSON files cannot be edited, appended to,
renamed or removed while satisfying this profile. A changed expectation fails
even if the current manifest is recalculated. New cases must satisfy the current
loader's schema and unique-ID checks, but only frozen IDs are executed in profile
mode. Invalid extra artifacts still fail corpus validation. A later canonical
spec revision likewise needs the pinned r37 checkout/artifact for this profile;
the spec hash is never silently relaxed.

- `profile` identifies the named target; `protocol` is `Totipo Vault Format v0`.
- `spec` pins the canonical path, positive revision number, and lowercase SHA-256.
- `vectors` identifies the corpus root and exact pinned manifest path/hash.
- `requirements.case_ids` is a nonempty, lexicographically sorted, unique list.
- `requirements.expected` requires every listed case to pass, with zero fail/blocked.
- Optional `requirements.category_counts` must exactly match the required IDs.

Unknown fields, duplicate JSON keys, null values, trailing data, invalid hashes,
and absolute or parent-traversing paths are rejected. The profile is reviewed
metadata, not a signed attestation: its contents and changes still require review.

IDs were enumerated using the existing Go `corpus.Load` implementation, which
recursively reads `vectors/v0/**/*.json`, verifies its complete canonical
manifest, validates schema-1 case bundles, and rejects duplicate IDs. Every
returned case is included; no artificial IDs were created for documentation or
status metadata. Counts were derived from the category component of each ID,
not entered as generation inputs:

| Category | Required cases |
|---|---:|
| bootstrap | 37 |
| dispatch | 56 |
| ed25519 | 33 |
| envelope | 13 |
| lifecycle | 15 |
| object-crypto | 2 |
| presentation | 11 |
| recovery | 28 |
| tlv | 65 |
| totp | 82 |
| transitions | 46 |
| **Total** | **388** |

## Portable conformance and integration evidence

The v0-rc1 normative conformance profile is distinct from Totipo reference
implementation / Linux integration evidence. Another implementation need not
use Go, Linux `renameat2`, the reference local-state format, particular ext4
syscalls, or reproduce Phase 3's crash-test count to execute this portable
388-case target. The [Phase 3 report](../review/phase3/IMPLEMENTATION_REPORT.md)
provides supporting implementation evidence, not additional profile cases.

Once tagged, treat the release profile as immutable. A later release candidate
gets a new profile such as `v0-rc2.json`. A demonstrable repository mistake
requires explicit release handling, never silent regeneration. Tests and CI
only validate profiles; they never rewrite them or normative vectors.

## Explicit candidate generation

The Go maintenance command shares the corpus loader with normal conformance:

```sh
# Preview JSON on stdout; creates no files.
go run ./conformance/cmd/totipo-requirements --profile v0-rc2

# Explicitly write a NEW candidate profile and exact manifest copy for review.
go run ./conformance/cmd/totipo-requirements --profile v0-rc2 --write
```

The original profile was created with `--profile v0-rc1 --write`. Generation
derives the revision from the canonical spec's status header and computes all
hashes, IDs and category counts. It does not claim tests passed; execute the
candidate profile separately. `--repo-root DIR` selects another repository copy.
The tool refuses to overwrite either output, and no Make target or CI step
invokes it. If a write fails partway, inspect the partial candidate files before
retrying; the command will not overwrite them.
