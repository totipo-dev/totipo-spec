# Totipo

Totipo is an encrypted, append-only TOTP vault format operating on a configured
durable store. Synchronization is optional and external.

The current normative specification is **v1/r16**, a design draft with moving
pre-release-candidate conformance evidence:
[Totipo Vault Format v1](spec/totipo-vault-format-v1.md).

The design uses complete-state TOKEN assertions, fixed 1024-byte encrypted objects,
keyed content addressing, a maximum of four explicit parents, and whole-state
conflicts. Cycles form causal-equivalence groups. Optional CLIENT_NAME and CLIENT_TIME
are informational and preserved exactly while each object is represented and
across all stages of one fold, without requiring persistent remembered history. Tombstones retain the full credential. Applications describe
observed state truthfully; known complete credentials remain available for TOTP
computation regardless of lifecycle or historical status.

A store contains canonical regular-file `vault` and directory `objects-v1/`. Only exact r16 TOKENs
contribute semantic state. Unknown sibling families are outside v1 interpretation;
future families define their own compatibility relationships. The vault root
provides authoring authority. VAULT_FINGERPRINT provides optional stable recognition.
No per-client persistent graph database is required. Store loss or rollback can
lose history; the protocol does not provide deletion or rollback resistance.

| Path | Role |
| --- | --- |
| `spec/` | Current normative protocol |
| `vectors/` | Exact byte, negative, graph, fold, and storage workflow cases |
| `requirements/` | Moving pre-RC pins |
| `conformance/` | Go reference consumer and explicit fixture generator |
| `review/` | Historical decisions and review evidence |
| `tools/` | Structural and schema checks |

Use `nix develop` or the existing direnv setup when available. Go 1.23+ and Python
3.9+ are supported. Run:

```sh
make check
make race
make fuzz
go -C conformance vet ./...
```

`make conformance` executes every manifest case. `make verify` checks schemas,
case hashes, physical case coverage, and exact requirements pins. Normal checks
never regenerate fixtures. Make uses a writable Go cache under `.direnv/`.
See the [case contract](vectors/FORMAT.md) and [Go consumer](conformance/README.md).

The [r16 rewrite report](review/V1_R16_REWRITE_REPORT.md) records the baseline,
five checkpoint hashes, per-case migration, and validation. Older reports remain
historical evidence. No release candidate is frozen by this rewrite.

Next: repin/reconcile `totipo-java` against r16 and evaluate what existing
implementation architecture/code should be kept, changed, simplified, or deleted.
An independent live implementation and platform durability evidence remain
necessary before an RC freeze; this repository's model is not a production client.
