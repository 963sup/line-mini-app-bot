"""Regression checks for embedded dispatch, with broken-source mutations."""
import json
from pathlib import Path
import unittest

from check_dispatch import check_dispatch

ROOT = Path(__file__).resolve().parents[2]
CONTRACT = "backend/internal/httpapi/contract.go"
ROUTES = "backend/internal/httpapi/routes.go"
GENERATED = "backend/internal/httpapi/generated/api.gen.go"


class DispatchChecks(unittest.TestCase):
    def setUp(self):
        self.operations = json.loads((ROOT / "api/capabilities.json").read_text())["operations"]
        paths = {CONTRACT, ROUTES, GENERATED} | {o["adapter"] for o in self.operations}
        self.sources = {p: (ROOT / p).read_text() for p in paths}

    def check(self):
        check_dispatch(self.sources.__getitem__, self.operations)

    def change(self, path, old, new):
        self.assertIn(old, self.sources[path])
        self.sources[path] = self.sources[path].replace(old, new, 1)

    def test_embedded_dispatch(self):
        self.check()

    def test_comments_and_formatting(self):
        self.change(CONTRACT, "type contractServer struct {", "type /* struct docs */ contractServer\nstruct {")
        self.change(ROUTES, "IdentityServer: handler", "IdentityServer /* bound capability */ :\n handler")
        self.check()

    def test_missing_embedded_interface(self):
        self.change(CONTRACT, "\tIdentityServer\n", "")
        with self.assertRaisesRegex(ValueError, "coverage"):
            self.check()

    def test_missing_interface_definition_not_satisfied_by_comment(self):
        start = self.sources[CONTRACT].index("type IdentityServer interface {")
        end = self.sources[CONTRACT].index("\n}", start) + 2
        text = self.sources[CONTRACT][start:end]
        self.sources[CONTRACT] = self.sources[CONTRACT].replace(text, "/* " + text + " */")
        with self.assertRaisesRegex(ValueError, "Missing Go interface"):
            self.check()

    def test_missing_method_not_satisfied_by_string(self):
        self.change(CONTRACT, "\tGetMiniAppConfig(w http.ResponseWriter, r *http.Request)\n", "")
        self.sources[CONTRACT] += '\nconst example = "GetMiniAppConfig(w http.ResponseWriter, r *http.Request)"\n'
        with self.assertRaisesRegex(ValueError, "coverage"):
            self.check()

    def test_generated_signature_mismatch(self):
        self.change(CONTRACT, "GetMiniAppConfig(w http.ResponseWriter, r *http.Request)", "GetMiniAppConfig(w http.ResponseWriter, r string)")
        with self.assertRaisesRegex(ValueError, "signature mismatch"):
            self.check()

    def test_mixed_adapter_owner(self):
        self.change(CONTRACT, "\tGetMiniAppConfig(w http.ResponseWriter, r *http.Request)\n", "")
        self.change(CONTRACT, "type RepositoryServer interface {", "type RepositoryServer interface {\nGetMiniAppConfig(w http.ResponseWriter, r *http.Request)")
        with self.assertRaisesRegex(ValueError, "mixes adapter owners"):
            self.check()

    def test_missing_handler_binding_not_satisfied_by_string(self):
        self.change(ROUTES, "contractServer{IdentityServer: handler}", "contractServer{}")
        self.sources[ROUTES] = self.sources[ROUTES].replace("routes := routeHandler(contractServer{})", 'example := "routeHandler ( contractServer { IdentityServer : handler } )"; _ = example; routes := routeHandler(contractServer{})')
        with self.assertRaisesRegex(ValueError, "handler binding"):
            self.check()

    def test_missing_public_route(self):
        self.change(ROUTES, 'public.Handle("GET /api/v1/miniapp/config", routes)', '// route removed')
        with self.assertRaisesRegex(ValueError, "public route"):
            self.check()

    def test_missing_generated_handler(self):
        self.change(ROUTES, "return generated.HandlerWithOptions(server,", "return anotherHandler(server,")
        with self.assertRaisesRegex(ValueError, "generated handler composition"):
            self.check()

    def test_missing_adapter_method(self):
        path = next(o["adapter"] for o in self.operations if o["handler"] == "GetMiniAppConfig")
        self.change(path, "func (a *Authenticator) GetMiniAppConfig(", "func (a *Authenticator) RemovedConfig(")
        with self.assertRaisesRegex(ValueError, "receiver adapter"):
            self.check()


if __name__ == "__main__":
    unittest.main()
