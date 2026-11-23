import os
from pathlib import Path
import re
import subprocess
import unittest


class SecretVersionsTests(unittest.TestCase):
    def test_all_bound_secret_versions_are_validated_before_deployment(self):
        text = (Path(__file__).resolve().parents[2] / "cloudbuild.yaml").read_text()
        script = text.split("      - |\n", 1)[1].split("\n  - id:", 1)[0].replace("$$", "$")
        env = {**os.environ, "LINE_MESSAGING_API_CHANNEL_ID": "1", "LINE_MINI_APP_ENV": "developing",
               "LINE_MINI_APP_DEVELOPING_CHANNEL_ID": "2", "LINE_MINI_APP_DEVELOPING_LIFF_ID": "2-app", "LINE_MINI_APP_DEVELOPING_ENDPOINT_URL": "https://example.com"}
        keys = ("SECRET_VERSION", "TOKEN_VERSION", "MINI_VERSION", "GEMINI_VERSION", "OPENAI_VERSION", "CWA_VERSION")
        env.update(dict.fromkeys(keys, "1"))
        self.assertEqual(subprocess.run(["bash", "-ceu", script], env=env, capture_output=True).returncode, 0)
        for key in keys:
            for value in ("", "latest", "0"):
                with self.subTest(key=key, value=value):
                    result = subprocess.run(["bash", "-ceu", script], env={**env, key: value}, capture_output=True)
                    self.assertNotEqual(result.returncode, 0)
        self.assertNotRegex(text, r"_\w+_VERSION:\s*['\"]latest")
