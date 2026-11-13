// Package browser serves the separately built Next.js static export.
package browser

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Root is explicit in containers; local commands run at either repository or backend root.
func Root() string {
	if root := os.Getenv("FRONTEND_DIR"); root != "" {
		return root
	}
	for _, root := range []string{"../frontend/miniapp/out", "frontend/miniapp/out", "../../../frontend/miniapp/out"} {
		if _, err := os.Stat(filepath.Join(root, "miniapp.html")); err == nil {
			return root
		}
	}
	// Package tests may run below backend/internal; use ancestor discovery only for local source trees.
	cwd, _ := os.Getwd()
	for dir := cwd; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
		root := filepath.Join(dir, "frontend", "miniapp", "out")
		if _, err := os.Stat(filepath.Join(root, "miniapp.html")); err == nil {
			return root
		}
	}
	return "../frontend/miniapp/out"
}

func ReadFile(name string) ([]byte, error) { return fs.ReadFile(os.DirFS(Root()), name) }

func Validate() error {
	for _, name := range []string{"miniapp.html", "miniapp/history.html", "miniapp/correction.html", "miniapp/review.html", "miniapp/punch.html", "miniapp/leave.html", "miniapp/journal.html", "miniapp/repositories.html", "miniapp/projects.html", "location.html"} {
		if _, err := ReadFile(name); err != nil {
			return fmt.Errorf("frontend artifact %s: %w (run pnpm build)", name, err)
		}
	}
	return nil
}

var scripts = regexp.MustCompile(`(?s)<script([^>]*)>(.*?)</script>`)

// Static Next.js bootstrap scripts receive exact-content hashes, never unsafe-inline.
func ScriptPolicy(html []byte) string {
	policy := "'self' https://static.line-scdn.net"
	for _, match := range scripts.FindAllSubmatch(html, -1) {
		if strings.Contains(string(match[1]), "src=") || len(match[2]) == 0 {
			continue
		}
		digest := sha256.Sum256(match[2])
		policy += " 'sha256-" + base64.StdEncoding.EncodeToString(digest[:]) + "'"
	}
	return policy
}

// Assets exposes only regular build files under the registered assets/_next prefixes.
func Assets() http.Handler {
	root := Root()
	files := http.FileServer(http.Dir(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		info, err := fs.Stat(os.DirFS(root), name)
		if err != nil || !info.Mode().IsRegular() {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		files.ServeHTTP(w, r)
	})
}
