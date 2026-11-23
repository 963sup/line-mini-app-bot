import runpy
import tempfile
import unittest
from pathlib import Path


check_routes = runpy.run_path(
    str(Path(__file__).resolve().parents[1] / "check/check-config.py")
)["check_capability_routes"]


class CapabilityRoutesTests(unittest.TestCase):
    def test_guide_must_be_reachable_from_selection_table(self):
        for heading in ("Select capabilities by task", "依任務選擇能力"):
            with self.subTest(heading=heading), tempfile.TemporaryDirectory() as directory:
                self.check_selection_table(Path(directory), heading)

    def check_selection_table(self, root, heading):
        index = root / "index.md"
        guides = [root / "planning.md", root / "explorer.md"]
        for guide in guides:
            guide.touch()
        table = f"## {heading}\n| Planning | [Planning](planning.md) |\n| Exploration | [Exploration](explorer.md) |\n"
        other_section = "\n## References\n| Reference | [Exploration](explorer.md) |\n"
        index.write_text(table.replace("(explorer.md)", "(planning.md)") + other_section)
        with self.assertRaisesRegex(ValueError, "explorer[.]md"):
            check_routes(index, guides)
        index.write_text(table + other_section)
        check_routes(index, guides)
