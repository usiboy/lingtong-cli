// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package table

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/lingtong/cli/internal/client"
	lterrors "github.com/lingtong/cli/internal/errors"
	"github.com/spf13/cobra"
)

func importFixture(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "订单 +&.csv")
	if err := os.WriteFile(path, []byte("商品,数量\n苹果,2\n"), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestImportCreatesTableAndWaits(t *testing.T) {
	var stages []string
	polls := 0
	fileURL := "https://example.com/订单 +&.csv?x=1&y=2"
	columns := []interface{}{map[string]interface{}{"id": float64(0), "title": "商品", "type": "text", "key": "item", "show": true, "primaryKey": false}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := decodeProxyRequest(t, r)
		stages = append(stages, p.Path)
		if r.Method != "POST" || r.URL.Path != "/gw/ai/proxy" || r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("wrong authenticated proxy transport")
		}
		var result interface{}
		switch p.Path {
		case "/basicdata/record/import/columnsSchema":
			if p.Method != "POST" || r.URL.Query().Get("fileUrl") != fileURL || p.Body != nil || p.Params != nil {
				t.Error("schema parsing query lost")
			}
			result = map[string]interface{}{"columnsSchema": columns, "total": 1}
		case "/basicdata/save":
			if p.Body["name"] != "订单 +&" || p.Body["appId"] != float64(165) || p.Body["source"] != float64(1) || p.Body["openHighMode"] != float64(0) {
				t.Errorf("bad table config: %v", p.Body)
			}
			result = map[string]interface{}{"id": 123}
		case "/basicdata/schema/save":
			if p.Body["basicDataId"] != float64(123) || !reflect.DeepEqual(p.Body["columnsSchema"], columns) {
				t.Errorf("inferred schema changed: %v", p.Body)
			}
			result = map[string]interface{}{"id": 456}
		case "/basicdata/record/import":
			if p.Method != "POST" || r.URL.Query().Get("fileUrl") != fileURL || r.URL.Query().Get("basicDataId") != "123" || p.Body != nil || p.Params != nil {
				t.Error("import must bind request parameters")
			}
			result = "task-123"
		case "/basicdata/record/getImportProgress":
			if p.Method != "GET" || r.URL.Query().Get("uuid") != "task-123" {
				t.Error("incorrect progress request")
			}
			polls++
			if polls > 1 {
				result = map[string]interface{}{"status": polls - 2, "current": 1, "total": 1}
			}
		default:
			t.Errorf("unexpected route: %s", p.Path)
		}
		writeProxyResponse(t, w, map[string]interface{}{"success": true, "result": result})
	}))
	defer server.Close()
	f, out, errOut := newTestFactory(server.URL)
	root := NewCmdTable(f)
	old, _, _ := root.Find([]string{"import"})
	root.RemoveCommand(old)
	uploads := 0
	root.AddCommand(newCmdTableImport(f, false, func(ctx context.Context, c *client.Client, file *os.File) (string, error) {
		uploads++
		return fileURL, nil
	}))
	root.SetArgs([]string{"import", "--app-id", "165", "--file", importFixture(t), "--poll-interval", "1ms"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Success bool         `json:"success"`
		Result  importResult `json:"result"`
	}
	if err := json.Unmarshal(out.Bytes(), &envelope); err != nil {
		t.Fatalf("stdout is not one JSON document: %s", out)
	}
	if !envelope.Success || !envelope.Result.Completed || !envelope.Result.Created || envelope.Result.BasicDataID != 123 || envelope.Result.UUID != "task-123" || *envelope.Result.Status != 2 || polls != 4 || uploads != 1 {
		t.Fatalf("unexpected result: %s", out)
	}
	if !strings.Contains(errOut.String(), "task-123") {
		t.Error("missing recoverable UUID on stderr")
	}
	want := []string{"/basicdata/record/import/columnsSchema", "/basicdata/save", "/basicdata/schema/save", "/basicdata/record/import"}
	if !reflect.DeepEqual(stages[:4], want) {
		t.Fatalf("wrong stage order: %v", stages)
	}
}

func TestImportExistingTableSubmitOnly(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		p := decodeProxyRequest(t, r)
		if p.Path != "/basicdata/record/import" || r.URL.Query().Get("basicDataId") != "99" {
			t.Errorf("unexpected request: %v", p)
		}
		writeProxyResponse(t, w, map[string]interface{}{"success": true, "result": "job"})
	}))
	defer server.Close()
	f, out, _ := newTestFactory(server.URL)
	parent := &cobra.Command{Use: "table"}
	parent.PersistentFlags().String("format", "json", "")
	parent.AddCommand(newCmdTableImport(f, true, func(context.Context, *client.Client, *os.File) (string, error) {
		return "https://example.com/file.csv", nil
	}))
	parent.SetArgs([]string{"import", "--basic-data-id", "99", "--file", importFixture(t), "--wait=false"})
	if err := parent.Execute(); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || !strings.Contains(out.String(), `"completed": false`) || !strings.Contains(out.String(), `"uuid": "job"`) {
		t.Fatalf("not submit-only: %s calls=%d", out, calls)
	}
}

func TestImportValidatesBeforeUpload(t *testing.T) {
	file := importFixture(t)
	for _, args := range [][]string{
		{"--app-id", "1"}, {"--file", file}, {"--app-id", "1", "--basic-data-id", "2", "--file", file},
		{"--basic-data-id", "-1", "--file", file}, {"--app-id", "1", "--file", file, "--timeout", "0"},
		{"--app-id", "1", "--file", file, "--poll-interval", "-1s"}, {"--app-id", "1", "--file", file, "--format", "xml"},
		{"--basic-data-id", "1", "--file", file, "--name", "bad"}, {"--app-id", "1", "--file", file, "--name", " "},
		{"--app-id", "1", "--file", "missing.xlsx"}, {"--app-id", "1", "--file", "file.txt"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			f, _, _ := newTestFactory("http://invalid")
			cmd := newCmdTableImport(f, false, func(context.Context, *client.Client, *os.File) (string, error) {
				t.Fatal("uploaded invalid input")
				return "", nil
			})
			cmd.Flags().String("format", "json", "")
			cmd.SetArgs(args)
			err := cmd.Execute()
			if err == nil || lterrors.CategoryOf(err) != lterrors.CategoryValidation {
				t.Fatalf("expected validation error: %v", err)
			}
		})
	}
	for _, size := range []int64{0, 50*1024*1024 + 1} {
		path := filepath.Join(t.TempDir(), "file.xlsx")
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err = f.Truncate(size); err != nil {
			t.Fatal(err)
		}
		f.Close()
		if f, err := openImportFile(path); err == nil {
			f.Close()
			t.Errorf("accepted size %d", size)
		}
	}
	dir := filepath.Join(t.TempDir(), "folder.xlsx")
	_ = os.Mkdir(dir, 0700)
	if f, err := openImportFile(dir); err == nil {
		f.Close()
		t.Error("accepted directory")
	}
}

func TestImportDryRunNeedsNoHostOrAuth(t *testing.T) {
	f, out, _ := newTestFactory("")
	f.Config.Token = ""
	root := NewCmdTable(f)
	root.SetArgs([]string{"import", "--app-id", "1", "--file", importFixture(t), "--dry-run"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), `"dryRun": true`) {
		t.Fatalf("missing plan: %s", out)
	}
}

func TestImportProgressFailureTimeoutAndUnknownStatus(t *testing.T) {
	for _, tc := range []struct {
		name, body, contains string
		timeout              bool
	}{
		{"failed", `{"success":true,"result":{"status":3,"message":"日期格式错误"}}`, "日期格式错误", false},
		{"unknown", `{"success":true,"result":{"status":99}}`, "unknown import status", false},
		{"missing_status", `{"success":true,"result":{}}`, "missing status", false},
		{"business_error", `{"success":false,"msg":"任务不存在"}`, "任务不存在", false},
		{"timeout", `{"success":true,"result":{"status":1}}`, "timed out", true},
		{"pending_null", `{"success":true,"result":null}`, "timed out", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				p := decodeProxyRequest(t, r)
				if p.Path != "/basicdata/record/getImportProgress" {
					t.Error("resubmitted task")
				}
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			f, out, _ := newTestFactory(server.URL)
			root := NewCmdTable(f)
			root.SetArgs([]string{"import-progress", "--uuid", "recover-me", "--timeout", "35ms", "--poll-interval", "2ms"})
			err := root.Execute()
			if err == nil || !strings.Contains(err.Error(), tc.contains) || !strings.Contains(err.Error(), "recover-me") {
				t.Fatalf("unexpected error: %v", err)
			}
			problem, _ := lterrors.ProblemOf(err)
			if problem == nil || problem.Retryable || !strings.Contains(problem.Hint, "table import-progress") {
				t.Fatalf("missing recovery hint: %v", problem)
			}
			if tc.timeout && problem.Subtype != lterrors.SubtypeNetworkTimeout {
				t.Errorf("wrong timeout type: %v", problem)
			}
			if out.Len() != 0 {
				t.Error("failure emitted success output")
			}
		})
	}
}

func TestImportRetainsCreatedTableOnSchemaFailure(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := decodeProxyRequest(t, r)
		calls = append(calls, p.Path)
		switch p.Path {
		case "/basicdata/record/import/columnsSchema":
			fmt.Fprint(w, `{"success":true,"result":{"columnsSchema":[{"id":1,"title":"商品","type":"text"}]}}`)
		case "/basicdata/save":
			fmt.Fprint(w, `{"success":true,"result":{"id":77}}`)
		case "/basicdata/schema/save":
			fmt.Fprint(w, `{"success":false,"message":"保存字段失败"}`)
		default:
			t.Error("should stop after failure")
		}
	}))
	defer server.Close()
	f, out, _ := newTestFactory(server.URL)
	cmd := newCmdTableImport(f, false, func(context.Context, *client.Client, *os.File) (string, error) {
		return "https://example.com/file.csv", nil
	})
	cmd.Flags().String("format", "json", "")
	cmd.SetArgs([]string{"--app-id", "1", "--file", importFixture(t)})
	err := cmd.Execute()
	problem, _ := lterrors.ProblemOf(err)
	if err == nil || !strings.Contains(err.Error(), "basicDataId=77") || !strings.Contains(problem.Hint, "retained") || len(calls) != 3 || out.Len() != 0 {
		t.Fatalf("incorrect partial failure: %v %v", err, calls)
	}
}

func TestImportProgressDeadlineCancelsHTTP(t *testing.T) {
	release := make(chan struct{})
	cancelled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		select {
		case <-r.Context().Done():
			close(cancelled)
		case <-release:
		}
	}))
	defer server.Close()
	defer close(release)
	f, _, _ := newTestFactory(server.URL)
	root := NewCmdTable(f)
	root.SetArgs([]string{"import-progress", "--uuid", "job", "--timeout", "20ms"})
	start := time.Now()
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "timed out") || time.Since(start) > time.Second {
		t.Fatalf("request not cancelled: %v", err)
	}
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Error("server did not observe request cancellation")
	}
}

func TestImportProgressSupportsEnvelopeJQAndSingleQuery(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		fmt.Fprint(w, `{"success":true,"result":{"status":1,"current":2,"total":10}}`)
	}))
	defer server.Close()
	f, out, _ := newTestFactory(server.URL)
	f.Envelope = true
	f.JqExpr = ".data.result.completed"
	root := NewCmdTable(f)
	root.SetArgs([]string{"import-progress", "--uuid", "job", "--wait=false"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if calls != 1 || strings.TrimSpace(out.String()) != "false" {
		t.Fatalf("calls=%d output=%s", calls, out)
	}
}
