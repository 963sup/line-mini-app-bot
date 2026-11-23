import json
from pathlib import Path
import runpy
import unittest

verify = runpy.run_path(str(Path(__file__).resolve().parents[1] / "check/check-go-test-results.py"))["verify"]


def events(*rows):
    return [json.dumps(row) for row in rows]


class GoTestResultTests(unittest.TestCase):
    def test_complete_packages_and_no_test_package(self):
        self.assertEqual(verify(events(
            {"Package": "app", "Test": "TestLive", "Action": "pass"},
            {"Package": "app", "Action": "pass"},
            {"Package": "values", "Action": "skip"}), ["app", "values"]), (2, 1))

    def test_skipped_case_cannot_hide_in_successful_package(self):
        with self.assertRaisesRegex(ValueError, "skip"):
            verify(events({"Package": "app", "Test": "TestStorage", "Action": "skip"},
                          {"Package": "app", "Test": "TestUnit", "Action": "pass"},
                          {"Package": "app", "Action": "pass"}), ["app"])

    def test_package_missing_failed_or_no_tests(self):
        for report in (events({"Package": "other", "Test": "TestUnit", "Action": "pass"},
                              {"Package": "other", "Action": "pass"}),
                       events({"Package": "app", "Action": "fail"}),
                       events({"Package": "app", "Action": "skip"})):
            with self.subTest(report=report), self.assertRaises(ValueError):
                verify(report, ["app"])

    def test_malformed_event_fails(self):
        for report in (["not JSON"], ["{}"]):
            with self.assertRaises(ValueError):
                verify(report, ["app"])
