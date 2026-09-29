# Requirements profiles

[`v1-pre-rc.json`](v1-pre-rc.json) is the **moving pre-RC** profile for v1/r16.
It pins the exact specification, entire manifest, manifest schema, case schema,
and every manifest case ID/hash. There are no conditional capabilities.

`make verify` checks schemas, exact physical case coverage, hashes, and profile
membership. `make conformance` executes every case. Neither regenerates fixtures.
Explicit maintenance is described in [FORMAT.md](../vectors/FORMAT.md).

Do not freeze an RC profile until independent implementation interoperability and
required security/platform evidence are complete. The r16 rewrite does not tag,
release, or freeze anything.
