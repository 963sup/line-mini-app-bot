package architecture

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
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
		// Issues use Repository-owned scope and label values, not its adapters.
		"business/issue/application/command": {"business/issue/application/port", "business/issue/domain/entity", "business/repository/domain/entity"},
		"business/issue/application/port":    {"business/issue/application/dto", "business/issue/domain/entity", "business/repository/domain/entity"},
		"business/issue/application/query":   {"business/issue/application/dto", "business/issue/application/port", "business/issue/domain/entity", "business/issue/domain/policy", "business/repository/domain/entity"},
		"business/issue/domain/entity":       {"business/issue/domain/policy", "business/repository/domain/entity", "business/repository/domain/entity/labels"},
		// Organization owns transfers; Membership owns opaque member references.
		"business/organization/application": {"business/organization/domain/entity", "business/membership/domain"},
		// Enterprise owns its operation errors; HTTP adapters preserve compatibility.
		"business/enterprise/application/command":         {"business/enterprise/application/dto", "business/enterprise/application/port", "business/enterprise/domain"},
		"business/workflow/application":                   {},
		"business/agent/application":                      {"business/agent/application/port"},
		"business/agent/application/port":                 {},
		"business/team/application":                       {"business/team/domain"},
		"linegroup/domain/entity":                         {},
		"linegroup/domain/event":                          {},
		"linegroup/domain/valueobject":                    {},
		"linegroup/application/dto":                       {"linegroup/domain/entity"},
		"linegroup/application/port":                      {"linegroup/application/dto", "linegroup/domain/entity", "linegroup/domain/event"},
		"linegroup/application/query":                     {"linegroup/application/dto", "linegroup/application/port", "linegroup/domain/entity", "linegroup/domain/valueobject"},
		"business/repository/application":                 {"business/repository/application/dto", "business/repository/application/port", "business/repository/domain/entity", "identity"},
		"business/repository/application/dto":             {"business/repository/application", "business/repository/application/port", "business/repository/domain/entity", "identity"},
		"business/repository/application/port":            {"business/repository/application", "business/repository/application/dto", "business/repository/domain/entity", "identity"},
		"business/repository/domain/entity":               {"identity"},
		"business/repository/domain/entity/wiki":          {"business/repository/domain/entity"},
		"business/repository/application/dto/space":       {"business/repository/application", "business/repository/application/dto", "business/repository/application/port", "business/repository/application/port/space", "business/repository/application/space", "business/repository/domain/entity", "business/repository/domain/entity/space"},
		"business/repository/application/port/space":      {"business/repository/application", "business/repository/application/dto", "business/repository/application/dto/space", "business/repository/application/port", "business/repository/application/space", "business/repository/domain/entity", "business/repository/domain/entity/space"},
		"business/repository/application/space":           {"business/repository/application", "business/repository/application/dto", "business/repository/application/dto/space", "business/repository/application/port", "business/repository/application/port/space", "business/repository/domain/entity", "business/repository/domain/entity/space"},
		"business/repository/domain/entity/space":         {"business/repository/domain/entity"},
		"business/repository/application/dto/milestones":  {"business/repository/application", "business/repository/application/dto", "business/repository/application/milestones", "business/repository/application/port", "business/repository/application/port/milestones", "business/repository/domain/entity", "business/repository/domain/entity/milestones"},
		"business/repository/application/milestones":      {"business/repository/application", "business/repository/application/dto", "business/repository/application/dto/milestones", "business/repository/application/port", "business/repository/application/port/milestones", "business/repository/domain/entity", "business/repository/domain/entity/milestones"},
		"business/repository/application/port/milestones": {"business/repository/application", "business/repository/application/dto", "business/repository/application/dto/milestones", "business/repository/application/milestones", "business/repository/application/port", "business/repository/domain/entity", "business/repository/domain/entity/milestones"},
		"business/repository/domain/entity/milestones":    {"business/repository/domain/entity"},
		"business/repository/application/port/statistics": {"business/repository/application", "business/repository/application/dto", "business/repository/application/port", "business/repository/application/statistics", "business/repository/domain/entity", "business/repository/domain/entity/statistics"},
		"business/repository/application/statistics":      {"business/repository/application", "business/repository/application/dto", "business/repository/application/port", "business/repository/application/port/statistics", "business/repository/domain/entity", "business/repository/domain/entity/statistics"},
		"business/repository/domain/entity/statistics":    {"business/repository/domain/entity"},
		"business/project/application":                    {"business/project/application/dto", "business/project/application/port", "business/project/domain/entity", "identity"},
		"business/project/application/dto":                {"business/project/application", "business/project/application/port", "business/project/domain/entity", "identity"},
		"business/project/application/port":               {"business/project/application", "business/project/application/dto", "business/project/domain/entity", "identity"},
		"business/project/domain/entity":                  {"linegroup/domain/entity", "linegroup/domain/valueobject", "identity"},
		"business/repository/application/dto/labels":      {"business/repository/application", "business/repository/application/dto", "business/repository/application/labels", "business/repository/application/port", "business/repository/application/port/labels", "business/repository/domain/entity", "business/repository/domain/entity/labels"},
		"business/repository/application/labels":          {"business/repository/application", "business/repository/application/dto", "business/repository/application/dto/labels", "business/repository/application/port", "business/repository/application/port/labels", "business/repository/domain/entity", "business/repository/domain/entity/labels"},
		"business/repository/application/port/labels":     {"business/repository/application", "business/repository/application/dto", "business/repository/application/dto/labels", "business/repository/application/labels", "business/repository/application/port", "business/repository/domain/entity", "business/repository/domain/entity/labels"},
		"business/repository/domain/entity/labels":        {"business/repository/domain/entity"},
		"business/discussion/domain/entity":               {"business/repository/domain/entity", "business/repository/domain/entity/labels"},
		"business/discussion/application":                 {"business/discussion/application/command", "business/discussion/application/port", "business/discussion/application/query"},
		"business/discussion/application/port":            {"business/discussion/application/dto", "business/discussion/domain/entity", "business/repository/application", "business/repository/application/dto", "business/repository/application/port", "business/repository/domain/entity"},
		"business/discussion/application/command":         {"business/discussion/application/dto", "business/discussion/application/port", "business/discussion/domain/entity", "business/repository/application", "business/repository/application/dto", "business/repository/application/port", "business/repository/domain/entity"},
		"business/discussion/application/query":           {"business/discussion/application/dto", "business/discussion/application/port", "business/discussion/domain/entity", "business/repository/application", "business/repository/application/dto", "business/repository/application/port", "business/repository/domain/entity"},
		"business/discussion/application/dto":             {"business/discussion/domain/entity"},
		"business/issue":                                  {"business/repository/application", "business/repository/application/dto", "business/repository/application/dto/labels", "business/repository/application/labels", "business/repository/application/port", "business/repository/application/port/labels", "business/repository/domain/entity", "business/repository/domain/entity/labels"},
		"business/knowledge/application/port":             {},

		"business/object/application":         {"business/object/application/command", "business/object/application/dto", "business/object/application/port", "business/object/application/query", "business/object/domain/entity", "business/object/domain/template", "business/operations/application", "business/repository/domain/entity"},
		"business/object/application/command": {"business/object/application/dto", "business/object/application/port", "business/object/domain/entity", "business/object/domain/service", "business/object/domain/template", "business/repository/domain/entity"},
		"business/object/application/query":   {"business/object/application/dto", "business/object/application/port", "business/object/domain/entity", "business/object/domain/service", "business/object/domain/template", "business/operations/application", "business/repository/domain/entity"},
		"business/object/application/dto":     {"business/object/domain/entity", "business/object/domain/template", "business/operations/application"},
		"business/object/application/port":    {"business/object/application/dto", "business/object/domain/entity", "business/object/domain/template", "business/operations/application", "business/repository/domain/entity"},
		"business/object/domain/template":     {},
		"business/object/domain/policy":       {},
		"business/object/domain/valueobject":  {},
		"business/object/domain/entity":       {"business/object/domain/valueobject", "business/object/domain/policy"},
		"business/object/domain/service":      {"business/object/domain/entity", "business/object/domain/template"},

		"business/attendance/domain/valueobject":              {},
		"business/attendance/domain/event":                    {},
		"business/attendance/domain/entity":                   {"business/attendance/domain/valueobject"},
		"business/attendance/domain/policy":                   {"business/attendance/domain/entity", "business/attendance/domain/valueobject"},
		"business/attendance/domain/service":                  {"business/attendance/domain/entity", "business/attendance/domain/event", "business/attendance/domain/valueobject"},
		"business/attendance/application/dto":                 {"business/attendance/domain/entity", "business/attendance/domain/service", "business/attendance/domain/valueobject"},
		"business/attendance/application/port":                {"business/attendance/application/dto", "business/attendance/domain/entity", "business/attendance/domain/policy", "business/attendance/domain/valueobject"},
		"business/attendance/application/command":             {"business/attendance/application/dto", "business/attendance/application/port", "business/attendance/domain/entity", "business/attendance/domain/policy", "business/attendance/domain/service", "business/attendance/domain/valueobject"},
		"business/attendance/application/query":               {"business/attendance/application/dto", "business/attendance/application/port", "business/attendance/domain/entity", "business/attendance/domain/service"},
		"business/summary/application/port":                   {},
		"business/summary/application/port/messagecollection": {"business/summary/domain/entity/messagecollection"},
		"business/summary/domain/entity/messagecollection":    {},
		"business/summary/application/port/summarization":     {"business/summary/application/port/messagecollection", "business/summary/application/summarization", "business/summary/domain/entity/messagecollection", "business/summary/domain/entity/summarization"},
		"business/summary/application/summarization":          {"business/summary/application/port/messagecollection", "business/summary/application/port/summarization", "business/summary/domain/entity/messagecollection", "business/summary/domain/entity/summarization"},
		"business/summary/domain/entity/summarization":        {"business/summary/domain/entity/messagecollection"},
		"business/weather/domain/entity/forecast":             {},
		"identity": {},
	}
	allowedTransport := map[string][]string{
		"transport/http/api":         {"transport/http", "transport/http/api/generated"},
		"transport/http/geolocation": {"transport/http/browser"},
		"transport/line/navigation":  {"transport/line"},
		"transport/line":             {"platform/line"},
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
		parts := strings.Split(dir, "/")
		businessCore := len(parts) >= 3 && parts[0] == "business" && (parts[2] == "application" || parts[2] == "domain")
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			dependency, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			if dir == "business/operations/application" && strings.HasPrefix(dependency, module) {
				t.Errorf("operation contract %s depends on business implementation %s", rel, dependency)
			}
			if businessCore {
				if !strings.HasPrefix(dependency, module) && (strings.Contains(strings.Split(dependency, "/")[0], ".") || dependency == "net/http" || dependency == "database/sql") {
					t.Errorf("%s imports infrastructure %s", rel, dependency)
				}
				target := strings.TrimPrefix(dependency, module)
				targetParts := strings.Split(target, "/")
				if strings.HasPrefix(dependency, module) && len(targetParts) >= 2 && targetParts[0] == "business" && targetParts[1] != parts[1] && !slices.Contains(core[dir], target) {
					t.Errorf("%s imports unapproved cross-domain dependency %s", rel, dependency)
				}
				if strings.HasPrefix(dependency, module) && (strings.Contains(target+"/", "/adapter/") || strings.HasPrefix(target+"/", "platform/") || strings.HasPrefix(target+"/", "transport/") || parts[2] == "domain" && strings.Contains(target+"/", "/application/")) {
					t.Errorf("%s imports outer layer %s", rel, dependency)
				}
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
			if strings.HasPrefix(dir, "business/summary/application/messagecollection") && strings.HasPrefix(dependency, module+"business/summary/application/summarization") {
				t.Errorf("message collection %s depends on summarization %s", rel, dependency)
			}
			if strings.Contains(dir+"/", "/adapter/http/") || strings.HasSuffix(dir, "/http") && dir != "transport/http" {
				for _, outgoing := range []string{"/firestore", "/cloudtasks", "/summarizer", "/cwa", "/poppler", "/openai", "/identity/line", "/identity/google"} {
					if strings.HasPrefix(dependency, module) && strings.Contains(dependency, outgoing) {
						t.Errorf("HTTP adapter %s constructs outgoing adapter %s", rel, dependency)
					}
				}
				for _, sdk := range []string{"cloud.google.com/", "google.golang.org/", "github.com/line/line-bot-sdk", "github.com/openai/"} {
					if strings.HasPrefix(dependency, sdk) {
						t.Errorf("HTTP adapter %s depends directly on SDK %s", rel, dependency)
					}
				}
			}
			if strings.HasPrefix(dir, "platform/") && strings.HasPrefix(dependency, module) && !strings.HasPrefix(dependency, module+"platform/") {
				t.Errorf("platform %s imports %s", rel, dependency)
			}
			if strings.HasPrefix(dependency, module+"bootstrap") {
				t.Errorf("%s imports composition root", rel)
			}
			if strings.HasPrefix(dir, "transport/") && strings.HasPrefix(dependency, module) && !slices.Contains(allowedTransport[dir], strings.TrimPrefix(dependency, module)) {
				t.Errorf("transport depends on application feature: %s", dependency)
			}
		}
		if _, isCore := core[dir]; isCore || businessCore {
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
