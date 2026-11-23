import copy
import json
from pathlib import Path
import runpy
import unittest
from unittest.mock import patch
from urllib.error import HTTPError

ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / 'scripts/db/ensure-firestore.py'
module = runpy.run_path(str(SCRIPT))
ensure_indexes = module['ensure_indexes']
ensure_field_policies = module['ensure_field_policies']
MANIFEST = json.loads((ROOT / 'firestore.indexes.json').read_text())
INDEXES = MANIFEST['indexes']


class FirestoreIndexesTests(unittest.TestCase):
    def test_main_waits_for_field_policies_before_indexes(self):
        main = module['main']
        calls = []

        def check_policies(project, database, policies):
            calls.append(('policies', project, database, policies))

        def check_indexes(project, database, indexes):
            calls.append(('indexes', project, database, indexes))

        with patch.dict(main.__globals__, ensure_field_policies=check_policies, ensure_indexes=check_indexes):
            main('test', 'db')
            self.assertEqual(calls, [
                ('policies', 'test', 'db', MANIFEST['fieldOverrides']),
                ('indexes', 'test', 'db', INDEXES),
            ])
            calls.clear()
            main('test', 'db', indexes_only=True)
            self.assertEqual(calls, [('indexes', 'test', 'db', INDEXES)])

        def fail_policies(*args):
            raise TimeoutError("Field policy not ready")

        with patch.dict(main.__globals__, ensure_field_policies=fail_policies, ensure_indexes=check_indexes):
            # A failed field policy must stop the pipeline before index creation.
            calls.clear()
            with self.assertRaises(TimeoutError):
                main('test', 'db')
            self.assertEqual(calls, [])

    def test_creates_all_groups_and_waits_for_ready(self):
        stored = {}
        created = []
        sleeps = []

        def call(url, body=None):
            base = url.split('?')[0]
            if body is not None:
                created.append((base, body))
                stored.setdefault(base, []).append(dict(copy.deepcopy(body), state='CREATING'))
                return {}
            return {'indexes': copy.deepcopy(stored.get(base, []))}

        def sleep(seconds):
            sleeps.append(seconds)
            for indexes in stored.values():
                for index in indexes:
                    index['state'] = 'READY'

        ensure_indexes('test', 'db', INDEXES, call=call, sleep=sleep)
        self.assertEqual(len(created), len(INDEXES))
        self.assertEqual({url.split('/')[-2] for url, _ in created}, {i['collectionGroup'] for i in INDEXES})
        self.assertEqual(sleeps, [10])
        # A rerun only reads; existing and extra indexes are preserved.
        ensure_indexes('test', 'db', INDEXES, call=call, sleep=lambda _: self.fail('already ready'))
        self.assertEqual(len(created), len(INDEXES))

    def test_lists_all_pages_before_deciding_to_create(self):
        index = INDEXES[0]
        calls = []

        def call(url, body=None):
            calls.append(url)
            self.assertIsNone(body)
            if 'pageToken=second' not in url:
                return {'indexes': [], 'nextPageToken': 'second'}
            return {'indexes': [dict(index, state='READY')]}

        ensure_indexes('test', 'db', [index], call=call)
        self.assertEqual(len(calls), 2)

    def test_create_conflict_waits_for_readiness(self):
        index = INDEXES[0]
        reads = 0
        sleeps = []

        def call(url, body=None):
            nonlocal reads
            if body is not None:
                raise HTTPError(url, 409, 'Already exists', {}, None)
            reads += 1
            return {'indexes': [] if reads == 1 else [dict(index, state='READY')]}

        ensure_indexes('test', 'db', [index], call=call, sleep=sleeps.append)
        self.assertEqual(sleeps, [10])

    def test_repair_permission_errors_and_timeout_block_release(self):
        index = INDEXES[0]
        with self.assertRaisesRegex(RuntimeError, 'needs repair'):
            ensure_indexes('test', 'db', [index], call=lambda *a: {'indexes': [dict(index, state='NEEDS_REPAIR')]})
        with self.assertRaises(HTTPError):
            def denied(url, body=None):
                raise HTTPError(url, 403, 'Forbidden', {}, None)
            ensure_indexes('test', 'db', [index], call=denied)
        ticks = iter([0, 0, 1801])
        with self.assertRaises(TimeoutError):
            ensure_indexes('test', 'db', [index], call=lambda *a: {'indexes': [dict(index, state='CREATING')]},
                   sleep=lambda _: None, clock=lambda: next(ticks))


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
        def call(url, body=None, *, method=None):
            calls.append((url, body, method))
            return {} if body is not None else next(states)
        policy = {"collectionGroup": "summaryMessages", "fieldPath": "ExpireAt", "ttl": True, "indexes": []}
        ensure_field_policies("test", "(default)", [policy], call=call, sleep=lambda _: None)
        self.assertEqual(calls[1][1], {"indexConfig": {"indexes": []}, "ttlConfig": {}})
        self.assertIn("updateMask=", calls[1][0])
        self.assertEqual(calls[1][2], "PATCH")
        self.assertEqual(len(calls), 4)

    def test_timeout_and_existing_ready_policy(self):
        policy = {"collectionGroup": "summaryJobs", "fieldPath": "ExpireAt", "ttl": True, "indexes": []}
        ensure_field_policies("test", "db", [policy], call=lambda url, body=None: {"ttlConfig": {"state": "ACTIVE"}, "indexConfig": {}})
        ticks = iter([0, 1801])
        with self.assertRaises(TimeoutError):
            ensure_field_policies("test", "db", [policy], call=lambda *args, **kwargs: {}, clock=lambda: next(ticks), sleep=lambda _: None)
