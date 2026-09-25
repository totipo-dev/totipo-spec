#!/usr/bin/env python3
from pathlib import Path

p = Path("spec/totipo-vault-format-v1.md")
s = p.read_text(encoding="utf-8")
norm = s[:s.index("## 59. Revision history")]

required = [
    "**Revision:** r9",
    "OPAQUE_ROUTABLE",
    "OPAQUE_UNSCOPED",
    "0200 DEVICE_ID",
    "AUTHORITATIVE_VAULT_READY",
    "CANDIDATE_USE_READY",
    "OPAQUE_CURRENT_TOKEN_HEADS(T)",
    "SIGNATURE_RESERVED_BYTES = 72",
    "469 + 14*36 = 973 bytes",
    "469 + 15*36 = 1009 bytes",
]
for x in required:
    assert x in norm, x

for x in [
    "ACCEPTED_HEAD_IDS",
    "ACCEPTED_TOKEN_HEADS",
    "ACCEPTANCE_PENDING",
    "SUPERSEDED_ID_SET",
    "REFERENCE_COVERS_FOR_REAFFIRMATION",
]:
    assert x not in norm, x

assert norm.count("| `0x0200` | `DEVICE_ID` |") == 1
assert "Unrelated tokens continue normal ordinary use and authorship" in norm
print("PASS: Totipo v1/r9 structural spec checks")
