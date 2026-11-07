"""Probe configured MCP transports without changing host configuration or credentials."""
import json
import os
from pathlib import Path
import queue
import signal
import subprocess
import threading
import tomllib
import urllib.request

ROOT = Path(__file__).resolve().parent.parent
CONFIG = tomllib.loads((ROOT / '.codex/config.toml').read_text())['mcp_servers']


def probe(name, config):
    process = None
    session = None
    protocol = '2025-03-26'
    inbox = queue.Queue()
    headers = {'Content-Type': 'application/json', 'Accept': 'application/json, text/event-stream'}
    token_var = config.get('bearer_token_env_var')
    if token_var:
        if not os.environ.get(token_var):
            raise RuntimeError(f'missing runtime credential: {token_var}')
        headers['Authorization'] = 'Bearer ' + os.environ[token_var]
    if 'command' in config:
        process = subprocess.Popen([config['command'], *config.get('args', [])], cwd=ROOT,
                                   stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                   stderr=subprocess.DEVNULL, text=True, start_new_session=True)
        def reader():
            for line in process.stdout:
                try:
                    inbox.put(json.loads(line))
                except ValueError:
                    continue
            inbox.put(None)
        threading.Thread(target=reader, daemon=True).start()

    def send(message):
        nonlocal session
        wanted = message.get('id')
        if process:
            process.stdin.write(json.dumps(message) + '\n')
            process.stdin.flush()
            if wanted is None:
                return None
            while True:
                item = inbox.get(timeout=45)
                if item is None:
                    raise RuntimeError('server exited before reply')
                if item.get('id') == wanted:
                    return item
        request_headers = dict(headers)
        if session:
            request_headers['Mcp-Session-Id'] = session
        if message.get('method') != 'initialize':
            request_headers['MCP-Protocol-Version'] = protocol
        req = urllib.request.Request(config['url'], data=json.dumps(message).encode(), headers=request_headers)
        with urllib.request.urlopen(req, timeout=25) as response:
            session = response.headers.get('Mcp-Session-Id', session)
            if wanted is None:
                return None
            if 'text/event-stream' in response.headers.get('Content-Type', ''):
                for line in response:
                    if line.startswith(b'data:'):
                        item = json.loads(line[5:])
                        if item.get('id') == wanted:
                            return item
            else:
                return json.load(response)
        raise RuntimeError('no JSON-RPC response')

    def request(number, method, params):
        result = send({'jsonrpc': '2.0', 'id': number, 'method': method, 'params': params})
        if not result or 'error' in result:
            raise RuntimeError(f'{method} failed')
        return result['result']

    try:
        init = request(1, 'initialize', {'protocolVersion': protocol, 'capabilities': {},
                                        'clientInfo': {'name': 'cloud-mcp-doctor', 'version': '1.0'}})
        protocol = init['protocolVersion']
        send({'jsonrpc': '2.0', 'method': 'notifications/initialized'})
        names = set()
        params = {}
        for number in range(2, 102):
            result = request(number, 'tools/list', params)
            names.update(tool['name'] for tool in result['tools'])
            if not result.get('nextCursor'):
                break
            params = {'cursor': result['nextCursor']}
        else:
            raise RuntimeError('tools/list pagination limit exceeded')
        missing = set(config.get('enabled_tools', [])) - names
        if missing:
            raise RuntimeError('configured tools not offered: ' + ', '.join(sorted(missing)))
        if not names:
            raise RuntimeError('empty tools list')
        print(f'PASS {name}: initialize + tools/list ({len(names)} tools)', flush=True)
    finally:
        if process:
            try:
                os.killpg(process.pid, signal.SIGTERM)
                process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                os.killpg(process.pid, signal.SIGKILL)
                process.wait()
            except ProcessLookupError:
                pass


if __name__ == '__main__':
    failures = 0
    for name, config in CONFIG.items():
        try:
            probe(name, config)
        except Exception as exc:
            failures += 1
            # Do not print response bodies or request headers containing credentials.
            detail = str(exc) if isinstance(exc, RuntimeError) else type(exc).__name__
            print(f'FAIL {name}: {detail}', flush=True)
    print('Transport probes do not prove the current chat host loaded these servers.', flush=True)
    raise SystemExit(bool(failures))
