"""Apply manifest TTL/index exemptions and wait for ACTIVE/READY before deployment."""
import json
from pathlib import Path
import runpy
import sys
import time
from urllib.parse import quote, urlencode
from urllib.request import Request, urlopen

_token = runpy.run_path(str(Path(__file__).with_name("ensure-attendance-indexes.py")))["token"]


def request(url, body=None):
    data = None if body is None else json.dumps(body).encode()
    req = Request(url, data=data, method="GET" if body is None else "PATCH",
                  headers={"Authorization": "Bearer " + _token(), "Content-Type": "application/json"})
    with urlopen(req, timeout=30) as response:
        return json.load(response)


def ensure(project, database, policies, call=request, sleep=time.sleep, clock=time.monotonic):
    base = f"https://firestore.googleapis.com/v1/projects/{quote(project, safe='')}/databases/{quote(database, safe='')}/collectionGroups"
    for policy in policies:
        if policy["fieldPath"] == "*":
            continue  # Wildcard index defaults are managed separately.
        url = f"{base}/{quote(policy['collectionGroup'], safe='')}/fields/{quote(policy['fieldPath'], safe='')}"
        body = {"indexConfig": {"indexes": policy.get("indexes", [])}}
        if policy.get("ttl"):
            body["ttlConfig"] = {}
        current = call(url)
        def ready(value):
            index = value.get("indexConfig", {})
            return (not policy.get("ttl") or value.get("ttlConfig", {}).get("state") == "ACTIVE") and "indexConfig" in value and not index.get("usesAncestorConfig", False) and not index.get("reverting", False) and index.get("indexes", []) == body["indexConfig"]["indexes"]
        if ready(current):
            continue
        call(url + "?" + urlencode({"updateMask": ",".join(body)}), body)
        deadline = clock() + 1800
        while True:
            state = call(url)
            if ready(state):
                break
            if state.get("ttlConfig", {}).get("state") == "NEEDS_REPAIR":
                raise RuntimeError(f"TTL needs repair: {policy['collectionGroup']}/{policy['fieldPath']}")
            if clock() >= deadline:
                raise TimeoutError(f"Field policy not ready: {policy['collectionGroup']}/{policy['fieldPath']}")
            sleep(5)
    print(f"Field policies verified on {project}/{database}; wildcard defaults are excluded")


if __name__ == "__main__":
    policies = json.loads((Path(__file__).resolve().parents[2] / "firestore.indexes.json").read_text())["fieldOverrides"]
    ensure(sys.argv[1], sys.argv[2] if len(sys.argv) > 2 else "(default)", policies)
