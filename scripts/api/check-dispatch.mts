import assert from "node:assert/strict";
import { dirname } from "node:path";

export interface DispatchOperation {
  operationId: string;
  audience: string;
  handler: string;
  adapter: string;
  method: string;
  path: string;
}

type Source = (path: string) => string;

function tokens(text: string): string[] {
  return (text.match(/\/\/[^\n]*|\/\*[\s\S]*?\*\/|"(?:\\.|[^"\\])*"|`[^`]*`|[A-Za-z_]\w*|\d+|[^\s]/g) ?? [])
    .filter(token => !token.startsWith("//") && !token.startsWith("/*"));
}

function structural(tokens: string[]): string {
  // Quoted code examples cannot satisfy structural checks.
  return tokens.map(token =>
    !/^["`]/.test(token) || token === '" "' ||
    /^"(?:(?:GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS) )?\/api\/[^"\\]*"$/.test(token)
      ? token : '"<literal>"').join(" ");
}

function group(tokens: string[], start: number, left = "{", right = "}"): [string[], number] {
  assert.equal(tokens[start], left);
  let depth = 0;
  for (let end = start; end < tokens.length; end++) {
    if (tokens[end] === left) depth++;
    if (tokens[end] === right) depth--;
    if (depth === 0) return [tokens.slice(start + 1, end), end + 1];
  }
  throw new Error("Unbalanced Go declaration");
}

function declaration(tokens: string[], name: string, kind: string): string[] {
  const index = tokens.findIndex((token, i) =>
    token === "type" && tokens[i + 1] === name && tokens[i + 2] === kind && tokens[i + 3] === "{");
  assert(index >= 0, `Missing Go ${kind}: ${name}`);
  return group(tokens, index + 3)[0];
}

function methods(body: string[]): Map<string, string[]> {
  const result = new Map<string, string[]>();
  for (let i = 0; i < body.length;) {
    const name = body[i];
    assert.equal(body[i + 1], "(", `Unsupported interface element: ${name}`);
    const [params, end] = group(body, i + 1, "(", ")");
    i = end;
    const signature = params.join(" ").split(",").map(field => field.trim()).filter(Boolean).map(field => {
      const parts = field.split(/\s+/);
      assert(parts.length >= 2, `Unsupported parameter in ${name}`);
      return parts.slice(1).join(" ").replaceAll("generated . ", "");
    });
    assert(!result.has(name), `Duplicate interface method: ${name}`);
    result.set(name, signature);
  }
  return result;
}

function functions(tokens: string[]) {
  const result = new Map<string, { params: string; body: string }>();
  for (let i = 0; i + 2 < tokens.length; i++) {
    if (tokens[i] !== "func" || tokens[i + 2] !== "(") continue;
    const [params, end] = group(tokens, i + 2, "(", ")");
    const start = tokens.indexOf("{", end);
    if (start >= 0) result.set(tokens[i + 1], { params: params.join(" "), body: structural(group(tokens, start)[0]) });
  }
  return result;
}

export function checkDispatch(source: Source, operations: DispatchOperation[]) {
  const contract = tokens(source("backend/internal/transport/http/api/contract.go"));
  const routes = functions(tokens(source("backend/internal/transport/http/api/routes.go")));
  const generated = methods(declaration(tokens(source("backend/internal/transport/http/api/generated/api.gen.go")), "ServerInterface", "interface"));
  const embedded = declaration(contract, "contractServer", "struct");
  assert(embedded.length && embedded.every(name => /^[A-Za-z_]\w*$/.test(name)),
    "contractServer must embed named capability interfaces");
  assert.equal(new Set(embedded).size, embedded.length, "Duplicate embedded capability interface");

  const publicOps = operations.filter(operation => operation.audience === "public");
  const byHandler = new Map(publicOps.map(operation => [operation.handler, operation]));
  assert.equal(byHandler.size, publicOps.length, "Duplicate public handler");
  const seen = new Set<string>();

  for (const capability of embedded) {
    const members = methods(declaration(contract, capability, "interface"));
    assert(members.size, `Empty capability interface: ${capability}`);
    const owners = new Set<string>();
    for (const [name, signature] of members) {
      const operation = byHandler.get(name);
      assert(operation && !seen.has(name), `Unowned or duplicate dispatch method: ${name}`);
      assert.deepEqual(generated.get(name), signature, `Generated signature mismatch: ${name}`);
      const adapter = structural(tokens(source(operation.adapter)));
      const match = adapter.match(new RegExp(String.raw`func \( \w+ (?:\* )?(\w+) \) ${name} \(`));
      assert(match, `Missing receiver adapter: ${name}`);
      owners.add(`${dirname(operation.adapter)}/${match[1]}`);
      seen.add(name);
    }
    assert.equal(owners.size, 1, `Capability interface mixes adapter owners: ${capability}`);
    let bindings = 0;
    for (const [name, { params, body }] of routes) {
      const parameter = params.match(new RegExp(String.raw`(?:^|, )([A-Za-z_]\w*) ${capability}(?: ,|$)`));
      if (!parameter) continue;
      assert(body.includes(`routeHandler ( contractServer { ${capability} : ${parameter[1]} } )`),
        `Missing capability handler binding: ${name}`);
      assert(body.includes("public . Handle (") && body.includes(", routes )"),
        `Missing public route registration: ${name}`);
      const prefixes = [...body.matchAll(/strings \. HasPrefix \( path , "([^"]+)" \)/g)];
      for (const operation of publicOps.filter(operation => members.has(operation.handler))) {
        const explicit = `public . Handle ( "${operation.method} ${operation.path}" , routes )`;
        const dynamic = prefixes.some(match => operation.path.startsWith(match[1])) &&
          ['Spec ( )', 'spec . Paths . Map ( )', 'item . Operations ( )',
            'public . Handle ( method + " " + path , routes )'].every(part => body.includes(part));
        assert(body.includes(explicit) || dynamic, `Missing capability public route: ${operation.operationId}`);
      }
      bindings++;
    }
    assert.equal(bindings, 1, `Missing or ambiguous route composition: ${capability}`);
  }
  assert.deepEqual(seen, new Set(byHandler.keys()), "Incomplete public dispatch coverage");
  assert.deepEqual(seen, new Set(generated.keys()), "Incomplete generated dispatch coverage");
  const handler = routes.get("routeHandler");
  assert(handler?.params === "server contractServer" &&
    handler.body.includes("return generated . HandlerWithOptions ( server ,"), "Missing generated handler composition");
  assert(contract.join(" ").includes("var _ generated . ServerInterface = contractServer { }"),
    "Missing generated interface compile-time assertion");
}
