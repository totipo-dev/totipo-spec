# Proposed current `main` tree after the reset

```text
.github/
spec/
    totipo-vault-format-v1.md
vectors/
    README.md
    CASE_PLAN.md
    manifest.schema.json
    manifest.json
requirements/
    README.md
review/
    V1_R8_FINAL_CONSISTENCY_ADVERSARIAL_REVIEW.md
    V1_R8_SIMPLIFICATION_REVIEW.md
    V1_R9_FORWARD_COMPAT_REVIEW.md
tools/
    check_spec.py
conformance/          # created by the agent
README.md
CONTRIBUTING.md
SECURITY.md
LICENSE
Makefile              # rebuilt by the agent
```

Generic developer-environment files may remain if rewritten for the v1 tree.

Historical v0 files are intentionally absent from current `main`; preserve them through Git history/tags/releases and `archive/v0`.
