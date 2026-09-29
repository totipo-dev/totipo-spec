This document is a non-normative design checkpoint for the future Totipo v1/r16 simplification. It records decisions made after `V1_R16_SIMPLIFICATION_DECISIONS.md`. The current normative specification remains v1/r15.

Where this checkpoint resolves a question explicitly left open by the first checkpoint, this later checkpoint records the current agreed design direction for the future r16 rewrite.

# Totipo v1/r16 state/policy decisions — second design checkpoint

MUST / MUST NOT language in this document records requirements for the future
r16 rewrite; it does not amend current r15. The normative specification remains
[`spec/totipo-vault-format-v1.md`](../spec/totipo-vault-format-v1.md).

Starting HEAD: `bb1c9bc38e6b5954a68634a9f55fc500de3db622`.
Before editing, `git status --short` was empty, and `make check` and `make verify`
passed. The first checkpoint was committed at this HEAD; no pending r16
normative rewrite was present.

## Relationship to the first checkpoint

[`review/V1_R16_SIMPLIFICATION_DECISIONS.md`](V1_R16_SIMPLIFICATION_DECISIONS.md)
remains the unchanged historical first checkpoint. It already settled:

- removal of DEVICE and provenance signatures/device identity;
- optional informational CLIENT_NAME / CLIENT_TIME;
- complete-state TOKENs, including tombstones, and flattened credential fields;
- removal of OBJECT_TYPE / OBJECT_VERSION and exact v1 grammar inside `objects-v1/`;
- explicit parent TLVs, field-size bounds, and guaranteed four-parent physical capacity;
- removal of opaque/future semantic machinery.

Those decisions remain intact. That checkpoint left graph-cycle/integrity
semantics, confirmation freshness, max-parent policy, candidate-use policy,
advisory history, VAULT_BINDING/local state, and platform implementation open.
This checkpoint resolves the first four, with exact same-OBJECT_ID defensive
failure wording deferred. Advisory history, local state, and platform decisions
remain open.

## Settled: cycles are causal-equivalence groups

Define a resolved parent edge from child to parent, within one TOKEN_ID. An
object A reaches B when a path of zero or more resolved parent edges leads from
A to B. The zero-length path makes reachability reflexive. For TOKEN objects A
and B of the same TOKEN_ID:

```text
A ≡c B
iff
    A reaches B
    and
    B reaches A
```

This is equivalence under mutual reachability. Its classes are the strongly
connected components (SCCs) of the resolved TOKEN graph. The mathematical
relation is the normative direction; implementations need not use any named
algorithm such as Tarjan's or Kosaraju's.

A resolved TOKEN cycle MUST NOT be a special fatal graph-integrity state or a
cause of vault-global/local-continuity poisoning. Within one causal-equivalence
class:

- no member supersedes or is preferred over another;
- all members occupy the same causal level;
- every object's complete value remains independently meaningful.

Causal equivalence is not TokenValue equality. With arrows denoting parent
references (child to parent):

```text
A -> B -> C
^         |
|---------|
```

there is one causal-equivalence group `{A, B, C}`. If all three complete
TokenValues are X, the current semantic value may be unambiguous X. If instead
`A = X`, `B = Y`, and `C = X`, X and Y remain an ordinary whole-state conflict
when this group is current.

### Current heads are maximal causal groups

Conceptually collapse each equivalence class to one node and omit internal
edges. The resulting component graph is a DAG. A distinct class H is a strict
descendant of G when a resolved parent path leads from H to G. A class is
maximal/current when no distinct class is its strict descendant.

> The current TOKEN head set is every TOKEN object belonging to every maximal causal-equivalence class.

If the only maximal class is `{A, B, C}`, then `HEADS = {A, B, C}`. All three
objects are equally current; none is chosen as a representative winner.

### Descendants supersede an entire group

If D is a strict descendant of any member of group G, D is causally after G as
a whole. For example, here the downward branch denotes a descendant, and D
lists A as a parent:

```text
A <-> B
 \
  D

D.parents = {A}
```

Because A and B are mutually reachable, D reaches the whole `{A, B}` group.
If D's group is maximal and there are no other maximal groups, `HEADS = {D}`.
D need not list every member of the old cycle as a direct parent. This follows
ordinary reachability through the equivalence-class model.

### Late-arriving cycle edges are ordinary recomputation

Using child-to-parent reference arrows, initially:

```text
A -> B
B -> C?
```

C is unavailable, so no cycle is currently established. Later valid C appears
and asserts `C -> A`. The graph now contains `A -> B -> C -> A`, forming one
causal-equivalence class. Recompute under ordinary snapshot/learned-fact rules.
This changes causal grouping; it MUST NOT mark vault continuity unknown,
persist a cycle failure marker, require rebaseline/reset, or globally block
unrelated tokens. This example assumes C's resolving facts were not already
known; it does not erase previously learned facts when bytes disappear.

### Defensive implementation behavior

Implementations must avoid unbounded recursion, use cycle-safe traversal,
reject malformed object encodings before graph interpretation, and remain
robust against hostile synchronized input. Valid resolved cycles are not
cryptographic corruption.

Non-normative rationale: valid cycles are expected to be extraordinarily
difficult to construct because object IDs commit to parent IDs. The r16
semantics MUST NOT depend on that cryptographic assumption.

The future rewrite should remove cycle-triggered machinery, including concepts
such as `RESOLVED_CYCLE` preventing ordinary semantic evaluation,
cycle-caused `LOCAL_CONTINUITY_UNKNOWN`, durable cycle failure markers,
cycle-specific rebaseline/reset, and vault-global readiness failure solely
because one token has a cycle.

### Same-OBJECT_ID contradiction remains exceptional

If two different successfully authenticated canonical plaintexts somehow
validate under the same OBJECT_ID, an implementation MUST NOT choose
arbitrarily between them. This is a cryptographic/integrity failure of that
object identity, distinct from cycle semantics and from ordinary graph states.
Keep it as a defensive exceptional condition. This checkpoint introduces no
elaborate durable recovery semantics; exact normative scoping and recovery
wording may be finalized during the r16 rewrite.

## Settled: authorship survives unavailable history

> A user must never be prevented from making a new complete TOKEN assertion merely because previously known history is currently unavailable from synchronized storage.

Unavailability limits what the client can currently inspect or explain. It does
not erase known causal facts or make authorship impossible. Every newly
authored TOKEN still carries a complete TokenValue.

### Known causal facts and current availability are distinct

The future model distinguishes:

```text
KNOWN CAUSAL FACTS
CURRENT AVAILABILITY
```

A client may know an OBJECT_ID, its TOKEN_ID, its parent claims, and possibly
its previously validated complete TokenValue even when synchronized storage
no longer supplies the bytes. Availability can fluctuate:

```text
available -> unavailable -> available
```

A learned valid causal fact does not become false when bytes disappear. This
does not authorize inventing ancestry or values that were never learned, or
authenticating corrupt current bytes using remembered facts.

This settles the semantic distinction, not durability scope. Whether all or any
such facts must survive process restart remains part of the advisory-history
and local-state discussion. Remembering a fact does not mean its bytes are
currently available.

### Writers may reference unavailable known parents

A writer that knows parent object ID B as a causal fact may include
`PARENT_ID = B` even when B's bytes are currently unavailable. Suppose the
user observed conflict `heads = {A, B}` and selected complete desired value Z.
If synchronized storage loses B before publication, the writer may still author:

```text
C:
    parents = {A, B}
    value = Z
```

C remains a valid assertion. On another client that does not possess B or the
facts needed to resolve it, `C -> B` is an unresolved parent edge until B
becomes available. A missing parent does not invalidate C.

The future rewrite must avoid any requirement equivalent to “every listed
parent must currently be readable.” A parent reference is a causal claim by
object ID. The child independently carries a complete TokenValue; missing
parent bytes may prevent complete ancestry resolution, but not interpretation
of the child's own value.

### Disappearing information does not invalidate an informed decision

If the user already considered `{A, B}` and selected Z, B disappearing does
not make that decision stale solely because B disappeared. There is less
current availability, not new contradictory information. Publication may
continue with `parents = {A, B}` and `value = Z`. Do not ask the same question
again solely because an already-considered branch became unavailable.

New information is different. If a genuinely new current causal alternative D
becomes known after the decision over `{A, B}`, the application must expose
that fact before claiming the new assertion resolves all known alternatives.
The previous decision did not include D; authorship is not forbidden. A
subsequent assertion may include `parents = {A, B, D}`, subject to fixed
parent/fold rules. If D disappears after the user considers it, disappearance
alone does not invalidate the updated decision.

### History never previously observed

Distinguish a previously understood B that is now unavailable from B known
only by an ID referenced by another object, whose contents were never
available. In the latter case, the client cannot claim to know what B asserted.
The application should truthfully disclose known current/causal IDs, which
complete values are known, and which referenced/history values are unavailable.
Even then, the user remains free to assert a new complete TokenValue. Missing
information calls for truthful disclosure, not a protocol authorship prohibition.

### Confirmation is descriptive, not an authorization gate

Confirmation/UI interaction ensures the user understands observed alternatives
and missing information before making a new assertion. It should no longer be
a protocol permission object that can permanently prohibit publication.

Once the user selects a complete desired TokenValue over the relevant known
causal context, disappearance of considered history does not stale the choice;
newly learned relevant history must be surfaced; and the user may always make
another complete assertion. Do not retain process-local “learned then
disappeared” semantic-generation machinery solely to invalidate a previously
informed decision. The precise application/UI API remains deferred.

Remove any separate rule equivalent to `current value unavailable => authoring
forbidden`. Report facts instead, such as `head A: complete value known` and
`head B: value unavailable`. The user/application decides whether to make a
new complete assertion. UI confirmation concerns informed choice, not protocol
permission.

## Settled: fixed maximum of four parents

The first checkpoint's parent-maximum question is now resolved:

```text
MAX_PARENTS = 4
0 <= PARENT_COUNT <= 4
```

This is a fixed grammar rule. Do not use value-dependent capacity or allow
additional parents because a particular value or metadata instance is smaller.
The maximum applies equally to ordinary updates, deletion/tombstone assertions,
restore assertions, conflict resolution, reaffirmation, and every fold stage.
No state/value class receives a larger budget. Parents remain explicit,
canonical, sorted, and duplicate-free as settled in the first checkpoint.

### Four parents always fit

The arithmetic in the first checkpoint's “Settled: field limits and guaranteed
parent capacity” section gives:

```text
maximum legal non-parent fields       861 bytes
four explicit parents: 4 * 36         144 bytes
                                     ---------
total                               1005 bytes
retained semantic payload maximum   1006 bytes
```

> Every legal TOKEN always fits four parents regardless of its allowed issuer, account, secret, CLIENT_NAME, or CLIENT_TIME sizes.

This invariant motivates the fixed maximum. It uses the retained payload
budget; whether the fixed 1024-byte envelope remains optimal is still open.

### Fixed linear-carry fold shape

For a frontier wider than four, use the existing linear-carry concept with
fixed capacity. For original heads `A B C D E F G H I J ...`:

```text
F1 <- A B C D
F2 <- F1 E F G
F3 <- F2 H I J
```

Here each line lists the fold object's direct parents, not a reachability-arrow
convention. The first stage incorporates four original heads. Each later stage
incorporates the previous fold head plus up to three additional original heads.
For `N > 4` original heads, the stage count is:

```text
ceil((N - 1) / 3)
```

Stage width must not depend on token field lengths. Do not reintroduce
signature-size reservation or DER-dependent capacity.

All stages carry the same complete desired TokenValue and the operation's
client metadata according to future r16 exact wording. They exist only to
incorporate more than four parents. Intermediate fold history is ordinary
TOKEN history. Folding does not alter semantic permission, and completion is
not an authorization boundary.

## Settled: protocol state is descriptive, not prescriptive

> Totipo describes the state it observes. It does not decide what the user is allowed to do with a complete known TokenValue.

Report factual state: current, conflicting, historical/stale,
deleted/tombstoned, partially unavailable, missing ancestry, complete value
known, or complete value unavailable. These descriptions MUST NOT become
credential-use permission.

The future rewrite should remove normative machinery whose purpose is to
authorize or prohibit using one known TokenValue, including candidate-use
readiness, candidate-use eligibility, candidate catalogs as permission sets,
ordinary-use versus explicit-candidate-use authorization, protocol-level
candidate-use warning gates, and historical candidate permission. Applications
may still provide useful views over known values. The protocol describes
those values' relationship to current state.

### TOTP computation applies to any complete known value

TOTP is mathematical computation over:

```text
ALGORITHM
DIGITS
PERIOD
SECRET_BYTES
time
```

When these fields form a complete known value, the protocol does not forbid
computation because the value is historical, stale, a conflict branch, or
marked tombstoned/deleted. Applications must describe it truthfully; the user
decides whether to use it.

### Tombstone is bookkeeping/presentation state

> TOMBSTONE means the token is marked deleted for bookkeeping and normal presentation purposes.

It does not mean the credential is cryptographically unavailable or invalid,
that TOTP computation is prohibited, that the user may not inspect/use the
value, or that the secret must be discarded. A tombstoned TOKEN remains a
complete TokenValue.

An application may hide tombstones from the ordinary active-token list, show
them in Trash / Deleted Items, and offer restore/revive. The user may still
choose to inspect issuer/account, generate a TOTP, copy/use the known
credential, compare it with another value, or restore it. These are application
decisions; tombstone is not a TOTP-use prohibition.

Restoration/revival is an ordinary new complete TOKEN assertion with
`STATUS = LIVE`. The user may restore the tombstone's complete value directly
or modify fields while restoring. No history reconstruction is required solely
to restore a complete readable tombstone.

### State descriptions must remain truthful

Applications MUST NOT present a historical value as the unique current value,
a conflict branch as conflict-free current state, or a tombstone as ordinary
active-list state without clearly indicating deleted status. An unavailable
current head must not be silently omitted when describing whether current
state is complete. These are correctness/presentation obligations, not
credential-use authorization.

### Example state descriptions

In these examples, X and Y denote complete semantic values; the tombstone
example spells out status separately to emphasize retained credential material.

**Unique current**

```text
heads = {A}
A value = X

Description: current complete value X
```

**Conflict**

```text
heads = {A, B}
A value = X
B value = Y

Description: current conflict between X and Y
```

**Equal concurrent values**

```text
heads = {A, B}
A value = X
B value = X

Description: one current semantic value X represented by two causal heads
```

**Partial unavailability**

```text
heads = {A, B}
A value = X
B value unavailable

Description: current value X is known from A;
another current head B is unavailable;
current state is not fully known
```

X is not forbidden to use. Unknown B cannot be assumed equal to X or assumed
to conflict with it.

**Tombstone**

```text
heads = {A}
A = TOMBSTONE + complete credential X

Description: token is currently marked deleted;
complete value X remains available
```

**Historical value**

Here the arrow denotes causal progression from ancestor to descendant, the
reverse of a child-to-parent reference arrow; B lists A as a parent:

```text
A -> B
A value = X
B value = Y

Description: Y is current
X is historical/stale
```

The user may still choose X.

### Composition with causal-equivalence cycles

For `A <-> B`, both objects are current heads if their causal group is maximal.
If A carries X and B carries Y, describe an ordinary current conflict. If both
carry X, describe one semantic current value with multiple equivalent causal
heads. No special “cycle warning state” is required for correctness, though
implementations may expose diagnostics.

## Newly resolved items from the first checkpoint

| First-checkpoint open item | Decision now |
| --- | --- |
| Graph-integrity cycle machinery | Replace fatal-cycle semantics with causal-equivalence/SCC semantics |
| Confirmation freshness | Disappearance of considered history does not stale the user's decision; new facts must be disclosed |
| Parent maximum | Fixed `MAX_PARENTS = 4` |
| Candidate-use policy | Protocol becomes descriptive; remove candidate authorization machinery |

## Open questions after this checkpoint

### A. Advisory / learned history durability

Learned causal facts and availability are distinct. Still open:

- what minimum learned state, if any, must survive process restart;
- whether cross-run remembered history is normative, advisory, or application-only;
- whether storage regression should produce warnings;
- whether local history may be discarded;
- how this interacts with unreliable synchronization.

This checkpoint does not decide those questions.

### B. VAULT_BINDING / local authoritative state

Still open. This checkpoint does not select local binding or authoritative
state requirements.

### C. Desktop/platform-specific storage implementation

Still open, including portable NIO versus platform-linux. No implementation
strategy is selected here.

### D. Fixed 1024-byte envelope

Still explicitly open from the first checkpoint. Capacity arithmetic above
uses the retained 1006-byte semantic payload maximum without deciding the
broader envelope question.

### E. Same-OBJECT_ID defensive failure wording

The high-level defensive exceptional-condition direction is recorded above.
Exact normative scoping/recovery text may be finalized during the r16 rewrite.

### F. Public Java API

Still open, including the precise confirmation/UI API. Other first-checkpoint
non-decisions, such as garbage collection and future-family migration, are not
resolved by this document.

## Explicitly superseded ideas for future r16

Future r16 should no longer preserve as normative requirements:

- cycles causing persistent continuity failure;
- cycles causing vault-global operation blocking;
- value-dependent parent capacity or more-than-four direct parents;
- candidate-use permission states;
- TOMBSTONE as a reason to prohibit TOTP generation;
- unavailable current values as a reason to categorically prohibit new assertions;
- disappearance of already-considered history as automatic confirmation staleness.

This does not amend current r15, nor assert that all these mechanisms remain
requirements in r15.

## Baseline terminology and scope notes

“TokenValue” denotes the complete semantic value named `TOKEN_VALUE` in r15,
including STATUS. `MAX_PARENTS = 4` is future design notation, not a claim about
the current grammar. CLIENT_NAME / CLIENT_TIME are first-checkpoint r16 fields.

Current r15 Section 22 requires acyclic resolved graphs and aborts evaluation
of an inconsistent snapshot/vault context for a cycle or same-ID contradiction.
It explicitly requires no permanent sticky cache flag after restart. Thus the
cycle failure markers, `RESOLVED_CYCLE`, and cycle-caused
`LOCAL_CONTINUITY_UNKNOWN` discussed above identify machinery/concepts to avoid
in r16, not a claim that r15 currently mandates those named durable states.

Current r15 Sections 23–25 derive topology and heads from accepted snapshot
evidence; remembered absent objects are not current heads and advisory topology
must not silently supply missing paths. Its Section 27 already distinguishes
missing remembered history from mandatory unavailable-state confirmation, but
still binds conflict confirmation to a changing semantic context. This future
checkpoint records the broader known-fact/availability distinction and informed
choice principles without deciding cross-run durability or rewriting r15's
accepted-snapshot rules in place.

Current r15's explicit candidate is a readable supported LIVE TOKEN; future
r16's descriptive model deliberately also allows user-chosen computation from
a complete tombstone. Candidate authorization labels in this document describe
the machinery's purpose, not necessarily exact r15 identifiers.

Diagram arrows are labeled where necessary: parent references run child to
parent, whereas the historical-value example uses ancestor-to-descendant
progression. This notation distinction does not change the causal relation.

Only this non-normative review artifact is added. The first checkpoint remains
byte-unchanged. No specification, vectors, schemas, requirements profile,
conformance implementation, generators, revision constants, README revision,
or release/tag metadata is changed. Normative rewriting and corpus changes
remain subsequent work.
