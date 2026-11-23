import runpy
import subprocess
import tempfile
import unittest
from pathlib import Path


check_links = runpy.run_path(
    str(Path(__file__).resolve().parents[1] / "check/check-config.py")
)["check_documentation_links"]


class DocumentationFixture:
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = Path(self.directory.name)
        subprocess.run(["git", "init", "-q", str(self.root)], check=True)

    def write(self, name, content):
        path = self.root / name
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content)
        return path

class DocumentationLinksTests(DocumentationFixture, unittest.TestCase):
    def test_nested_guidance_and_untracked_files_are_checked(self):
        for name in ("api/AGENTS.md", "backend/README.md", "frontend/README.md"):
            with self.subTest(name=name):
                path = self.write(name, "# Guide\n[Missing](missing.md)\n")
                with self.assertRaisesRegex(ValueError, name + ":2"):
                    check_links(self.root)
                path.write_text("# Guide\n")

    def test_fragments_local_and_cross_file(self):
        self.write("README.md", '# Root\n[Self](#root)\n[Other](docs/topic.md#重複-1)\n[Named](docs/topic.md#stable)\n')
        target = self.write("docs/topic.md", '# 重複\n## 重複\n<a id="stable"></a>\n')
        self.assertEqual(check_links(self.root), 2)
        target.write_text('# Renamed\n<a id="stable"></a>\n')
        with self.assertRaisesRegex(ValueError, "README.md:3.*重複-1"):
            check_links(self.root)

    def test_examples_upstream_ignored_and_directory_links(self):
        self.write("README.md", '# Root\n[Directory](backend)\n[Web](https://example.com/missing)\n'
                   '```md\n[Example](missing.md)\n```\n`[Inline](missing.md)`\n'
                   '~~~~md\n```\n[Example](missing.md)\n~~~~\n')
        self.write("backend/README.md", "# Backend\n")
        self.write(".gitignore", "ignored/\n")
        for name in ("ignored/README.md", ".agents/skills/upstream/SKILL.md",
                     ".claude/skills/upstream/SKILL.md",
                     "docs/development/tooling/frontend/shadcn-reference/snapshot/pages/upstream.md"):
            self.write(name, "[Upstream](missing.md)\n")
        self.assertEqual(check_links(self.root), 2)

    def test_escaped_paths_and_explicit_anchor(self):
        self.write("README.md", '# Root\n[Target](<docs/topic name.md#explicit%20anchor>)\n')
        self.write("docs/topic name.md", '# Topic\n<a id="explicit anchor"></a>\n')
        self.assertEqual(check_links(self.root), 2)


class ExtendedDocumentationLinksTests(DocumentationFixture, unittest.TestCase):
    def test_reference_html_and_mermaid_targets(self):
        self.write("docs/topic.md", '# Target\n<a id="stable"></a>\n')
        examples = {
            "reference": '[Topic][ref]\n[ref]: docs/topic.md#stable "Title"\n',
            "html": '<a\n href="docs/topic.md#stable">Topic</a>\n',
            "diagram": '```mermaid\nflowchart LR\n a["A"]\n click a "docs/topic.md#stable"\n```\n',
        }
        for name, content in examples.items():
            with self.subTest(name=name):
                self.write("README.md", '# Root\n' + content)
                self.assertEqual(check_links(self.root), 2)
                self.write("README.md", '# Root\n' + content.replace("#stable", "#missing"))
                with self.assertRaisesRegex(ValueError, "Broken documentation anchor"):
                    check_links(self.root)
                self.write("README.md", '# Root\n' + content.replace("docs/topic.md#stable", "missing.md"))
                with self.assertRaisesRegex(ValueError, "Broken documentation link"):
                    check_links(self.root)

    def test_standalone_mermaid_and_literal_examples(self):
        self.write("README.md", '# Root\n```html\n<a href="missing.md">Example</a>\n```\n'
                   '~~~text\n[ref]: missing.md\n~~~\n`<a href="missing.md">literal</a>`\n')
        diagram = self.write("flow.mermaid", 'flowchart LR\n a["A"]\n click a href "README.md#root"\n')
        self.assertEqual(check_links(self.root), 2)
        diagram.write_text('flowchart LR\n a["A"]\n click a "missing.md"\n')
        with self.assertRaisesRegex(ValueError, "flow.mermaid:3"):
            check_links(self.root)


if __name__ == "__main__":
    unittest.main()
