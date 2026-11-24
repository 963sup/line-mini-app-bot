package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestLINEFlexUsesAreDocumentedExceptions(t *testing.T) {
	allowed := map[string]bool{
		"business/attendance/adapter/line/attendance_history.go": true,
		"business/attendance/adapter/line/board.go":              true,
	}
	root := filepath.Join("internal")
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		ast.Inspect(file, func(node ast.Node) bool {
			literal, ok := node.(*ast.CompositeLit)
			if !ok {
				return true
			}
			selector, ok := literal.Type.(*ast.SelectorExpr)
			if ok && selector.Sel.Name == "FlexMessage" && !allowed[rel] {
				t.Errorf("%s creates an undocumented Flex message; use a native template or record the concrete limitation in docs/standards/line-interactions.md", rel)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestLINEResponsibilitiesHaveSingleAuthorities(t *testing.T) {
	const module = "github.com/963sup/line-mini-app-bot/internal/"
	root := filepath.Join("internal")
	outbound := map[string]bool{
		"ReplyMessageWithHttpInfo": true, "PushMessageWithHttpInfo": true,
		"CreateRichMenu": true, "CreateRichMenuAlias": true, "DeleteRichMenu": true,
		"RichMenuBatchWithHttpInfo": true, "SetDefaultRichMenu": true,
		"SetRichMenuImage": true, "UpdateRichMenuAlias": true,
	}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		dir := filepath.ToSlash(filepath.Dir(rel))
		parts := strings.Split(dir, "/")
		for _, spec := range file.Imports {
			dependency, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			target := strings.TrimPrefix(dependency, module)
			if strings.HasPrefix(dir+"/", "platform/line/") && (strings.HasPrefix(target, "transport/") || strings.HasPrefix(target, "business/")) {
				t.Errorf("%s reverses the LINE platform dependency toward %s", rel, target)
			}
			if strings.HasPrefix(dir+"/", "transport/line/") && strings.HasPrefix(target, "business/") {
				t.Errorf("%s makes LINE transport depend on business code %s", rel, target)
			}
			if len(parts) >= 4 && parts[0] == "business" && parts[2] == "adapter" && parts[3] == "line" && strings.HasPrefix(target, "business/") {
				targetParts := strings.Split(target, "/")
				if len(targetParts) > 1 && targetParts[1] != parts[1] {
					t.Errorf("%s couples the %s LINE adapter to the %s context", rel, parts[1], targetParts[1])
				}
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			if outbound[selector.Sel.Name] && !strings.HasPrefix(dir+"/", "platform/line/") {
				t.Errorf("%s performs LINE outbound operation %s outside platform/line", rel, selector.Sel.Name)
			}
			if selector.Sel.Name == "ParseQuery" && rel != "transport/line/postback.go" {
				t.Errorf("%s parses postback query data outside transport/line/postback.go", rel)
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
