"""One HTTP exchange; the doctor owns this worker's lifetime and deadline."""
import json
import sys
import urllib.request


def exchange(payload):
    message = payload['message']
    wanted = message.get('id')
    request = urllib.request.Request(payload['url'], data=json.dumps(message).encode(),
                                     headers=payload['headers'])
    with urllib.request.urlopen(request, timeout=payload['timeout']) as response:
        session = response.headers.get('Mcp-Session-Id', payload['session'])
        if wanted is None:
            return {'session': session, 'reply': None}
        if 'text/event-stream' in response.headers.get('Content-Type', ''):
            for line in response:
                if line.startswith(b'data:'):
                    reply = json.loads(line[5:])
                    if reply.get('id') == wanted:
                        return {'session': session, 'reply': reply}
        else:
            reply = json.load(response)
            if isinstance(reply, dict) and reply.get('id') == wanted:
                return {'session': session, 'reply': reply}
    return {'error': 'no JSON-RPC response'}


if __name__ == '__main__':
    try:
        result = exchange(json.load(sys.stdin))
    except Exception as error:
        # Do not serialize HTTP bodies, URLs, headers or credential-bearing input.
        result = {'error': type(error).__name__}
    json.dump(result, sys.stdout)
