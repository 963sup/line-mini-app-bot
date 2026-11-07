"""Probe configured MCP transports without changing host configuration or credentials."""
import argparse
import json
import os
from pathlib import Path
import queue
import signal
import subprocess
import threading
import tomllib
import urllib.request
from urllib.parse import urlsplit, urlunsplit
import shlex

ROOT = Path(__file__).resolve().parent.parent
CONFIG = tomllib.loads((ROOT / '.codex/config.toml').read_text())['mcp_servers']

# Explicit host-managed OAuth services; arbitrary HTTP 401s are not skips.
HOST_OAUTH_SERVERS = {"fly-mcp"}


def host_verification(name, config):
    endpoint = urlsplit(config.get("url", ""))
    if endpoint.scheme != "https" or not endpoint.hostname:
        raise RuntimeError("host OAuth server requires a valid HTTPS URL")
    host = endpoint.hostname
    if ":" in host:
        host = "[" + host + "]"
    if endpoint.port:
        host += ":" + str(endpoint.port)
    safe_url = urlunsplit(("https", host, endpoint.path, "", ""))
    print(f"HOST_REQUIRED {name}: 需在宿主內驗證（未執行直接連線檢查）", flush=True)
    print(f"  MCP URL: {safe_url} (服務端點，不是 OAuth 授權 URL)", flush=True)
    print(f"  Login: codex mcp login {shlex.quote(name)} --no-browser", flush=True)
    print("  由支援的 Codex 宿主產生實際授權 URL；登入後呼叫宿主內唯讀工具驗證。", flush=True)


def check_servers(configs, server=None, call=None, arguments=None, direct=False):
    counts = {"PASS": 0, "FAIL": 0, "HOST_REQUIRED": 0, "SKIP": 0}
    required_pending = False
    for name, config in configs.items():
        if server and name != server:
            continue
        if config.get("enabled", True) is False:
            counts["SKIP"] += 1
            print(f"SKIP {name}: disabled", flush=True)
            continue
        try:
            if name in HOST_OAUTH_SERVERS and not direct:
                host_verification(name, config)
                counts["HOST_REQUIRED"] += 1
                required_pending |= bool(config.get("required", False))
            else:
                probe(name, config, call, arguments)
                counts["PASS"] += 1
        except Exception as exc:
            counts["FAIL"] += 1
            detail = str(exc) if isinstance(exc, RuntimeError) else type(exc).__name__
            print(f"FAIL {name}: {detail}", flush=True)
    print("Summary: " + ", ".join(f"{key}={value}" for key, value in counts.items()), flush=True)
    print("Transport probes do not prove the current chat host loaded these servers.", flush=True)
    if required_pending:
        print("Required host OAuth verification is pending; readiness is incomplete.", flush=True)
    return int(bool(counts["FAIL"] or required_pending))


def probe(name, config, call=None, arguments=None):
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
                item = inbox.get(timeout=config.get('startup_timeout_sec', 45))
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
        if call:
            if call not in names:
                raise RuntimeError('requested tool is not offered')
            result = request(102, 'tools/call', {'name': call, 'arguments': arguments or {}})
            if result.get('isError'):
                raise RuntimeError('tool returned an error')
            print(json.dumps(result), flush=True)
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
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--server', choices=sorted(CONFIG))
    parser.add_argument('--direct', action='store_true', help='Explicitly probe host-OAuth servers without host credentials (401 remains a failure)')
    parser.add_argument('--call', help='Explicit read-only tool call; requires --server')
    parser.add_argument('--arguments', default='{}', help='JSON tool arguments; never include credentials')
    args = parser.parse_args()
    if args.call and not args.server:
        parser.error('--call requires --server')
    if args.call and (args.server != 'openaiDeveloperDocs' or args.call not in {'search_openai_docs', 'fetch_openai_doc'}):
        parser.error('--call currently supports only OpenAI Docs search/fetch')
    arguments = json.loads(args.arguments)
    if not isinstance(arguments, dict):
        parser.error('--arguments must be a JSON object')
    raise SystemExit(check_servers(CONFIG, args.server, args.call, arguments, args.direct))
