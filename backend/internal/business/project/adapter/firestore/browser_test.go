package firestore_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	db "cloud.google.com/go/firestore"
	orgdb "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/firestore"
	orghttp "github.com/963sup/line-mini-app-bot/internal/business/organization/adapter/http"
	projectdb "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/firestore"
	projecthttp "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/http"
	sources "github.com/963sup/line-mini-app-bot/internal/business/project/adapter/repository"
	projectapp "github.com/963sup/line-mini-app-bot/internal/business/project/application"
	repodb "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/firestore"
	repohttp "github.com/963sup/line-mini-app-bot/internal/business/repository/adapter/http"
	repoapp "github.com/963sup/line-mini-app-bot/internal/business/repository/application"
	"github.com/google/uuid"
)

func TestProjectSettingsBrowserRoundTrip(t *testing.T) {
	out := os.Getenv("MINIAPP_BROWSER_EXPORT")
	if out == "" {
		t.Skip("set MINIAPP_BROWSER_EXPORT to the built miniapp out directory")
	}
	host, _, err := net.SplitHostPort(os.Getenv("FIRESTORE_EMULATOR_HOST"))
	if err != nil || !(host == "localhost" || net.ParseIP(host).IsLoopback()) {
		t.Fatal("requires a loopback Firestore emulator")
	}
	if _, err := os.Stat(filepath.Join(out, "miniapp", "projects.html")); err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()
	client, err := db.NewClient(ctx, "project-browser-test")
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	channel := uuid.NewString()
	actor := "U" + strings.Repeat("a", 32)
	orgs := orgdb.Store{DB: client, Channel: channel}
	org, err := orgs.Create(ctx, actor, "Audit team")
	if err != nil {
		t.Fatal(err)
	}
	store := &projectdb.Store{Client: client, ChannelID: channel}
	repos := &repoapp.Service{Store: &repodb.Store{Client: client, ChannelID: channel}}
	mux := http.NewServeMux()
	projecthttp.Register(mux, &projectapp.Service{AuthorizeOrganization: store.AuthorizeOrganization, Store: store, Sources: sources.Sources{Repositories: repos}}, auth{})
	repohttp.Register(mux, repos, auth{})
	organizations := orghttp.Handler{Store: orgs, Auth: auth{}}
	mux.HandleFunc("GET /api/v1/me/organization-context", organizations.GetOrganizationContext)
	mux.HandleFunc("GET /miniapp/projects", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(out, "miniapp", "projects.html"))
	})
	mux.Handle("/", http.FileServer(http.Dir(out)))
	server := httptest.NewServer(mux)
	defer server.Close()
	cmd := exec.CommandContext(ctx, "python3", "testdata/project_browser.py", server.URL, actor, org.ID)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, output)
	}
}
