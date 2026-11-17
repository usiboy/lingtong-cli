package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lingtong/cli/internal/openapi"
	"github.com/spf13/cobra"
)

func TestDashboardQueryUsesPostBody(t *testing.T) {
	for _, args := range [][]string{{"--body", `{"id":3001}`}, {"--id", "3001"}} {
		t.Run(args[0], func(t *testing.T) {
			called := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				if r.Method != http.MethodPost || r.URL.Path != "/gw/ai/proxy" || r.URL.RawQuery != "" {
					t.Errorf("unexpected query request: %s %s", r.Method, r.URL.String())
				}
				var request struct {
					Path   string `json:"path"`
					Method string `json:"method"`
					Body   struct {
						ID int `json:"id"`
					} `json:"body"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Path != "/dashboard/query" || request.Method != "POST" || request.Body.ID != 3001 {
					t.Errorf("query body: %+v, %v", request, err)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"success":true,"result":[]}`))
			}))
			defer server.Close()
			spec, err := openapi.EmbeddedSpec()
			if err != nil {
				t.Fatal(err)
			}
			root := &cobra.Command{Use: "test"}
			RegisterServiceCommands(root, newTestFactory(server.URL), spec)
			root.SetArgs(append([]string{"service", "dashboard", "query"}, args...))
			if err := root.Execute(); err != nil {
				t.Fatal(err)
			}
			if !called {
				t.Fatal("query was not sent")
			}
		})
	}
}
