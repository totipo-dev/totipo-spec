# Recovery/security-memory traces

`phase2.json` contains 28 explicitly reviewed state-machine fixtures. Each trace event supplies `op`, string-valued `args`, and `expected` outcome/checkpoint facts. The case's `input.binding` initializes an established symbolic binding, or an empty string initializes UNESTABLISHED.

Both the primary durable-memory simulator and independent durable-fact-ledger reviewer consume every trace. They model exact-ID/vault/scope-bound pending evidence and abandonment, future observations/bypass, regression, conservative durable replacement, partial batches, presentation recovery, confirmation freshness, establishment, and capacity/migration. See [the review](../../../review/phase2/REVIEW.md) for assumptions, section derivations and limits.

`valid` means upstream authentication/semantic validation succeeded. `terminal` requires an identity-grounded proof input. Binding labels stand for equal HMAC-derived bindings. These are abstract state-machine facts, not substitutes for cryptographic validation or OS durability tests.
