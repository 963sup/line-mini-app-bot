import copy
import json
from pathlib import Path
import runpy
import unittest
from unittest.mock import patch
from urllib.error import HTTPError

ROOT = Path(__file__).resolve().parents[2]
SCRIPT = ROOT / 'scripts/db/ensure-firestore-indexes.py'
module = runpy.run_path(str(SCRIPT))
ensure = module['ensure']
INDEXES = json.loads((ROOT / 'firestore.indexes.json').read_text())['indexes']


class FirestoreIndexesTests(unittest.TestCase):
    def test_main_passes_entire_manifest(self):
        main = module['main']
        with patch.dict(main.__globals__, ensure=lambda project, database, indexes: self.assertEqual(indexes, INDEXES)):
            main('test', 'db')

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

        ensure('test', 'db', INDEXES, call=call, sleep=sleep)
        self.assertEqual(len(created), len(INDEXES))
        self.assertEqual({url.split('/')[-2] for url, _ in created}, {i['collectionGroup'] for i in INDEXES})
        self.assertEqual(sleeps, [10])
        # A rerun only reads; existing and extra indexes are preserved.
        ensure('test', 'db', INDEXES, call=call, sleep=lambda _: self.fail('already ready'))
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

        ensure('test', 'db', [index], call=call)
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

        ensure('test', 'db', [index], call=call, sleep=sleeps.append)
        self.assertEqual(sleeps, [10])

    def test_repair_permission_errors_and_timeout_block_release(self):
        index = INDEXES[0]
        with self.assertRaisesRegex(RuntimeError, 'needs repair'):
            ensure('test', 'db', [index], call=lambda *a: {'indexes': [dict(index, state='NEEDS_REPAIR')]})
        with self.assertRaises(HTTPError):
            def denied(url, body=None):
                raise HTTPError(url, 403, 'Forbidden', {}, None)
            ensure('test', 'db', [index], call=denied)
        ticks = iter([0, 0, 1801])
        with self.assertRaises(TimeoutError):
            ensure('test', 'db', [index], call=lambda *a: {'indexes': [dict(index, state='CREATING')]},
                   sleep=lambda _: None, clock=lambda: next(ticks))
