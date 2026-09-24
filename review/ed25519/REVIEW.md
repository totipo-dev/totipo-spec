# Strict Ed25519 corpus review

Authority: Totipo r36 Section 16. Review performed explicitly during Phase 2 by the implementing agent: mathematical derivation and isolated arithmetic cross-check, followed by independent consumption using a different curve implementation. This is an author review with independent calculations, not a claim of an external human cryptographic audit.

`review.go` uses only the Go standard library's integer, hash and artifact utilities. It does not import `crypto/ed25519`, `filippo.io/edwards25519`, or any conformance package. It represents affine points over p=2^255−19, with d=−121665/121666, computes square roots with integer modular arithmetic, adds points using the twisted Edwards affine equations, and multiplies by unreduced integer scalars. Re-encoding, [8]P, [L]P, S bounds and the cofactorless equation are separately checked.

The 33 fixed cases contain **4 ACCEPT and 29 REJECT** expectations. Expected reason labels are manually specified predicates in the maintenance source. The reviewer checks those predicates; it does not assign expectations from the standard verifier's result. Ordinary review is read-only:

```sh
go run ./review/ed25519/review.go
```

Explicit creation/maintenance is `go run ./review/ed25519/review.go --write-vectors`. Normal tests and CI never invoke that mode. The committed expectations are in `vectors/v0/ed25519/phase2.json`.

## Mathematical review

The prime subgroup order is L=2^252+27742317777372353535851937790883648493. The reviewer uses that integer directly, never a scalar constructor that would reduce L to zero.

- `rfc8032-empty`: copies RFC 8032 section 7.1 TEST 1. The isolated signing routine reproduces the complete published public key and signature before any additional signature is emitted. Source: [RFC 8032](https://www.rfc-editor.org/rfc/rfc8032.html#section-7.1).
- `totipo-token`, `totipo-device`: copy the existing externally pinned signature, public key and signature input exactly from the object fixtures. They remain reviewed-pinned provenance; no signature is regenerated to replace them.
- `independent-positive`: the isolated affine signer uses a public 32-byte `42` seed and the literal public message in the fixture. Its deterministic RFC 8032 construction produces a prime-subgroup A/R, an in-range S, and the required equation. It is separate from the standard signer and the implementation under test.
- `a-length-*`, `sig-length-*`: exact byte counts alone establish rejection before point arithmetic.
- `a-nondecode`, `r-nondecode`: the recorded y yields a nonsquare x²=(y²−1)/(dy²+1) modulo p. The reviewer searches a small public y only to construct this input; the fixed expected rejection follows from the recorded nonsquare predicate.
- `*-noncanonical-y`: y=p+1 reduces to identity y=1 and cannot round-trip to its supplied bytes. `*-negative-zero`: x=0 with sign bit one re-encodes with sign zero. Canonical-byte equality rejects both families, independently of later small-order checks.
- `a-identity`, `r-identity`, `a-order-two`, `r-order-two`: identity and T=(0,−1) have [8]P=identity. `identity-forgery` deliberately satisfies the verification equation with A=R=identity and S=0; the mandatory small-order gate still rejects it.
- `a-mixed-order`: A=B+T. Because B has order L and L is odd, [L](B+T)=T, while [8](B+T)=[8]B is not identity. Thus this is specifically a subgroup failure, not a small-order failure.
- `r-mixed-order`: R=2B+T has the same nontrivial order-two component and fails true prime-subgroup membership.
- `mixed-equation-0`, `mixed-equation-1`: A=B+T, R=2B+uT, S=(2+k) mod L. The public message is selected so k mod 2=u. Consequently R+[k]A=(2+k)B+(u+k)T=[S]B, so the cofactorless equation holds. A still has [L]A=T and must be rejected. With u=1, R is also mixed order. The reviewer separately asserts the equation before asserting subgroup rejection. These cases defeat a purported subgroup test that reduces the integer L modulo L.
- `s-equal-l`, `s-greater-l`, `s-maximum`: explicit little-endian integers violate S<L without relying on an equation check.
- `modified-public-key`, `modified-r`: negate an accepted prime-subgroup point, preserving canonicality, non-small order and subgroup membership. The final equation fails.
- `modified-s`, `modified-message`: retain otherwise valid point/scalar domains but break the signature equation.
- `canonical-invalid-equation`: A=R=B and S=0 are individually well-formed; the recorded message does not satisfy the required equation.

Both the arithmetic reviewer and the strict Go consumer agree on every fixed expectation. The consumer uses pinned `filippo.io/edwards25519 v1.1.0` for point decoding/addition and explicit full-integer [L] multiplication. It checks canonical encodings by round-trip, excludes small-order points, checks both subgroup memberships and S<L, then uses standard PureEd25519 verification for the equation. `golang.org/x/crypto/ed25519` is only a wrapper around the standard verifier and does not expose the needed point arithmetic.

## Standard-library comparison

On Go 1.26.7, standard `crypto/ed25519.Verify` differs on exactly three cases: `identity-forgery`, `mixed-equation-0`, and `mixed-equation-1`. It accepts each; Totipo rejects them. Both accept all four positives. The other 26 negatives are rejected by both. Wrong public-key lengths are rejected before calling the standard API, which would otherwise panic.

This comparison is diagnostic. It is not used to establish any expected disposition. Source review and test evidence establish confidence in this corpus; they do not make these 33 cases an exhaustive proof of all curve arithmetic or a production security audit.
