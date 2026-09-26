# Totipo v1/r11 — Full Specification Review

**Scope:** `spec/totipo-vault-format-v1.md` (r11, 2868 lines), full-document adversarial review: internal consistency, arithmetic, cryptographic construction, cross-references, and state-machine completeness.

**Verdict:** The r11 body is coherent and largely self-consistent. All byte-level arithmetic, the TLV registry, the crypto constructions, the readiness gates, and the revision history check out. The findings below are localized specification defects and ambiguities — none require wire-format or crypto changes, but F1–F3 should be resolved before v1-rc1 freeze.

---

## Findings

### F1 (High) — `OPAQUE_UNSCOPED` evidence: "active" is undefined and there is no deactivation path in v1

§24.2:

```text
AUTHORITATIVE_VAULT_READY = ... and no OPAQUE_UNSCOPED evidence is active
```

Nothing in the document defines when opaque-unscoped evidence is *active* or how it ever becomes inactive. §37.3 says it "pauses" authoritative operations vault-wide — but per §37.6, once durably recorded:

- synchronized disappearance does not erase it;
- only "causal descendants, compatible reprocessing/upgrade, migration, or explicit continuity reset/re-establishment" change the durable interpretation.

None of those exits can apply: an opaque-unscoped object's scope is unknowable, so no descendant can resolve an edge to it (§22 requires same-TOKEN_ID/same-DEVICE_ID resolution), "compatible reprocessing" and "migration" are undefined in v1, and §34.3 baseline re-establishment is entered from `LOCAL_CONTINUITY_UNKNOWN`, not offered as a voluntary remedy — and it rebuilds opaque-unscoped records from currently available objects, so it is unclear whether it can ever drop a durably recorded one at all.

**Consequence:** any `K_root` holder can permanently degrade every v1 client of that vault to candidate-only mode by publishing a single authenticated 1024-byte object with an unknown `OBJECT_TYPE` or a routing-prefix-violating future version. That author is inside the documented threat model (§3), but the spec currently neither states the permanence nor provides any user-facing override.

**Recommendations:**

1. define "active" (presumably: durably recorded and not superseded);
2. state in §37.3 that within v1 the block is permanent, not a "pause";
3. either define an explicit user-acknowledged downgrade/supersede procedure, or explicitly accept the fail-closed trade-off in prose (as §54's philosophy suggests is intended), so reviewers and implementers do not invent their own recovery paths.

### F2 (High) — §49 DEVICE convergence excludes unverified current heads; no DEVICE analogue of TOKEN reaffirmation

TOKEN writers must parent exactly `KNOWN_CURRENT_TOKEN_HEADS(T)` — all heads regardless of provenance status (§32) — and §21.2 defines reaffirmation for rejected-provenance current TOKENs. §49 instead defines the rename parent set as "the complete current supported **provenance-verified** DEVICE frontier", and fold finalization requires "every original **verified** current DEVICE head" to be an ancestor.

An assertion-valid but provenance-`REJECTED`/`UNRESOLVED` current DEVICE head is therefore never incorporated: it remains a maximal current head forever, with no convergence path defined.

**Attack:** a `K_root` holder publishes a root `DEVICE { DEVICE_ID = victim's real derived ID, PUBLIC_KEY_X963 = victim's genuine 65-byte key, DISPLAY_NAME = attacker text, SIGNATURE = garbage }`. It is assertion-valid (the §20 derivation matches by construction), provenance-rejected, permanently current, and immune to every rename the victim performs. §31 leaves the presentation status of non-verified, non-opaque current heads undefined (only the opaque case is specified), so §55's "differing names => presentation conflict" may or may not fire depending on the implementation.

**Recommendation:** pick one and state it:

- (a) mirror §32 — a rename parents **all** current supported DEVICE heads regardless of provenance status (parenting is causality, not endorsement; this converges the hostile node into history); or
- (b) declare assertion-valid-but-unverified current DEVICE heads presentation-inert (no authenticated name, excluded from conflict counting) and rename-incorporable.

Either resolves the ambiguity; (a) is closer to the TOKEN design philosophy.

### F3 (Medium) — §24.7 "Detected corruption ... enters LOCAL_CONTINUITY_UNKNOWN": scope ambiguous

§24.7:

```text
Reappearing bytes for one durable OBJECT_ID must reproduce the same immutable routing record.
Detected corruption, global-ID inconsistency, or resolved cycles enters LOCAL_CONTINUITY_UNKNOWN.
```

If "detected corruption" is read as "synchronized bytes at a durably known OBJECT_ID fail AEAD/ID verification", then the hostile medium — which per §3 may freely *replace* files — triggers `LOCAL_CONTINUITY_UNKNOWN` (blocking all credential use, including candidate use, per §34.2/§38) with a single byte flip, and §34.3 cannot even rebuild that node since its bytes no longer authenticate. That is a trivial permanent DoS far beyond the spec's own degradation philosophy (§54: "loss of semantic certainty should normally degrade capability").

The consistent reading — supported by §22 ("unreadability, AEAD failure, keyed-ID mismatch ... leaves an otherwise unknown edge unresolved"), §55 ("detected **graph-record** corruption enters LOCAL_CONTINUITY_UNKNOWN"), and the fact that any bytes which *do* authenticate at a known ID are cryptographically guaranteed to reproduce the same record — is that "detected corruption" means corruption of the **local durable graph record** (or an authenticated record inconsistent with durable memory).

**Recommendation:** state the distinction explicitly in §24.7: synchronized-file AEAD/ID failure at a known ID ⇒ file treated as unreadable/absent, durable node retained, value marked unavailable if current; local security-memory corruption/inconsistency ⇒ `LOCAL_CONTINUITY_UNKNOWN`.

### F4 (Medium) — ECDSA nonce-quality guidance absent (§16)

§16: "ECDSA signatures need not be deterministic. Randomized signatures are valid and expected." Verification-wise this is correct, but a reused or predictable per-signature nonce discloses the device private key (classic ECDSA k-reuse attacks; real-world platform RNG failures exist). The impact is bounded (provenance key, not `K_root`), yet the spec leans on platform crypto (JCA/Keystore/CryptoKit) whose nonce handling varies.

**Recommendation:** add to §16: implementations SHOULD use deterministic ECDSA (RFC 6979) where available, or per-signature nonces drawn from the platform CSPRNG; and a detected repeated nonce across two signatures by one key SHOULD be surfaced as a provenance anomaly.

### F5 (Medium) — No writer requirement to publish a DEVICE alongside first TOKEN authorship

§16.1 lists three sources of the P-256 public key for TOKEN provenance. Only the `DEVICE` object makes the key available to *other* clients — "locally bound key material" helps only the author. Without a SHOULD/MUST, a conforming writer can leave its TOKEN provenance permanently `UNRESOLVED` for every peer while remaining conforming, quietly degrading the whole provenance feature.

**Recommendation:** a conforming writer SHOULD publish an assertion-valid `DEVICE` for its device key no later than its first TOKEN publication for that vault; add a §55 conformance bullet ("TOKEN provenance resolves for a fresh peer after DEVICE publication").

### F6 (Low) — `K_signature_context` in the signed message is never motivated (§19)

The provenance signature input embeds a vault-derived secret (`K_signature_context`). This is the load-bearing vault-binding of provenance: `DEVICE_ID` is vault-independent (`SHA-256(domain || pubkey)`), so without the context, a signature valid in vault A would verify in vault B for the same device identity, letting a malicious member of one vault forge attributed provenance in another. §52 provides exactly this kind of rationale for the object-ID/encryption construction; provenance deserves the same paragraph so implementers do not "optimize" the context away and reviewers can check the intent.

### F7 (Low, editorial) — Stale "r10" in §56

The opening sentence of "Open work after r11" reads "r10 retains the complete-state/durable-graph and opaque-routing architecture...". It should say r11. This is the one place where r11's own summary claim ("aligns revision history with the current body") is violated.

### F8 (Low) — Minor editorial notes

1. §40: "Newly generated secrets MUST contain at least 20 CSPRNG-generated bytes" reads oddly; suggest "MUST be at least 20 bytes generated by the platform CSPRNG" (imports remain `1..128`).
2. §10: `VAULT_BINDING = HMAC-SHA-256(K_root, domain)` uses `K_root` directly as an HMAC key while every other derived value goes through HKDF from `PRK`. Cryptographically fine; a one-line note (or uniform derivation) would remove the asymmetry.
3. §6: the empty password is format-valid and readers must accept it; consider a SHOULD that creation UIs warn on empty/weak passwords (policy already MAY require stronger).
4. §42.1 vs §15/§12.4: any non-1 `OBJECT_VERSION` with a preserved routing prefix is `OPAQUE_ROUTABLE` whether or not the value is actually allocated by a published Totipo spec — i.e., reader-side routability is not an endorsement of allocation governance. One sentence (and optionally an "unassigned-as-of-r11" marker in the durable record) would prevent a hostile writer using an unassigned version from being misread as official future-version evidence.
5. §4: accidental multi-vault sharing of one `objects-v1/` namespace silently yields AEAD failures classified as invalid storage evidence. Consider a SHOULD to surface unauthenticated-candidate counts as diagnostics (not warnings) so users notice misconfiguration.

---

## Verified consistent (no action)

- **Arithmetic:** 87-byte bootstrap (10+1+16+12+32+16; header = bytes 0..38 = 39); 1024 = 1008 plaintext + 16 tag; 1006-byte semantic capacity; `TOKEN_BYTES_RESERVED = 215` (5+5+6+12+36+36+5+4+4+26+76) + issuer + account + secret + 36×parents, giving 855/999/1035 for max fields with 0/4/5 parents and 1005 for the typical case; `DEVICE_BYTES_RESERVED = 213` + name + 36×parents, giving 469+504=973 (fits) vs 469+540=1009 (folds) at 14/15 parents; 72-byte DER ECDSA P-256 maximum (SEQ(2) + INT(35)×2); 36-byte parent TLV cost; 12-byte `AUTHOR_TIME` cost; fold rates (4-then-3 for TOKEN, 14-then-13 for DEVICE) guarantee forward progress over any finite frontier.
- **Registry/grammar agreement:** §42 tags ↔ §44 grammars ↔ §12.1–12.3 frozen routing prefixes all agree, including strictly increasing tag order, terminal `FF01`, reserved `0x0003`, and the DEVICE_ID derivation repeated identically in §2/§12.3/§16/§20.
- **Crypto:** Argon2id parameters match the RFC 9106 second recommended profile; all §11 HKDF domain strings cover every use (§10/§13/§14/§19); per-object key from the full OBJECT_ID with nonce from its first 12 bytes has no two-time-pad or cross-key nonce-reuse hazard (§52's analysis is sound); mandatory post-decryption keyed-ID recomputation and canonical zero padding close plaintext-substitution and alternate-padding attacks; `K_object_root`-rooted expand is valid HKDF usage.
- **State machine:** §24.2 readiness predicates are used consistently in §32/§33/§35/§38; AUTHOR_TIME non-causality is enforced uniformly (§20.1/§23/§53 and the §55 vectors); the §45 capacity numbers agree with the §55 conformance bullets and §56 item 6; the r7→r9 DEVICE fan-in evolution (15→14 after serialized DEVICE_ID) is internally consistent across §45.3 and §58.
- **Document structure:** sections 1..58 sequential with no gaps or duplicates; every "Section N" cross-reference resolves correctly; no residual normative v0-migration text remains (only historical references, consistent with the r11 governance changes).

## Suggested disposition before v1-rc1

- F1, F2, F3: resolve in the next revision (all are prose/state-machine clarifications; none touch wire format).
- F4, F5, F6: add the missing normative guidance and rationale paragraphs.
- F7, F8: fold into the same editorial pass.
