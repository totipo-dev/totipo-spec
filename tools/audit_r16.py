#!/usr/bin/env python3
"""Read-only case migration evidence from committed r15 to the r16 worktree."""
import hashlib
import json
from pathlib import Path
import subprocess

BASE = '0a01fc1f0493ec1fefc0c4b89f8b22490d586852'

def committed(path):
    return subprocess.check_output(['git', 'show', f'{BASE}:{path}'])

def digest(b):
    return hashlib.sha256(b).hexdigest()

def audit():
    old = json.loads(committed('vectors/manifest.json'))
    new = json.loads(Path('vectors/manifest.json').read_bytes())
    after = {e['id']: e for e in new['cases']}
    replacements = {
        'routing.token-v1': ['encoding.token-root'],
        'encoding.author-time-zero': ['metadata.client-time-zero'],
        'encoding.author-time-u64max': ['metadata.client-time-u64max'],
        'timestamp.zero': ['metadata.client-time-zero'],
        'timestamp.normal': ['metadata.client-time-normal'],
        'timestamp.i64max': ['metadata.client-time-u64max'],
        'timestamp.u64max': ['metadata.client-time-u64max'],
        'timestamp.fold-common-time': ['fold.width-10'],
        'timestamp.equal-value-different-times': ['graph.equal-concurrent', 'metadata.client-time-normal'],
        'size.token-max-5-fold': ['encoding.five-parents', 'fold.width-5'],
        'graph.cycle-integrity-failure': ['graph.cycle-equal', 'graph.cycle-conflicting', 'graph.late-cycle'],
        'graph.global-object-id-conflict': ['graph.same-id-defensive'],
        'graph.reappearance-mismatch': ['graph.same-id-defensive'],
        'graph.corrupt-current-bytes-excluded': ['storage.failed-aead'],
        'storage.wrong-size-not-opaque': ['storage.wrong-size'],
        'snapshot.incomplete-usable': ['storage.incomplete-diagnostics'],
        'snapshot.late-arrival-before-publication': ['graph.late-parent', 'storage.missing-parent-publication'],
        'snapshot.hidden-history-after-publication': ['graph.conflicting-concurrent'],
        'publication.ambiguous-retry-sibling': ['storage.ambiguous-publication', 'storage.exact-existing-publication'],
    }
    replacements = {f'v1.{k}.001': [f'v1.{v}.001' for v in vs] for k, vs in replacements.items()}
    rows = []
    for e in old['cases']:
        cid = e['id']
        target = after.get(cid)
        if target:
            if target['sha256'] == e['sha256']:
                classification = 'retain but metadata/section refs change'
                reason = 'RFC known-answer file byte-identical; manifest section changes to 19.'
            elif e['category'] in ('crypto', 'encoding', 'size'):
                classification = 'regenerate for changed r16 wire bytes'
                reason = 'Flattened exact TOKEN grammar; deterministic crypto regenerated where applicable.'
            elif e['category'] == 'bootstrap':
                classification = 'replace with new r16 equivalent'
                reason = 'Same password/root/wrap bytes; replace binding field with fingerprint derivation.'
            else:
                classification = 'replace with new r16 equivalent'
                reason = 'Retain relevant behavior in simplified graph/storage schema; preserve per-head metadata and remove policy fields.'
            targets = [cid]
        elif cid in replacements:
            classification = 'replace with new r16 equivalent'
            reason = 'Retire old shape/expectations; replacement covers surviving r16 behavior.'
            targets = replacements[cid]
        else:
            classification = 'remove because concept is deleted'
            targets = []
            if any(x in cid for x in ('device', 'provenance', 'der-')):
                reason = 'Removed identity/signing grammar or signature-dependent capacity.'
            elif any(x in cid for x in ('future', 'opaque', 'routing.')):
                reason = 'Removed in-family semantic routing and compatibility state.'
            elif 'binding' in cid:
                reason = 'Mandatory local establishment removed; recognition has separate fingerprint fixtures.'
            elif any(x in cid for x in ('history', 'remembered', 'cache', 'known-id-corrupt')):
                reason = 'Protocol remembered-history/cache diagnostics removed.'
            else:
                reason = 'Protocol use/authorship permission or confirmation-generation machinery removed; descriptive rules remain normative.'
        for t in targets:
            assert t in after, t
        rows.append({'id': cid, 'old_sha256': e['sha256'], 'classification': classification,
                     'reason': reason, 'replacement_ids': targets,
                     'new_sha256': target['sha256'] if target else None})
    old_ids = {e['id'] for e in old['cases']}
    return {'baseline_commit': BASE, 'old_count': len(old['cases']), 'new_count': len(new['cases']),
            'baseline_hashes': {p: digest(committed(p)) for p in (
                'spec/totipo-vault-format-v1.md', 'vectors/manifest.json', 'requirements/v1-pre-rc.json')},
            'cases': rows, 'added': [e for e in new['cases'] if e['id'] not in old_ids]}

if __name__ == '__main__':
    print(json.dumps(audit(), indent=2))
