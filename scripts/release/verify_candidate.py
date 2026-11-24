"""Smoke-test the immutable tagged candidate before updating navigation or traffic."""
import json
import sys
from pathlib import Path
from urllib.error import HTTPError
from urllib.request import urlopen


def verify(service, tag):
    targets = [t for t in service["status"]["traffic"] if t.get("tag") == tag]
    if len(targets) != 1:
        raise ValueError("Expected one candidate traffic tag")
    candidate = targets[0]
    url = candidate["url"]
    with urlopen(url + "/api/v1/miniapp/config", timeout=15) as response:
        if not json.load(response).get("liffId"):
            raise ValueError("Candidate has no LIFF configuration")
    with urlopen(url + "/miniapp/object-recognition", timeout=15) as response:
        if response.status != 200 or "text/html" not in response.headers["Content-Type"]:
            raise ValueError("Candidate cannot serve the Mini App")
    try:
        with urlopen(url + "/api/v1/me", timeout=15):
            raise ValueError("Anonymous request unexpectedly succeeded")
    except HTTPError as error:
        if error.code != 401:
            raise
    return candidate["revisionName"]


if __name__ == "__main__":
    source, tag, target = sys.argv[1:]
    Path(target).write_text(verify(json.loads(Path(source).read_text()), tag))
