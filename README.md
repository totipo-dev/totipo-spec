# Totipo

Totipo is an encrypted, append-only, multi-device TOTP vault format designed for synchronization over untrusted storage. This repository contains the v0 specification, language-neutral conformance vectors, a Go conformance/reference implementation, Linux integration/crash/recovery evidence, and retained review/provenance material.

The v0 wire format is **frozen candidate/release-candidate material**. The normative protocol text is [spec/totipo-vault-format-v0.md](spec/totipo-vault-format-v0.md) (r36).

Although the project is named Totipo, frozen protocol literals intentionally retain their historical names, including `TOTP-VAULT` and `TOTP-Vault/v0/...`. They are part of the frozen byte and cryptographic format and MUST NOT be renamed without a protocol version/change.

## Repository map

| Location | Role |
|---|---|
| [spec/](spec/totipo-vault-format-v0.md) | Normative protocol text |
| [vectors/v0/](vectors/v0/README.md) | Normative/reviewed, language-neutral conformance artifacts and manifest |
| [conformance/](conformance/README.md) | Executable Go reference/conformance implementation; not production code |
| [conformance/reference/](conformance/reference/README.md) | Linux storage, durable local state, and integration tests |
| [review/](review/README.md) | Historical provenance, methodology, phase reports, and execution evidence |

## Current evidence

The normative abstract corpus reports **388 PASS / 0 FAIL / 0 BLOCKED**. Separately, Linux/ext4 integration tests exercise filesystem publication, durable local state, process crash/restart, multi-process races, and hostile filesystem entries. See the [Phase 2](review/phase2/IMPLEMENTATION_REPORT.md) and [Phase 3](review/phase3/IMPLEMENTATION_REPORT.md) reports for scope and limitations.

Linux integration evidence does not establish production readiness, power-loss correctness, or cross-platform storage conformance. macOS/Windows storage behavior is not yet claimed. No external security audit is claimed.

## Run the checks

Use **Go 1.23 or later**; CI tests the latest 1.23 patch and current stable Go. Linux integration tests require the documented local filesystem facilities; race tests also require a supported C compiler. From the repository root:

```sh
mkdir -m 700 -p .phase3-test-tmp
TMPDIR="$PWD/.phase3-test-tmp" go -C conformance test -count=1 -timeout=180s ./...
TMPDIR="$PWD/.phase3-test-tmp" go -C conformance test -race -count=1 -timeout=300s ./...
go run ./conformance/cmd/totipo-conformance ./vectors/v0
(cd vectors/v0 && sha256sum -c manifest.sha256)
```

The repository-local `TMPDIR` keeps Linux integration tests on the repository's filesystem. For portable-only commands, review inventories, or fuzzing, see [conformance/README.md](conformance/README.md) and [conformance/FUZZING.md](conformance/FUZZING.md).

See [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md), and the unresolved [license selection](LICENSE-TODO.md) before contributing, reporting security issues, or reusing the material.
