"""Stop superseded builds at safe boundaries, before deployment side effects."""
import argparse
import json
import subprocess


def gcloud(project, region, *args):
    result = subprocess.run(
        ['gcloud', *args, '--project=' + project, '--region=' + region,
         '--quiet', '--format=json'],
        capture_output=True, text=True, check=True, timeout=45,
    )
    return json.loads(result.stdout or '{}')


def superseded(current, latest):
    if not current.get('id') or not current.get('createTime'):
        raise ValueError('Current build identity/time missing')
    if not latest or latest[0].get('buildTriggerId') != current.get('buildTriggerId'):
        raise ValueError('Latest build lookup returned no matching trigger')
    if not latest[0].get('createTime'):
        raise ValueError('Latest build time missing')
    # Cloud Build timestamps use RFC3339 UTC; normalize optional fractional seconds.
    from datetime import datetime
    timestamp = lambda value: datetime.fromisoformat(value.replace('Z', '+00:00'))
    return timestamp(latest[0]['createTime']) > timestamp(current['createTime'])


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--project', required=True)
    parser.add_argument('--region', required=True)
    parser.add_argument('--build', required=True)
    parser.add_argument('--dry-run', action='store_true')
    args = parser.parse_args()
    current = gcloud(args.project, args.region, 'builds', 'describe', args.build)
    trigger = current.get('buildTriggerId')
    if not trigger:
        print('Manual build: no trigger ordering to enforce')
        return
    latest = gcloud(
        args.project, args.region, 'builds', 'list', '--filter=trigger_id="' + trigger + '"',
        '--sort-by=~createTime', '--limit=1',
    )
    if not superseded(current, latest):
        print('Latest build: continuing')
        return
    print('Superseded by ' + latest[0]['id'] + '; stopping before deployment')
    if not args.dry_run:
        gcloud(args.project, args.region, 'builds', 'cancel', args.build)
    # Fail closed even if cancellation is asynchronous; no following steps may run.
    raise SystemExit(1)


if __name__ == '__main__':
    main()
