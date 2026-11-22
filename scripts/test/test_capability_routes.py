import runpy
import tempfile
import unittest
from pathlib import Path


check_routes = runpy.run_path(
    str(Path(__file__).resolve().parents[1] / "check/check-config.py")
)["check_capability_routes"]


class CapabilityRoutesTests(unittest.TestCase):
    def test_guide_must_be_reachable_from_selection_table(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            index = root / "index.md"
            guides = [root / "planning.md", root / "explorer.md"]
            for guide in guides:
                guide.touch()
            table = "## 依任務選擇能力\n| 規劃 | [規劃](planning.md) |\n| 探索 | [探索](explorer.md) |\n"
            other_section = "\n## 相關資料\n| 參考 | [探索](explorer.md) |\n"
            index.write_text(table.replace("(explorer.md)", "(planning.md)") + other_section)
            with self.assertRaisesRegex(ValueError, "explorer[.]md"):
                check_routes(index, guides)
            index.write_text(table + other_section)
            check_routes(index, guides)
