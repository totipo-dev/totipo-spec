"""Exercise the checker CLI against isolated files, including editorial drift."""
import json
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest

BASE = Path(__file__).resolve().parent.parent
ARTIFACTS = {
    'spec_sha256': 'spec/totipo-vault-format-v1.md',
    'manifest_sha256': 'vectors/manifest.json',
    'schema_sha256': 'vectors/manifest.schema.json',
    'case_schema_sha256': 'vectors/case.schema.json',
}


class ProfileIntegrityTests(unittest.TestCase):
    def setUp(self):
        temp = tempfile.TemporaryDirectory()
        self.addCleanup(temp.cleanup)
        self.root = Path(temp.name)
        for name in [*ARTIFACTS.values(), 'requirements/v1-pre-rc.json',
                     'review/V1_PRE_R16_REVISION_HISTORY.md']:
            target = self.root / name
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(BASE / name, target)

    def check(self):
        return subprocess.run(
            [sys.executable, str(BASE / 'tools/check_spec.py')],
            cwd=self.root, capture_output=True, text=True)

    def test_current_artifacts_pass(self):
        result = self.check()
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_actual_byte_changes_fail(self):
        for field, name in ARTIFACTS.items():
            with self.subTest(artifact=name):
                path = self.root / name
                original = path.read_bytes()
                # A trailing newline preserves JSON semantics and spec anchors.
                path.write_bytes(original + b'\n')
                result = self.check()
                self.assertNotEqual(result.returncode, 0)
                self.assertIn(field + ' mismatch', result.stderr)
                path.write_bytes(original)

    def test_stale_and_missing_pins_fail(self):
        path = self.root / 'requirements/v1-pre-rc.json'
        original = json.loads(path.read_bytes())
        for field in ARTIFACTS:
            for value in ('0' * 64, None):
                with self.subTest(field=field, value=value):
                    profile = original.copy()
                    if value is None:
                        del profile[field]
                    else:
                        profile[field] = value
                    path.write_text(json.dumps(profile), encoding='utf-8')
                    result = self.check()
                    self.assertNotEqual(result.returncode, 0)
                    self.assertIn(field + ' mismatch', result.stderr)


if __name__ == '__main__':
    unittest.main()
