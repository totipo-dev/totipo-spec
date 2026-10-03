#!/usr/bin/env python3
"""Structural and semantic anchors for the current r18 specification.

Historical revision entries are deliberately excluded from retired-term checks.
Wire correctness is exercised by the conformance corpus, not prose fingerprints.
"""
from pathlib import Path
import re

s = Path('spec/totipo-vault-format-v1.md').read_text(encoding='utf-8')
norm, history = s.split('## 21. Revision history', 1)
assert [int(n) for n in re.findall(r'^## (\d+)\.', s, re.M)] == list(range(1, 22))
assert '**Revision:** r18' in norm and '**Protocol version:** 1' in norm
archive = Path('review/V1_PRE_R16_REVISION_HISTORY.md').read_text(encoding='utf-8')
assert '../review/V1_PRE_R16_REVISION_HISTORY.md' in history
for n in range(1, 19):
    assert f'### v1/r{n}\n' in (archive if n < 16 else history)
for term in ('DEVICE', 'DEVICE_ID', 'P-256', 'ECDSA', 'DER', 'SIGNATURE',
             'AUTHOR_DEVICE_ID', 'AUTHOR_TIME', 'PROVENANCE', 'OPAQUE_ROUTABLE',
             'OPAQUE_UNSCOPED', 'OBJECT_TYPE', 'OBJECT_VERSION', 'VAULT_BINDING',
             'READY', 'PROCESSING_INCOMPLETE', 'HISTORY_REGRESSION',
             'HISTORY_MEMORY_LOST', 'candidate-use', 'K_signature_context'):
    assert not re.search(r'(?<![A-Z_])'+re.escape(term)+r'(?![A-Z_])', norm), term
assert 'provenance' not in norm.lower()
assert 'opaque' not in norm.lower()
for anchor in ('TOKEN is the sole semantic object grammar', 'MAX_PARENTS = 4',
               '0 <= PARENT_COUNT <= 4', '1005', '1006', '1024',
               'VAULT_FINGERPRINT', 'totipo/v1/vault-fingerprint',
               'A ≡c B iff A reaches B and B reaches A',
               'CURRENT != BASE', 'Synchronization is\noptional and external',
               'MUST NOT require exhaustive enumeration',
               'Parent availability is not a publication prerequisite'):
    assert anchor in norm, anchor
registry = re.findall(r'^\| (0x[0-9a-f]{4}) \| ([A-Z_]+) \|', norm, re.M)
assert registry == [(f'0x{i:04x}', name) for i, name in enumerate((
    'TOKEN_ID', 'PARENT_COUNT', 'PARENT_ID', 'STATUS', 'ISSUER', 'ACCOUNT',
    'ALGORITHM', 'DIGITS', 'PERIOD', 'SECRET_BYTES', 'CLIENT_NAME', 'CLIENT_TIME'), 1)]
# Every active numeric section reference must resolve.
for ref in re.findall(r'Section (\d+)', norm):
    assert 1 <= int(ref) <= 20, ref
# Storage-layout code block defines a single exact bootstrap path.
layout = re.search(r'```text\nconfigured Totipo store\n(.*?)```', norm, re.S).group(1)
assert layout.splitlines()[0].strip() == 'vault'
assert not re.search(r'^\s*VAULT\s*$', layout, re.M)
assert not re.search(r'`VAULT`[^.\n]*(?:pathname|path is)', norm)
fold = norm.split('## 17.', 1)[1].split('## 18.', 1)[0]
preserved = dict(re.findall(r'^\| (TokenValue|CLIENT_NAME|CLIENT_TIME) \| ([^|]+) \|$', fold, re.M))
assert set(preserved) == {'TokenValue', 'CLIENT_NAME', 'CLIENT_TIME'}
for field in ('CLIENT_NAME', 'CLIENT_TIME'):
    assert 'same exact presence' in preserved[field]
values = norm.split('## 16.', 1)[1].split('## 17.', 1)[0]
value_tuple = re.search(r'TokenValue is the complete tuple `([^`]+)`', values, re.S).group(1)
assert 'CLIENT_NAME' not in value_tuple and 'CLIENT_TIME' not in value_tuple
heads = norm.split('## 15.', 1)[1].split('## 16.', 1)[0]
assert all(field in heads for field in ('CLIENT_NAME', 'CLIENT_TIME', 'TokenValue'))
assert re.search(r'MUST NOT erase metadata from head objects', heads)
# Entry types are part of the canonical storage contract, not host syscall APIs.
store = norm.split('## 3.', 1)[1].split('## 4.', 1)[0]
entry_types = dict(re.findall(r'^\| `(vault|objects-v1)` \| ([^|]+) \|', store, re.M))
assert entry_types == {'vault': 'regular file only', 'objects-v1': 'directory only'}
flat_store = ' '.join(store.split())
assert re.search(r'symlink.*?MUST NOT be deliberately followed or interpreted', flat_store)
assert re.search(r'symlink.*?MUST NOT be deliberately traversed', flat_store)
for syscall in ('openat2', 'O_NOFOLLOW', 'O_PATH', 'statx', 'SecureDirectoryStream'):
    assert syscall not in norm, syscall
assert 'optional implementation hardening' in norm
metadata = ' '.join(norm.split('## 12.', 1)[1].split('## 13.', 1)[0].split())
assert re.search(r'For as long as.*?represented, returned, exposed, or retained.*?MUST be preserved exactly', metadata)
assert 'does not require persistent retention' in metadata
assert 'no advisory-history, remembered-head, protocol-cache, or cross-run persistence requirement' in metadata
writing = ' '.join(fold.split())
assert re.search(r'Unavailability of history alone MUST NOT prohibit authorship when the TOKEN_ID and complete desired TokenValue.*?known or supplied', writing)
assert 'generate a new CSPRNG TOKEN_ID' in writing and 'empty parent set' in writing
# Application policy strength and scope boundaries remain distinct from core bytes.
def section_text(number):
    return ' '.join(norm.split(f'## {number}.', 1)[1].split(f'## {number + 1}.', 1)[0].split())

password = section_text(5)
assert 'interactive application MUST require explicit confirmation' in password
assert 'before creating a vault with an empty password' in password
creation = section_text(7)
for anchor in ('during available observation', 'application observes',
               'SHOULD prominently warn', 'SHOULD require explicit user confirmation',
               'SHOULD recommend checking synchronization and provider state',
               'MUST NOT require exhaustive enumeration',
               'MUST NOT by itself make fresh creation impossible',
               'unauthenticated contextual evidence'):
    assert anchor in creation, anchor
assert re.search(r'MUST NOT describe tombstoning or deleting a token as securely erasing', section_text(16))
assert re.search(r'MUST NOT describe password rewrap as a complete security reset or as recovery from suspected K_root compromise', section_text(8))
scopes = section_text(20)
for scope in ('Core protocol', 'Store/writer', 'Application'):
    assert f'| {scope} |' in scopes
assert re.search(r'\| Core protocol \|[^|]*metadata preservation', scopes)
for anchor in ('per-object metadata rule in Section 12',
               'Section 17 additionally governs operation-wide fold metadata',
               'MUST satisfy the applicable core requirements',
               'core library need not guarantee the behavior of a GUI',
               'portable corpus success alone does not establish application conformance',
               'No stronger atomic primitive is imposed by this scope'):
    assert anchor in scopes, anchor
print('PASS: Totipo v1/r18 structural checks')
