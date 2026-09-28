#!/usr/bin/env python3
"""Structural r15 invariants and r14-frozen wire/crypto section fingerprints."""
from pathlib import Path
import hashlib
import re

s = Path("spec/totipo-vault-format-v1.md").read_text(encoding="utf-8")
norm, history = s.split("## 58. Revision history", 1)
assert [int(n) for n in re.findall(r"^## (\d+)\.", s, re.M)] == list(range(1,59))
assert "**Revision:** r15" in norm and "**Protocol version:** 1" in norm
assert "## 56. Open work after r15" in norm
for revision in ("r15", "r14", "r13", "r12"):
    assert "### v1/" + revision in history
for obsolete in ("LOCAL_CONTINUITY_UNKNOWN", "KNOWLEDGE_PERSISTENCE_BLOCKED",
                 "BASE_OPERATION_SAFE", "AUTHORITATIVE_VAULT_READY", "CANDIDATE_USE_READY",
                 "OPAQUE_UNSCOPED_RECORD", "KNOWN_CURRENT_TOKEN_HEADS", "CONFIRMATION_EPOCH",
                 "PENDING(VAULT_BINDING)", "UNAVAILABLE_SUPPORTED_CURRENT_HEADS"):
    assert obsolete not in norm, obsolete
for concept in (
    r"Implementations MAY retain advisory local history",
    r"Implementations without advisory history remain baseline conforming",
    r"mere absence of an advisory-history feature is not HISTORY_MEMORY_LOST",
    r"need not retain cross-run graph history",
    r"claiming this capability.*?MUST satisfy the conditional advisory-history cases",
    r"MUST produce the applicable diagnostic state",

    r"No local observation proves global completeness",
    r"No successful local publication proves peer propagation",
    r"ACCEPTED_SNAPSHOT.*?successfully read, envelope-authenticated, keyed-ID authenticated",
    r"PARENTS\(new\) = H",
    r"Later discovery of history.*?does not retroactively invalidate",
    r"optional history is advisory",
    r"Loss or corruption of advisory history.*?does not invalidate",
    r"VAULT_BINDING.*?authoritative local state",
    r"multiple distinct complete TokenValue.*?requires explicit user confirmation",
    r"Immediately before publication, recompute this TOKEN-specific semantic context",
    r"Unrelated TOKEN changes.*?do not stale confirmation",
    r"opaque-routable current TOKEN head.*?blocks ordinary use and v1 semantic authorship for T",
    r"MUST NOT globally block ordinary v1 TOKEN use/authorship",
    r"Cache-update failure.*?does not.*?block authorship",
    r"Exact-byte equality is enough.*?no fsync",
    r"No persistent reconciliation state",
    r"MAY.*?retain.*?opaque bytes|Exact opaque bytes MAY be retained",
    r"PROCESSING_INCOMPLETE.*?Successfully accepted objects remain usable",
    r"present but malformed.*?MUST NOT silently become absence",
    r"New synchronized objects SHOULD.*?file flush/force",
    r"VAULT, local VAULT_BINDING, and private-key custody MUST be crash-safe",
    r"runtime.*?optional|optional.*?Runtime",
    r"Wrong-size files never become opaque evidence",
):
    assert re.search(concept,norm,re.S), concept
# Reject recognizable retired affirmative gates, independently of identifier spelling.
for retired in (
    r"every implementation (?:MUST|must) retain (?:local|advisory|remembered) history",
    r"all implementations (?:MUST|must) retain (?:local|advisory|remembered) history",

    r"MUST.*?durably persist.*?before authoritative operations",
    r"A changed frontier.*?aborts/restarts",
    r"Any non-batch safety-relevant observation advances",
    r"ordinary.*?MUST remain blocked.*?local continuity",
    r"exact authenticated object retention is.*?required",
):
    assert not re.search(retired,norm,re.I), retired
assert "469 + 14*36 = 973 bytes" in norm
assert "469 + 15*36 = 1009 bytes" in norm
assert "SIGNATURE_RESERVED_BYTES = 72" in norm
assert "Every valid object in objects-v1/ is exactly 1024 bytes." in norm
assert "v1 defines no migration protocol from v0" in norm

def section(n):
    return re.search(r"^## " + str(n) + r"\..*?(?=^## \d+\.|\Z)",s,re.M|re.S).group()
# These fingerprints come from committed r14, not the modified specification.
FROZEN_SECTIONS = {5: 'c8b20b9feaa654e61c37a8c3449a35da4cd925ccb53e4e00afbd7fa8ed0cb7c0', 6: 'e6538a284e3cdf873498aa14e277db9bbd8d8654dde268466879cd4c952f0d70', 7: 'd48c001a6e0169ac75693cbb7c0b761920050a108975ea437a4b11e15f9f6f27', 11: '8df42fdedc5b49f3c820321f2b6b9d6fa3eb5bb36ac3cf1097f412c3a53f4cc1', 13: '747986a1b1b80c93b9f92be3b1446717efb86261557055608662b9fa8968aef1', 14: '01856280166ac57bd0c2174016a4628e08286c0238372d47ccae95d8fb717551', 15: '233e03d828e3ea30b4a43b75461553cc4341e533f4d59a087c99b3fbc1d19580', 17: '10c67a08bf6969280361689d316ba49ec8d22633a7d8a2b84f5425c2e3de8274', 19: 'aba7b0ea7ceb1c2dc17f20c14d39bf2298b3e9748bec7e7cb88b7188221f054a', 20: 'fd4c4cd01741a53f9a85e3cbe0e2d9cfef6ca000b0b61fa86c4ae8894fcb289f', 21: '93477ac2b29fc7977fe2e9b43fafd1cc36f987feb24e5af2389795486a6cbeda', 40: 'c5526e6883ef1769e06bfc0bf8eb0a041d4b1ab093c031b01570390beb0185c5', 41: '9898d0971ac1d7c437b16b8643567b69d2f54b8cb8dc83b9aca795c1059dd534', 42: '384a9d8065d3340b46bf53b507c0c464210403ca2ba36433ee6e0b4168949ce8', 43: '2eda2fe7aee91c43f8ef583a6659a51e0c8fa326eb03fd1ca26c6dc8e5d2dca9', 44: '5a75da630a2a4aef59b7e3ba015ce60dc12b452620e3cc832dab5f97238addb4', 45: 'fea2758fb4b3428e6dec9b3eae74211272e4b56fe79d3b5991d18fcd92fed842', 46: '49d7e747f95bd20654b085a995444487d7e90ff7e26c9fa2d103912b8e7b615b', 52: '109ddb731cd889fc72d570fb8ad143cbf46e62254894feac045a34fe2d2bb79b'}
FROZEN_CODE_BLOCKS = {8: '65dac5f04b6d10118b963f7d9def76c92cb4f0bd9ee997b50b567d3e71f094e8', 12: 'b8fca02a670a2cb134fa0e7ea90acdd685e954ec887b93fae59d00acd10a1ee5', 16: '2a9847d972b2b221863764f435be14e1c5bcfa9f7efced4b2b8f923d727876bc', 18: 'd520206ab7ad2b2e0c1411bb512a479c5e4fa29263ecdc20a8c8f92ebd27d384'}
for n,want in FROZEN_SECTIONS.items():
    assert hashlib.sha256(section(n).encode()).hexdigest()==want, f"frozen section {n} drift"
for n,want in FROZEN_CODE_BLOCKS.items():
    blocks="\n".join(re.findall(r"```.*?```",section(n),re.S))
    assert hashlib.sha256(blocks.encode()).hexdigest()==want, f"frozen wire blocks {n} drift"
print("PASS: Totipo v1/r15 structural and frozen-wire spec checks")
