# v1/r17 hardening report

Starting worktree: clean (`git status --short` produced no output). Baseline HEAD:
`54f38d13d62dc811c98489eb1867d63a9bef235a`, commit subject `Update to r16`.
`git describe --tags --always` returned `v0-rc1-18-g54f38d1`;
`git tag --points-at HEAD` produced no output. Thus the committed baseline is r16,
not an r16 release tag. The specification and profile remain moving pre-RC design
draft evidence. HEAD and tags have not changed. No commit, push, tag, or release
was performed, and `totipo-java` was not touched.

External review found that “hostile” observed bytes could be read as a promise to
detect malicious withholding, deletion, or rollback of valid history. Before r17,
§2 required validation and denied cryptographic rollback/deletion resistance, but
did not immediately distinguish those two properties. After r17, §§1–2 explicitly
separate authenticity/integrity of observed representations from completeness and
freshness of the configured-store view. An old valid representation can still
authenticate; v1 cannot generally establish unavailable newer history without
independent evidence. This is not a claim that rollback is harmless.

Exact normative text changes:

| Location | Change |
| --- | --- |
| Preamble | Identify revision 17 and unchanged r16 behavior/wire bytes. |
| §1 Goals and authority | One early sentence separating observed validity from complete/freshest history. |
| §2 Threat model and storage boundary | Untrusted entries/bytes must earn meaning through validation; explicit store-view limitation, old valid representations, identity versus set authentication, preserved publication obligations, and informative deployment seam. |
| §14 Observation and diagnostics | Make safe disappearance/older-view interpretation explicit: no crash or fabricated facts, truthful unresolved references, ordinary recomputation, no cross-run memory requirement. |
| §21 Revision history | Record threat-model clarification, deployment-layer guarantees, and no format/crypto/graph/publication/semantic or case-outcome change. |

Content validation remains hostile/untrusted: all applicable path/name/type,
bounded-read, size, AEAD, padding/length, keyed OBJECT_ID, exact TOKEN grammar,
semantic-bound, and VAULT authentication checks remain required. Stale valid
material does not become forged material, and the freshness limitation does not
permit accepting malformed/tampered observations.

The informative §2 deployment note explicitly permits storage or additional
application mechanisms to supply freshness, rollback detection, retained history,
auditability, and stronger deletion resistance. Deployments must choose such
properties explicitly, according to that layer's trust and retention model.
Version history alone is not cryptographic rollback protection. Synchronization
remains optional; a configured local store is still supported. Successful local
publication retains the full existing contract, without promising later complete
history. No new journal, monotonic counter, rollback database, or history capability
is introduced.

The full terminology audit and all eight cold-reader review questions are recorded
in [the focused review](V1_R17_STORE_FRESHNESS_CLARIFICATION.md). No surviving current
normative sentence promises cryptographic detection of withheld newer history.
§8's immediate exact-byte comparison with BASE remains distinct from historical
freshness; §9 still explicitly denies fingerprint freshness/rollback protection.
README.md, CONTRIBUTING.md, requirements/README.md, conformance/README.md,
vectors/README.md, and vectors/FORMAT.md were checked. Current revision labels
were updated, README explains the boundary, and historical r16 report links remain.

The corpus impact decision, made before vector edits, was **NO semantic expectation
changes**. All 90 cases were inspected, including their nested graph/storage/workflow
expectations. The r16→r17 byte comparison then checked every physical case against
`git show 54f38d13d62dc811c98489eb1867d63a9bef235a:<path>` and its manifest SHA-256.
All 90 files are byte-identical; zero additions, deletions, ID/path changes, verdict
changes, nested expectation changes, or embedded object/plaintext/bootstrap changes.
The complete manifest `cases` array and profile `required_cases` array are identical.
Every case remains required, with no new capability or skip/deferred state.

| Operation | Cases | r16→r17 result |
| --- | ---: | --- |
| bootstrap | 4 | Byte-identical |
| post-aead | 3 | Byte-identical |
| crypto | 17 | Byte-identical |
| dispatch | 23 | Byte-identical |
| fold | 6 | Byte-identical |
| graph | 13 | Byte-identical |
| workflow | 14 | Byte-identical |
| storage | 7 | Byte-identical |
| totp | 3 | Byte-identical |
| Total | 90 | Identical IDs, bytes, and expectations |

In particular, `graph.intermediate-disappears` still yields heads A/C, conflict,
and unresolved B; `graph.reappearance` keeps B current, reports unresolved A while
absent, and resolves A on return. Wrong size/failed AEAD remain INVALID_STORAGE;
authenticated invalid grammar remains INVALID. Incomplete observations retain
valid objects and diagnostics. Publication with missing parents still succeeds
when durable; exact-existing publication still needs no new persistence barrier;
non-exact/unreadable targets remain failures. Ambiguous TOKEN publication, VAULT
creation, and replacement remain FAILED. VAULT exact BASE match/replacement,
STALE mismatch, orphan-tolerant creation, and unreadable failure are unchanged.

The SHA-256 of concatenated manifest-order `ID + " " + case SHA256 + "\n"`
records is identical for r16 and r17:
`da4646a7aaec040ca24f6ed2f922a7e9f51f2a3ce77ef0e8a56f37e44ba68316`.
A reproducible case audit from the repository root is:

```sh
python3 - <<'PY'
import hashlib, json, subprocess
from pathlib import Path
base = '54f38d13d62dc811c98489eb1867d63a9bef235a'
def old(p):
    return subprocess.check_output(['git', 'show', f'{base}:{p}'])
m = json.loads(Path('vectors/manifest.json').read_bytes())
assert m['cases'] == json.loads(old('vectors/manifest.json'))['cases']
assert len(m['cases']) == 90
assert {str(p) for p in Path('vectors/cases').rglob('*.json')} == {
    'vectors/' + e['path'] for e in m['cases']}
for e in m['cases']:
    p = 'vectors/' + e['path']
    b = Path(p).read_bytes()
    assert b == old(p)
    assert hashlib.sha256(b).hexdigest() == e['sha256']
print('PASS: 90 byte-identical cases and identical manifest case entries')
PY
```

Byte/semantic invariance evidence:

| Contract | Evidence |
| --- | --- |
| TOKEN bytes, tags, widths, field limits, MAX_PARENTS=4 | §12, object/TLV implementation, and all exact/negative fixtures unchanged. |
| VAULT bytes and authentication | §§4–6, crypto implementation, and all four bootstrap fixtures unchanged. |
| HKDF, keyed OBJECT_ID, AES-GCM, Argon2, VAULT_FINGERPRINT | §§4–6 and §§9–10, crypto code, and intermediate expected values unchanged. |
| Fixed 1024-byte envelope | §11 and all physical/encryption bytes unchanged. |
| Graph/SCC/parents and same-ID defense | §§13, 15–16, graph code, and all 13 graph cases unchanged; §14 only clarifies existing observation handling. |
| Fold and metadata | §§12, 16–17, code, and all six fold cases unchanged. |
| Storage publication and VAULT operations | §§7–8 and §18, storage code, and all 14 workflow cases unchanged. |
| Discovery/cache/history behavior | Existing §2 cache text and §§3, 12–13, 15–17 unchanged; no new retained-state requirement. |
| TOTP | §19, implementation, and three RFC fixtures with 18 answers unchanged. |
| All conformance outcomes | All 90 case files, including every nested expected result, byte-identical. |

The section comparison proves §§3–13 and §§15–20 are byte-identical, as is all
pre-existing revision history. Every Go source is either byte-identical or differs
only by literal `r16`→`r17` substitutions. These substitutions update generator,
consumer output, revision pins, the revision-tampering test input, and a package
comment. No parser, graph, crypto, vault, storage, or vector-runner algorithm
changed. The structural checker only changes the current revision and history
range; no validation assertion was weakened.

The manifest changes only `spec_revision`. The profile changes only
`spec_revision`, `spec_sha256`, `manifest_sha256`, and `schema_sha256`.
The manifest schema changes exactly one revision-metadata constant:
`properties.spec_revision.const`, from `r16` to `r17`. This is structurally
required to accept the r17 manifest; there is no semantic schema change.
The case schema is byte-identical, including its inherited r16 title.
Its unchanged SHA-256 is
`d38618f53dcf0a558c389831e8838a07248066beff392812f3d277cc974e25c9`.

The first post-edit suite exposed the missed manifest-schema constant:
`make check` and `make verify` failed with `$.spec_revision: const`; the other
requested commands passed. The correction updated only that metadata constant
and regenerated its profile hash. The complete suite was then rerun. No failing
baseline was modified: every baseline command passed before any repository edit.

Exact validation commands and results (exit 0 is PASS):

| Command | Clean r16 baseline | Final r17 |
| --- | --- | --- |
| `python3 tools/check_spec.py` | PASS | PASS |
| `go -C conformance test ./...` | PASS | PASS |
| `go -C conformance vet ./...` | PASS | PASS |
| `make check` | PASS | PASS |
| `make conformance` | PASS | PASS |
| `make verify` | PASS | PASS |
| `make race` | PASS | PASS |
| `make fuzz` | PASS | PASS |
| `git diff --check` | PASS | PASS |
| `test -z "$(gofmt -l conformance)"` | PASS | PASS |

The formatting command was also run separately before editing. `make check`
includes all 10 Python tests and uncached Go tests. Both baseline and final
`make conformance` report **90/90**; `make verify` validates both schemas, physical
coverage, hashes, and moving pins. `make fuzz` runs FuzzDispatch, FuzzOpen, and
FuzzArrivalAndDisappearance for 10 seconds each with parallelism 2; all passed.

Nix: `command -v nix` returned exit 1 with no output. The `nix` executable is not
available on PATH, so Nix verification was not run. The repository documents
`nix develop` and the same Make/Go checks; flake.nix defines a development shell
and formatter, not a separate flake checks suite. No Nix files were changed.

Reproducibility: ran `go run ./conformance/cmd/generate-vectors -root .` twice
before the schema-pin correction and twice after it. Each repeated run preserved
every byte of all vector files and the moving profile, and preserved
`git status --short` exactly. The final generator reports 90 cases; the unchanged
manifest case entries establish ID/order/expectation stability. No untracked
generated files or fixture drift remain. Only the two requested review documents
are untracked. Normal validation does not regenerate cases.

Exact metadata SHA-256 audit:

| Artifact | r16 | r17 |
| --- | --- | --- |
| `spec/totipo-vault-format-v1.md` | `96f362d68251a801b2c433bd6435c0056d5b070bfe1258a12ae12bdf7bb7990e` | `f8d2ab02c97e8ac54847048a06cb088359db982fec3f176fd473ce0f223d43cf` |
| `vectors/manifest.json` | `d03821fa2a0c41145bd8b82c6ea46800ab1a847b50c92fb9996b3c585d6f0a33` | `94fff22842573e15b254bb0653970761c15cf122192947a36fccc48344a84f08` |
| `vectors/manifest.schema.json` | `3b54847e7a21bacc7956406c0d79dc172a0fb6415e9eb9c0335a511c1987664a` | `e7a5d8ec0392e0248ab867375b7c907a7ca1298595acdb14ecd3edcbe66df476` |
| `requirements/v1-pre-rc.json` | `687faf7757de7c63da68c28889bf3d0976b1858f9647b67569e3d8dfafc59c56` | `ec65793e4734086cb79ad1bfbe96df4030c743b4b00d56bcdfa8cc90bfcbcb9e` |

Remaining issues: no known blocker to releasing r17 as a clarification-only
revision. Nix validation remains unavailable in this environment. Existing pre-RC
needs for independent interoperability and live platform durability evidence remain;
these checks do not establish production backend durability. No release action
has been taken.

**The claim holds:** r17 is a threat-model clarification only; it changes no v1
wire format, cryptography, graph semantics, publication behavior, or
conformance-case outcome.

Final uncommitted `git status --short`:

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
?? review/V1_R17_HARDENING_REPORT.md
?? review/V1_R17_STORE_FRESHNESS_CLARIFICATION.md
```

Final `git diff --stat` (Git excludes the two untracked review documents):

```text
 CONTRIBUTING.md                             |  2 +-
 Makefile                                    |  2 +-
 README.md                                   | 13 +++--
 conformance/README.md                       |  2 +-
 conformance/cmd/generate-vectors/main.go    |  8 +--
 conformance/cmd/totipo-conformance/main.go  |  4 +-
 conformance/internal/vectors/runner.go      |  4 +-
 conformance/internal/vectors/runner_test.go |  2 +-
 conformance/internal/vectors/types.go       |  2 +-
 requirements/README.md                      |  2 +-
 requirements/v1-pre-rc.json                 |  8 +--
 spec/totipo-vault-format-v1.md              | 77 +++++++++++++++++++++++++----
 tools/check_spec.py                         |  8 +--
 vectors/FORMAT.md                           |  4 +-
 vectors/README.md                           |  2 +-
 vectors/manifest.json                       |  2 +-
 vectors/manifest.schema.json                |  2 +-
 17 files changed, 103 insertions(+), 41 deletions(-)
```
