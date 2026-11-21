"""Ordering must never let an old build proceed on incomplete API evidence."""
import importlib.util
from pathlib import Path
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('latest', Path(__file__).with_name('cloud-build-latest.py'))
latest = importlib.util.module_from_spec(spec)
spec.loader.exec_module(latest)


class BuildOrderingTests(unittest.TestCase):
    def setUp(self):
        self.current = {'id': 'old', 'buildTriggerId': 'trigger', 'createTime': '2026-11-21T12:00:00Z'}

    def test_newer_build_supersedes_even_if_failed(self):
        newer = dict(self.current, id='new', createTime='2026-11-21T12:00:00.100Z', status='FAILURE')
        self.assertTrue(latest.superseded(self.current, [newer]))

    def test_same_build_continues(self):
        self.assertFalse(latest.superseded(self.current, [self.current]))

    def test_older_build_does_not_supersede(self):
        older = dict(self.current, id='older', createTime='2026-11-21T11:59:59.999Z')
        self.assertFalse(latest.superseded(self.current, [older]))

    def test_lookup_missing_or_wrong_trigger_fails_closed(self):
        for builds in ([], [dict(self.current, buildTriggerId='other')], [dict(self.current, createTime='')]):
            with self.subTest(builds=builds), self.assertRaises(ValueError):
                latest.superseded(self.current, builds)


    def test_cancellation_targets_only_current_build(self):
        newer = dict(self.current, id='new', createTime='2026-11-21T12:00:01Z')
        with patch.object(latest, 'gcloud', side_effect=[self.current, [newer], {}]) as call, patch(
            'sys.argv', ['guard', '--project=p', '--region=global', '--build=old']
        ), self.assertRaises(SystemExit) as stopped:
            latest.main()
        self.assertEqual(stopped.exception.code, 1)
        self.assertEqual(call.call_args.args[-3:], ('builds', 'cancel', 'old'))

    def test_dry_run_never_cancels(self):
        newer = dict(self.current, id='new', createTime='2026-11-21T12:00:01Z')
        with patch.object(latest, 'gcloud', side_effect=[self.current, [newer]]) as call, patch(
            'sys.argv', ['guard', '--project=p', '--region=global', '--build=old', '--dry-run']
        ), self.assertRaises(SystemExit):
            latest.main()
        self.assertEqual(call.call_count, 2)


if __name__ == '__main__':
    unittest.main()
