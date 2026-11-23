"""Create missing manifest Firestore indexes and wait for readiness before deployment."""
import json
import os
from pathlib import Path
import sys
import time
from urllib.error import HTTPError
from urllib.parse import quote, urlencode
from urllib.request import Request, urlopen


def token():
    if os.environ.get("GOOGLE_OAUTH_ACCESS_TOKEN"):
        return os.environ["GOOGLE_OAUTH_ACCESS_TOKEN"]
    url = "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token"
    with urlopen(Request(url, headers={"Metadata-Flavor": "Google"}), timeout=15) as response:
        return json.load(response)["access_token"]


def request(url, body=None):
    data = None if body is None else json.dumps(body).encode()
    headers = {"Authorization": "Bearer " + token(), "Content-Type": "application/json"}
    with urlopen(Request(url, data=data, headers=headers), timeout=30) as response:
        return json.load(response)


def key(index):
    return index["queryScope"], json.dumps(index["fields"], sort_keys=True)


def ensure(project, database, required, call=request, sleep=time.sleep, clock=time.monotonic):
    groups = {}
    for index in required:
        groups.setdefault(index["collectionGroup"], []).append(index)
    base = f"https://firestore.googleapis.com/v1/projects/{quote(project, safe='')}/databases/{quote(database, safe='')}/collectionGroups"
    deadline = clock() + 1800
    while clock() < deadline:
        ready = True
        for group, definitions in groups.items():
            url = f"{base}/{quote(group, safe='')}/indexes"
            indexes = []
            page = {}
            while True:
                result = call(url + "?" + urlencode(page))
                indexes.extend(result.get("indexes", []))
                if not result.get("nextPageToken"):
                    break
                page = {"pageToken": result["nextPageToken"]}
            states = {key(i): i["state"] for i in indexes}
            for index in definitions:
                state = states.get(key(index))
                if state != "READY":
                    ready = False
                if state is None:
                    try:
                        call(url, {k: index[k] for k in ("queryScope", "fields")})
                    except HTTPError as error:
                        if error.code != 409:
                            raise
                elif state == "NEEDS_REPAIR":
                    raise RuntimeError(f"Index needs repair: {group} {key(index)}")
        if ready:
            print(f"{len(required)} Firestore indexes READY on {project}/{database}")
            return
        sleep(10)
    raise TimeoutError("Firestore indexes did not become READY")


def main(project, database):
    manifest = Path(__file__).resolve().parents[2] / "firestore.indexes.json"
    ensure(project, database, json.loads(manifest.read_text())["indexes"])


if __name__ == "__main__":
    main(sys.argv[1], sys.argv[2] if len(sys.argv) > 2 else "(default)")
