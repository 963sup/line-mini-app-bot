import { spawnSync } from 'node:child_process';
import { mkdtempSync, readFileSync, rmSync, copyFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
const root = fileURLToPath(new URL('../', import.meta.url));
const temporary = mkdtempSync(join(tmpdir(), 'line-api-generate-'));
const check = process.argv.includes('--check');
function run(command, args, cwd = root) {
 const result = spawnSync(command, args, { cwd, stdio: 'inherit' });
 if (result.error) throw result.error;
 if (result.status !== 0) throw new Error(`${command} failed (${result.status})`);
}
try {
 run('pnpm', ['exec', 'redocly', 'lint', '--config', 'api/redocly.yaml', 'api/contract/openapi.yaml']);
 run('pnpm', ['exec', 'redocly', 'bundle', '--config', 'api/redocly.yaml', 'api/contract/openapi.yaml', '-o', join(temporary, 'openapi.yaml')]);
 run('pnpm', ['exec', 'openapi-typescript', join(temporary, 'openapi.yaml'), '-o', join(temporary, 'schema.ts')]);
 run('oapi-codegen', ['-config', resolve(root, 'api/codegen/go.yaml'), '-o', join(temporary, 'api.gen.go'), join(temporary, 'openapi.yaml')], join(root, 'backend'));
 for (const [file, target] of [['schema.ts', 'api/client/src/generated/schema.ts'], ['api.gen.go', 'backend/internal/httpapi/generated/api.gen.go']]) {
  if (check) { if (!readFileSync(join(temporary, file)).equals(readFileSync(join(root, target)))) throw new Error(`Generated file is stale: ${target}`); }
  else copyFileSync(join(temporary, file), join(root, target));
 }
} finally { rmSync(temporary, { recursive: true, force: true }); }
