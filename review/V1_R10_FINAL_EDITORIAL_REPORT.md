# V1 r10 final editorial report

## Change

Baseline: `86169f943752fa2b696300acf96d4b762acfc808` (`r10 updates`), on `main`.
The tracked working tree was clean; baseline checks passed all 77 cases.

Added this sentence in §12.6 immediately before the existing no-compatibility-promise statement:

> Rolling-upgrade interoperability with v1 is cooperative, not enforceable by v1 alone: an older v1 client can observe future-family state only to the extent that the future-family writer publishes the required authenticated v1-family compatibility projection.

Only `spec_sha256` changed in the moving profile. The structural checker needed
no change, and vector generation was not run.

Documentation accuracy fixes: the r9 implementation report now places its tree
and 68/71-case milestone descriptions in historical context, preserving its
historical tables. The r10 integration report received an append-only acceptance
update identifying the accepted commit and verified CI run; its original
execution-time statements remain intact.

Exact changed files:

```text
spec/totipo-vault-format-v1.md
requirements/v1-pre-rc.json
review/V1_VECTOR_IMPLEMENTATION_REPORT.md
review/V1_R10_INTEGRATION_REPORT.md
review/V1_R10_FINAL_EDITORIAL_REPORT.md (new)
```

## Protocol impact

- Revision remains **r10**; no new revision-history entry.
- Wire bytes changed: **NO**.
- Semantics changed: **NO**.
- Storage namespace changed: **NO**.
- Vector cases changed: **NO**.
- Compatibility mechanism changed: **NO**.
- Unknown sibling namespaces still provide no authenticated signal or warning.
- No algorithm, parser, encoder, Nix, CI, or Makefile changes.
- Ambiguity or blocker found: **none**.

## Evidence

Spec SHA-256:

```text
old: e3bb61f079e7b436c609b5fbba47db4116b2a01e3c16f88bf8d33408bd96e485
new: 056303e547b33912967f612e07dd5e1a3f4fd3bc599599348504f60c1b9943a2
```

An external temporary baseline captured spec, manifest, profile, and all case-file
hashes before editing. Comparison confirms **0 of 77 case hashes changed**, zero
embedded semantic/object byte changes, and an unchanged manifest. Profile comparison
confirms that only the spec checksum changed; revision, IDs, paths, expectations,
and case hashes are unchanged.

PASS: structural spec check, Go tests, Go vet, `make check`, conformance **77/77**,
`make verify`, and `git diff --check`. Optional race/fuzz runs were not repeated
for this prose-only change. Nix is unavailable in this environment and was not modified.

The [accepted r10 integration CI run](https://github.com/totipo-dev/totipo-spec/actions/runs/36207631688)
was verified through the GitHub Actions API as completed/success for baseline
commit `86169f943752fa2b696300acf96d4b762acfc808`. This editorial change has not
been pushed or run remotely. Changes remain ready for the user's manual commit;
the local instruction file is excluded from staging/commits.

## Remaining work

r10 is editorially complete. Stop spec-repository work and start the first live
Totipo implementation as the independent vector consumer. No release, r11, frozen
RC profile, historical-ref change, or push was made in this pass.
