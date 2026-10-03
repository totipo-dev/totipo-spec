# Go v1/r18 reference/conformance consumer

This Go 1.23+ module is the repository's reference consumer, not a production client.
Run `make check`, `make race`, and `make fuzz` from the repository root. The CLI
accepts `-root /path/to/repository` and `-verify-only`. Every manifest case is
required; there is no optional capability selection. These cases provide core
protocol and modeled store-operation evidence under the specification’s Section 20
scopes. Passing them does not establish application conformance or live backend
crash durability.

Packages separate TLV framing, exact TOKEN grammar, object/root cryptography,
TOTP computation, causal-equivalence graph evaluation, and storage outcomes.
Storage bytes pass physical-size, AEAD, semantic-length, zero-padding, and keyed-ID
checks before grammar validation. Required post-AEAD negative fixtures prove that
authentication alone cannot bypass length, padding, or keyed-ID checks. Invalid grammar contributes no TOKEN state.

`graph.Store` models a supplied valid observed set, not persistent security memory.
It uses cycle-safe reachability to compute all members of maximal SCCs. Missing
parents remain unresolved. Head records and retained historical nodes preserve exact
client metadata independently of the semantic-value projection for as long as
the object is represented; this creates no persistent history obligation. Fold construction
copies one operation’s complete value and exact metadata into every stage. Values
in graph vectors are symbolic complete tuples;
real TokenValue extraction is tested against canonical objects. Defensive same-ID
failure excludes that identity from the model without poisoning unrelated tokens.
The model is deliberately small, not a scalable graph-index implementation.

`internal/storage` models exact namespaces, candidate names, observation diagnostics,
immutable publication, initial VAULT creation, and exact compare-before-replace.
Observed bootstrap entries must be regular files and object namespaces directories;
wrong types are not followed or traversed. Backend booleans describe trusted API outcomes. They do not prove file/namespace
persistence, staging behavior, host syscall races, or remote completeness. A live
adapter must implement the specification's durability and truthful reporting rules.

Tests never regenerate fixtures. For an intentional reviewed protocol update:

```sh
go run ./conformance/cmd/generate-vectors -root .
make verify
make conformance
```

The generator constructs deterministic TOKEN bytes using public test roots and
preserves the independent RFC TOTP answers. Semantic graph/workflow expectations
are specified explicitly, not queried from the evaluator. Review the diff and
pins; independent interoperability remains necessary before RC freeze.
See the [case contract](../vectors/FORMAT.md) and
[r16 rewrite report](../review/V1_R16_REWRITE_REPORT.md).
