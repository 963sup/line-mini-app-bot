#!/usr/bin/env python3
"""Read Cloud Run release metadata, excluding plaintext secrets and user data."""
import argparse
import datetime
import json
import re
import subprocess
import sys


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--project', required=True)
    parser.add_argument('--region', required=True)
    parser.add_argument('--service', required=True)
    args = parser.parse_args()
    for value in [args.project, args.region, args.service]:
        if not re.fullmatch(r'[a-z][a-z0-9-]{0,62}', value):
            parser.error('Use explicit project, region and service IDs')
    common = [f'--project={args.project}', f'--region={args.region}', '--format=json', '--quiet']

    def read(*command):
        result = subprocess.run(['gcloud', *command, *common], capture_output=True, text=True, timeout=60)
        if result.returncode:
            # Raw SDK errors may contain environment details; preserve a safe classification only.
            raise RuntimeError('Cloud Run metadata unavailable; check active gcloud identity, project and IAM')
        return json.loads(result.stdout)

    try:
        service = read('run', 'services', 'describe', args.service)
        status = service.get('status', {})
        revisions = {}
        names = {item.get('revisionName') for item in status.get('traffic', [])}
        names.update([status.get('latestCreatedRevisionName'), status.get('latestReadyRevisionName')])
        for name in sorted(n for n in names if n):
            revision = read('run', 'revisions', 'describe', name)
            spec = revision.get('spec', {})
            containers = []
            for container in spec.get('containers', []):
                env = container.get('env', [])
                containers.append({'image': container.get('image'), 'resources': container.get('resources'),
                                   'envNames': sorted(e['name'] for e in env),
                                   'secretReferences': [{'env': e['name'], 'reference': e['valueFrom']} for e in env if 'valueFrom' in e]})
            revisions[name] = {'serviceAccount': spec.get('serviceAccountName'),
                               'imageDigest': revision.get('status', {}).get('imageDigest'),
                               'conditions': revision.get('status', {}).get('conditions'),
                               'containers': containers}
        print(json.dumps({'capturedAt': datetime.datetime.now(datetime.timezone.utc).isoformat(),
                          'project': args.project, 'region': args.region, 'service': args.service,
                          'url': status.get('url'), 'traffic': status.get('traffic', []),
                          'latestReadyRevision': status.get('latestReadyRevisionName'),
                          'revisions': revisions}, ensure_ascii=False, indent=2))
        return 0
    except (RuntimeError, subprocess.TimeoutExpired, ValueError) as error:
        print(str(error), file=sys.stderr)
        return 1


if __name__ == '__main__':
    sys.exit(main())
