"""Apply Firestore field policies and composite indexes before deployment."""
import argparse
import json
import os
from pathlib import Path
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


def request(url, body=None, *, method=None):
    data = None if body is None else json.dumps(body).encode()
    headers = {"Authorization": "Bearer " + token(), "Content-Type": "application/json"}
    with urlopen(Request(url, data=data, headers=headers, method=method), timeout=30) as response:
        return json.load(response)


def collection_groups_url(project, database):
    project = quote(project, safe="")
    database = quote(database, safe="")
    return f"https://firestore.googleapis.com/v1/projects/{project}/databases/{database}/collectionGroups"


def ensure_field_policies(project, database, policies, call=request, sleep=time.sleep, clock=time.monotonic):
    base = collection_groups_url(project, database)
    for policy in policies:
        if policy["fieldPath"] == "*":
            continue  # Wildcard index defaults are managed separately.
        url = f"{base}/{quote(policy['collectionGroup'], safe='')}/fields/{quote(policy['fieldPath'], safe='')}"
        body = {"indexConfig": {"indexes": policy.get("indexes", [])}}
        if policy.get("ttl"):
            body["ttlConfig"] = {}

        def ready(value):
            index = value.get("indexConfig", {})
            return (
                (not policy.get("ttl") or value.get("ttlConfig", {}).get("state") == "ACTIVE")
                and "indexConfig" in value
                and not index.get("usesAncestorConfig", False)
                and not index.get("reverting", False)
                and index.get("indexes", []) == body["indexConfig"]["indexes"]
            )

        if ready(call(url)):
            continue
        call(url + "?" + urlencode({"updateMask": ",".join(body)}), body, method="PATCH")
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


def key(index):
    return index["queryScope"], json.dumps(index["fields"], sort_keys=True)


def ensure_indexes(project, database, required, call=request, sleep=time.sleep, clock=time.monotonic):
    groups = {}
    for index in required:
        groups.setdefault(index["collectionGroup"], []).append(index)
    base = collection_groups_url(project, database)
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


def main(project, database, *, indexes_only=False):
    manifest_path = Path(__file__).resolve().parents[2] / "firestore.indexes.json"
    manifest = json.loads(manifest_path.read_text())
    if not indexes_only:
        ensure_field_policies(project, database, manifest["fieldOverrides"])
    ensure_indexes(project, database, manifest["indexes"])


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("project")
    parser.add_argument("database", nargs="?", default="(default)")
    parser.add_argument("--indexes-only", action="store_true", help="Skip field policies")
    args = parser.parse_args()
    main(args.project, args.database, indexes_only=args.indexes_only)
