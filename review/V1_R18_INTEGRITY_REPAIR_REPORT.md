# v1/r18 integrity repair report

## Summary

Baseline: clean committed `main`, HEAD
`de1c9e2cb60005db140ef6ba15f1856be2da0c87`, independently confirmed against
`https://github.com/totipo-org/totipo-spec` main with `git ls-remote`.
The current spec revision is **r18**. The committed normative r18 bytes were
already authoritative, but the moving requirements profile retained the spec
hash from before the final editorial edit. This repair regenerates the profile
and closes the checker gap. Normative spec unchanged; generated/profile
integrity metadata repaired. No new revision, commit, tag, release, or push.

## Root cause

A final editorial edit to the r18 revision-history wording occurred after the
previous profile/hash regeneration. The profile therefore pinned the earlier
spec bytes. The hardening report also retained the earlier spec/profile hashes
and validation claims. The old `python3 tools/check_spec.py` passed against the
mismatched committed baseline because it checked prose structure and anchors,
not the profile's artifact hashes. The final manual history wording is intentional
and remains exactly as committed.

## Before/after hashes

All hashes below were computed independently from raw file bytes. Baseline
artifacts were also read using `git show HEAD:<path>` and compared byte-for-byte
with the initially clean worktree. No supplied audit hash was accepted without
recomputation.

| Value | Before | After |
| --- | --- | --- |
| Actual spec SHA-256 | `8a357e75f3ddd92efa954fde2ffc9af33f40a2c2396d6afbf1d5de5f00bc4f8a` | unchanged |
| Profile `spec_sha256` | `e413b142adbbb080eb8d8a825226e084e55ae75fd2685bc8bcb980a8768dadb1` | `8a357e75f3ddd92efa954fde2ffc9af33f40a2c2396d6afbf1d5de5f00bc4f8a` |
| Profile file SHA-256 | `6ff8c082f0bf26e7de0e396f4846992d237f7e08d5abbb80b124a4647f0a0ef6` | `4c7954cd2b59aa0afbbe3c22080cadddf72135d884b186a671266c29d58418df` |

| Artifact / profile pin | Actual SHA-256, unchanged and matching regenerated profile |
| --- | --- |
| `vectors/manifest.json` / `manifest_sha256` | `bd2b52adc05b26e09790f5f7367761b2b86ba3cf8d97d10ed187fcf7213fcf02` |
| `vectors/manifest.schema.json` / `schema_sha256` | `f6dfef831f9b391ef8c9e675024b9cb9cb6cc352858c182439ee8847e55c3647` |
| `vectors/case.schema.json` / `case_schema_sha256` | `d38618f53dcf0a558c389831e8838a07248066beff392812f3d277cc974e25c9` |

The established command `go run ./conformance/cmd/generate-vectors -root .`
changed only `requirements/v1-pre-rc.json.spec_sha256`. Every other profile
field, including `spec_revision: r18` and all `required_cases`, is unchanged.
After inspecting that one-line diff, a second generator run produced zero
additional diff: every file under `vectors/` and the profile was byte-identical
to its pre-second-run snapshot, and `git diff` output was identical.

## Regression prevention

`tools/check_spec.py` now computes SHA-256 directly from the spec, manifest,
manifest schema, and case schema files being validated and compares each with
its profile pin. Missing or mismatched pins exit nonzero with the field, path,
stored value, actual digest, and regeneration command. It neither regenerates
artifacts nor derives expected hashes from other profile fields. All existing
structural checks remain. `make check` already invokes this checker first.

`tools/test_check_spec.py` tests the actual CLI in isolated temporary directories:
current files pass; a raw-byte edit to each artifact fails (including a trailing
spec newline that preserves structural anchors); stale and missing values for
each of the four pins fail. Three new tests with parameterized subtests bring
the Python suite to 13 passing tests. Fixtures do not modify the repository.

## Corpus preservation

Exactly **90 case files**, all byte-identical to committed HEAD. All case
SHA-256 values and every expected outcome remain unchanged. The entire manifest
is byte-identical, so its `cases` array, order, hashes, and outcomes are unchanged.
The profile's `required_cases` IDs, order, and case hashes are unchanged.

SHA-256 of manifest-order `ID + " " + case SHA256 + "\n"` remains:

```text
da4646a7aaec040ca24f6ed2f922a7e9f51f2a3ce77ef0e8a56f37e44ba68316
```

## Validation

| Command/check | Result |
| --- | --- |
| `git status --short`, `git branch --show-current`, `git rev-parse HEAD`, `git remote -v` | Initially clean main; baseline and authoritative organization confirmed. |
| `git ls-remote origin refs/heads/main` | SSH transport unavailable (`ssh` executable absent). |
| `git ls-remote https://github.com/totipo-org/totipo-spec.git refs/heads/main` | PASS: remote main equals baseline HEAD. |
| Baseline raw-byte hash audit (`python3` with `hashlib`, `git show HEAD:<path>`) | PASS: exact reported mismatch reproduced; other artifact hashes match audit. |
| `python3 tools/check_spec.py` before repair | Passed despite stale pin, reproducing the validation gap. |
| `go run ./conformance/cmd/generate-vectors -root .` (twice) | PASS: 90 cases; only spec pin changed; second run zero additional diff. |
| `python3 tools/check_spec.py` after repair | PASS: structural anchors and all four artifact pins. |
| `make check` | PASS: 13 Python tests, uncached Go tests, 90/90 cases, schema and integrity verification. |
| `make conformance` | PASS: 90/90. |
| `make verify` | PASS: schemas, manifest, profile, and 90 case files. |
| `make race` | PASS: all Go packages. |
| `make fuzz` (first attempt) | FAIL: FuzzDispatch ended at 10 seconds with `context deadline exceeded`; remaining targets were not reached on that attempt. |
| `make fuzz` (unchanged retry) | PASS: FuzzDispatch, FuzzOpen, FuzzArrivalAndDisappearance, each configured for 10 seconds and two workers. No code or test changes between attempts. |
| `go -C conformance vet ./...` | PASS. |
| `test -z "$(gofmt -l conformance)"` | PASS. |
| `git diff --check` | PASS, including final report edits. |
| Independent final hash/corpus audit below | PASS: pins match actual files; 90 cases and outcomes unchanged; corpus digest unchanged. |
| `command -v nix` | Unavailable; documented Nix development-shell validation not run. Installed Go/Python used. |

The first fuzz timeout is retained here rather than silently discarded. Its cause
was not established; the identical full retry passed. Cross-platform CI and live
provider durability/UI testing were not run for this metadata/checker repair.

Reproducible independent audit (also run after report edits):

```sh
python3 - <<'PY'
import hashlib, json, subprocess
from pathlib import Path

def committed(path):
    return subprocess.check_output(['git', 'show', 'HEAD:' + path])

def sha(data):
    return hashlib.sha256(data).hexdigest()

profile = json.loads(Path('requirements/v1-pre-rc.json').read_bytes())
for field, path in (
    ('spec_sha256', 'spec/totipo-vault-format-v1.md'),
    ('manifest_sha256', 'vectors/manifest.json'),
    ('schema_sha256', 'vectors/manifest.schema.json'),
    ('case_schema_sha256', 'vectors/case.schema.json'),
):
    data = Path(path).read_bytes()
    assert data == committed(path), path
    assert sha(data) == profile[field], field
    print(field, sha(data))
old = json.loads(committed('requirements/v1-pre-rc.json'))
assert {k: v for k, v in profile.items() if k != 'spec_sha256'} == {
    k: v for k, v in old.items() if k != 'spec_sha256'}
manifest = json.loads(Path('vectors/manifest.json').read_bytes())
cases = manifest['cases']
assert len(cases) == 90
assert cases == json.loads(committed('vectors/manifest.json'))['cases']
assert {str(p) for p in Path('vectors/cases').rglob('*') if p.is_file()} == {
    'vectors/' + e['path'] for e in cases}
for entry in cases:
    path = 'vectors/' + entry['path']
    data = Path(path).read_bytes()
    assert data == committed(path), path
    assert sha(data) == entry['sha256'], path
corpus = sha(''.join(e['id'] + ' ' + e['sha256'] + '\n' for e in cases).encode())
assert corpus == 'da4646a7aaec040ca24f6ed2f922a7e9f51f2a3ce77ef0e8a56f37e44ba68316'
print('PASS: 90 unchanged cases/outcomes and corpus', corpus)
print('Profile file SHA-256:', sha(Path('requirements/v1-pre-rc.json').read_bytes()))
PY
```

## Semantic compatibility

The entire normative spec, including final manual revision-history wording, is
byte-identical to baseline HEAD. All protocol source code and vector files are
unchanged. This repair changes none of: wire format; cryptography; TOKEN grammar;
object framing; size limits; graph semantics; fold semantics; current-state
semantics; TOTP semantics; storage protocol behavior; application requirements;
or vector outcomes. No normative r18 protocol/application behavior changed.

## Files changed

| File | Reason |
| --- | --- |
| `requirements/v1-pre-rc.json` | Generator repairs stale spec pin. |
| `tools/check_spec.py` | Validate four profile hashes against actual artifact bytes. |
| `tools/test_check_spec.py` (new) | CLI regression coverage for artifact drift and missing/stale pins. |
| `review/V1_R18_HARDENING_REPORT.md` | Correct hashes, qualify historical validation, record repair and current status. |
| `review/V1_R18_INTEGRITY_REPAIR_REPORT.md` (new) | This focused audit and validation evidence. |

## `git diff --stat`

Against baseline HEAD; Git excludes the two untracked new files from this stat.

```text
 requirements/v1-pre-rc.json       |  2 +-
 review/V1_R18_HARDENING_REPORT.md | 93 +++++++++++++++++++--------------------
 tools/check_spec.py               | 19 +++++++-
 3 files changed, 64 insertions(+), 50 deletions(-)
```

## `git status --short`

```text
 M requirements/v1-pre-rc.json
 M review/V1_R18_HARDENING_REPORT.md
 M tools/check_spec.py
?? review/V1_R18_INTEGRITY_REPAIR_REPORT.md
?? tools/test_check_spec.py
```
