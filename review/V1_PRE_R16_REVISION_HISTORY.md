# Historical v1 revision notes before r16

This document preserves the r1–r15 revision entries formerly included in the
[specification](../spec/totipo-vault-format-v1.md#21-revision-history). It is
non-normative historical material, not an alternative current grammar. Removed
DEVICE, signature, provenance, opaque-state, and persistence designs below do not
apply to current v1. The entries are preserved verbatim.

### v1/r15

This revision intentionally changes state semantics. Synchronized storage is explicitly unreliable and nontransactional. Writers use accepted local snapshots; later hidden history is normal concurrency. Optional local history becomes advisory regression evidence, with cache loss/corruption discarded rather than blocking operations. Durable graph transactions, history-persistence gates, global discovery/unscoped gates, mandatory opaque retention, and unavailable remembered-state confirmation are removed. Conflict confirmation tracks TOKEN-specific semantic decisions, not unrelated vault events. Ambiguous publication is normal uncertainty with permitted retries. Synchronized-object fsync becomes reliability guidance; exact-existing publication needs exact bytes. First DEVICE success uses publication acknowledgement. Local binding establishment no longer uses pending records; VAULT, binding, and private custody retain strong durability. Runtime writer self-roundtrips are optional hardening.

No wire-format, crypto, routing-prefix, envelope-family, object-size, capacity-formula, TOTP algorithm, or identifier construction changes.

### v1/r14

Fourteenth v1 design draft.

Local-filesystem threat-boundary simplification from r13:

- clarifies that the synchronization medium remains fully untrusted for contents, history, freshness, ordering, completeness, and availability;
- makes the local OS/filesystem execution environment part of the trusted computing base for baseline v1 conformance;
- removes the baseline requirement to prove immunity to an actively malicious same-privilege process racing pathname, file-type, directory-identity, or inode-content changes between individual filesystem operations;
- retains direct-family namespace confinement, exact filename rules, bounded reads, observed symlink/special-file exclusion, conservative handling of synchronization churn, and full cryptographic validation of synchronized bytes;
- retains immutable no-overwrite object publication, crash-safe VAULT replacement, durability acknowledgement, and publication-before-durable-graph ordering;
- classifies ordinary synchronization races as availability/freshness conditions causing retry/rescan/incomplete processing rather than invented authenticated state;
- permits stronger hostile-local-filesystem hardening as an implementation-specific defense rather than an interoperability requirement;
- does not change TOKEN/DEVICE bytes, cryptographic domains/construction, routing prefixes, `objects-v1/` layout, object size, capacity formulas, graph/state semantics, provenance semantics, or TOTP algorithms.

### v1/r13

Thirteenth v1 design draft.

Durable-recovery/provenance-completeness changes from r12:

- defines the minimum `OPAQUE_UNSCOPED_RECORD` and requires retention of the exact authenticated 1024-byte object so a later compatible implementation can reprocess it after synchronized deletion;
- makes failure to durably retain that exact opaque-unscoped object a `KNOWLEDGE_PERSISTENCE_BLOCKED` condition;
- explicitly warns that continuity reset/re-baseline may cause assertions previously known to be historical to re-enter the derived current/conflict frontier when prior ancestry knowledge is discarded;
- requires provenance status to be recomputed for affected known TOKEN assertions when matching DEVICE/public-key material later becomes available;
- aligns first-DEVICE conformance wording with the normative report-success gate: TOKEN bytes may be published first, but first TOKEN success is not reported until the matching DEVICE advertisement is durable;
- defines `resource-complete` discovery/baseline scans;
- clarifies that the four-parent fold in Section 47 is illustrative and state-size-dependent;
- clarifies DEVICE presentation wording and cleans section-separator/editorial artifacts;
- no TOKEN/DEVICE wire-format, crypto, routing-prefix, storage-family, capacity-formula, or TOTP algorithm changes.

### v1/r12

Twelfth v1 design draft.

State-machine/provenance hardening changes from r11:

- defines when `OPAQUE_UNSCOPED` evidence is active and makes synchronized disappearance explicitly non-clearing;
- adds an explicit user-directed continuity reset/re-baseline path that can abandon prior-epoch opaque-unscoped evidence only with acknowledgement of lost continuity guarantees;
- clarifies that present opaque-unscoped evidence is rediscovered and remains blocking after re-baseline;
- makes supported provenance-UNRESOLVED/REJECTED DEVICE heads presentation-inert but still part of DEVICE causality;
- requires DEVICE rename/convergence to incorporate every current supported DEVICE head regardless of provenance status;
- distinguishes hostile synchronized-byte corruption/unavailability from corruption/inconsistency of local durable security memory;
- requires cryptographically secure native/platform ECDSA nonce generation and forbids ad-hoc Totipo nonce generation;
- requires durable publication of the local DEVICE advertisement by the time first TOKEN publication for that device/vault is reported successful;
- documents that `K_signature_context` vault-binds otherwise vault-independent device provenance;
- clarifies empty-password creation UX, TOTP secret-generation wording, direct-HMAC `VAULT_BINDING` rationale, and the distinction between `OBJECT_VERSION` routability and official allocation;
- fixes stale revision/open-work wording and updates conformance requirements;
- no TOKEN/DEVICE wire-format, object-crypto, envelope-family, routing-prefix, size/capacity, or TOTP algorithm changes.

### v1/r11

Eleventh v1 design draft.

Governance, compatibility-clarification, and historical-cleanup changes from r10:

- clarifies that rolling-upgrade interoperability with future envelope families is cooperative and depends on the future writer publishing the required v1-family compatibility projection;
- assigns `OBJECT_VERSION = 0x01` to the current semantic grammar, leaves every other value unassigned, requires published-spec allocation for interoperable use, and defines no private-use range;
- records v0 as undeployed historical design work rather than a supported predecessor;
- removes the normative v0 migration procedure and renumbers subsequent sections;
- cleans stale/completed pre-RC work wording;
- no TOKEN/DEVICE semantic encoding, routing-prefix, storage-family, cryptographic, vector-expectation, or runtime-semantic changes.

### v1/r10

Tenth v1 design draft.

Envelope-family namespace changes from r9:

- renames the current object namespace from flat `objects/` to exact top-level `objects-v1/`;
- defines `objects-v1/` as the Totipo v1 envelope/storage-family namespace;
- clarifies that `OBJECT_VERSION` versions semantics inside this family rather than selecting the storage family;
- freezes 1024-byte object size as a v1-family property rather than a promise that every future Totipo family uses 1024-byte files;
- unknown sibling namespaces such as `objects-v2/` are outside v1 discovery and have no authenticated semantic meaning by themselves;
- differently sized files inside `objects-v1/` are invalid current-family storage evidence and never become opaque future-version evidence;
- `OPAQUE_ROUTABLE` / `OPAQUE_UNSCOPED` apply only after successful authentication of a valid v1-family envelope;
- a future envelope/storage family may use a separate sibling namespace and arbitrary future layout defined by that future specification;
- a future family claiming rolling-upgrade interoperability with v1 must publish authenticated v1-family compatibility assertions into `objects-v1/`;
- old v1 clients observe future-family state only through those compatibility assertions;
- future-family namespace names cannot be used by an untrusted sync provider to manufacture warnings/blocks;
- privacy/filesystem/conformance wording updated around family namespace isolation;
- no change to TOKEN/DEVICE semantic bytes, crypto construction, routing prefix, or 1024-byte v1-family object bytes beyond the storage path change.

### v1/r9

Ninth v1 design draft.

Forward-compatible rolling-upgrade changes from r8:

- freezes common TOKEN/DEVICE routing metadata across semantic versions in the v1 envelope family;
- TOKEN routing includes TOKEN_ID and AUTHOR_DEVICE_ID;
- DEVICE_ID is explicitly serialized as frozen DEVICE routing identity;
- v1 DEVICE validates serialized DEVICE_ID against the existing public-key derivation;
- defines supported-valid, opaque-routable, opaque-unscoped, and invalid compatibility classes;
- future TOKEN/DEVICE objects with valid frozen routing are retained as opaque durable graph nodes;
- opaque nodes participate in ancestry/current-head selection without parsing their body;
- opaque current TOKEN degrades only that token and keeps supported candidate use available;
- unrelated tokens continue normal operation;
- v1 authorship is blocked only for a token whose current frontier includes opaque future TOKEN semantics;
- opaque future DEVICE affects presentation/provenance scope, not TOKEN authority;
- opaque-unscoped authenticated future evidence blocks authoritative operations but not candidate use;
- later supported/readable descendants can causally supersede opaque future nodes;
- semantic version number never establishes ordering;
- common readiness predicates centralize authoritative vs candidate-use gates;
- maximum-display DEVICE planned fan-in drops from 15 to 14 parents because explicit DEVICE_ID adds 36 semantic bytes;
- degradation philosophy is explicit: uncertainty reduces assurance/capability rather than automatically making authenticated candidate material unusable.

### v1/r8

Eighth v1 design draft.

Pre-vector adversarial hardening from r7:

- resolved-edge cycles are graph-integrity failures and cannot participate in current-head suppression;
- one `OBJECT_ID` is globally unique across all semantic object types and immutable graph records;
- inconsistent TOKEN/DEVICE reuse of one ID forces local continuity recovery;
- writer capacity planning reserves 72 bytes for DER ECDSA signatures;
- writers may not exploit/retry shorter signatures to squeeze in extra parents;
- maximum-size TOKEN plans 4 parents; maximum-size DEVICE plans 15 parents;
- TOKEN and DEVICE wide-frontier fold progress is explicitly guaranteed;
- ordinary current use and semantic authorship remain `DISCOVERY_STATE=READY` only;
- explicit candidate use may proceed under `PROCESSING_INCOMPLETE` from an already authenticated LIVE candidate, with mandatory warning, when no stronger gate is active;
- unreadable candidate observations remain unprocessed rather than being treated as safely classified;
- unsupported-version blocking explicitly includes candidate use;
- credential-generation freshness snapshots are explicit for ordinary and candidate use;
- `AUTHOR_TIME` must be preserved as unsigned/raw u64 and platform date-conversion overflow cannot affect validity;
- extreme `AUTHOR_TIME` values remain structurally valid and non-causal;
- privacy/overview wording updated;
- at least two independent implementations are required before v1-rc1 freeze.

### v1/r7

Seventh v1 design draft.

Pre-vector hardening changes from r6:

- defines `DISCOVERY_STATE = READY | PROCESSING_INCOMPLETE`;
- visible/accepted-but-unprocessed object candidates block ordinary TOTP use, explicit candidate credential use, and semantic authorship until processed;
- broadens degraded fallback into explicit candidate credential use;
- user may explicitly generate from any available vault-authenticated LIVE TOKEN candidate when ordinary use is blocked by whole-state conflict or unavailable current values;
- candidate may be a current conflict alternative, an available current head while another is missing, or historical recovery material;
- candidate use never changes current-state semantics and cannot bypass unsupported-version, continuity, discovery, or knowledge-persistence gates;
- parent current-path/intrinsic-invalid observations leave unknown parent identities unresolved rather than creating rejected edges;
- durable DEVICE graph records retain `PUBLIC_KEY_X963`;
- known-object reappearance must match immutable durable graph identity/topology/public-key/timestamp records;
- detected durable-graph corruption enters local continuity recovery rather than silently dropping records;
- multi-object folds explicitly are not synchronized transactions;
- historical provenance UI caveat restored: current device friendly name is not historical friendly-name metadata;
- adds required common `AUTHOR_TIME` `u64be` field to TOKEN and DEVICE;
- `AUTHOR_TIME=0` means unknown; positive values are reported Unix seconds;
- author time is authenticated/signed informational metadata only and never affects causality, freshness, head selection, or conflict resolution;
- all objects in one multi-object fold use one common captured AUTHOR_TIME;
- TOKEN maximum-size formula becomes 215 + issuer + account + secret + 36×parents; maximum fields still fit four parents in the 1024-byte envelope;
- maximum-size DEVICE now fits 15 parents with AUTHOR_TIME.

### v1/r6

Sixth v1 design draft.

Major simplification from r5:

- removes durable accepted-head and superseded-ID summaries;
- introduces append-only durable authenticated TOKEN/DEVICE topology as primary local continuity/security memory;
- once a valid node and parent claims are durably learned, synchronized file deletion cannot erase that topology;
- current TOKEN heads are maximal nodes in the durable known graph;
- complete TOKEN value availability is tracked separately from graph/topology knowledge;
- a known current head may remain current while its value bytes are unavailable;
- older available TOKENs may be explicitly used in degraded stale mode with strong warning, without changing current-state semantics;
- reaffirmation is an ordinary complete TOKEN parenting all durable known current heads, including unavailable current heads;
- a remote TOKEN parenting a locally known unavailable current head advances normally when learned;
- removes `REFERENCE_COVERS_FOR_REAFFIRMATION`, reaffirmation records, `SUPERSEDED_ID_SET`, `ACCEPTED_HEAD_IDS`, and `ACCEPTANCE_PENDING`;
- bounded folds rely on durable intermediate graph nodes instead of local supersession summaries;
- detected graph/security-memory rollback enters `LOCAL_CONTINUITY_UNKNOWN`; baseline re-establishment rebuilds a fresh durable graph from a resource-complete current scan;
- first-open and pending-VAULT establishment procedures are explicit;
- introduces explicit degraded stale credential use for an older available LIVE TOKEN when a newer/current TOKEN value is unavailable.

### v1/r5

Fifth v1 design draft.

Consistency-hardening changes from r4:

- formally distinguishes `OBSERVED_TOKEN_HEADS`, `CURRENT_TOKEN_HEADS`, and durably `ACCEPTED_TOKEN_HEADS`;
- introduces `ACCEPTANCE_PENDING` and blocks unsafe authorship/credential consumption while a safety-relevant frontier change is not durably accepted;
- defines active staged multi-object TOKEN folds so an operation's own intermediate publications do not stale their confirmation;
- external safety-relevant observations during a fold stale confirmation and force recomputation;
- replaces reaffirmation-specific local historical state with one per-token `SUPERSEDED_ID_SET`;
- every successful multi-object TOKEN fold adds every original required input ID and every intermediate fold TOKEN except FINAL to `SUPERSEDED_ID_SET`;
- intermediate fold TOKENs such as `M1` are explicitly superseded after successful finalization;
- one-object ordinary writes do not require extra superseded bookkeeping unless performing missing-history reaffirmation;
- ordinary fold coverage requires resolved causal ancestry; missing-ID reference coverage is restricted to reaffirmation;
- unsupported authenticated semantic versions explicitly block both ordinary authorship and ordinary TOTP consumption;
- full crash-safe `VAULT` publication ordering restored;
- common `OBJECT_VERSION` / `OBJECT_TYPE` TLV prefix frozen across the v1 semantic-version family;
- explicit `LOCAL_CONTINUITY_UNKNOWN` baseline re-establishment procedure added;
- DEVICE wide-frontier rename fold defined;
- stale earlier-revision terminology cleaned up.

### v1/r4

Fourth v1 design draft.

Adversarial-hardening changes from r3:

- automatic remote handoff removed; another writer's TOKEN cannot clear this client's local known accepted-head-loss warning while the remembered object remains unavailable;
- local user reaffirmation is the sole mechanism for clearing such local missing-history evidence without restoration of resolved causal history;
- `REFERENCE_COVERS_FOR_REAFFIRMATION` narrowed to local reaffirmation finalization;
- every intermediate reference-chain TOKEN must be available and valid; only the exact locally remembered terminal head ID may be unavailable;
- finalized reaffirmation permanently records exact reaffirmed historical TOKEN IDs in local security memory;
- exact reaffirmed historical IDs are excluded from this established client's current frontier even if synchronized causal proof is later deleted;
- newly discovered descendants of reaffirmed historical IDs are never suppressed and may create new conflicts;
- loss of the newer covering head creates fresh known accepted-head loss; old historical state is not silently resurrected;
- durable local security memory must be independent of the hostile synchronized namespace;
- rollback/reset/deletion/loss of local security-memory freshness enters `LOCAL_CONTINUITY_UNKNOWN` and loses prior rollback-continuity claims until explicit baseline re-establishment;
- complete-state confirmations now bind to a monotonic token confirmation-context epoch and become stale on change-and-return or process restart;
- multi-object convergence/reaffirmation batches explicitly bind one confirmation to the entire original required identity set;
- optional same-root exact-VAULT representation fingerprint warning restored for password/bootstrap rollback visibility.

### v1/r3

Third v1 design draft.

Changes from r2:

- `TOKEN` no longer repeats the 65-byte P-256 public key;
- `TOKEN` carries `AUTHOR_DEVICE_ID[32]`;
- `DEVICE_ID = SHA-256("totipo/v1/device-id" || PUBLIC_KEY_X963)`;
- `DEVICE` carries the canonical 65-byte X9.63 public key and self-signature;
- TOKEN provenance may be verified using matching DEVICE/local key material;
- missing matching public-key material is `PROVENANCE_UNRESOLVED`;
- zero-length or bad signatures are `PROVENANCE_REJECTED` without erasing TOKEN state;
- maximum-field TOKEN size drops from r2's 876 bytes to 843 bytes before parents;
- a maximum-field TOKEN now fits four parents in the 1024-byte envelope;
- defined `REFERENCE_COVERS` separately from resolved causal ancestry;
- known-history reaffirmation now uses signed reference coverage;
- replaced r2 §35.2 migration/capacity contradiction with bounded linear reaffirmation folding;
- added durable `superseded-ID security record` and explicit remembered-head replacement so completed user intent survives later loss of an intermediate batch object without masking future loss of the new covering head;
- clarified that wide-frontier convergence is a linear fold, not a combinatorial merge search;
- updated DEVICE causality/presentation to use derived `DEVICE_ID`.

### v1/r2

Second v1 design draft.

Changes from r1:

- semantic object envelope reduced from 2048 to 1024 bytes;
- device provenance changed from strict deterministic Ed25519 to ECDSA P-256 with SHA-256;
- deterministic signing requirement removed;
- P-256 signer public key encoded as fixed 65-byte ANSI X9.63 uncompressed point;
- conforming writer signature encoded as DER ECDSA, with `1..72` byte field bound;
- TOKEN assertion validity separated from provenance verification;
- provenance status introduced: `VERIFIED`, `UNRESOLVED`, `REJECTED`;
- failed or malformed provenance no longer erases vault-authenticated TOKEN state;
- only provenance-verified DEVICE objects provide authenticated friendly names;
- bounded ordinary-TOKEN convergence batches added for frontiers/reaffirmations that cannot fit in one 1024-byte object.

### v1/r1

Initial v1 design draft.

Major differences from frozen v0:

- new protocol line and lowercase `totipo/v1/...` namespaces;
- `TOKEN_UPDATE` renamed to `TOKEN`;
- `DEVICE_UPDATE` renamed to `DEVICE`;
- all token state fields required in every `TOKEN`;
- complete token state directly authoritative;
- parent availability no longer gates token value validity;
- parent edges classified `RESOLVED`, `UNRESOLVED`, or `REJECTED`;
- ordinary conflict is whole-token value disagreement;
- no automatic disjoint-field composition;
- known accepted-head loss uses ordinary TOKEN reaffirmation;
- no recovery marker, generation ID, field CRDT, or lifecycle witness machinery.
