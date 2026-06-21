// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package basicdata

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	lterrors "github.com/lingtong/cli/internal/errors"
	"github.com/lingtong/cli/internal/output"
)

// proxyRequest captures what the proxy client sent.
type proxyRequest struct {
	rawURL    string
	proxyBody map[string]interface{}
}

// newTestServer returns a server that records the request and replies with resp.
func newTestServer(t *testing.T, resp string, captured *proxyRequest) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured.rawURL = r.URL.String()
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured.proxyBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
}

func newTestFactory(serverURL string) (*cmdutil.Factory, *bytes.Buffer) {
	out := &bytes.Buffer{}
	f := &cmdutil.Factory{
		Config:    &config.Config{Host: serverURL, Token: "t"},
		IOStreams: &output.IOStreams{In: &bytes.Buffer{}, Out: out, ErrOut: io.Discard},
	}
	return f, out
}

// run executes a freshly-built basicdata command with args and returns the error.
func run(f *cmdutil.Factory, args ...string) error {
	cmd := NewCmdBasicdata(f)
	cmd.SetArgs(args)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return cmd.Execute()
}

func bodyOf(t *testing.T, p *proxyRequest) map[string]interface{} {
	t.Helper()
	b, ok := p.proxyBody["body"].(map[string]interface{})
	if !ok {
		t.Fatalf("proxy body missing or wrong type: %+v", p.proxyBody)
	}
	return b
}

func TestRecordList(t *testing.T) {
	var cap proxyRequest
	srv := newTestServer(t, `{"result":{"list":[],"total":0}}`, &cap)
	defer srv.Close()
	f, out := newTestFactory(srv.URL)

	err := run(f, "record", "list",
		"--basic-data-id", "123",
		"--filter", `{"1":"系统订单"}`,
		"--ids", "1001,1002",
		"--page", "2", "--page-size", "50",
		"--order-by", "created", "--order-asc")
	if err != nil {
		t.Fatalf("list error = %v", err)
	}
	if cap.proxyBody["path"] != "/basicdata/record/listNew" {
		t.Errorf("path = %v", cap.proxyBody["path"])
	}
	body := bodyOf(t, &cap)
	if body["basicDataId"].(float64) != 123 {
		t.Errorf("basicDataId = %v", body["basicDataId"])
	}
	if body["pageNum"].(float64) != 2 || body["pageSize"].(float64) != 50 {
		t.Errorf("paging = %v/%v", body["pageNum"], body["pageSize"])
	}
	if _, ok := body["text"].(map[string]interface{}); !ok {
		t.Errorf("filter not parsed into text object: %v", body["text"])
	}
	if ids, ok := body["ids"].([]interface{}); !ok || len(ids) != 2 {
		t.Errorf("ids = %v", body["ids"])
	}
	if body["column"] != "created" || body["asc"] != true {
		t.Errorf("sort = %v/%v", body["column"], body["asc"])
	}
	if !bytes.Contains(out.Bytes(), []byte("result")) {
		t.Errorf("output missing result: %s", out.String())
	}
}

func TestRecordListRequiresBasicDataID(t *testing.T) {
	f, _ := newTestFactory("http://localhost:1")
	err := run(f, "record", "list")
	if err == nil {
		t.Fatal("expected error when --basic-data-id missing")
	}
	if output.ExitCodeOf(lterrors.Classify(err)) != output.ExitValidation {
		t.Errorf("exit code = %d, want %d", output.ExitCodeOf(lterrors.Classify(err)), output.ExitValidation)
	}
}

func TestRecordListInvalidFilter(t *testing.T) {
	var cap proxyRequest
	srv := newTestServer(t, `{}`, &cap)
	defer srv.Close()
	f, _ := newTestFactory(srv.URL)
	err := run(f, "record", "list", "--basic-data-id", "1", "--filter", "{bad")
	if err == nil || lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Errorf("expected validation error for bad --filter, got %v", err)
	}
}

func TestRecordCount(t *testing.T) {
	var cap proxyRequest
	srv := newTestServer(t, `{"result":42}`, &cap)
	defer srv.Close()
	f, _ := newTestFactory(srv.URL)

	if err := run(f, "record", "count", "--schema-id", "2546", "--version", "1"); err != nil {
		t.Fatalf("count error = %v", err)
	}
	if cap.proxyBody["path"] != "/basicdata/record/count" {
		t.Errorf("path = %v", cap.proxyBody["path"])
	}
	if cap.proxyBody["method"] != "GET" {
		t.Errorf("method = %v, want GET", cap.proxyBody["method"])
	}
	// GET query params are appended to the proxy URL.
	if got := cap.rawURL; !bytes.Contains([]byte(got), []byte("schemaId=2546")) {
		t.Errorf("query missing schemaId: %s", got)
	}
}

func TestRecordCountRequiresVersion(t *testing.T) {
	f, _ := newTestFactory("http://localhost:1")
	err := run(f, "record", "count", "--schema-id", "1")
	if err == nil {
		t.Fatal("expected error when --version missing")
	}
}

func TestRecordSave(t *testing.T) {
	var cap proxyRequest
	srv := newTestServer(t, `{"success":true}`, &cap)
	defer srv.Close()
	f, _ := newTestFactory(srv.URL)

	if err := run(f, "record", "save",
		"--basic-data-id", "123", "--schema-id", "1", "--data", `{"0":"a","1":"b"}`); err != nil {
		t.Fatalf("save error = %v", err)
	}
	if cap.proxyBody["path"] != "/basicdata/record/save" {
		t.Errorf("path = %v", cap.proxyBody["path"])
	}
	body := bodyOf(t, &cap)
	if body["basicDataId"].(float64) != 123 || body["schemaId"].(float64) != 1 {
		t.Errorf("ids = %v/%v", body["basicDataId"], body["schemaId"])
	}
	if d, ok := body["data"].(map[string]interface{}); !ok || d["0"] != "a" {
		t.Errorf("data = %v", body["data"])
	}
	if _, present := body["version"]; present {
		t.Errorf("version should be omitted when not set, got %v", body["version"])
	}
}

func TestRecordSaveInvalidData(t *testing.T) {
	f, _ := newTestFactory("http://localhost:1")
	err := run(f, "record", "save", "--basic-data-id", "1", "--schema-id", "1", "--data", "{bad")
	if err == nil || lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Errorf("expected validation error for bad --data, got %v", err)
	}
}

func TestRecordUpdate(t *testing.T) {
	var cap proxyRequest
	srv := newTestServer(t, `{"success":true}`, &cap)
	defer srv.Close()
	f, _ := newTestFactory(srv.URL)

	if err := run(f, "record", "update",
		"--basic-data-id", "123", "--schema-id", "1", "--id", "999",
		"--data", `{"1":"x"}`, "--version", "5"); err != nil {
		t.Fatalf("update error = %v", err)
	}
	if cap.proxyBody["path"] != "/basicdata/record/update" {
		t.Errorf("path = %v", cap.proxyBody["path"])
	}
	body := bodyOf(t, &cap)
	if body["id"].(float64) != 999 {
		t.Errorf("id = %v", body["id"])
	}
	if body["version"].(float64) != 5 {
		t.Errorf("version = %v", body["version"])
	}
}

func TestRecordDeleteNeedsConfirmation(t *testing.T) {
	f, _ := newTestFactory("http://localhost:1")
	err := run(f, "record", "delete", "--schema-id", "1", "--id", "5")
	if err == nil {
		t.Fatal("expected confirmation error without --yes")
	}
	if lterrors.CategoryOf(err) != lterrors.CategoryConfirmation {
		t.Errorf("category = %v, want confirmation", lterrors.CategoryOf(err))
	}
	if got := output.ExitCodeOf(err); got != output.ExitConfirmationRequired {
		t.Errorf("exit code = %d, want %d", got, output.ExitConfirmationRequired)
	}
}

func TestRecordDeleteConfirmed(t *testing.T) {
	var cap proxyRequest
	srv := newTestServer(t, `{"success":true}`, &cap)
	defer srv.Close()
	f, _ := newTestFactory(srv.URL)

	if err := run(f, "record", "delete", "--schema-id", "2546", "--id", "33832272", "--yes"); err != nil {
		t.Fatalf("delete error = %v", err)
	}
	if cap.proxyBody["path"] != "/basicdata/record/batchDelete" {
		t.Errorf("path = %v", cap.proxyBody["path"])
	}
	body := bodyOf(t, &cap)
	ids, ok := body["ids"].([]interface{})
	if !ok || len(ids) != 1 || ids[0].(float64) != 33832272 {
		t.Errorf("ids = %v", body["ids"])
	}
}

func TestRecordBatchDelete(t *testing.T) {
	var cap proxyRequest
	srv := newTestServer(t, `{"success":true}`, &cap)
	defer srv.Close()
	f, _ := newTestFactory(srv.URL)

	if err := run(f, "record", "batch-delete", "--schema-id", "1", "--ids", "10,20,30", "--yes"); err != nil {
		t.Fatalf("batch-delete error = %v", err)
	}
	body := bodyOf(t, &cap)
	ids, ok := body["ids"].([]interface{})
	if !ok || len(ids) != 3 {
		t.Errorf("ids = %v", body["ids"])
	}
}

func TestRecordBatchDeleteInvalidIDs(t *testing.T) {
	f, _ := newTestFactory("http://localhost:1")
	err := run(f, "record", "batch-delete", "--schema-id", "1", "--ids", "10,abc", "--yes")
	if err == nil || lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Errorf("expected validation error for bad --ids, got %v", err)
	}
}

func TestRecordBatchDeleteNeedsConfirmation(t *testing.T) {
	f, _ := newTestFactory("http://localhost:1")
	err := run(f, "record", "batch-delete", "--schema-id", "1", "--ids", "10,20")
	if lterrors.CategoryOf(err) != lterrors.CategoryConfirmation {
		t.Errorf("category = %v, want confirmation", lterrors.CategoryOf(err))
	}
}
