"""Local regressions for MCP request deadlines and transport cleanup."""
import importlib.util
import json
import os
import pathlib
import signal
import subprocess
import sys
import tempfile
import threading
import time
import unittest
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
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


HTTP_RUNNER = '''import importlib.util,json,subprocess,sys
from unittest import mock
spec=importlib.util.spec_from_file_location('doctor',sys.argv[1])
doctor=importlib.util.module_from_spec(spec);spec.loader.exec_module(doctor)
children=[];popen=subprocess.Popen
def spawn(*args,**kwargs):
    child=popen(*args,**kwargs);children.append(child);return child
result={}
try:
    with mock.patch.object(doctor.subprocess,'Popen',side_effect=spawn):
        doctor.probe_server('local',{'url':sys.argv[2],'startup_timeout_sec':float(sys.argv[3]),
            'env_http_headers':{'X-Test-Secret':'MCP_TEST_SECRET'}})
except RuntimeError as error:
    result['error']=str(error)
except KeyboardInterrupt:
    result['error']='KeyboardInterrupt'
result['reaped']=all(child.poll() is not None for child in children)
result['secret_in_argv']=any('sensitive-fixture' in str(child.args) for child in children)
print(json.dumps(result))
'''


@unittest.skipUnless(hasattr(os, 'killpg'), 'requires POSIX process cleanup')
class HttpDeadlineTests(unittest.TestCase):
    def probe(self, mode):
        requests = []
        disconnected = threading.Event()
        received = threading.Event()
        # Deadline cases must expire; healthy/error classification cases should
        # not accidentally become startup benchmarks on a loaded CI host.
        stalls = ('notifications', 'wrong_ids', 'partial_line', 'json_stall',
                  'headers', 'notification_headers')
        request_timeout = 5 if mode == 'interrupt' else 2 if mode in stalls else 3
        outer_deadline = time.monotonic() + 10

        class Handler(BaseHTTPRequestHandler):
            def log_message(self, *_):
                pass

            def do_POST(self):
                message = json.loads(self.rfile.read(int(self.headers['Content-Length'])))
                requests.append((message, dict(self.headers)))
                received.set()
                notification_stall = (mode == 'notification_headers' and
                                      message['method'] == 'notifications/initialized')
                try:
                    if mode in ('headers', 'interrupt') or notification_stall:
                        self.rfile.read(1)
                        return
                    self.send_response(202 if 'id' not in message else
                                       403 if mode == 'http_error' else 200)
                    self.send_header('Content-Type', 'text/event-stream' if mode in
                                     ('notifications', 'wrong_ids', 'partial_line', 'sse', 'sse_eof')
                                     else 'application/json')
                    if 'id' in message:
                        self.send_header('Mcp-Session-Id', 'local-session')
                    self.end_headers()
                    if mode in ('notifications', 'wrong_ids', 'partial_line'):
                        value = ({'jsonrpc': '2.0', 'method': 'notifications/message'}
                                 if mode == 'notifications' else {'id': 999, 'result': {}})
                        chunk = b'x' if mode == 'partial_line' else b'data: ' + json.dumps(value).encode() + b'\n\n'
                        while True:
                            self.wfile.write(chunk)
                            self.wfile.flush()
                            time.sleep(.005)
                    if mode == 'json_stall':
                        self.wfile.write(b'{"result":')
                        self.wfile.flush()
                        self.rfile.read(1)
                        return
                    if mode == 'http_error':
                        self.wfile.write(b'sensitive-fixture')
                        return
                    if mode == 'sse_eof':
                        self.wfile.write(b'data: {"method":"notifications/message"}\n\n'
                                         b'data: {"id":999,"result":{}}\n\n')
                        return
                    if 'id' not in message:
                        return
                    result = ({'protocolVersion': '2025-06-18'} if message['method'] == 'initialize'
                              else {'tools': [{'name': 'read_only'}]})
                    reply = {'jsonrpc': '2.0', 'id': message['id'], 'result': result}
                    if mode in ('json_error', 'json_wrong_error', 'json_missing_error'):
                        reply.pop('result')
                        reply['error'] = {'code': -32600, 'message': 'sensitive-fixture'}
                    if mode in ('json_wrong', 'json_wrong_error'):
                        reply['id'] = 999
                    elif mode in ('json_missing', 'json_missing_error'):
                        reply.pop('id')
                    payload = json.dumps(reply).encode()
                    if mode == 'sse':
                        payload = b'data: {"method":"notifications/message"}\n\ndata: ' + payload + b'\n\n'
                    self.wfile.write(payload)
                except (BrokenPipeError, ConnectionResetError):
                    pass
                finally:
                    if mode != 'notification_headers' or notification_stall:
                        disconnected.set()

        server = ThreadingHTTPServer(('127.0.0.1', 0), Handler)
        server.daemon_threads = True
        thread = threading.Thread(target=server.serve_forever, daemon=True)
        thread.start()
        runner = subprocess.Popen([sys.executable, '-c', HTTP_RUNNER, str(SOURCE),
                                   f'http://127.0.0.1:{server.server_port}', str(request_timeout)],
                                  stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True,
                                  start_new_session=True,
                                  env={**os.environ, 'MCP_TEST_SECRET': 'sensitive-fixture'})
        try:
            if mode == 'interrupt':
                self.assertTrue(received.wait(min(request_timeout + 2,
                                                  max(0, outer_deadline - time.monotonic()))),
                                'worker must begin its HTTP request')
                runner.send_signal(signal.SIGINT)
            try:
                stdout, stderr = runner.communicate(timeout=max(0, outer_deadline - time.monotonic()))
            except subprocess.TimeoutExpired:
                os.killpg(runner.pid, signal.SIGKILL)
                runner.communicate()
                self.fail('HTTP doctor exceeded outer safety deadline')
            self.assertEqual(runner.returncode, 0, stderr)
            result = json.loads(stdout.splitlines()[-1])
            self.assertTrue(result['reaped'], 'HTTP worker must be reaped')
            self.assertFalse(result['secret_in_argv'])
            self.assertNotIn('sensitive-fixture', stdout + stderr)
            self.assertTrue(disconnected.wait(2), 'server must observe the closed response')
            return result, requests
        finally:
            if runner.poll() is None:
                os.killpg(runner.pid, signal.SIGKILL)
                runner.communicate()
            server.shutdown()
            server.server_close()
            thread.join()

    def test_streaming_and_stalled_responses_have_one_deadline(self):
        for mode in ('notifications', 'wrong_ids', 'partial_line', 'json_stall', 'headers'):
            with self.subTest(mode=mode):
                result, _ = self.probe(mode)
                self.assertEqual(result['error'], 'initialize: timed out waiting for HTTP response')

    def test_json_and_sse_preserve_session_protocol_and_notification(self):
        for mode in ('json', 'sse'):
            with self.subTest(mode=mode):
                result, requests = self.probe(mode)
                self.assertNotIn('error', result)
                self.assertEqual([r[0]['method'] for r in requests],
                                 ['initialize', 'notifications/initialized', 'tools/list'])
                for _, headers in requests:
                    self.assertEqual(headers.get('X-Test-Secret'), 'sensitive-fixture')
                for _, headers in requests[1:]:
                    self.assertEqual(headers.get('Mcp-Session-Id'), 'local-session')
                    self.assertEqual(headers.get('Mcp-Protocol-Version'), '2025-06-18')

    def test_http_errors_do_not_expose_response_or_credentials(self):
        for mode, error in (('http_error', 'HTTPError'), ('sse_eof', 'no JSON-RPC response')):
            with self.subTest(mode=mode):
                result, _ = self.probe(mode)
                self.assertEqual(result['error'], error)

    def test_json_responses_are_correlated_before_success_or_error_classification(self):
        cases = (('json', None), ('json_error', 'initialize failed'),
                 ('json_wrong', 'no JSON-RPC response'),
                 ('json_wrong_error', 'no JSON-RPC response'),
                 ('json_missing', 'no JSON-RPC response'),
                 ('json_missing_error', 'no JSON-RPC response'))
        for mode, expected in cases:
            with self.subTest(mode=mode):
                result, _ = self.probe(mode)
                self.assertEqual(result.get('error'), expected)

    def test_notification_timeout_names_its_phase_and_closes_its_own_socket(self):
        result, requests = self.probe('notification_headers')
        self.assertEqual(result['error'],
                         'notifications/initialized: timed out waiting for HTTP response')
        self.assertEqual([r[0]['method'] for r in requests],
                         ['initialize', 'notifications/initialized'])

    def test_interrupted_request_reaps_its_worker_and_closes_the_socket(self):
        result, _ = self.probe('interrupt')
        self.assertEqual(result['error'], 'KeyboardInterrupt')


if __name__ == '__main__':
    unittest.main()
