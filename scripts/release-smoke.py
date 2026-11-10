#!/usr/bin/env python3
"""Probe deployed/static-container HTTP behavior without tokens or database writes."""
import argparse
import json
import re
import sys
import urllib.error
import urllib.parse
import urllib.request
from html.parser import HTMLParser


class Assets(HTMLParser):
    def __init__(self):
        super().__init__()
        self.paths = set()

    def handle_starttag(self, tag, attrs):
        values = dict(attrs)
        path = values.get('src') if tag == 'script' else values.get('href') if tag == 'link' else None
        if path and path.startswith(('/_next/', '/assets/')):
            self.paths.add(path)


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--base-url', required=True)
    parser.add_argument('--allow-disabled-attendance', action='store_true', help='Local container only: explicitly accept a disabled attendance store (503)')
    args = parser.parse_args()
    target = urllib.parse.urlsplit(args.base_url)
    if target.username or target.password or target.query or target.fragment or target.path not in ('', '/'):
        parser.error('Provide an origin URL without credentials, path, query or fragment')
    if target.scheme != 'https' and not (target.scheme == 'http' and target.hostname in ('localhost', '127.0.0.1', '::1')):
        parser.error('HTTPS is required except for local containers')
    if args.allow_disabled_attendance and target.hostname not in ('localhost', '127.0.0.1', '::1'):
        parser.error('Disabled attendance is only permitted for a local container probe')
    base = args.base_url.rstrip('/')
    opener = urllib.request.build_opener(NoRedirect())
    results = []

    def probe(path, expected, check=None):
        try:
            request = urllib.request.Request(base + path, headers={'User-Agent': 'repository-release-smoke/1'})
            try:
                response = opener.open(request, timeout=20)
            except urllib.error.HTTPError as error:
                response = error
            with response:
                status, headers = response.status, response.headers
                body = response.read(2_000_001)
            assert status in (expected if isinstance(expected, tuple) else (expected,)), f'HTTP {status}, expected {expected}'
            assert len(body) <= 2_000_000, 'response exceeds probe limit'
            if check:
                check(headers, body)
            results.append({'path': path, 'passed': True, 'status': status})
            return body
        except Exception as error:
            # Never log response bodies, tokens, profiles or server error payloads.
            results.append({'path': path, 'passed': False, 'error': str(error)})
            return b''

    def page(headers, body):
        assert headers.get_content_type() == 'text/html', 'not HTML'
        assert 'no-store' in headers.get('Cache-Control', ''), 'page cache policy missing'
        assert headers.get('X-Content-Type-Options') == 'nosniff', 'nosniff missing'
        assert headers.get('Referrer-Policy') == 'no-referrer', 'referrer policy missing'
        policy = headers.get('Content-Security-Policy', '')
        assert "default-src 'none'" in policy and "'unsafe-inline'" not in policy, 'CSP missing or relaxed'
        assert b'__NEXT_DATA__' in body, 'Next page artifact missing'

    assets = set()
    for path in ['/miniapp', '/miniapp/history', '/miniapp/correction', '/miniapp/review', '/miniapp/punch']:
        body = probe(path, 200, page)
        parsed = Assets()
        parsed.feed(body.decode('utf8', errors='replace'))
        assets.update(parsed.paths)
    for path in sorted(assets):
        probe(path, 200)
    for path in ['/location', '/assets/miniapp.js', '/assets/attendance.js', '/assets/app-api.js', '/assets/miniapp.css']:
        probe(path, 200)
    probe('/api/v1/me', 401)
    probe('/api/v1/attendance/groups/page', (401, 503) if args.allow_disabled_attendance else 401)
    probe('/_next/nonexistent-release-check.js', 404)
    probe('/assets/nonexistent-release-check.js', 404)
    probe('/assets/%2e%2e/package.json', 404)
    print(json.dumps({'baseURL': base, 'passed': all(r['passed'] for r in results),
                      'disabledAttendanceAllowed': args.allow_disabled_attendance,
                      'scope': 'HTTP only; no LINE login, device GPS, database or workflow acceptance',
                      'checks': results}, ensure_ascii=False, indent=2))
    return 0 if all(r['passed'] for r in results) else 1


if __name__ == '__main__':
    sys.exit(main())
