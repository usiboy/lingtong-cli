// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package factory

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

type capture struct {
	rawURL string
	proxy  map[string]interface{}
}

func newServer(t *testing.T, resp string, cap *capture) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cap.rawURL = r.URL.String()
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &cap.proxy)
		_, _ = w.Write([]byte(resp))
	}))
}

func newFactory(url string) *cmdutil.Factory {
	return &cmdutil.Factory{
		Config:    &config.Config{Host: url, Token: "t"},
		IOStreams: &output.IOStreams{In: &bytes.Buffer{}, Out: &bytes.Buffer{}, ErrOut: io.Discard},
	}
}

func run(f *cmdutil.Factory, args ...string) error {
	cmd := NewCmdFactory(f)
	cmd.SetArgs(args)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return cmd.Execute()
}

func TestFactoryList(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"result":{"list":[]}}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "list", "--page", "2", "--page-size", "30"); err != nil {
		t.Fatalf("list error = %v", err)
	}
	if cap.proxy["path"] != "/factory/list" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
	if !bytes.Contains([]byte(cap.rawURL), []byte("pageNum=2")) {
		t.Errorf("missing pageNum: %s", cap.rawURL)
	}
}

func TestFactoryGetRequiresID(t *testing.T) {
	err := run(newFactory("http://localhost:1"), "get")
	if err == nil || output.ExitCodeOf(lterrors.Classify(err)) != output.ExitValidation {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestFactorySave(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"success":true}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "save", "--data", `{"name":"c1"}`); err != nil {
		t.Fatalf("save error = %v", err)
	}
	if cap.proxy["path"] != "/factory/save" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
	body, _ := cap.proxy["body"].(map[string]interface{})
	if body["name"] != "c1" {
		t.Errorf("body = %v", cap.proxy["body"])
	}
}

func TestFactorySaveInvalidData(t *testing.T) {
	err := run(newFactory("http://localhost:1"), "save", "--data", "{bad")
	if err == nil || lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestFactoryDeleteNeedsConfirmation(t *testing.T) {
	err := run(newFactory("http://localhost:1"), "delete", "--id", "12")
	if lterrors.CategoryOf(err) != lterrors.CategoryConfirmation {
		t.Errorf("category = %v, want confirmation", lterrors.CategoryOf(err))
	}
	if output.ExitCodeOf(err) != output.ExitConfirmationRequired {
		t.Errorf("exit = %d, want 10", output.ExitCodeOf(err))
	}
}

func TestFactoryDeleteConfirmed(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"success":true}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "delete", "--id", "12", "--yes"); err != nil {
		t.Fatalf("delete error = %v", err)
	}
	if cap.proxy["path"] != "/factory/delete" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
	// integer body
	if cap.proxy["body"].(float64) != 12 {
		t.Errorf("body = %v, want 12", cap.proxy["body"])
	}
}

func TestFactoryHTTPRun(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"ok":true}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "http", "run", "--data", `{"id":88}`); err != nil {
		t.Fatalf("http run error = %v", err)
	}
	if cap.proxy["path"] != "/factory/http/run" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
}

func TestFactoryHTTPDeleteNeedsConfirmation(t *testing.T) {
	err := run(newFactory("http://localhost:1"), "http", "delete", "--id", "88")
	if lterrors.CategoryOf(err) != lterrors.CategoryConfirmation {
		t.Errorf("category = %v, want confirmation", lterrors.CategoryOf(err))
	}
}

func TestFactoryHTTPList(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"result":{}}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "http", "list", "--factory-id", "12"); err != nil {
		t.Fatalf("http list error = %v", err)
	}
	if cap.proxy["path"] != "/factory/http/list" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
	if !bytes.Contains([]byte(cap.rawURL), []byte("factoryId=12")) {
		t.Errorf("missing factoryId: %s", cap.rawURL)
	}
}

func TestFactoryScriptListRequiresFactoryID(t *testing.T) {
	err := run(newFactory("http://localhost:1"), "script", "list")
	// cobra's required-flag error is classified to validation by the root handler.
	if err == nil || output.ExitCodeOf(lterrors.Classify(err)) != output.ExitValidation {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestFactoryScriptDeleteConfirmed(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"success":true}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "script", "delete", "--id", "55", "--yes"); err != nil {
		t.Fatalf("script delete error = %v", err)
	}
	if cap.proxy["path"] != "/factory/script/delete" || cap.proxy["body"].(float64) != 55 {
		t.Errorf("path/body = %v / %v", cap.proxy["path"], cap.proxy["body"])
	}
}

func TestFactoryGet(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"result":{"id":12}}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "get", "--id", "12"); err != nil {
		t.Fatalf("get error = %v", err)
	}
	if cap.proxy["path"] != "/factory/get" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
	if !bytes.Contains([]byte(cap.rawURL), []byte("id=12")) {
		t.Errorf("missing id query: %s", cap.rawURL)
	}
}

func TestFactoryHTTPGet(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"result":{"id":88}}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "http", "get", "--id", "88"); err != nil {
		t.Fatalf("http get error = %v", err)
	}
	if cap.proxy["path"] != "/factory/http/get" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
}

func TestFactoryScriptGetAndList(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"result":{}}`, &cap)
	defer srv.Close()
	f := newFactory(srv.URL)
	if err := run(f, "script", "get", "--id", "55"); err != nil {
		t.Fatalf("script get error = %v", err)
	}
	if cap.proxy["path"] != "/factory/script/get" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
	if err := run(f, "script", "list", "--factory-id", "12"); err != nil {
		t.Fatalf("script list error = %v", err)
	}
	if cap.proxy["path"] != "/factory/script/list" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
}

func TestFactoryHTTPDeleteConfirmed(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"success":true}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "http", "delete", "--id", "88", "--yes"); err != nil {
		t.Fatalf("http delete error = %v", err)
	}
	if cap.proxy["path"] != "/factory/http/delete" || cap.proxy["body"].(float64) != 88 {
		t.Errorf("path/body = %v / %v", cap.proxy["path"], cap.proxy["body"])
	}
}
