#!/usr/bin/env python3
from pathlib import Path
import re

p = Path("spec/totipo-vault-format-v1.md")
s = p.read_text(encoding="utf-8")
norm = s[:s.index("## 58. Revision history")]

assert [int(n) for n in re.findall(r"^## (\d+)\.", s, re.M)] == list(range(1, 59))
for heading in [
    "## 54. Core invariants",
    "## 55. Required conformance evidence for v1",
    "## 56. Open work after r11",
    "## 57. v0 concepts intentionally absent from v1",
]:
    assert heading in norm, heading

allocation = norm.split("#### OBJECT_VERSION allocation", 1)[1].split("### 42.2", 1)[0]
for concept in [
    r"published.*specification.*process",
    r"semantic grammars.*envelope family",
    r"r11\s+assigns\s+`0x01`",
    r"[Aa]ll other values.*unassigned",
    r"MUST NOT.*independently assign.*interoperable/shared-vault.*published.*specification",
    r"no private-use or experimental.*OBJECT_VERSION.*range",
]:
    assert re.search(concept, allocation, re.S), concept

historical = norm.split("**Historical note:**", 1)[1].split("\n\n", 1)[0]
for concept in [
    r"v0 draft",
    r"never released.*implemented/deployed",
    r"not a supported predecessor of v1",
    r"v1 defines no migration protocol from v0",
]:
    assert re.search(concept, historical), concept
assert not re.search(r"^## .*Migration from v0", norm, re.M)
assert "Migration creates a new v1 vault" not in norm

assert re.search(r"^### v1/r11$", s.split("## 58. Revision history", 1)[1], re.M)

required = [
    "**Revision:** r11",
    "**Revision 11 summary:**",
    "objects-v1/",
    "Every valid object in objects-v1/ is exactly 1024 bytes.",
    "OBJECT_VERSION versions semantics inside the v1 envelope family",
    "Only direct regular-file children of `objects-v1/`",
    "Unknown sibling family names/files are not authenticated semantic evidence.",
    "A candidate in `objects-v1/` with any file length other than exactly 1024 bytes is invalid current v1-family storage evidence.",
    "It MUST NOT by itself create `OPAQUE_ROUTABLE` or `OPAQUE_UNSCOPED`",
    "A future envelope family that claims rolling-upgrade interoperability with v1 MUST maintain a v1-family compatibility projection in `objects-v1/`",
    "A future-family writer claiming rolling compatibility MUST ensure the required v1-family compatibility assertion becomes durable no later than it reports the corresponding future-family semantic action successful.",
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
print("PASS: Totipo v1/r11 structural spec checks")
