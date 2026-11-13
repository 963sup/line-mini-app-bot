"""Local subprocess regressions for the doctor's stdio response deadline."""
import importlib.util
import pathlib
import signal
import subprocess
import sys
import tempfile
import unittest
from unittest import mock


SOURCE = pathlib.Path(__file__).with_name('mcp-doctor.py')
spec = importlib.util.spec_from_file_location('mcp_doctor', SOURCE)
doctor = importlib.util.module_from_spec(spec)
spec.loader.exec_module(doctor)

SERVER = '''import json, sys, time
mode = sys.argv[1]
print('READY', flush=True)
for line in sys.stdin:
    message = json.loads(line)
    if 'id' not in message:
        continue
    if mode == 'exit':
        sys.exit(0)
    if mode == 'silent':
        time.sleep(30)
    if mode in ('notifications', 'wrong_ids'):
        while True:
            value = ({'jsonrpc': '2.0', 'method': 'notifications/message'}
                     if mode == 'notifications' else {'jsonrpc': '2.0', 'id': 999, 'result': {}})
            print(json.dumps(value), flush=True)
            time.sleep(.005)
    print(json.dumps({'jsonrpc': '2.0', 'method': 'notifications/message'}), flush=True)
    result = ({'protocolVersion': '2025-03-26'} if message['method'] == 'initialize'
              else {'tools': [{'name': 'read_only'}]})
    print(json.dumps({'jsonrpc': '2.0', 'id': message['id'], 'result': result}), flush=True)
'''


@unittest.skipUnless(hasattr(signal, 'setitimer'), 'requires POSIX subprocess timers')
class StdioDeadlineTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.server = pathlib.Path(self.directory.name) / 'server.py'
        self.server.write_text(SERVER)

    def probe(self, mode):
        children = []
        popen = subprocess.Popen

        def spawn(*args, **kwargs):
            child = popen(*args, **kwargs)
            children.append(child)
            # Keep host startup scheduling outside the request's short budget.
            self.assertEqual(child.stdout.readline().strip(), 'READY')
            return child

        def safety_limit(*_):
            raise AssertionError('doctor exceeded the outer test safety deadline')

        previous = signal.signal(signal.SIGALRM, safety_limit)
        signal.setitimer(signal.ITIMER_REAL, 5)
        try:
            with mock.patch.object(doctor.subprocess, 'Popen', side_effect=spawn):
                return doctor.probe_server(mode, {
                    'command': sys.executable,
                    'args': ['-u', str(self.server), mode],
                    'startup_timeout_sec': .3,
                })
        finally:
            signal.setitimer(signal.ITIMER_REAL, 0)
            signal.signal(signal.SIGALRM, previous)
            self.assertTrue(children, 'probe must launch its configured child')
            self.assertTrue(all(child.poll() is not None for child in children),
                            'every outcome must reap the server process')
            for child in children:
                child.stdin.close()
                child.stdout.close()

    def test_notifications_and_wrong_ids_cannot_extend_request_deadline(self):
        for mode in ('notifications', 'wrong_ids', 'silent'):
            with self.subTest(mode=mode):
                with self.assertRaisesRegex(RuntimeError, 'initialize: timed out'):
                    self.probe(mode)

    def test_matching_response_after_notification_completes_handshake(self):
        self.probe('normal')

    def test_early_exit_preserves_its_error(self):
        with self.assertRaisesRegex(RuntimeError, 'server exited before reply'):
            self.probe('exit')


if __name__ == '__main__':
    unittest.main()
