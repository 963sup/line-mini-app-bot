import runpy
import tempfile
from pathlib import Path
import unittest

check_sources = runpy.run_path(str(Path(__file__).resolve().parents[1] / "check/check-config.py"))["check_cloud_build_sources"]

class CloudBuildSourceTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = Path(self.directory.name)
        (self.root / "scripts/db").mkdir(parents=True)
        (self.root / "cloudbuild.yaml").write_text('args: ["scripts/db/policy.py"]\n')
        (self.root / "scripts/db/policy.py").write_text('from pathlib import Path\nhelper = Path(__file__).with_name("indexes.py")\n')
        (self.root / "scripts/db/indexes.py").write_text('import json\n')
        self.rules = "**\n!scripts/\n!scripts/db/\n!scripts/db/policy.py\n!scripts/db/indexes.py\n"
        (self.root / ".gcloudignore").write_text(self.rules)

    def test_entry_and_sibling_dependency_are_uploaded(self):
        self.assertEqual(check_sources(self.root), 2)

    def test_each_required_file_missing_from_upload_fails(self):
        for file in ("policy.py", "indexes.py"):
            with self.subTest(file=file):
                (self.root / ".gcloudignore").write_text(self.rules.replace("!scripts/db/" + file + "\n", ""))
                with self.assertRaisesRegex(ValueError, "excluded.*" + file):
                    check_sources(self.root)

    def test_missing_dependency_fails(self):
        (self.root / "scripts/db/indexes.py").unlink()
        with self.assertRaisesRegex(ValueError, "Missing Cloud Build source"):
            check_sources(self.root)

    def test_unsupported_sdk_include_is_explicit(self):
        (self.root / ".gcloudignore").write_text("#!include:.gitignore\n")
        with self.assertRaisesRegex(ValueError, "include directives"):
            check_sources(self.root)
