package cloudstorage

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/api/option"
	storage "google.golang.org/api/storage/v1"
)

func TestSourceIsolationAndIntegrity(t *testing.T) {
	data := []byte("source bytes")
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	var corrupt bool
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method == "POST" {
			if r.URL.Query().Get("ifGenerationMatch") != "0" {
				t.Error("missing create-only precondition")
			}
			w.WriteHeader(412)
			return
		}
		if corrupt {
			_, _ = w.Write([]byte("wrong version"))
			return
		}
		_, _ = w.Write(data)
	}))
	defer server.Close()
	ctx := context.Background()
	client, err := storage.NewService(ctx, option.WithEndpoint(server.URL+"/"), option.WithoutAuthentication())
	if err != nil {
		t.Fatal(err)
	}
	s := Sources{client, "test", "channel"}
	key, _ := s.key("repo", "actor", digest)
	for _, other := range [][2]string{{"repo", "other"}, {"other", "actor"}} {
		k, _ := s.key(other[0], other[1], digest)
		if k == key {
			t.Fatal("scope collision")
		}
	}
	if err := s.Put(ctx, "repo", "actor", digest, []byte("wrong")); err == nil || calls != 0 {
		t.Fatal("accepted mismatched content")
	}
	if err := s.Put(ctx, "repo", "actor", digest, data); err != nil {
		t.Fatal(err)
	}
	if result, err := s.Read(ctx, "repo", "actor", digest); err != nil || string(result) != string(data) {
		t.Fatal(result, err)
	}
	corrupt = true
	if _, err = s.Read(ctx, "repo", "actor", digest); err == nil {
		t.Fatal("accepted corrupted data")
	}
}
