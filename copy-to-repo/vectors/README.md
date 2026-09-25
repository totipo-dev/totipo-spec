# v1 Conformance Vectors

This directory is for the **current Totipo v1 protocol only**.

Historical v0 vectors are intentionally not copied into the current `main` working tree. They remain available through historical Git references.

## Principles

1. Case IDs are stable once consumed by an implementation.
2. Byte vectors and semantic/state vectors are distinct.
3. The manifest lists only vector cases that actually exist.
4. Exact byte vectors are normative where marked `normative: true`.
5. Semantic/state cases should be language-neutral.
6. Vector generation must be deterministic except where a fixture deliberately supplies a fixed externally-generated signature.
7. Writer-capacity cases reserve 72 DER-signature bytes even when the actual fixture signature is shorter.

## Suggested structure

```text
vectors/
    manifest.json
    manifest.schema.json
    routing/
    encoding/
    bootstrap/
    crypto/
    provenance/
    token/
    device/
    graph/
    candidate/
    timestamp/
    negative/
```

Start with routing/encoding because r9's frozen routing prefix is the compatibility contract future semantic versions must preserve.

Do not create a frozen release requirements profile until the vector set has stabilized.
