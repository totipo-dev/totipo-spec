# Requirements Profiles

[`v1-pre-rc.json`](v1-pre-rc.json) is a **moving** profile for the implemented
v1/r14 corpus. It pins every current case ID and case-file SHA-256, plus hashes of
the specification, manifest, manifest schema, and case schema. `make verify`
checks these pins; `make conformance` additionally executes every case.

The profile is created only from existing case files. IDs are stable once used
by consumers. Changes to expected bytes or requirements need explicit review;
normal checks never update the profile.

Do not freeze `v1-rc1.json` until the first live Totipo implementation has
independently consumed the corpus, byte expectations have stabilized, and the
remaining specification/security and platform reviews have no blocker. This
repository provides one reference consumer, not two independent implementations.

The r14 profile pins the same 90 cases and case hashes as r13: all 85 r12 cases
plus five recovery and provenance cases. It remains moving pre-RC evidence; no RC profile is frozen.
