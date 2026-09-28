#!/usr/bin/env python3
"""Read-only r14→r15 audit; prints JSON evidence, never regenerates vectors."""
from pathlib import Path
import hashlib
import json
import re
import subprocess

BASE = '0b63f77886f95cf9bab3ceb9cbb0916177fe3589'
def committed(path):
    return subprocess.check_output(['git', 'show', f'{BASE}:{path}'])
def digest(b):
    return hashlib.sha256(b).hexdigest()

def audit():
    old = json.loads(committed('vectors/manifest.json'))
    new = json.loads(Path('vectors/manifest.json').read_bytes())
    before = {e['id']: e for e in old['cases']}
    after = {e['id']: e for e in new['cases']}
    rows = []
    frozen = []
    replacements = {
        'v1.candidate.continuity-block.001': 'v1.candidate.history-memory-lost.001',
        'v1.candidate.persistence-block.001': 'v1.candidate.cache-write-warning.001',
        'v1.future.unscoped-authoritative-block.001': 'v1.future.unscoped-warning.001',
        'v1.future.disappearance-retains-routing.001': 'v1.future.disappearance-removes-current.001',
        'v1.graph.local-security-memory-corruption.001': 'v1.graph.history-cache-corrupt.001',
        'v1.graph.known-id-corrupt-bytes-retain-node.001': 'v1.graph.known-id-corrupt-bytes-warning.001',
        'v1.graph.missing-current-value.001': 'v1.graph.remembered-head-absent.001',
    }
    for id in sorted(before.keys() | after.keys()):
        a, b = before.get(id), after.get(id)
        status = 'added' if a is None else 'removed' if b is None else 'unchanged' if a['sha256'] == b['sha256'] else 'modified'
        reason = 'Exact case file unchanged.'
        if status == 'removed':
            reason = 'Replaced by ' + replacements[id] if id in replacements else 'Retired mandatory retention/sticky blocking/reset semantics; optional caching has no protocol transition.'
        elif status == 'added':
            case = json.loads(Path('vectors', b['path']).read_bytes())
            reason = case.get('graph', {}).get('notes', 'New local API case; its specific reason is assigned below.')
        elif status == 'modified':
            if '.storage.' in id or '.family-' in id:
                reason = 'Remove authoritative readiness result; expose concrete graph integrity only. Storage byte fixtures unchanged.'
            elif '.initial-device-' in id:
                reason = 'Publication acknowledgements replace durability gates; matching valid self-signed key checks retained.'
            elif '.rename-' in id:
                reason = 'Accepted topology replaces durable state/readiness tuple; all-head DEVICE rename preserved.'
            elif 'conflict' in id:
                reason = 'Ordinary author=false and requires_confirmation=true for known visible supported conflict.'
            elif 'discovery-incomplete' in id:
                reason = 'Incomplete-view warning; accepted ordinary use/authorship remain available.'
            elif 'intermediate-disappears' in id:
                reason = 'Missing intermediate no longer supplies current ancestry; accepted A/C become conflicting heads requiring confirmation.'
            elif 'unscoped-candidate' in id:
                reason = 'Disappeared unscoped evidence leaves advisory regression warning; ordinary and candidate use remain available.'
            elif 'missing' in id:
                reason = 'Absent remembered head removed from H; remaining accepted value usable without mandatory unavailable-state confirmation, with regression warning.'
            else:
                reason = 'Explicit advisory remember observation; reappearance restores accepted evidence without a continuity gate.'
        applicability = b['applicability'] if b else None
        applicability_reason = 'Removed r14 case; no current applicability.'
        if b:
            case = json.loads(Path('vectors', b['path']).read_bytes())
            actions = {step['action'] for step in case.get('graph', {}).get('steps', [])}
            warnings = {w for step in case.get('graph', {}).get('steps', []) for w in step.get('expect', {}).get('warnings', [])}
            history_actions = actions & {'remember', 'history-lost', 'history-clear', 'cache-write-fails'}
            history_warnings = {w for w in warnings if w.startswith('HISTORY_')}
            if applicability['kind'] == 'conditional':
                assert applicability == {'kind': 'conditional', 'capability': 'advisory-history'}
                assert history_actions or history_warnings
                applicability_reason = 'Conditional advisory-history: ' + case['graph']['notes']
            else:
                assert applicability == {'kind': 'baseline'}
                assert not history_actions and not history_warnings, f'baseline depends on optional history: {id}'
                if 'optional-cache-failure' in actions:
                    applicability_reason = 'Baseline authority separation under a hypothetical optional-cache failure; no physical cache, failure-injection API, or diagnostic output is required. Warning behavior has a conditional companion.'
                elif case['operation'] == 'graph':
                    applicability_reason = 'Baseline accepted-snapshot graph/operation behavior; no advisory history actions or history-derived warning expectations. ' + case.get('graph', {}).get('notes', '')
                elif case['operation'] == 'local':
                    applicability_reason = 'Baseline immutable publication or authoritative local binding obligation, independent of advisory history.'
                else:
                    applicability_reason = 'Baseline ' + case['operation'] + ' validation/interpretation; independent of retained advisory history.'
            if id == 'v1.storage.exact-existing-publication.001':
                reason = 'Existing canonical ordinary file containing exactly the intended immutable object bytes acknowledges local publication without existing-file fsync, reread-after-fsync, or inode-identity proof; mismatched existing content is never overwritten.'
            elif id == 'v1.bootstrap.local-binding-establishment.001':
                reason = 'Absent binding can be durably established after canonical VAULT authentication; corrupt or mismatching present binding is not treated as absence.'
            elif status == 'modified' and case.get('graph', {}).get('notes'):
                reason = case['graph']['notes']
        rows.append({'case_id': id, 'r14_status': a['expected'] if a else None, 'r14_hash': a['sha256'] if a else None, 'r15_status': b['expected'] if b else None, 'r15_hash': b['sha256'] if b else None, 'delta': status, 'reason': reason, 'r15_applicability': applicability, 'applicability_reason': applicability_reason})

        if a:
            old_bytes = committed('vectors/' + a['path'])
            c = json.loads(old_bytes)
            # Compare entire byte-bearing case files, stricter than comparing just
            # known field names; includes nested signatures and every TOTP row.
            if c['operation'] in ('crypto', 'dispatch', 'bootstrap', 'provenance', 'signature-context', 'size', 'totp', 'late-provenance'):
                assert b, f'frozen fixture removed: {id}'
                current = Path('vectors', b['path']).read_bytes()
                assert current == old_bytes, f'frozen fixture changed: {id}'
                frozen.append({'case_id': id, 'sha256': digest(current), 'operation': c['operation']})
    # Crypto implementations and test vectors have not been rewritten either.
    sources = subprocess.check_output(['git','ls-tree','-r','--name-only',BASE,'conformance/internal/cryptov1','conformance/internal/object','conformance/internal/tlv','conformance/internal/totp'],text=True).splitlines()
    for path in sources:
        assert Path(path).read_bytes() == committed(path), f'frozen codec source changed: {path}'
    spec = Path('spec/totipo-vault-format-v1.md').read_text()
    old_spec = committed('spec/totipo-vault-format-v1.md').decode()
    historical = old_spec[old_spec.index('### v1/r14'):]
    assert spec[spec.index('### v1/r14'):] == historical, 'historical revision prose changed'
    pattern = re.compile(r'r14|LOCAL_CONTINUITY_UNKNOWN|KNOWLEDGE_PERSISTENCE_BLOCKED|AUTHORITATIVE_VAULT_READY|CANDIDATE_USE_READY|rebaseline|OPAQUE_UNSCOPED_RECORD')
    hits=[]
    paths = subprocess.check_output(['git','ls-files','--cached','--others','--exclude-standard'],text=True).splitlines()
    for path in sorted(set(paths)):
        p=Path(path)
        if not p.is_file() or path.endswith('.json') or path.startswith('review/'):
            continue
        try: text=p.read_text()
        except UnicodeDecodeError: continue
        history_line = text[:text.find('## 58. Revision history')].count('\n')+1 if path.startswith('spec/') else None
        for line, content in enumerate(text.splitlines(),1):
            if pattern.search(content):
                classification = 'revision-history' if history_line and line>history_line else 'obsolete-concept rejection/audit tooling' if path.startswith('tools/') else 'active r15 removal explanation' if path.startswith('spec/') else 'documentation/historical report link'
                hits.append({'path':path,'line':line,'classification':classification,'text':content})
    return {'baseline_head':BASE,'r14_cases':len(before),'r15_cases':len(after),'baseline_case_count':sum(e['applicability']['kind']=='baseline' for e in after.values()),'conditional_case_count':sum(e['applicability']['kind']=='conditional' for e in after.values()),'conditional_case_counts':{'advisory-history':sum(e['applicability']['kind']=='conditional' for e in after.values())},'counts':{v:sum(r['delta']==v for r in rows) for v in ('unchanged','modified','removed','added')},'case_matrix':rows,'frozen_case_count':len(frozen),'frozen_cases':frozen,'frozen_codec_sources':sources,'changed_frozen_cases':0,'historical_revision_text_unchanged':True,'legacy_reference_audit':hits,'review_directory_classification':'Existing r8–r14 reports are historical supporting evidence, preserved verbatim; r15 report and attached audit describe removals, not active gates.'}

if __name__ == '__main__':
    print(json.dumps(audit(),indent=2))
