import contextlib
import importlib.util
import io
from pathlib import Path
import unittest
from unittest.mock import patch
from urllib.error import HTTPError

spec = importlib.util.spec_from_file_location("doctor", Path(__file__).resolve().parents[1] / "scripts/mcp-doctor.py")
doctor = importlib.util.module_from_spec(spec)
spec.loader.exec_module(doctor)


class HostOAuthTests(unittest.TestCase):
    def check(self, configs, **kwargs):
        output = io.StringIO()
        with contextlib.redirect_stdout(output):
            code = doctor.check_servers(configs, **kwargs)
        return code, output.getvalue()

    def test_optional_host_oauth_prints_safe_endpoint_without_probing(self):
        with patch.object(doctor, "probe") as probe:
            code, output = self.check({"fly-mcp": {"url": "https://user:secret@mcp.fly.io/path?token=secret#private"}})
        probe.assert_not_called()
        self.assertEqual(code, 0)
        self.assertIn("HOST_REQUIRED=1", output)
        self.assertIn("https://mcp.fly.io/path", output)
        self.assertNotIn("secret", output)
        self.assertNotIn("private", output)
        self.assertIn("codex mcp login fly-mcp --no-browser", output)

    def test_required_host_oauth_is_not_ready(self):
        code, output = self.check({"fly-mcp": {"url": "https://mcp.fly.io", "required": True}})
        self.assertEqual(code, 1)
        self.assertIn("PASS=0", output)

    def test_unrelated_401_is_not_hidden(self):
        error = HTTPError("https://example.test", 401, "Unauthorized", {}, None)
        with patch.object(doctor, "probe", side_effect=error):
            code, output = self.check({"other": {"url": "https://example.test"}})
        self.assertEqual(code, 1)
        self.assertIn("FAIL=1", output)
        self.assertIn("HOST_REQUIRED=0", output)

    def test_explicit_direct_probe_preserves_failure(self):
        with patch.object(doctor, "probe", side_effect=RuntimeError("unauthorized")) as probe:
            code, output = self.check({"fly-mcp": {"url": "https://mcp.fly.io"}}, direct=True)
        probe.assert_called_once()
        self.assertEqual(code, 1)
        self.assertIn("HOST_REQUIRED=0", output)


if __name__ == "__main__":
    unittest.main()
