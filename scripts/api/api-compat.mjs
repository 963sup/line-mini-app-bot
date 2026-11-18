import { spawnSync } from 'node:child_process';
import { mkdtempSync, mkdirSync, writeFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
const root = fileURLToPath(new URL('../../', import.meta.url));
function command(binary, args, options = {}) {
 const result = spawnSync(binary, args, { cwd: root, encoding: 'utf8', ...options });
 if (result.error) throw result.error;
 return result;
}
const args = process.argv.slice(2);
if (args.length && (args.length !== 2 || args[0] !== '--base')) throw new Error('Usage: pnpm api:compat [--base git-ref]');
const requested = args[1] || process.env.API_BASE_REF;
const temp = mkdtempSync(join(tmpdir(), 'line-api-compat-'));
function compare(base, current = 'api/contract/openapi.yaml') {
 const result = command('oasdiff', ['breaking', '--fail-on', 'ERR', base, current], { stdio: 'inherit' });
 if (result.status !== 0) throw new Error('API breaking change detected or comparison failed');
}
try {
 // Frozen first complete contract also protects the initial migration before it is committed.
 compare('api/compatibility/baseline/openapi.yaml');
 compare('api/compatibility/baseline/internal-openapi.yaml', 'api/contract/internal-openapi.yaml');
 if (requested && !/^0+$/.test(requested)) {
  const revision = command('git', ['rev-parse', '--verify', `${requested}^{commit}`]);
  if (revision.status !== 0) throw new Error('API base revision is unavailable; fetch the base before checking');
  const sha = revision.stdout.trim();
  const tree = command('git', ['ls-tree', '-r', '--name-only', sha, '--', 'api/contract/']);
  if (tree.status !== 0) throw new Error('Could not inspect API base revision');
  // Documentation can live beside contracts; only materialize supported spec sources.
  const paths = tree.stdout.trim().split('\n').filter(path => /\.(yaml|yml|json)$/.test(path));
  if (paths.includes('api/contract/openapi.yaml')) {
   for (const path of paths) {
    if (!path.startsWith('api/contract/') || path.includes('..') || !/\.(yaml|yml|json)$/.test(path)) throw new Error('Unsupported API contract source path');
    const file = command('git', ['show', `${sha}:${path}`]);
    if (file.status !== 0) throw new Error('Could not read API base source');
    const target = join(temp, path.slice('api/contract/'.length));
    mkdirSync(dirname(target), { recursive: true }); writeFileSync(target, file.stdout);
   }
   compare(join(temp, 'openapi.yaml'));
   if (paths.includes('api/contract/internal-openapi.yaml')) compare(join(temp, 'internal-openapi.yaml'), 'api/contract/internal-openapi.yaml');
   else console.log('Base predates internal OpenAPI; frozen initial internal contract checked.');
  } else console.log('Base predates OpenAPI; frozen initial contract checked.');
 }
 console.log('PASS: API backward compatibility');
} finally { rmSync(temp, { recursive: true, force: true }); }
