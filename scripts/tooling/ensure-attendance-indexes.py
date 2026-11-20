"""Create missing attendance indexes and wait for readiness before deployment."""
import json
import os
from pathlib import Path
import sys
import time
from urllib.error import HTTPError
from urllib.parse import urlencode
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
    return index["queryScope"], tuple((f["fieldPath"], f["order"]) for f in index["fields"])


def main(project, database):
    manifest = Path(__file__).resolve().parents[2] / "deploy/firestore/indexes.json"
    required = [i for i in json.loads(manifest.read_text())["indexes"] if i["collectionGroup"] == "attendance"]
    url = f"https://firestore.googleapis.com/v1/projects/{project}/databases/{database}/collectionGroups/attendance/indexes"
    deadline = time.monotonic() + 1800
    while time.monotonic() < deadline:
        indexes = []
        page = {}
        while True:
            result = request(url + "?" + urlencode(page))
            indexes.extend(result.get("indexes", []))
            if not result.get("nextPageToken"):
                break
            page = {"pageToken": result["nextPageToken"]}
        states = {key(i): i["state"] for i in indexes if all("order" in f for f in i["fields"])}
        for index in required:
            state = states.get(key(index))
            if state is None:
                try:
                    request(url, {k: index[k] for k in ("queryScope", "fields")})
                except HTTPError as error:
                    if error.code != 409:
                        raise
            elif state == "NEEDS_REPAIR":
                raise RuntimeError(f"Index needs repair: {key(index)}")
        if all(states.get(key(i)) == "READY" for i in required):
            print(f"{len(required)} attendance indexes READY on {project}/{database}")
            return
        time.sleep(10)
    raise TimeoutError("Attendance indexes did not become READY")


if __name__ == "__main__":
    main(sys.argv[1], sys.argv[2] if len(sys.argv) > 2 else "(default)")
