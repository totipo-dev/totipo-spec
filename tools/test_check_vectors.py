import copy
import json
import unittest
from pathlib import Path
from check_vectors import Invalid, read, validate, pairs

BASE = Path(__file__).resolve().parent.parent / 'vectors'

class SchemaContractTests(unittest.TestCase):
    def setUp(self):
        self.schema = read(BASE / 'case.schema.json')
        self.case = read(BASE / 'cases/graph/v1.graph.sequential.001.json')

    def test_required_false_result_cannot_be_omitted(self):
        del self.case['graph']['steps'][-1]['expect']['conflicting']
        with self.assertRaises(Invalid):
            validate(self.case, self.schema)

    def test_head_records_required_and_exact_metadata_types(self):
        c = read(BASE / 'cases/graph/v1.graph.equal-concurrent.001.json')
        validate(c, self.schema)
        head = c['graph']['steps'][-1]['expect']['head_objects'][0]
        head['client_time'] = 2**64
        with self.assertRaises(Invalid):
            validate(c, self.schema)
        del c['graph']['steps'][-1]['expect']['head_objects']
        with self.assertRaises(Invalid):
            validate(c, self.schema)

    def test_unsigned_timestamp_exactness(self):
        c = read(BASE / 'cases/metadata/v1.metadata.client-time-u64max.001.json')
        validate(c, self.schema)
        for bad in (-1, 18446744073709551616, 1.5, True):
            c['input']['client_time'] = bad
            with self.assertRaises(Invalid):
                validate(c, self.schema)

    def test_nested_unknown_and_mixed_payloads(self):
        self.case['graph']['steps'][-1]['expect']['typo'] = False
        with self.assertRaises(Invalid):
            validate(self.case, self.schema)
        c = read(BASE / 'cases/graph/v1.graph.sequential.001.json')
        c['root_hex'] = '00' * 32
        with self.assertRaises(Invalid):
            validate(c, self.schema)

    def test_manifest_revision_and_retired_extension(self):
        schema = read(BASE / 'manifest.schema.json')
        m = read(BASE / 'manifest.json')
        validate(m, schema)
        m['cases'][0]['applicability'] = {'kind': 'baseline'}
        with self.assertRaises(Invalid):
            validate(m, schema)
        del m['cases'][0]['applicability']
        m['spec_revision'] = 'r99'
        with self.assertRaises(Invalid):
            validate(m, schema)

    def test_post_aead_shape(self):
        c = read(BASE / 'cases/crypto/v1.crypto.nonzero-padding.001.json')
        validate(c, self.schema)
        c['post_aead']['defect'] = 'failed-aead'
        with self.assertRaises(Invalid):
            validate(c, self.schema)
        c['post_aead']['defect'] = 'nonzero-padding'
        c['post_aead']['object_hex'] = c['post_aead']['object_hex'][:-2]
        with self.assertRaises(Invalid):
            validate(c, self.schema)

    def test_duplicate_members(self):
        with self.assertRaises(Invalid):
            json.loads('{"id": 1, "id": 2}', object_pairs_hook=pairs)

if __name__ == '__main__':
    unittest.main()
