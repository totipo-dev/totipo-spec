# Phase 2 explicit semantic and state-machine review

Authority: `spec/totipo-vault-format-v0.md` r36. Expected results were authored from the cited sections, then checked against separate implementations. This document records the agent's explicit author review. No outside human review or independent organization is implied by `spec-derived-reviewed`.

Every newly authored normative fixture is enumerated in [CASE_INDEX.md](CASE_INDEX.md), including its provenance class and exact section locator. The maintenance sources in `tools/phase2/` contain literal expected states/traces and do not import any conformance implementation or oracle. The historical vectors were not regenerated or retagged.

## Token expectations

The primary model follows implicit field parents and joins stored derived field-head sets. The separate `referenceoracle` projects all field assertions in an explicit causal closure and removes nonmaximal assertions. For fully valid history these definitions coincide: each new F assertion incorporates every maximal earlier F revision, so typed F ancestry is precisely causal incorporation restricted to revisions asserting F. The reference code does not call the primary JOIN, ancestry, lifecycle or validation helpers.

The reference oracle reconstructs transitions by comparing a status assertion to maximal prior status assertions; scans every historical credential assertion; tests raw witnesses by causal incomparability; tests confirmation through accepted resolution ancestry; and filters maximal unresolved credential witnesses. Its validation stages and provisional candidate checks are independently implemented.

The 32 new token scenarios and 15 retained Phase 1 scenarios are consumed by both evaluators. New cases have explicit review provenance; retained cases retain their original artifact labels. Manual derivations:

| Case family | Sections | Reviewed expectation |
|---|---|---|
| Creation with a missing required assertion | 28 | INVALID; no token state is produced |
| Single and combined field edits | 23–31 | Only asserted fields get new heads; all omitted fields are inherited |
| Disjoint concurrent edits | 27 | Compose independent maximal field heads |
| Same-field alternatives and complete merge | 24,30,31 | Distinct lineages persist regardless of scalar equality; a new assertion replaces every visible alternative |
| Omitted/hidden branch, selected observation view | 22,50,56 | Historical validity remains local to signed context; observation selection does not add causality |
| Missing dependency and signer labels | 46,56,57 | Unavailable parent gives PENDING; signer equality/difference does not order token history |
| Deletion/restoration races | 32–36 | Explainable status-head/credential-witness tuples retain every witnessing transition |
| Multiple status/witness branches | 35,36 | Maximal unresolved credential revisions are selected independently per status head |
| Ordinary status conflict without credential witness | 32 | Ordinary field conflict, not a lifecycle witness |
| Transition with multiple status parents | 32 | Differing from any status parent makes a transition |
| Same-value and multi-branch lifecycle resolution | 34,38,51 | STATUS-only candidate covers incorporated credential history for its own POST; accepted resolution then supplies final coverage |
| Status descendant of a resolution | 34 | Coverage propagates through typed STATUS ancestry |
| PRE nonempty and STATUS plus any other field | 37,51 | INVALID before provisional coverage; existing witness remains |
| PRE empty combined update | 50 | Ordinary rules apply; no automatic resolution classification |
| Tombstone/credential restrictions | 41 | Any tombstone head requires LIVE in the same ordinary credential update; TOMBSTONE+credential is forbidden |
| Historical non-head witness | 33–36 | Causal incorporation by a later credential write does not erase an earlier unresolved witness |

Resolution-coverage assertions explicitly include root credential revisions, not just current witnesses. Availability and observation filters select input history before either evaluator; they do not repair dependencies or rewrite parent lists.

Three deliberately incorrect reference variants are retained and distinguished by named fixtures: current credential heads only (`historical-nonhead-witness`), coverage in the reverse direction (`stale-witness-reappears`), and coverage not inherited by status descendants (`coverage-inherited-by-status-descendant`). A separate forced nonempty candidate POST mutation proves failed provisional coverage cannot enter accepted history.

## Durable-memory expectations

The 28 memory traces use symbolic exact IDs, binding labels, token/device scopes and separately persisted records. The primary implementation uses explicit evidence maps and direct graph traversal. The independent memory reviewer uses a flat durable-fact ledger, transitive-closure relations and a fixed-point reconstruction of available histories. They share only event DTOs; neither calls the other's state transition, ancestry, writer gate or query functions.

Every event has a fixed expected outcome, and relevant checkpoints assert durable evidence and writer/observation facts. Both evaluators must match every assertion. Fields omitted from a checkpoint are unconstrained there; they are not implicitly empty. Manual derivations:

- Sections 56/56.1: pending evidence blocks only its token; observation processing awaits persistence; disappearance/restart preserve durable evidence. Exact durable abandonment permits writing without deleting evidence or claiming completeness. Warning acknowledgement also preserves the evidence. New IDs require new recovery decisions. Later fully valid abandoned history is included normally.
- Section 56.2: pathname failures never establish immutable invalidity. Only an explicitly identity-grounded proof event permits a durable terminal-invalidity handoff. That event assumes the upstream proof, rather than claiming arbitrary damaged bytes prove invalidity.
- Sections 59/63: future observations are vault-wide, durable and independent of object-file presence. Exact bypass changes writer policy only; it does not erase observations or establish completeness. Clearing a future record requires explicit compatible processing and a durable superseding handoff.
- Section 64: each remembered head must be reconstructible beneath current history. An unrelated branch or empty reconstruction is degraded; another healthy token remains writable. Pending abandonment cannot clear degradation. Restoring authenticated dependencies restores reconstructibility without forgetting remembered evidence.
- Section 63.1: replacement evidence is persisted before prior evidence is removed. Crash before persistence retains the old head; crash after persistence may leave conservative duplicates. A partial non-atomic batch persists only its selected record, so another token's pending evidence cannot be discarded. The trace would fail if one flush were treated as committing every staged record.
- Section 49.1: pending presentation evidence/abandonment is device-key scoped and never blocks token authorship. A later-valid abandoned presentation branch rejoins the presentation frontier. Presentation remembered-head replacement obeys the same durability ordering.
- Sections 39/71: confirmation binds current token context, pending recovery, desired status and an event-sensitive future epoch. Before-linearization changes force blocking/staleness; restart loses unpublished confirmations. After-linearization observations do not retract publication.
- Sections 8.1/10.1: establishment uses a serialized compare-and-set reservation; PENDING is persisted before no-replace installation. Canonical binding must match again when ESTABLISHED is persisted. A crash preserves durable pending binding, and a conflicting root cannot silently complete establishment.
- Section 65.1: 32 current heads remain encodable; 33 remain intact and block source authorship. Migration locally selects live values and materializes a fresh parentless destination token with fresh binding/device labels, without publishing a fake source resolution or claiming continuity, retirement or erasure.

Binding labels represent equality of the Section 10.1 HMAC-derived bindings, not an alternative binding algorithm. `valid` events assume upstream authenticated full validation; `terminal` proof events assume identity-grounded evidence. The state machines do not perform cryptography or model hostile OS syscalls.

## Presentation and application policy

Eleven presentation fixtures derive current names from maximal fully valid same-key DEVICE_UPDATE histories (Sections 21/49). Multiple roots and equal-valued forks remain distinct heads. Missing dependencies are PENDING; duplicate/redundant/wrong-type/cross-key dependencies are INVALID. Name value coalescing does not erase lineage. Token validity is a separate model; presentation recovery traces verify token writer independence.

Fourteen application-use fixtures implement Section 68.1 separately from consensus validity. Degraded/incomplete/lifecycle-conflicted, ambiguous/non-LIVE status, and ambiguous/incomplete credential state block ordinary current-token use/export and unqualified active presentation. Byte-equal credential alternatives can coalesce for use without changing lineage. Metadata conflicts remain surfaced but need not block use. Recovery discoverability remains true even for tombstoned conflicted tokens.

## Publication review scope

The small abstract Store interface requires exclusive/no-follow temporary creation, complete writes, file durability, candidate-bootstrap validation, serialized gate checking/install, and directory durability. Deterministic tests inject interruptions at every stage, reject preexisting/symlink/special entries and unsafe path components, preserve no-replace bootstrap installation, and permit only same-root atomic replacement. The simulator annotates bootstrap records with their already-validated binding; real unwrap is covered by the separate bootstrap codec.

These are abstract ordering and policy tests. Real OS adapters, filesystem/power-loss experiments and multi-process execution are still outside this implementation. Storage publication completion is not, by itself, user-visible token acceptance or vault establishment: durable remembered-head/establishment completion remains a separate required step modeled by the traces.

## Review outcome

Manual expectations agree with both token evaluators and both memory evaluators. Presentation/use-profile expectations agree with their independent manual derivations and consumers. No normative contradiction was found in the covered rules. No pre-existing reviewed expectation or protocol byte was changed. Fixture review does not claim exhaustiveness, production readiness, or an outside cryptographic/security audit.
