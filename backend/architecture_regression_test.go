package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Run the real scanner on a miniature repository, so a new core package cannot
// bypass the same boundary merely by being absent from the named package map.
func TestUnregisteredBusinessContextBoundaries(t *testing.T) {
	scanner, err := os.ReadFile("architecture_test.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, source, dependency string
		denied                   bool
	}{
		{"new domain cannot import another domain", "business/future/domain", "business/repository/domain/entity", true},
		{"new application cannot import another use case", "business/future/application", "business/project/application", true},
		{"existing unlisted domain cannot import another domain", "business/membership/domain", "business/organization/domain/entity", true},
		{"approved issue repository scope", "business/issue/domain/entity", "business/repository/domain/entity/labels", false},
		{"known package cannot add another context", "business/issue/domain/entity", "business/organization/domain/entity", true},
		{"same context remains local", "business/future/application", "business/future/domain", false},
		{"approved ownership member reference", "business/organization/application", "business/membership/domain", false},
		{"enterprise cannot borrow organization core errors", "business/enterprise/application/command", "business/organization/domain/entity", true},
		{"enterprise owns its operation errors", "business/enterprise/application/command", "business/enterprise/domain", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			files := map[string][]byte{
				"go.mod":                              []byte("module boundary-fixture\n\ngo 1.27\n"),
				"architecture_test.go":                scanner,
				"internal/" + tc.source + "/probe.go": []byte("package probe\nimport _ \"github.com/963sup/line-mini-app-bot/internal/" + tc.dependency + "\"\n"),
			}
			for path, content := range files {
				target := filepath.Join(root, path)
				if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(target, content, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.CommandContext(t.Context(), "go", "test", "-count=1", "-run", "^TestDependencyBoundaries$", ".")
			cmd.Dir = root
			output, err := cmd.CombinedOutput()
			if tc.denied {
				if err == nil || !strings.Contains(string(output), "unapproved cross-domain dependency") {
					t.Fatalf("expected boundary rejection, got %v:\n%s", err, output)
				}
			} else if err != nil {
				t.Fatalf("valid boundary rejected: %v\n%s", err, output)
			}
		})
	}
}
