# v1/r17 observed content and store freshness clarification

External review identified an ambiguity in r16 §2: “External store bytes MUST be
treated as hostile and validated” could imply that v1 detects malicious withholding
or rollback. The settled design authenticates and validates observed material;
it does not authenticate the completeness or freshness of the supplied store view.
An older valid object or VAULT representation can authenticate even when newer
valid history is unavailable. This limitation does not make rollback harmless.

The normative edits are confined to §1 (early boundary statement), §2 (untrusted
observations, validation versus freshness, and publication versus later retention),
and §14 (safe interpretation of disappearance and older views). The preamble
identifies r17 and §21 adds its revision-history entry. §§3–13 and §§15–20 are
byte-identical to r16. No new protocol mechanism or semantic rule is intended.

Observed candidate material still receives every applicable path/name/type,
bounded-read, physical-size, AEAD, padding/length, keyed-ID, exact grammar,
semantic-bound, and VAULT authentication check. Missing parents remain unresolved;
no ancestry is invented or skipped. Available validated inputs determine observable
state; reappearance permits ordinary recomputation. A fresh client need not remember
prior runs, and lack of freshness evidence cannot supply protocol facts.

The informative deployment note in §2 locates additional freshness, rollback
detection, retained history, auditability, and stronger deletion resistance at the
storage/application layer. Deployments select or provide those properties explicitly.
Their guarantees depend on trust and retention assumptions; provider version history
alone proves no cryptographic rollback protection. Local directories, removable
media, network mounts, and externally synchronized stores remain valid deployments.
Synchronization is optional. Local publication obligations are unchanged.

Before editing vectors, all 90 r16 cases and their expectations were inspected.
No expectation needs to change. In particular, intermediate disappearance leaves
ancestry unresolved, reappearance resolves it ordinarily, authenticated invalid
material remains rejected, and ambiguous publication/VAULT outcomes remain failures.
Repeated generator runs confirmed all 90 case JSON files remain byte-identical.
The manifest revision, manifest-schema revision constant, and moving-profile
revision/specification/manifest/schema hashes are metadata changes only. See the [hardening report](V1_R17_HARDENING_REPORT.md).

The dedicated cold-reader review reached these conclusions:

| Question | Finding |
| --- | --- |
| Does an untrusted store imply detection of withheld newer objects? | No; §§1–2 explicitly deny complete/freshest history guarantees. |
| Do authenticated bytes imply freshness? | No; §2 explicitly permits older valid representations to authenticate. |
| Does VAULT_FINGERPRINT prove freshness? | No; unchanged §9 explicitly excludes it, and §8 requires exact BASE comparison. |
| Does immutable content addressing prevent rollback? | No; §2 separates authenticated identity from the observed set. |
| Is synchronization required for security? | No; §§1–2 keep it optional and external. |
| Does rollback invalidate otherwise valid old bytes? | No; §2 distinguishes old valid material from failed authentication/validation. |
| Are malformed/tampered observations still rejected? | Yes; §2 preserves all applicable checks in §§3, 6, 11–13. |
| Can another layer supply stronger properties? | Yes; the informative §2 note makes these explicit deployment choices, not v1 semantics. |

The terminology audit covered every occurrence in the current normative text of
hostile, untrusted, adversarial, malicious, rollback, replay, fresh, freshness,
complete, completeness, withhold, withholding, delete, deletion, history,
synchronization, synchronized, and provider. No surviving statement claims
cryptographic protection of store-view completeness/freshness. “Freshness comparison”
in §8 denotes the existing immediate comparison against BASE, not proof of the
latest historical wrapper. “Fresh” randomness/persistence operations, complete
TokenValues, and tombstone lifecycle deletion remain distinct concepts. Historical
§21 entries retain their historical wording and are explicitly non-current.

Non-goals: wire or crypto changes; new storage, graph, SCC, parent, fold, metadata,
TOTP, discovery, publication, VAULT, cache, or same-ID behavior; rollback databases,
generation counters, advisory-history capabilities, monotonic storage, or local
journals; mandatory cross-run history; vendor recommendations; Java changes;
commits, tags, pushes, releases, or an RC freeze. This is no new security feature
or design checkpoint.
