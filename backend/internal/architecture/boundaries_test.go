package architecture

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

func TestDependencyBoundaries(t *testing.T) {
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	core := map[string][]string{
		"business/group":                          {},
		"business/repository":                     {"identity"},
		"business/repository/space":               {"business/repository"},
		"business/repository/milestones":          {"business/repository"},
		"business/repository/statistics":          {"business/repository"},
		"business/project":                        {"business/group/application/query", "business/group/domain/entity", "business/group/domain/valueobject", "identity"},
		"business/repository/labels":              {"business/repository"},
		"business/discussion/domain/entity":       {"business/repository", "business/repository/labels"},
		"business/discussion/application":         {"business/discussion/application/command", "business/discussion/application/port", "business/discussion/application/query"},
		"business/discussion/application/port":    {"business/discussion/application/dto", "business/discussion/domain/entity", "business/repository"},
		"business/discussion/application/command": {"business/discussion/application/dto", "business/discussion/application/port", "business/discussion/domain/entity", "business/repository"},
		"business/discussion/application/query":   {"business/discussion/application/dto", "business/discussion/application/port", "business/discussion/domain/entity", "business/repository"},
		"business/discussion/application/dto":     {"business/discussion/domain/entity"},
		"business/issue":                          {"business/repository", "business/repository/labels"},
		"business/objectimport/automatic":         {"business/object/application", "business/object/application/dto", "business/object/application/port", "business/object/domain/entity", "business/objectimport", "business/repository"},
		"business/objectimport/parsing":           {"business/objectimport", "business/repository"},
		"business/objectimport":                   {"business/object/application", "business/object/application/dto", "business/object/application/port", "business/object/domain/entity", "business/repository"},
		"business/object/application":             {"business/object/application/dto", "business/object/application/port", "business/object/domain/entity", "business/repository"},
		"business/object/application/dto":         {"business/object/application", "business/object/application/port", "business/object/domain/entity", "business/repository"},
		"business/object/application/port":        {"business/object/application", "business/object/application/dto", "business/object/domain/entity", "business/repository"},
		"business/object/domain/entity":           {"business/repository"},
		"business/subobject":                      {"business/object/application", "business/object/application/dto", "business/object/application/port", "business/object/domain/entity", "business/repository"},
		"business/attendance/domain/valueobject":  {},
		"business/attendance/domain/event":        {},
		"business/attendance/domain/entity":       {"business/attendance/domain/valueobject"},
		"business/attendance/domain/policy":       {"business/attendance/domain/entity", "business/attendance/domain/valueobject"},
		"business/attendance/domain/service":      {"business/attendance/domain/entity", "business/attendance/domain/event", "business/attendance/domain/valueobject"},
		"business/attendance/application/dto":     {"business/attendance/domain/entity", "business/attendance/domain/service", "business/attendance/domain/valueobject"},
		"business/attendance/application/port":    {"business/attendance/application/dto", "business/attendance/domain/entity", "business/attendance/domain/policy", "business/attendance/domain/valueobject"},
		"business/attendance/application/command": {"business/attendance/application/dto", "business/attendance/application/port", "business/attendance/domain/entity", "business/attendance/domain/policy", "business/attendance/domain/service", "business/attendance/domain/valueobject"},
		"business/attendance/application/query":   {"business/attendance/application/dto", "business/attendance/application/port", "business/attendance/domain/entity", "business/attendance/domain/service"},
		"business/summary":                        {},
		"business/summary/messagecollection":      {},
		"business/summary/summarization":          {"business/summary/messagecollection"},
		"business/weather/forecast":               {},
		"identity":                                {},
	}
	const module = "github.com/963sup/line-mini-app-bot/internal/"
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		dir := filepath.ToSlash(filepath.Dir(rel))
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			dependency, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			// Every attendance package, including a newly introduced subdomain, must
			// keep framework and adapter dependencies outside the inner layers.
			attendance := "business/attendance/"
			for _, layer := range []string{"domain", "application"} {
				inside := dir == attendance+layer || strings.HasPrefix(dir, attendance+layer+"/")
				if !inside {
					continue
				}
				if _, known := core[dir]; !known {
					t.Errorf("unclassified attendance %s package: %s", layer, dir)
				}
				internal := strings.HasPrefix(dependency, module+attendance+"domain/")
				if layer == "application" {
					internal = internal || strings.HasPrefix(dependency, module+attendance+"application/")
				}
				external := strings.Contains(strings.Split(dependency, "/")[0], ".")
				if external && !internal || dependency == "net/http" || dependency == "database/sql" {
					t.Errorf("%s imports %s outside the attendance %s boundary", rel, dependency, layer)
				}
			}
			if allowedDependencies, isCore := core[dir]; isCore {
				allowed := false
				for _, target := range allowedDependencies {
					if dependency == module+target {
						allowed = true
						break
					}
				}
				if strings.Contains(strings.Split(dependency, "/")[0], ".") && !allowed {
					t.Errorf("core %s imports non-core implementation %s", rel, dependency)
				}
			}
			if strings.HasPrefix(dir, "business/summary/messagecollection") && strings.HasPrefix(dependency, module+"business/summary/summarization") {
				t.Errorf("message collection %s depends on summarization %s", rel, dependency)
			}
			if strings.HasSuffix(dir, "/http") && dir != "transport/http" {
				for _, outgoing := range []string{"/firestore", "/cloudtasks", "/summarizer", "/cwa", "/identity/line", "/identity/google"} {
					if strings.HasPrefix(dependency, module) && strings.Contains(dependency, outgoing) {
						t.Errorf("HTTP adapter %s constructs outgoing adapter %s", rel, dependency)
					}
				}
				for _, sdk := range []string{"cloud.google.com/", "google.golang.org/", "github.com/line/line-bot-sdk", "github.com/sashabaranov/"} {
					if strings.HasPrefix(dependency, sdk) {
						t.Errorf("HTTP adapter %s depends directly on SDK %s", rel, dependency)
					}
				}
			}
			if dir == "httpapi" && strings.HasPrefix(dependency, module) && dependency != module+"httpapi/generated" && dependency != module+"transport/http" {
				t.Errorf("OpenAPI wiring depends on feature implementation: %s", dependency)
			}
			if strings.HasPrefix(dir, "transport/") && strings.HasPrefix(dependency, module) {
				t.Errorf("transport depends on application feature: %s", dependency)
			}
		}
		if _, isCore := core[dir]; isCore {
			ast.Inspect(file, func(node ast.Node) bool {
				field, ok := node.(*ast.Field)
				if ok && field.Tag != nil {
					tag, _ := strconv.Unquote(field.Tag.Value)
					if strings.Contains(tag, "firestore:") || strings.Contains(tag, "json:") {
						t.Errorf("core %s owns serialization tag %s", rel, tag)
					}
				}
				return true
			})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
