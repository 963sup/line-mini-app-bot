"""Fail-closed structural checks for the embedded Go API dispatcher.

This is a deliberately small Go token reader, not a type checker. It ignores
comments/formatting, checks method signatures and the supported route composition,
and leaves full Go type checking to the generated interface assertion and build.
No Go executable is required by this Python/Node preparation check.
"""
import re

TOKEN = re.compile(r'//[^\n]*|/\*[\s\S]*?\*/|"(?:\\.|[^"\\])*"|`[^`]*`|[A-Za-z_][A-Za-z_0-9]*|\d+|[^\s]')


def tokens(text):
    return [m.group() for m in TOKEN.finditer(text) if not m.group().startswith(("//", "/*"))]


def structural(ts):
    # Keep only exact public path literals needed for routing. A quoted code
    # example must never satisfy a binding/receiver/composition requirement.
    return " ".join(t if not t.startswith(('"', '`')) or t == '" "'
                    or re.fullmatch(r'"(?:(?:GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS) )?/api/[^"\\]*"', t)
                    else '"<literal>"' for t in ts)


def group(ts, start, left="{", right="}"):
    assert ts[start] == left
    depth = 0
    for end in range(start, len(ts)):
        depth += ts[end] == left
        depth -= ts[end] == right
        if depth == 0:
            return ts[start + 1:end], end + 1
    raise ValueError("Unbalanced Go declaration")


def declaration(ts, name, kind):
    for i in range(len(ts) - 3):
        if ts[i:i + 4] == ["type", name, kind, "{"]:
            return group(ts, i + 3)[0]
    raise ValueError(f"Missing Go {kind}: {name}")


def methods(body):
    result = {}
    i = 0
    while i < len(body):
        name = body[i]
        if i + 1 >= len(body) or body[i + 1] != "(":
            raise ValueError(f"Unsupported interface element: {name}")
        params, i = group(body, i + 1, "(", ")")
        # These generated HTTP interfaces have named parameters and no results.
        # Reject unknown forms rather than silently accepting partial coverage.
        signature = []
        for field in " ".join(params).split(","):
            field = field.strip().split()
            if not field:
                continue
            if len(field) < 2:
                raise ValueError(f"Unsupported parameter in {name}")
            signature.append(" ".join(field[1:]).replace("generated . ", ""))
        if name in result:
            raise ValueError(f"Duplicate interface method: {name}")
        result[name] = tuple(signature)
    return result


def functions(ts):
    result = {}
    for i in range(len(ts) - 2):
        if ts[i] != "func" or ts[i + 2] != "(":
            continue  # receiver methods are checked separately via the inventory
        params, end = group(ts, i + 2, "(", ")")
        while end < len(ts) and ts[end] != "{":
            end += 1
        if end < len(ts):
            result[ts[i + 1]] = (" ".join(params), structural(group(ts, end)[0]))
    return result


def check_dispatch(source, operations):
    contract = tokens(source("backend/internal/httpapi/contract.go"))
    routes = functions(tokens(source("backend/internal/httpapi/routes.go")))
    generated = methods(declaration(tokens(source("backend/internal/httpapi/generated/api.gen.go")), "ServerInterface", "interface"))
    embedded = declaration(contract, "contractServer", "struct")
    if not embedded or any(not re.fullmatch(r"[A-Za-z_]\w*", t) for t in embedded):
        raise ValueError("contractServer must embed named capability interfaces")
    if len(set(embedded)) != len(embedded):
        raise ValueError("Duplicate embedded capability interface")
    public = {o["handler"]: o for o in operations if o["audience"] == "public"}
    if len(public) != sum(o["audience"] == "public" for o in operations):
        raise ValueError("Duplicate public handler")
    seen = set()
    for interface in embedded:
        members = methods(declaration(contract, interface, "interface"))
        if not members:
            raise ValueError(f"Empty capability interface: {interface}")
        owners = set()
        for name, signature in members.items():
            if name not in public or name in seen:
                raise ValueError(f"Unowned or duplicate dispatch method: {name}")
            if generated.get(name) != signature:
                raise ValueError(f"Generated signature mismatch: {name}")
            operation = public[name]
            adapter = structural(tokens(source(operation["adapter"])))
            match = re.search(r"func \( \w+ (?:\* )?(\w+) \) " + re.escape(name) + r" \(", adapter)
            if not match:
                raise ValueError(f"Missing receiver adapter: {name}")
            owners.add((operation["adapter"].rsplit("/", 1)[0], match[1]))
            seen.add(name)
        if len(owners) != 1:
            raise ValueError(f"Capability interface mixes adapter owners: {interface}")
        bindings = []
        for name, (params, body) in routes.items():
            parameter = re.search(r"(?:^|, )([A-Za-z_]\w*) " + re.escape(interface) + r"(?: ,|$)", params)
            if not parameter:
                continue
            binding = f"routeHandler ( contractServer {{ {interface} : {parameter[1]} }} )"
            if binding not in body:
                raise ValueError(f"Missing capability handler binding: {name}")
            if "public . Handle (" not in body or ", routes )" not in body:
                raise ValueError(f"Missing public route registration: {name}")
            # Explicit routes, or the generated attendance-style path iterator.
            for method in members:
                operation = public[method]
                path = operation["path"]
                explicit = f'public . Handle ( "{operation["method"]} {path}" , routes )'
                prefixes = re.findall(r'strings \. HasPrefix \( path , "([^"]+)" \)', body)
                dynamic = (any(path.startswith(p) for p in prefixes)
                           and 'Spec ( )' in body and 'spec . Paths . Map ( )' in body
                           and 'item . Operations ( )' in body
                           and 'public . Handle ( method + " " + path , routes )' in body)
                if explicit not in body and not dynamic:
                    raise ValueError(f"Missing capability public route: {operation['operationId']}")
            bindings.append(name)
        if len(bindings) != 1:
            raise ValueError(f"Missing or ambiguous route composition: {interface}")
    if seen != set(public) or seen != set(generated):
        raise ValueError(f"Incomplete generated dispatch coverage: {sorted(set(public) ^ seen)}")
    handler = routes.get("routeHandler", ("", ""))
    if handler[0] != "server contractServer" or 'return generated . HandlerWithOptions ( server ,' not in handler[1]:
        raise ValueError("Missing generated handler composition")
    if 'var _ generated . ServerInterface = contractServer { }' not in " ".join(contract):
        raise ValueError("Missing generated interface compile-time assertion")
