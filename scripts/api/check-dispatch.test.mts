import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { test } from "node:test";
import { checkDispatch, type DispatchOperation } from "./check-dispatch.mts";

const root = resolve(import.meta.dirname, "../..");
const contract = "backend/internal/transport/http/api/contract.go";
const routes = "backend/internal/transport/http/api/routes.go";
const generated = "backend/internal/transport/http/api/generated/api.gen.go";
const operations: DispatchOperation[] = JSON.parse(readFileSync(resolve(root, "api/capabilities.json"), "utf8")).operations;
const original = new Map([...new Set([contract, routes, generated, ...operations.map(op => op.adapter)])]
  .map(path => [path, readFileSync(resolve(root, path), "utf8")]));

function scenario(name: string, mutate: (change: (path: string, before: string, after: string) => void) => void,
  expected?: RegExp) {
  test(name, () => {
    const sources = new Map(original);
    const source = (path: string) => {
      const text = sources.get(path);
      assert(text !== undefined, `Missing test source: ${path}`);
      return text;
    };
    mutate((path, before, after) => {
      const text = source(path);
      assert(text.includes(before), `Missing fixture text: ${before}`);
      sources.set(path, text.replace(before, () => after));
    });
    const check = () => checkDispatch(source, operations);
    if (expected) assert.throws(check, expected);
    else check();
  });
}

const method = "GetMiniAppConfig(w http.ResponseWriter, r *http.Request)";
scenario("embedded dispatch", () => {});
scenario("comments and formatting", change => {
  change(contract, "type contractServer struct {", "type /* struct docs */ contractServer\nstruct {");
  change(routes, "IdentityServer: handler", "IdentityServer /* bound capability */ :\n handler");
});
scenario("missing embedded interface", change => {
  change(contract, "\tIdentityServer\n", "");
}, /coverage/);
scenario("a comment cannot replace an interface declaration", change => {
  const declaration = original.get(contract)?.match(/type IdentityServer interface \{[\s\S]*?\n\}/)?.[0];
  assert(declaration);
  change(contract, declaration, `/* ${declaration} */`);
}, /Missing Go interface/);
scenario("a string cannot replace an interface method", change => {
  change(contract, `\t${method}\n`, "");
  change(contract, "type contractServer struct {", `const example = "${method}"\ntype contractServer struct {`);
}, /coverage/);
scenario("generated signature mismatch", change => {
  change(contract, method, "GetMiniAppConfig(w http.ResponseWriter, r string)");
}, /signature mismatch/);
scenario("mixed adapter owner", change => {
  change(contract, `\t${method}\n`, "");
  change(contract, "type RepositoryServer interface {", `type RepositoryServer interface {\n${method}`);
}, /mixes adapter owners/);
scenario("a string cannot replace a handler binding", change => {
  change(routes, "routes := routeHandler(contractServer{IdentityServer: handler})",
    'example := "routeHandler ( contractServer { IdentityServer : handler } )"; _ = example; routes := routeHandler(contractServer{})');
}, /handler binding/);
scenario("missing public route", change => {
  change(routes, 'public.Handle("GET /api/v1/miniapp/config", routes)', "// route removed");
}, /public route/);
scenario("missing generated handler", change => {
  change(routes, "return generated.HandlerWithOptions(server,", "return anotherHandler(server,");
}, /generated handler composition/);
scenario("missing adapter method", change => {
  const operation = operations.find(op => op.handler === "GetMiniAppConfig");
  assert(operation);
  change(operation.adapter, "func (a *Authenticator) GetMiniAppConfig(", "func (a *Authenticator) RemovedConfig(");
}, /receiver adapter/);
