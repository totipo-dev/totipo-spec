# Totipo v0-rc1 — Release-Profile and r37 Promotion Agent Instructions

You are working in the local `totipo-dev/totipo-spec` repository.

This task prepares the repository for a **reviewable `v0-rc1` release candidate**.

Do **not** commit, push, tag, publish a GitHub Release, or merge dependency-update PRs.

Do **not** redesign the protocol or change normative vector expectations.

The local working tree currently has:

- the existing canonical r36 spec at `spec/totipo-vault-format-v0.md`;
- an uncommitted r37 spec file next to it, expected to be named something like `spec/totipo-vault-format-v0-r37.md`.

Revision 37 is intended to be **status/evidence bookkeeping only**. It must not change wire bytes, cryptography, semantic validity, accepted history, limits, recovery rules, or writer behavior.

The goal is to:

1. verify r37 really is bookkeeping-only;
2. promote r37 into the stable canonical spec path;
3. remove the duplicate revisioned r37 working copy;
4. introduce a machine-readable, versioned conformance requirement set for `v0-rc1`;
5. teach the conformance runner to execute that frozen requirement profile;
6. update project documentation;
7. leave everything uncommitted for human review.

---

# 1. Preserve the current baseline

Before changing anything, record:

```bash
git status --short
git diff -- spec/
sha256sum spec/totipo-vault-format-v0.md
(cd vectors/v0 && sha256sum -c manifest.sha256)
go -C conformance test -count=1 ./...
go -C conformance test -race -count=1 ./...
go run ./conformance/cmd/totipo-conformance ./vectors/v0
```

Also run all current review-inventory checks documented by the repository.

Expected normative corpus result:

```text
388 PASS
0 FAIL
0 BLOCKED
```

Do not regenerate vectors.

Record the current:

- r36 canonical spec SHA-256;
- `vectors/v0/manifest.sha256` SHA-256;
- current list/count of normative case IDs.

---

# 2. Verify r37 before promoting it

Locate the uncommitted r37 file beside the canonical spec.

Compare it with:

```text
spec/totipo-vault-format-v0.md
```

The intended r37 changes are bookkeeping/status only.

Verify specifically:

- Sections 1–73 are byte-for-byte identical to r36.
- Section 74 changes no required test semantics or protocol expectations; only stale freeze-status wording may change.
- Section 75 changes only release/evidence bookkeeping.
- Header/revision/status text may change.
- No frozen literal changes.
- No numeric/tag/enum/limit changes.
- No semantic transition changes.
- No recovery or durability-rule changes.
- No signature/KDF/AEAD/domain-string changes.
- No object/bootstrap byte-layout changes.

If any non-bookkeeping difference is found, **stop promotion** and create:

```text
R37_PROMOTION_ISSUE.md
```

describing the exact diff and why it is not bookkeeping-only.

Do not silently accept unexpected spec changes.

The expected reviewed r37 SHA-256 from the preparation work is:

```text
aa281a757be6c324e11b492d195f1b4e5c1604fe9bdf7d20cff9adbfe62bc409
```

If the local r37 file does not have that hash, do not assume it is wrong automatically; compare contents and report the reason for the difference. But do not promote an unexplained mismatch.

---

# 3. Promote r37 into the canonical spec path

The repository should continue to have one stable canonical spec path:

```text
spec/totipo-vault-format-v0.md
```

Do **not** make users or tooling follow revision-specific filenames.

Once r37 is verified:

1. replace the contents of `spec/totipo-vault-format-v0.md` with the verified r37 contents;
2. remove the separate uncommitted `spec/totipo-vault-format-v0-r37.md` copy;
3. update any repository references that explicitly say the canonical spec is revision 36;
4. do not create an archive copy of r36 unless the repository already has a deliberate historical-spec policy.

Git history is the revision archive.

After promotion:

```bash
sha256sum spec/totipo-vault-format-v0.md
```

should identify the exact canonical r37 bytes.

Do not rename frozen protocol literals such as:

```text
TOTP-VAULT
TOTP-Vault/v0/...
```

---

# 4. Add a versioned requirements directory

Create:

```text
requirements/
```

Add:

```text
requirements/README.md
requirements/v0-rc1.json
```

The purpose of `requirements/v0-rc1.json` is to define exactly what it meant to conform to **Totipo v0-rc1**, even if new tests are added later.

Future additions to `vectors/v0/` must not retroactively expand the `v0-rc1` requirement set.

---

# 5. Define `requirements/v0-rc1.json`

Use a simple, language-neutral JSON schema.

The profile must include at least:

```json
{
  "profile": "v0-rc1",
  "protocol": "Totipo Vault Format v0",
  "spec": {
    "path": "spec/totipo-vault-format-v0.md",
    "revision": 37,
    "sha256": "<canonical-r37-sha256>"
  },
  "vectors": {
    "root": "vectors/v0",
    "manifest_path": "vectors/v0/manifest.sha256",
    "manifest_sha256": "<sha256-of-manifest-file>"
  },
  "requirements": {
    "case_ids": [
      "..."
    ],
    "expected": {
      "pass": 388,
      "fail": 0,
      "blocked": 0
    }
  }
}
```

You may add useful metadata such as:

```text
created_from
description
category_counts
required_categories
```

but keep the schema small and auditable.

Do not include volatile data such as:

- timestamps required for validation;
- local filesystem paths;
- current Go version;
- CI run IDs;
- machine-specific integration output.

---

# 6. Freeze exact required case IDs

Enumerate the current normative vector corpus.

Collect every normative case ID from `vectors/v0/**/*.json` according to the existing vector schema.

Requirements:

- exactly 388 required IDs;
- IDs sorted lexicographically in `requirements/v0-rc1.json`;
- no duplicates;
- every required ID exists exactly once in the current corpus;
- no missing required category;
- the existing current corpus produces the expected 388/0/0 result.

Do not merely freeze category counts.

The exact case-ID set is what prevents later test additions from silently changing `v0-rc1`.

If the current corpus contains non-normative/status JSON files that the existing runner does not count as normative cases, do not add artificial IDs for them. Follow the existing corpus-loader semantics.

Document the derivation in `requirements/README.md`.

---

# 7. Freeze useful category counts too

In addition to the exact ID set, include category counts as a review aid.

Use the current Phase 2/3 normative corpus categories and actual loader output, not hand-entered guesses.

Expected current total:

```text
388
```

The current known category breakdown from Phase 2 is approximately:

```text
TLV                                      65
Envelope/file-length gates               13
Object crypto                             2
Bootstrap/password                       37
Dispatch/processing order                56
TOTP                                     82
Strict Ed25519                           33
Retained lifecycle scenarios             15
Transitions/application-use policy       46
Recovery/security-memory traces          28
DEVICE_UPDATE presentation               11
```

Verify these against the actual corpus before recording them.

Do not change vectors to make the counts match this list.

---

# 8. Decide what the profile does NOT require

`v0-rc1.json` is the **normative protocol/vector conformance profile**.

Do not make Linux/ext4 Phase 3 integration evidence a mandatory requirement for every independent Totipo implementation.

In particular, do not require another implementation to reproduce:

- the reference implementation's ext4 syscall choices;
- Linux `renameat2`;
- a particular local-state storage format;
- the exact Phase 3 crash-test count;
- Go-specific behavior.

Those are reference/integration evidence, not portable protocol-vector requirements.

The requirements README should explicitly distinguish:

```text
v0-rc1 normative conformance profile
vs.
Totipo reference implementation / Linux integration evidence
```

You may link to the Phase 3 report as supporting release evidence, but do not put platform-specific cases into the portable 388-case requirement set unless they are already normative vectors.

---

# 9. Add requirements-profile validation code

Create a small reusable package, for example:

```text
conformance/internal/requirements/
```

It should:

1. load a requirements JSON file;
2. validate its schema;
3. verify the canonical spec file SHA-256;
4. verify `vectors/v0/manifest.sha256` SHA-256;
5. verify every listed case ID exists exactly once;
6. reject duplicate required IDs;
7. reject required IDs absent from the corpus;
8. verify expected category counts if present;
9. verify the expected total;
10. allow additional corpus cases not listed in the profile without making the profile invalid.

That last point is important:

> Future vectors may be added to `vectors/v0/`; running `v0-rc1` must continue to run only the IDs frozen into `requirements/v0-rc1.json`.

---

# 10. Extend the conformance CLI

Extend:

```bash
go run ./conformance/cmd/totipo-conformance ...
```

to support a requirements profile.

Preferred interface:

```bash
go run ./conformance/cmd/totipo-conformance \
  --requirements requirements/v0-rc1.json
```

or, if the existing positional corpus path must remain:

```bash
go run ./conformance/cmd/totipo-conformance \
  ./vectors/v0 \
  --requirements requirements/v0-rc1.json
```

Choose the form that best preserves existing CLI compatibility.

Behavior with `--requirements`:

- validate spec hash;
- validate vector-manifest hash;
- load the full corpus;
- select exactly the required case IDs;
- run only those IDs for profile conformance;
- report PASS/FAIL/BLOCKED;
- return nonzero if:
  - profile integrity fails;
  - a required case is missing;
  - a required case fails;
  - a required case is blocked;
  - expected counts do not match.

Extra newly-added corpus cases must not affect `v0-rc1` results.

Preserve the existing no-profile behavior for development/current-corpus testing.

---

# 11. Support filters without weakening the profile

If the CLI already supports:

```text
--filter
```

then filtering a requirements profile is useful for diagnostics.

However:

```bash
--requirements requirements/v0-rc1.json --filter ...
```

must not be reported as a complete profile-conformance success.

Clearly report it as a filtered/partial run.

A full `v0-rc1` conformance claim requires executing the complete frozen requirement set.

---

# 12. Add tests for profile stability

Add tests proving:

## Valid profile

The current canonical r37 + current frozen vector corpus passes `v0-rc1`.

## Spec changed

Mutating a temporary copy of the spec causes requirements validation failure.

## Manifest changed

Mutating the vector manifest causes validation failure.

## Required case missing

Removing one required case causes failure.

## Duplicate required ID

Duplicate ID in the profile causes failure.

## Unknown required ID

Causes failure.

## Extra new vector case

Add a temporary valid extra corpus case not listed in `v0-rc1`.

Expected:

```text
current-corpus mode sees it
v0-rc1 mode ignores it
v0-rc1 still means exactly the original frozen ID set
```

## Later changed expectation

A required case whose expected normative data changes should be detected through the vector-manifest hash/profile integrity path.

---

# 13. Add a requirements-profile generation tool, but never automatic regeneration

Create an explicit maintenance command such as:

```bash
go run ./tools/requirements/main.go \
  --profile v0-rc1 \
  --write
```

or an equivalent tool.

It may:

- enumerate normative case IDs;
- calculate hashes;
- calculate category counts;
- write a candidate profile.

It must require an explicit write flag.

Normal tests and CI must never regenerate:

```text
requirements/v0-rc1.json
```

The committed profile is reviewed release metadata.

Once `v0-rc1` is tagged, treat that profile as immutable.

A later release candidate gets a new profile such as:

```text
requirements/v0-rc2.json
```

Do not edit `v0-rc1.json` after release except to correct a demonstrable repository mistake, which would itself require explicit release handling.

---

# 14. Add Makefile targets if the repo uses Make

If a Makefile already exists, add intuitive targets such as:

```text
make conformance
make conformance-v0-rc1
make verify-requirements
```

Do not introduce Make solely for this task if the repository does not already use it.

The full profile target should execute the frozen profile, not the moving full corpus.

---

# 15. Update CI

Add a CI step that runs:

```bash
go run ./conformance/cmd/totipo-conformance \
  --requirements requirements/v0-rc1.json
```

using the final supported invocation.

This should be separate from the ordinary "run current corpus" test where practical.

CI should therefore detect both:

1. regressions in the current development corpus;
2. regressions against the immutable `v0-rc1` profile.

Do not regenerate the requirements file in CI.

Preserve pinned GitHub Action SHAs.

---

# 16. Update the root README

Update the current root README to say revision 37 is the canonical release-candidate spec.

Prefer wording such as:

```text
Current protocol status: Totipo Vault Format v0 release candidate, revision 37.
```

Add a short conformance section explaining:

```text
requirements/v0-rc1.json
```

pins the exact portable conformance target for `v0-rc1`.

Include the command to run it.

Explain:

- `vectors/v0/` may grow over time;
- `requirements/v0-rc1.json` does not;
- extra future tests do not retroactively alter the `v0-rc1` target.

Do not claim the Git tag/release already exists.

---

# 17. Update conformance documentation

Update:

```text
conformance/README.md
vectors/v0/README.md
```

as appropriate.

Document the distinction:

```text
moving current v0 corpus
frozen release profile
```

Recommended terminology:

- **corpus** = all currently committed v0 normative vector cases;
- **requirements profile** = immutable subset/hash target for a named release;
- **reference integration evidence** = implementation/platform evidence, not portable vector conformance.

---

# 18. Handle r37 supporting files

If the local working tree also contains r37 preparation artifacts such as:

```text
totipo-vault-v0-r37-release-note.md
totipo-vault-v0-r37-SHA256SUMS.txt
```

do not automatically put them in the repository root.

Use project structure intentionally.

A good location for a release-preparation note is:

```text
review/releases/v0-rc1/
```

or:

```text
review/process/
```

If these artifacts duplicate information now captured in Git history + requirements metadata, it is acceptable not to commit them.

Do not commit a redundant second canonical spec copy.

The one authoritative spec path remains:

```text
spec/totipo-vault-format-v0.md
```

---

# 19. Add release evidence metadata separately if useful

Optionally create:

```text
requirements/v0-rc1-evidence.json
```

ONLY if it provides useful release evidence without confusing portable conformance.

If created, it may reference hashes for:

- Phase 2 report/review inventory;
- Phase 3 source/results;
- Linux integration evidence.

It must clearly say:

```text
informational release evidence
not required for portable v0-rc1 protocol conformance
```

This file is optional.

Do not put machine-specific transient data into the normative requirements profile.

---

# 20. Check r37 references throughout the repo

Search for stale wording such as:

```text
revision 36
r36
byte-level-freeze candidate
Still to freeze
before byte-level v0 freeze
```

Do not blindly replace historical review reports.

Rules:

- current project-facing docs should identify r37;
- historical reports should remain historically accurate;
- historical filenames/content should not be rewritten merely for consistency.

---

# 21. Do not change normative vectors

All current files under:

```text
vectors/v0/
```

must remain byte-for-byte unchanged during this task.

The requirements profile is new release metadata around them.

Verify:

```bash
(cd vectors/v0 && sha256sum -c manifest.sha256)
```

and compare the vector tree/hash against baseline.

If a vector must change, stop and report why; that is outside this release-profile task.

---

# 22. Preserve all protocol/conformance tests

Run:

```bash
gofmt -l conformance tools
go -C conformance vet ./...
go -C conformance test -count=1 ./...
go -C conformance test -race -count=1 ./...
go run ./conformance/cmd/totipo-conformance ./vectors/v0
go run ./conformance/cmd/totipo-conformance \
  --requirements requirements/v0-rc1.json
(cd vectors/v0 && sha256sum -c manifest.sha256)
```

Also run all existing review inventory checks and Phase 3 source checks.

If integration tests need the documented repository-local `TMPDIR`, use it.

Expected moving-corpus result remains:

```text
388 PASS
0 FAIL
0 BLOCKED
```

Expected `v0-rc1` profile result must also be:

```text
388 PASS
0 FAIL
0 BLOCKED
```

at the moment this profile is created.

---

# 23. Verify spec and profile hashes explicitly

The final report must state:

```text
canonical spec SHA-256
vector manifest file SHA-256
requirements/v0-rc1.json SHA-256
number of required IDs
category counts
```

The canonical r37 spec is expected to hash to:

```text
aa281a757be6c324e11b492d195f1b4e5c1604fe9bdf7d20cff9adbfe62bc409
```

If it does not, explain the discrepancy.

Do not conceal it by updating the profile silently.

---

# 24. Do not tag or release

Do NOT:

- commit;
- push;
- create `v0-rc1` Git tag;
- create GitHub Release;
- merge Dependabot PRs;
- change repository settings.

Leave the complete working tree for human review.

---

# 25. Produce a review report

Create:

```text
V0_RC1_REQUIREMENTS_REPORT.md
```

Include:

## r37 promotion

- r36 original canonical hash;
- r37 source path/hash before promotion;
- verification that r37 was bookkeeping-only;
- final canonical path/hash;
- whether the duplicate revisioned r37 file was removed.

## Requirements profile

- path;
- profile SHA-256;
- spec SHA-256;
- vector-manifest SHA-256;
- exact required case count;
- category counts;
- how IDs were enumerated.

## CLI changes

- exact invocation;
- full-profile behavior;
- partial/filter behavior;
- behavior with extra future vectors.

## Tests

- current-corpus result;
- `v0-rc1` profile result;
- profile-integrity negative tests;
- race tests;
- integration tests where applicable.

## Files changed

List every added/modified/deleted file.

## Release readiness

State remaining human actions:

- review working tree;
- commit;
- obtain green CI on the release commit;
- optionally review/update GitHub repository metadata;
- create/tag `v0-rc1`;
- create GitHub Release.

Do not claim the tag/release exists yet.

---

# Acceptance criteria

This task is complete when:

1. r37 is verified as bookkeeping-only.
2. `spec/totipo-vault-format-v0.md` contains canonical r37.
3. no duplicate revisioned r37 spec remains in the working tree.
4. current normative vectors remain unchanged.
5. `requirements/v0-rc1.json` freezes exactly the current 388 required case IDs.
6. the profile pins canonical r37 and the vector manifest by SHA-256.
7. later extra vectors do not retroactively alter `v0-rc1`.
8. the CLI can execute exactly the frozen profile.
9. CI is configured to verify the frozen profile.
10. current corpus and frozen profile both pass 388/0/0.
11. README/conformance docs explain the release profile.
12. no commit, push, tag, release, dependency upgrade, or protocol redesign occurs.

The human maintainer will review and commit these changes before any `v0-rc1` tag is created.
