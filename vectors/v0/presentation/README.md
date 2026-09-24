# DEVICE_UPDATE presentation history

`phase2.json` contains 11 post-authentication presentation-history fixtures. Updates name a symbolic ID, signer, parent list, name and object kind. Expected output includes validation per ID, maximal heads per signer and distinct displayed name values. Equal displayed names retain distinct heads.

The presentation validator does not affect token consensus validity. Pending presentation durability, abandonment, re-entry and rollback are exercised separately in recovery traces.
