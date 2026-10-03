# v1/r18 hardening report

## Summary

r18 is an application-safety, conformance-scope, compromise-recovery, and editorial
hardening revision. It intentionally strengthens application requirements without
changing portable protocol semantics or conformance-vector outcomes. The baseline
was clean r17 HEAD `3bc3f3507578b9ebdd43129c8c567cf2ea16fae5`.
No commit, tag, release, or push was performed; `totipo-java` was not touched.

## Changes made

| Location in specification | Change |
| --- | --- |
| Preamble and §21 | Identify r18 and its application/editorial scope; preserve r17 and r16 history. |
| §§3, 7 | Keep orphan-tolerant creation and no exhaustive-enumeration requirement. Recommend a prominent possible-existing-vault warning and explicit confirmation when canonical `vault` is absent and canonical-ID-looking entries are observed during available observation. Recommend checking sync/provider state and finding the bootstrap. Names remain unauthenticated contextual evidence and cannot veto initialization. |
| §20 | Define core, store/writer, and application conformance, precise claims and applicable operations. A library cannot guarantee a GUI; data exposure alone is not application conformance. All portable profile cases remain required. |
| §15 | Explain HMAC parent commitments and why genuine resolved cycles are cryptographically unexpected, without claiming impossibility or changing SCC interpretation, traversal, or abstract tests. |
| §13 | Explain ordinary single-path naming, hypothetical keyed-ID collision/failure, and inconsistent provider observations over time. Preserve authentication and the exact defensive identity-exclusion rule. |
| §5 | Require explicit interactive empty-password creation confirmation; recommend explaining offline guessing. Empty passwords remain format-valid and readable; no portable complexity policy is added. |
| §16 | Require truthful deletion-versus-erasure wording. Tombstones retain complete credentials and can remain current graph assertions; logical deletion does not erase history or provider copies. |
| §8 | Consolidate historical-wrapper and same-root rewrap limitations. Prohibit describing rewrap as a complete security reset or root-compromise recovery. |
| §8.1 | Add explicitly informative fresh-vault recovery guidance, issuer credential rotation, and provider-aware retirement. Re-encrypting an exposed secret does not revoke the credential. |
| §§7, 8, 18, 20 | Retain platform-neutral strongest-reasonable crash-safe publication/replacement and durability-result contracts. No platform recipe or new normative syscall requirement was added; no suitable implementation-notes section existed. |
| §20 | Remove time-sensitive Java reconciliation status without replacing it with a new project TODO. |
| §21 and historical archive | Move detailed r1–r15 entries verbatim to a clearly non-normative archive in `review/`, with links in both directions. Removed designs cannot be read as current requirements. |
| §§12, 15–17, 20 | Keep §12 authoritative for exact per-object metadata preservation, refer to it from heads/current/history descriptions, and keep §17 authoritative for operation-wide fold preservation. Retain local anti-erasure and per-stage reminders needed to implement the algorithms safely. |

Summary and cross-document review covered the namespace/creation summary, TOKEN
registry and size arithmetic, graph/current-state descriptions, fold table and
algorithm, observation/publication synopses, §20 evidence checklist, README, and
vector/consumer documentation. Hostile or unavailable evidence does not gain
semantic authority; wrong-sized files still fail validation and cannot become
alternative opaque state. Filenames alone never authenticate. Complete-state,
metadata, publication, stale/incomplete-view, and unresolved-parent invariants
remain intact. No device, signature, persistent object, or root-migration machinery
was introduced.

Final wording review made exactly two spec corrections: §7 now bases possible
existing-vault evidence on entries **observed during available observation**,
consistent with its no-exhaustive-enumeration rule; §20 now says **crash-safe
publication/replacement behavior under Sections 7, 8, and 18**, replacing
“crash-safe `vault` replacement and applicable atomic facilities.” Section 3,
the conformance-scope architecture, and the historical archive were unchanged.
Orphan warning and confirmation remain SHOULD-level; empty-password confirmation
remains MUST-level. No storage primitive or initialization veto was introduced.

The structural checker previously lacked several requested application anchors.
Focused checks now cover empty-password confirmation, observed orphan evidence
and its SHOULD policy, anti-initialization-DoS wording, erasure/rewrap disclosures,
and conformance-scope boundaries. Existing checks were retained without weakening.

## Semantic compatibility

| Contract | Changed? |
| --- | --- |
| Wire format | No |
| Cryptography, algorithms, domains, and constants | No |
| TOKEN encoding, grammar, and tags | No |
| Object framing | No |
| Size limits | No |
| Graph semantics | No |
| Fold semantics and parent rules | No |
| Current-state semantics | No |
| TOTP computation | No |
| Existing vector expected outcomes | No |

Every fenced code block before revision history and every TOKEN registry row is
identical to r17. Entire §§2, 4, 6, 9, 10, 11, 14, 18, and 19 are byte-identical.
The same-ID normative rule is byte-identical. Review of §§15–17 confirms that
changes only add informative/application prose or replace repeated metadata text
with references; SCC, head, equality, ancestry, and fold algorithms are unchanged.
Every Go source is identical after accounting for literal `r17` → `r18` labels;
there are no evaluator, crypto, parser, storage, or TOTP implementation changes.
The archived r1–r15 entries are byte-identical to their former spec text.

## Conformance-corpus verification

**90 cases; all 90 pre-r18 case files are byte-identical, including every input,
expected result, note, and encrypted byte.** Manifest case entries and their order
are unchanged. No case was added, removed, or semantically updated. SHA-256 of
manifest-order `ID + " " + case SHA256 + "\n"` remains:

```text
da4646a7aaec040ca24f6ed2f922a7e9f51f2a3ce77ef0e8a56f37e44ba68316
```

Only revision metadata changed in generated artifacts: manifest `spec_revision`;
manifest-schema revision constant; and profile `spec_revision`, `spec_sha256`,
`manifest_sha256`, and `schema_sha256`. The case schema and its pin are unchanged.
The explicit established generator updated these artifacts. A subsequent run
compared SHA-256 for every vector file and the profile before/after and produced
no differences or new files.

The following suite and baseline audit were rerun for the final wording review.
The final generator was run twice with no further diff on the second run.

| Command/check | Result |
| --- | --- |
| `make check` | PASS: spec structure, 10 Python tests, uncached Go tests, all 90 cases, schemas, physical coverage, hashes, and profile pins. |
| `make race` | PASS. |
| `make fuzz` | PASS: FuzzDispatch, FuzzOpen, and FuzzArrivalAndDisappearance, each 10 seconds with parallelism 2. |
| `go -C conformance vet ./...` | PASS. |
| `test -z "$(gofmt -l conformance)"` | PASS. |
| `git diff --check` | PASS. |
| `go run ./conformance/cmd/generate-vectors -root .` | PASS: 90 cases; repeated final generation byte-identical. |
| Baseline comparison audit | PASS: 90 case bytes, manifest entries, all expected outcomes, Go sources modulo revision labels, protocol code blocks, registry, unchanged sections, and archived history. |

The case comparison can be reproduced from this uncommitted worktree with:

```sh
python3 - <<'PY'
import hashlib, json, subprocess
from pathlib import Path
def old(p):
    return subprocess.check_output(['git', 'show', 'HEAD:' + p])
m = json.loads(Path('vectors/manifest.json').read_bytes())
assert len(m['cases']) == 90
assert m['cases'] == json.loads(old('vectors/manifest.json'))['cases']
assert {str(p) for p in Path('vectors/cases').rglob('*.json')} == {
    'vectors/' + e['path'] for e in m['cases']}
for e in m['cases']:
    p = 'vectors/' + e['path']
    b = Path(p).read_bytes()
    assert b == old(p)
    assert hashlib.sha256(b).hexdigest() == e['sha256']
print('PASS: 90 byte-identical cases and unchanged expectations')
PY
```

Nix was not run: `command -v nix` found no executable on PATH. The installed Go
and Python tools ran the repository's relevant validation suite. The flake defines
a development shell/formatter, not a separate flake checks suite. Cross-platform
CI and real provider crash tests were not run locally; abstract storage cases do
not establish live backend durability or application UI behavior.

Final wording-review SHA-256 values:

| Artifact | SHA-256 |
| --- | --- |
| `spec/totipo-vault-format-v1.md` | `e413b142adbbb080eb8d8a825226e084e55ae75fd2685bc8bcb980a8768dadb1` |
| `requirements/v1-pre-rc.json` | `6ff8c082f0bf26e7de0e396f4846992d237f7e08d5abbb80b124a4647f0a0ef6` |

Relative to the preceding uncommitted r18 draft, the only regenerated artifact
change is the profile’s `spec_sha256`; all vector artifacts remain unchanged.

## Repository consistency

Current spec, README, contributor instructions, Make help, consumer documentation,
vector contract documentation, generator/consumer labels, revision rejection test,
manifest/schema/profile, and structure checker now track r18. A repository search
found remaining active-document r17 mentions only in the comparison preamble,
historical §21 entry, and historical report link. Existing historical reports and
audit tools were preserved. The structure checker now verifies r1–r15 entries in
the linked archive and r16–r18 entries in the spec, retaining all existing protocol
anchors and retired-concept checks. No corpus capability selection was introduced.

## Files changed

| File | Reason |
| --- | --- |
| `spec/totipo-vault-format-v1.md` | r18 clarifications, application safeguards, scopes, and concise history. |
| `review/V1_PRE_R16_REVISION_HISTORY.md` (added) | Preserve detailed historical entries verbatim with non-normative labeling. |
| `review/V1_R18_HARDENING_REPORT.md` (added) | This review and validation evidence. |
| `README.md` | Current revision and r18 report link. |
| `CONTRIBUTING.md` | Current normative revision link label. |
| `Makefile` | Current revision in help text. |
| `conformance/README.md` | Current revision and scope/evidence limitations. |
| `conformance/cmd/generate-vectors/main.go` | Revision labels and metadata constants only. |
| `conformance/cmd/totipo-conformance/main.go` | Current revision in result messages only. |
| `conformance/internal/vectors/runner.go` | Current revision validation constants only. |
| `conformance/internal/vectors/runner_test.go` | Current baseline revision in tampering test only. |
| `conformance/internal/vectors/types.go` | Current revision in package comment only. |
| `requirements/README.md` | Current profile revision label. |
| `requirements/v1-pre-rc.json` | Regenerated revision and affected artifact hashes. |
| `tools/check_spec.py` | Current revision, linked history-location checks, and focused application-policy/scope anchors. |
| `vectors/FORMAT.md` | Current revision labels only. |
| `vectors/README.md` | Current revision label only. |
| `vectors/manifest.json` | Revision metadata only. |
| `vectors/manifest.schema.json` | Current revision constant only. |

No files were deleted. No vector case changed.

## Review notes

Final review confirmed the intended application policy strength: orphan
warnings/confirmation are SHOULD requirements, while empty-password confirmation
is MUST. The observation-relative and platform-neutral wording corrections are
resolved. The tombstone wording explicitly distinguishes logical deletion from
current graph membership and credential usability. Scope claims require their
applicable normative behavior and cannot turn passing the portable corpus into
an application-conformance claim. These application changes are intentionally
stronger even though portable semantics are unchanged.

Nix availability, independent interoperability, actual UI behavior, and live
platform durability remain outside the evidence collected here. No known portable
semantic discrepancy or failing validation remains. All changes await review in
the uncommitted worktree.

## `git diff --stat`

Git excludes the two untracked added review documents from this stat.

```text
 CONTRIBUTING.md                             |   2 +-
 Makefile                                    |   2 +-
 README.md                                   |   4 +-
 conformance/README.md                       |   7 +-
 conformance/cmd/generate-vectors/main.go    |   8 +-
 conformance/cmd/totipo-conformance/main.go  |   4 +-
 conformance/internal/vectors/runner.go      |   4 +-
 conformance/internal/vectors/runner_test.go |   2 +-
 conformance/internal/vectors/types.go       |   2 +-
 requirements/README.md                      |   2 +-
 requirements/v1-pre-rc.json                 |   8 +-
 spec/totipo-vault-format-v1.md              | 466 +++++++++-------------------
 tools/check_spec.py                         |  40 ++-
 vectors/FORMAT.md                           |   4 +-
 vectors/README.md                           |   2 +-
 vectors/manifest.json                       |   2 +-
 vectors/manifest.schema.json                |   2 +-
 17 files changed, 210 insertions(+), 351 deletions(-)
```

## `git status --short`

Final uncommitted status:

```text
 M CONTRIBUTING.md
 M Makefile
 M README.md
 M conformance/README.md
 M conformance/cmd/generate-vectors/main.go
 M conformance/cmd/totipo-conformance/main.go
 M conformance/internal/vectors/runner.go
 M conformance/internal/vectors/runner_test.go
 M conformance/internal/vectors/types.go
 M requirements/README.md
 M requirements/v1-pre-rc.json
 M spec/totipo-vault-format-v1.md
 M tools/check_spec.py
 M vectors/FORMAT.md
 M vectors/README.md
 M vectors/manifest.json
 M vectors/manifest.schema.json
?? review/V1_PRE_R16_REVISION_HISTORY.md
?? review/V1_R18_HARDENING_REPORT.md
```
