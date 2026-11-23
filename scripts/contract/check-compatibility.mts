import { spawnSync } from 'node:child_process';
import { mkdtempSync, readFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('../../', import.meta.url));
const ignored = new Set(['description', 'summary', 'example', 'examples', 'externalDocs', 'tags']);
function normalize(value: unknown, preserveKeys = false): unknown {
  if (Array.isArray(value)) return value.map(entry => normalize(entry));
  if (value && typeof value === 'object') return Object.fromEntries(Object.entries(value)
    .filter(([key]) => preserveKeys || !ignored.has(key)).sort(([a], [b]) => a.localeCompare(b))
    .map(([key, entry]) => [key, normalize(entry, ['properties', 'schemas', 'responses', 'headers', 'content', 'encoding', 'securitySchemes', 'paths', '$defs'].includes(key))]));
  return value;
}
type Document = { paths?: Record<string, Record<string, unknown>>; security?: unknown; servers?: unknown; components?: { securitySchemes?: unknown } };
export function incompatibleOperations(before: Document, after: Document): string[] {
  const changes: string[] = [];
  const same = (a: unknown, b: unknown) => JSON.stringify(normalize(a)) === JSON.stringify(normalize(b));
  for (const key of ['security', 'servers'] as const) if (!same(before[key], after[key])) changes.push(key);
  if (!same(before.components?.securitySchemes, after.components?.securitySchemes)) changes.push('securitySchemes');
  for (const [path, item] of Object.entries(before.paths || {})) {
    for (const [method, operation] of Object.entries(item)) {
      if (!same(operation, after.paths?.[path]?.[method])) changes.push(`${method.toUpperCase()} ${path}`);
    }
  }
  return changes;
}

function run(command: string, args: string[], cwd = root, input?: Buffer): Buffer {
  const result = spawnSync(command, args, { cwd, input, maxBuffer: 64 * 1024 * 1024 });
  if (result.error) throw result.error;
  if (result.status !== 0) throw new Error(`${command} failed: ${result.stderr.toString()}`);
  return result.stdout;
}
function main() {
  const base = process.argv[process.argv.indexOf('--base') + 1];
  if (!process.argv.includes('--base') || !base || base.startsWith('-')) throw new Error('Specify --base <Git commit>; no implicit compatibility baseline');
  const sha = run('git', ['rev-parse', '--verify', `${base}^{commit}`]).toString().trim();
  const temporary = mkdtempSync(join(tmpdir(), 'api-compatibility-'));
  try {
    run('tar', ['-x', '-C', temporary], root, run('git', ['archive', sha, 'api/contract']));
    for (const entry of ['openapi', 'internal-openapi']) {
      const documents = [join(temporary, 'api/contract', `${entry}.yaml`), join(root, 'api/contract', `${entry}.yaml`)]
        .map((source, index) => {
          const output = join(temporary, `${entry}-${index}.json`);
          run('pnpm', ['exec', 'redocly', 'bundle', '--config', 'api/codegen/redocly.yaml', '--dereferenced', '--ext', 'json', source, '-o', output]);
          return JSON.parse(readFileSync(output, 'utf8')) as Document;
        });
      const changes = incompatibleOperations(documents[0], documents[1]);
      if (changes.length) throw new Error(`Contract changes require compatibility review against ${sha} (${entry}):\n${changes.join('\n')}`);
    }
    console.log(`PASS: existing public/internal operations unchanged against ${sha}; additive operations and documentation permitted`);
  } finally { rmSync(temporary, { recursive: true, force: true }); }
}
if (process.argv[1] && fileURLToPath(import.meta.url) === resolve(process.argv[1])) main();
