# Totipo

Totipo is an encrypted, append-only, multi-device TOTP vault format designed for synchronization over untrusted storage. This repository contains the v0 specification, language-neutral conformance vectors, a Go conformance/reference implementation, Linux integration/crash/recovery evidence, and retained review/provenance material.

Current protocol status: **Totipo Vault Format v0 release candidate, revision 37**. The normative protocol text is [spec/totipo-vault-format-v0.md](spec/totipo-vault-format-v0.md). Revision 37 updates status/evidence bookkeeping; the r36 protocol bytes and rules are unchanged. The `v0-rc1` tag and release have not yet been created.

Although the project is named Totipo, frozen protocol literals intentionally retain their historical names, including `TOTP-VAULT` and `TOTP-Vault/v0/...`. They are part of the frozen byte and cryptographic format and MUST NOT be renamed without a protocol version/change.

## Repository map

| Location | Role |
|---|---|
| [spec/](spec/totipo-vault-format-v0.md) | Normative protocol text |
| [vectors/v0/](vectors/v0/README.md) | Normative/reviewed, language-neutral conformance artifacts and manifest |
| [requirements/](requirements/README.md) | Frozen portable conformance profiles for named releases |
| [conformance/](conformance/README.md) | Executable Go reference/conformance implementation; not production code |
| [conformance/reference/](conformance/reference/README.md) | Linux storage, durable local state, and integration tests |
| [review/](review/README.md) | Historical provenance, methodology, phase reports, and execution evidence |

## Current evidence

The normative abstract corpus reports **388 PASS / 0 FAIL / 0 BLOCKED**. Separately, Linux/ext4 integration tests exercise filesystem publication, durable local state, process crash/restart, multi-process races, and hostile filesystem entries. See the [Phase 2](review/phase2/IMPLEMENTATION_REPORT.md) and [Phase 3](review/phase3/IMPLEMENTATION_REPORT.md) reports for scope and limitations.

Linux integration evidence does not establish production readiness, power-loss correctness, or cross-platform storage conformance. macOS/Windows storage behavior is not yet claimed. No external security audit is claimed.

## Release conformance

[requirements/v0-rc1.json](requirements/v0-rc1.json) pins the exact 388-case portable conformance target for `v0-rc1`. The current corpus under `vectors/v0/` may grow; the frozen profile does not. Extra future cases do not retroactively expand the release target. Linux reference integration evidence is separate from portable profile conformance.

```sh
go run ./conformance/cmd/totipo-conformance --requirements requirements/v0-rc1.json
# Or: make conformance-v0-rc1
```

## Run the checks

Use **Go 1.23 or later** and **GNU Make**; CI tests the latest 1.23 patch and current stable Go. Linux integration tests require the documented local filesystem facilities; race tests also require a supported C compiler. From the repository root:

```sh
make check        # Run all checks below

make test         # Full suite, including Linux integration/crash tests
make race         # Full suite with the race detector
make conformance  # Expected: PASS 388 / FAIL 0 / BLOCKED 0
make conformance-v0-rc1  # Frozen release profile: 388 / 0 / 0
make verify-requirements # Profile integrity, without executing cases
make verify       # Vector and review checksums
```

The Makefile creates `.phase3-test-tmp/` and sets the repository-local `TMPDIR` to keep Linux integration tests on the repository's filesystem. The Nix development environment includes GNU Make; reload it after updating the flake. For the underlying commands, portable-only checks, or fuzzing, see [conformance/README.md](conformance/README.md) and [conformance/FUZZING.md](conformance/FUZZING.md).

Licensed under the [Apache License 2.0](LICENSE). See [CONTRIBUTING.md](CONTRIBUTING.md) and [SECURITY.md](SECURITY.md) for contribution and security reporting guidance.
