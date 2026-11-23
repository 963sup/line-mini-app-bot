import assert from 'node:assert/strict';
import { test } from 'node:test';
import { incompatibleOperations } from '../contract/check-compatibility.mts';

const original = { paths: { '/items': { post: { requestBody: { schema: { type: 'object', properties: { description: { type: 'string' } }, required: ['description'] } }, responses: { '200': { schema: { type: 'string', enum: ['a', 'b'] } } } } } } };
test('adding an operation and changing prose preserves existing contract', () => {
  const after = structuredClone(original);
  Object.assign(after.paths, { '/new': { get: { description: 'new' } } });
  Object.assign(after.paths['/items'].post, { summary: 'new prose' });
  assert.deepEqual(incompatibleOperations(original, after), []);
});
test('removed operations, changed fields named description and response enums fail closed', () => {
  assert.deepEqual(incompatibleOperations(original, { paths: {} }), ['POST /items']);
  const field = structuredClone(original);
  field.paths['/items'].post.requestBody.schema.properties.description.type = 'number';
  assert.deepEqual(incompatibleOperations(original, field), ['POST /items']);
  const response = structuredClone(original);
  response.paths['/items'].post.responses['200'].schema.enum.pop();
  assert.deepEqual(incompatibleOperations(original, response), ['POST /items']);
});
test('authentication changes are compared even without operation changes', () => {
  assert.deepEqual(incompatibleOperations({ security: [{ bearer: [] }] }, { security: [] }), ['security']);
  assert.deepEqual(incompatibleOperations({ components: { securitySchemes: { bearer: { type: 'http' } } } }, {}), ['securitySchemes']);
});
