"""Independent stdlib checks of declared r16 wire fields and key intermediates.

AES/Argon2 are exercised by Go. These checks don't claim independent AEAD evidence.
"""
import base64
import hashlib
import hmac
import struct
import unittest
from pathlib import Path
from check_vectors import read

BASE = Path(__file__).resolve().parent.parent / 'vectors'

def mac(k, p):
    return hmac.new(k, p, hashlib.sha256).digest()

def expand(k, info):
    return mac(k, info + b'\x01')

class IndependentWireTests(unittest.TestCase):
    def test_fields_and_key_intermediates(self):
        for entry in read(BASE / 'manifest.json')['cases']:
            c = read(BASE / entry['path'])
            if 'crypto' not in c:
                continue
            with self.subTest(case=c['id']):
                p = bytes.fromhex(c['semantic_hex'])
                root = bytes.fromhex(c['root_hex'])
                prk = mac(bytes(32), root)
                kid = expand(prk, b'totipo/v1/object-id')
                kroot = expand(prk, b'totipo/v1/object-key-root')
                oid = mac(kid, p)
                x = c['crypto']
                self.assertEqual(x['id_key_hex'], kid.hex())
                self.assertEqual(x['object_root_key_hex'], kroot.hex())
                self.assertEqual(x['object_id'], oid.hex())
                self.assertEqual(x['object_key_hex'], expand(kroot, b'totipo/v1/object-key' + oid).hex())
                self.assertEqual(x['nonce_hex'], oid[:12].hex())
                self.assertEqual(x['aad_hex'], (b'totipo/v1/object' + oid).hex())
                padded = struct.pack('>H', len(p)) + p + bytes(1006-len(p))
                self.assertEqual(x['padded_plaintext_hex'], padded.hex())
                self.assertEqual(len(bytes.fromhex(x['object_hex'])), 1024)
                if 'input' not in c:
                    continue
                fields = []
                offset = 0
                while offset < len(p):
                    tag, n = struct.unpack_from('>HH', p, offset)
                    fields.append((tag, p[offset+4:offset+4+n]))
                    offset += 4+n
                self.assertEqual(offset, len(p))
                v = c['input']
                binary = lambda s: base64.b64decode(s, validate=True)
                parents = [binary(s) for s in v['parents']]
                self.assertLessEqual(len(parents), 4)
                expected = [(1, binary(v['identity'])), (2, struct.pack('>H', len(parents)))]
                expected += [(3, parent) for parent in parents]
                expected += [(4, bytes([v['status']])), (5, v['issuer'].encode()),
                             (6, v['account'].encode()), (7, bytes([v['algorithm']])),
                             (8, bytes([v['digits']])), (9, struct.pack('>I', v['period'])),
                             (10, binary(v['secret']))]
                if 'client_name' in v:
                    expected.append((11, v['client_name'].encode()))
                if 'client_time' in v:
                    expected.append((12, struct.pack('>Q', v['client_time'])))
                self.assertEqual(fields, expected)

    def test_fingerprint_and_maximum(self):
        seen = {}
        for entry in read(BASE / 'manifest.json')['cases']:
            c = read(BASE / entry['path'])
            if 'bootstrap' in c:
                fp = mac(bytes.fromhex(c['root_hex']), b'totipo/v1/vault-fingerprint').hex()
                self.assertEqual(fp, c['bootstrap']['fingerprint_hex'])
                if c['root_hex'] in seen:
                    self.assertEqual(fp, seen[c['root_hex']])
                seen[c['root_hex']] = fp
        c = read(BASE / 'cases/size/v1.size.token-max-4.001.json')
        self.assertEqual(c['crypto']['semantic_length'], 1005)
        self.assertEqual(bytes.fromhex(c['crypto']['padded_plaintext_hex'])[-1], 0)

    def test_post_aead_fixture_inputs(self):
        for defect in ('nonzero-padding', 'object-id-mismatch', 'semantic-length-invalid'):
            c = read(BASE / f'cases/crypto/v1.crypto.{defect}.001.json')
            x = c['post_aead']
            p = bytes.fromhex(c['semantic_hex'])
            plain = bytes.fromhex(x['encryption_plaintext_hex'])
            self.assertEqual(len(plain), 1008)
            self.assertEqual(len(bytes.fromhex(x['object_hex'])), 1024)
            root = bytes.fromhex(c['root_hex'])
            kid = expand(mac(bytes(32), root), b'totipo/v1/object-id')
            correct_id = mac(kid, p).hex()
            n = struct.unpack_from('>H', plain)[0]
            canonical = struct.pack('>H', len(p)) + p + bytes(1006-len(p))
            if defect == 'nonzero-padding':
                self.assertEqual(x['object_id'], correct_id)
                self.assertEqual(n, len(p))
                self.assertEqual(plain[:2+n], canonical[:2+n])
                self.assertNotEqual(plain[2+n:], canonical[2+n:])
            elif defect == 'object-id-mismatch':
                self.assertNotEqual(x['object_id'], correct_id)
                self.assertEqual(plain, canonical)
            else:
                self.assertEqual(x['object_id'], correct_id)
                self.assertEqual(n, 1007)
                self.assertEqual(plain[2:], canonical[2:])
