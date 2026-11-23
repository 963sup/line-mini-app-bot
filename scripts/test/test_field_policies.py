import json
from pathlib import Path
import runpy
import unittest

ROOT = Path(__file__).resolve().parents[2]
ensure = runpy.run_path(str(ROOT / "scripts/db/ensure-field-policies.py"))["ensure"]


class FieldPoliciesTests(unittest.TestCase):
    def test_manifest_contains_summary_retention_and_index_exemptions(self):
        policies = json.loads((ROOT / "firestore.indexes.json").read_text())["fieldOverrides"]
        for group, text in (("summaryMessages", "Text"), ("summaryJobs", "Result")):
            self.assertIn({"collectionGroup": group, "fieldPath": "ExpireAt", "ttl": True, "indexes": []}, policies)
            self.assertIn({"collectionGroup": group, "fieldPath": text, "indexes": []}, policies)

    def test_waits_for_ttl_and_explicit_index_configuration(self):
        calls = []
        states = iter([{"indexConfig": {"usesAncestorConfig": True}},
                       {"ttlConfig": {"state": "CREATING"}, "indexConfig": {}},
                       {"ttlConfig": {"state": "ACTIVE"}, "indexConfig": {}}])
        def call(url, body=None):
            calls.append((url, body))
            return {} if body is not None else next(states)
        policy = {"collectionGroup": "summaryMessages", "fieldPath": "ExpireAt", "ttl": True, "indexes": []}
        ensure("test", "(default)", [policy], call=call, sleep=lambda _: None)
        self.assertEqual(calls[1][1], {"indexConfig": {"indexes": []}, "ttlConfig": {}})
        self.assertIn("updateMask=", calls[1][0])
        self.assertEqual(len(calls), 4)

    def test_timeout_and_existing_ready_policy(self):
        policy = {"collectionGroup": "summaryJobs", "fieldPath": "ExpireAt", "ttl": True, "indexes": []}
        ensure("test", "db", [policy], call=lambda url, body=None: {"ttlConfig": {"state": "ACTIVE"}, "indexConfig": {}})
        ticks = iter([0, 1801])
        with self.assertRaises(TimeoutError):
            ensure("test", "db", [policy], call=lambda *args: {}, clock=lambda: next(ticks), sleep=lambda _: None)
