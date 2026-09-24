# Review and provenance

**Material under `review/` is provenance, methodology, historical evidence, or implementation-review material. It is not normative protocol text unless the specification explicitly incorporates it. Normative protocol text lives under [spec/](../spec/totipo-vault-format-v0.md); normative conformance vectors live under [vectors/](../vectors/v0/README.md).**

| Directory | Role |
|---|---|
| [source-vectors/](source-vectors/) | Retained historical vector inputs and original checksum files; sources for the normalized conformance corpus |
| [source-code/](source-code/) | Historical implementations and generators retained for provenance and cross-checking, not the protocol definition |
| [reports/](reports/) | Original protocol review and interoperability reports |
| [ed25519/](ed25519/REVIEW.md) | Isolated mathematical review code and derivations for the strict Ed25519 corpus |
| [phase1/](phase1/IMPLEMENTATION_REPORT.md) | Initial byte/crypto implementation checkpoint and its historical limitations |
| [phase2/](phase2/IMPLEMENTATION_REPORT.md) | Abstract semantic/recovery implementation report, fixture review, case index, and review inventory |
| [phase3/](phase3/README.md) | Linux integration report, recorded local execution results, and hashes identifying tested source |
| [process/](process/README.md) | Archived implementation instructions and the original seed review index |
| [releases/v0-rc1/](releases/v0-rc1/) | Release-candidate preparation [instructions](releases/v0-rc1/V0_RC1_AGENT_INSTRUCTIONS.md) and [requirements report](releases/v0-rc1/V0_RC1_REQUIREMENTS_REPORT.md); review evidence, not a published release |

Some historical files intentionally retain older `totp-vault-*` names for provenance. Frozen literals such as `TOTP-VAULT` and `TOTP-Vault/v0/...` are likewise unchanged. Historical reports describe their own checkpoints, not necessarily today's coverage or CI configuration; see the [project README](../README.md) for the current overview.

From the repository root, verify retained evidence with:

```sh
sha256sum -c conformance/review-inventory.sha256
sha256sum -c review/phase2/inventory.sha256
sha256sum -c review/phase3/source.sha256
```

The first inventory still protects all 17 original seed files. Its original `review/README.md` entry now points to the byte-identical [archived seed index](process/SEED_REVIEW_README.md); this current index is not a historical input. Phase 2 review hashes and Phase 3 source hashes remain separate.

Original `source-vectors/*SHA256SUMS.txt` files retain their authoring-environment paths. Some referenced inputs are absent from this repository, as recorded in the [Phase 1 report](phase1/IMPLEMENTATION_REPORT.md); do not rewrite those historical lists or mistake them for directly runnable current-layout inventories.

The hashed historical reports also retain three pre-existing links from that older layout. Current counterparts are [R31Review.java](source-code/R31Review.java), [r31-object-vectors.txt](source-vectors/r31-object-vectors.txt), and [BootstrapV0.java](source-code/BootstrapV0.java). The maintainer added the bootstrap Java source during repository cleanup; it was absent at the recorded Phase 1 checkpoint. It is not covered by the original 17-file inventory, and no original Java-source hash was available to establish identity with the historical implementation. Its received hash is recorded in the [cleanup report](../REPO_CLEANUP_REPORT.md). The original report bytes and checksum lists remain preserved.
