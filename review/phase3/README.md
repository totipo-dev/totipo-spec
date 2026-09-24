# Phase 3 local execution evidence

`local-results.json` records the successful Linux/ext4 test run on 2026-09-24.
It was extracted from `go test -json`; the helper-only test and parent test groups
are excluded from leaf counts. Crash cases are real subprocess exits followed by
fresh adapter/client reconstruction, not in-memory error simulations.

`source.sha256` identifies the Go integration source tested. Verify from the
repository root with `sha256sum -c review/phase3/source.sha256`.

These are non-normative implementation observations, not protocol expectations,
external review, or evidence of a remote CI run. The historical input bytes,
Phase 2 review inventory, and normative vectors remain unchanged; the historical
inventory records the seed index's move into `review/process/`. See the
[Phase 3 implementation report](IMPLEMENTATION_REPORT.md) for commands, boundary
assumptions, and limits.
