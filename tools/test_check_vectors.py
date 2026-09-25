import copy
import json
import unittest
from pathlib import Path

from check_vectors import Invalid, read, validate


class SchemaContractTests(unittest.TestCase):
    def setUp(self):
        base = Path(__file__).resolve().parent.parent / "vectors"
        self.schema = read(base / "case.schema.json")
        self.case = read(base / "cases/graph/v1.graph.sequential.001.json")

    def test_required_false_result_cannot_be_omitted(self):
        for field in ("candidate", "ordinary", "integrity_failure"):
            case = copy.deepcopy(self.case)
            del case["graph"]["steps"][-1]["expect"][field]
            with self.assertRaises(Invalid):
                validate(case, self.schema)

    def test_unsigned_timestamp_exactness(self):
        node = self.case["graph"]["steps"][0]["node"]
        node["author_time"] = 18446744073709551615
        validate(self.case, self.schema)
        for bad in (-1, 18446744073709551616, 1.5, True):
            node["author_time"] = bad
            with self.assertRaises(Invalid):
                validate(self.case, self.schema)

    def test_nested_unknown_field(self):
        self.case["graph"]["steps"][-1]["expect"]["ordinary_typo"] = False
        with self.assertRaises(Invalid):
            validate(self.case, self.schema)

    def test_duplicate_members(self):
        from check_vectors import pairs
        with self.assertRaises(Invalid):
            json.loads('{"id": 1, "id": 2}', object_pairs_hook=pairs)


if __name__ == "__main__":
    unittest.main()
