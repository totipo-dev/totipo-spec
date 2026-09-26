# V1 r10 integration report

## Baseline

- Starting r9-clean commit: `3b5f808020779c8df3e3323f5ba82e761f4d045c` (`cleanup`), on `main`.
- Old corpus: **71 cases**. Baseline `make check`, `make conformance`, and
  `make verify` all passed before editing.
- Per-case IDs, paths and SHA-256 values, plus manifest/spec hashes, were captured
  outside the repository at `/tmp/totipo-r10-baseline.json`. The recorded Git
  commit permanently identifies the same baseline.
- Baseline manifest SHA-256:
  `c0e7aa0e3633e789f3c6649c21b217f08474904b696809caea9005abd931cd15`.
- Baseline spec SHA-256:
  `99253a68cc926e43f9b16595dd060911f9067f3d15d23293638dd554455e62bb`.

## Specification integration

The normative spec was replaced with the exact supplied r10 artifact, and the
supplied focused review was copied unchanged to
[`V1_R10_ENVELOPE_FAMILY_REVIEW.md`](V1_R10_ENVELOPE_FAMILY_REVIEW.md).
Current-spec SHA-256 is
`e3bb61f079e7b436c609b5fbba47db4116b2a01e3c16f88bf8d33408bd96e485`.

The storage path changes from `objects/` to exact top-level `objects-v1/`.
`OBJECT_VERSION` remains semantic versioning inside the fixed v1 envelope family.
A future envelope family claiming rolling compatibility publishes authenticated
v1-family compatibility assertions; its sibling directory alone conveys no
trusted future-version evidence. TOKEN/DEVICE semantic encodings, routing,
cryptographic domains, IDs, provenance, capacity rules and envelope bytes remain
unchanged.

The structural checker retains r9 invariants and adds the r10 namespace, size,
sibling isolation and compatibility-projection requirements. The only surviving
flat-path reference outside review history is the supplied spec's revision-history
entry describing the rename. It is historical, not a live path assumption.

The supplied spec uses intentional two-space Markdown hard line breaks. A
file-specific `.gitattributes` whitespace rule permits that syntax without editing
the supplied bytes. Other whitespace checks, fixture checks and CI commands remain.

**Spec ambiguity found: none.** No normative edits beyond the supplied r10 were made.

## Go implementation changes

Added a small `internal/storage` environment evaluator, using the existing crypto,
object parser and graph consumer. There was no host-filesystem adapter to migrate.

- Discovery selects only direct regular-file children of `objects-v1/` with exact
  64-character lowercase hexadecimal names. It never invokes content readers for
  siblings, nested paths, traversal spellings, temporary/conflict names, symlinks,
  or special files.
- Wrong sizes and failed envelope authentication yield `INVALID_STORAGE`, not
  opaque graph records. Unreadable candidates or unsafe family-directory kinds
  return errors requiring incomplete discovery rather than successful classification.
- Valid authenticated future objects retain supported/opaque dispatch. Tests cover
  routable TOKEN/DEVICE objects and authenticated opaque-unscoped evidence.
- Compatibility projections are authenticated through existing referenced fixtures
  and inserted into the existing graph model. The future-family representation is
  never read or parsed. Without a projection, its existence changes no v1 state.
- Fixture references must resolve to existing hash-verified envelope cases in the
  same vault. Missing/recursive references, duplicate environment paths, and mixed
  content descriptions are rejected. Incorrect graph expectations are tested.

This is an in-memory filesystem/environment model, not a production filesystem
implementation. A live adapter must still implement stable bounded no-follow reads,
namespace-rebinding protection, durable persistence and concurrent filesystem safety.
The model rejects unsafe observed kinds; it does not claim OS race protection.
No dependencies, second implementation, or future-family format were introduced.

## New vector cases

| Case ID | Purpose |
|---|---|
| `v1.storage.objects-v1.001` | Supported TOKEN and routable future DEVICE discovery in the exact family namespace |
| `v1.storage.nonobject-name-ignored.001` | Wrong names, nested/traversal paths, symlinks and special files never reach semantic processing |
| `v1.storage.wrong-size-not-opaque.001` | 0/1023/1025-byte candidates are invalid storage, without persistent semantic blocking |
| `v1.storage.unknown-sibling-ignored.001` | Plausible or valid encrypted content in arbitrary sibling namespaces contributes no evidence |
| `v1.future.family-compat-shadow.001` | Only the authenticated v1 compatibility assertion causes scoped opaque degradation |
| `v1.future.family-no-shadow.001` | A future-family representation without projection contributes no authenticated future state |

New total: **77 cases**, exactly six additions (28 byte, 6 negative, 43 semantic).
All new entries have real files, stable IDs, expected results, spec sections and
hashes. The case schema supports environment inputs, authenticated fixture
references and explicit state expectations. The manifest schema already allowed
these categories and was not changed. The moving r10 pre-RC profile includes every
manifest case; no RC profile, release or tag was created.

No-shadow explicitly means that family is **not providing rolling-upgrade
compatibility to v1 for that state**. Candidate selection retains the preexisting
explicit-use warning; unknown sibling names add no warning/block. Expectations are
authored from the spec, not generated by the evaluator. Referenced 1024-byte crypto
fixtures are reused, not duplicated. Generator/consumer ancestry remains shared.

## Regression evidence

- Existing pre-r10 case-file hashes changed: **0 of 71**.
- Existing embedded semantic/object byte changes: **0**; entire case files match.
- Metadata-only exceptions: **none**.
- Deliberate regeneration: **all 81 JSON artifact hashes stable**, including all
  77 cases, manifest, schemas and moving profile.
- Both integrated source artifacts match their supplied files byte-for-byte.
- Archive/v0 remains at `5157f14e1af9d28b6f15db20a51cd384f97c8a37`.
  No history, archive refs or historical tags/releases were changed.

## Validation

PASS: structural spec checks; Go tests and vet; `make check`; `make conformance`
(**77/77**); `make verify`; `make race`; bounded `make fuzz`; formatting;
`git diff --check`; documentation links; baseline hash comparison.

The first graph fuzz run completed roughly 226,000 inputs then returned
`context deadline exceeded` without a failing-input reproducer. Repeating the
unchanged complete `make fuzz` command passed all three targets. No test was
weakened, timeout increased, or graph implementation changed.

Nix files, dependencies, lockfile, and CI workflow are unchanged. `nix` is not
available inside this agent environment, so `nix develop --command make check`
could not run. All existing CI-facing commands passed locally using the current
environment and writable Go build cache. This unpushed integration has not run
remote CI.

Changes remain ready for the user's manual commit. Local input/instruction files
are excluded from implementation artifacts and staging. Nothing was pushed.

## Remaining work

Commit/push the integration and confirm remote CI green, then start the first live
Totipo implementation as the independent vector consumer. Preserve the remaining
crash/persistence, operation-freshness, native-platform P-256, hostile-filesystem,
and external review work before RC freeze. The future-family writer's durable
compatibility-publication ordering remains a live-implementation obligation.

## Exact files modified/added

Modified:

```text
.gitattributes
CONTRIBUTING.md
Makefile
README.md
conformance/README.md
conformance/cmd/totipo-conformance/main.go
conformance/cmd/totipo-vector-gen/main.go
conformance/internal/vectors/runner.go
conformance/internal/vectors/runner_test.go
conformance/internal/vectors/types.go
requirements/README.md
requirements/v1-pre-rc.json
spec/totipo-vault-format-v1.md
tools/check_spec.py
vectors/FORMAT.md
vectors/README.md
vectors/case.schema.json
vectors/manifest.json
```

Added:

```text
conformance/cmd/totipo-vector-gen/storage.go
conformance/internal/storage/storage.go
conformance/internal/storage/storage_test.go
conformance/internal/vectors/storage.go
review/V1_R10_ENVELOPE_FAMILY_REVIEW.md
review/V1_R10_INTEGRATION_REPORT.md
vectors/cases/future/v1.future.family-compat-shadow.001.json
vectors/cases/future/v1.future.family-no-shadow.001.json
vectors/cases/storage/v1.storage.nonobject-name-ignored.001.json
vectors/cases/storage/v1.storage.objects-v1.001.json
vectors/cases/storage/v1.storage.unknown-sibling-ignored.001.json
vectors/cases/storage/v1.storage.wrong-size-not-opaque.001.json
```

## Post-integration acceptance

The r10 integration was subsequently committed and pushed as
`86169f943752fa2b696300acf96d4b762acfc808` (`r10 updates`). The
[remote Totipo v1 conformance run](https://github.com/totipo-dev/totipo-spec/actions/runs/36207631688)
completed successfully for that exact commit, verified through the GitHub Actions
API during the final editorial pass. The execution-time statements above describe
the earlier handoff. This acceptance record does not claim CI for the later editorial
change, which has not been pushed.
