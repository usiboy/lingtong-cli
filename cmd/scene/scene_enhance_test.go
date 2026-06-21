// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package scene

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
	cmd := NewCmdScene(f)
	cmd.SetArgs(args)
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return cmd.Execute()
}

func TestSceneDeleteNeedsConfirmation(t *testing.T) {
	err := run(newFactory("http://localhost:1"), "delete", "--scene-id", "123")
	if lterrors.CategoryOf(err) != lterrors.CategoryConfirmation {
		t.Errorf("category = %v, want confirmation", lterrors.CategoryOf(err))
	}
	if output.ExitCodeOf(err) != output.ExitConfirmationRequired {
		t.Errorf("exit = %d, want 10", output.ExitCodeOf(err))
	}
}

func TestSceneDeleteConfirmed(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"success":true}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "delete", "--scene-id", "123", "--yes"); err != nil {
		t.Fatalf("delete error = %v", err)
	}
	if cap.proxy["path"] != "/scene/delete" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
	if cap.proxy["body"].(float64) != 123 {
		t.Errorf("body = %v, want 123", cap.proxy["body"])
	}
}

func TestSceneCopy(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"result":456}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "copy", "--scene-id", "123"); err != nil {
		t.Fatalf("copy error = %v", err)
	}
	if cap.proxy["path"] != "/scene/copy" || cap.proxy["body"].(float64) != 123 {
		t.Errorf("path/body = %v / %v", cap.proxy["path"], cap.proxy["body"])
	}
}

func TestSceneOpen(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"success":true}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "open", "--scene-id", "123"); err != nil {
		t.Fatalf("open error = %v", err)
	}
	if cap.proxy["path"] != "/scene/open" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
}

func TestSceneUpdate(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"success":true}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "update", "--data", `{"id":123,"name":"x"}`); err != nil {
		t.Fatalf("update error = %v", err)
	}
	if cap.proxy["path"] != "/scene/model/update" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
	body, _ := cap.proxy["body"].(map[string]interface{})
	if body["name"] != "x" {
		t.Errorf("body = %v", cap.proxy["body"])
	}
}

func TestSceneUpdateRequiresData(t *testing.T) {
	err := run(newFactory("http://localhost:1"), "update")
	// cobra's required-flag error is classified to validation by the root handler.
	if err == nil || output.ExitCodeOf(lterrors.Classify(err)) != output.ExitValidation {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestScenePublishWithSceneID(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"success":true}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "publish", "--scene-id", "123"); err != nil {
		t.Fatalf("publish error = %v", err)
	}
	if cap.proxy["path"] != "/scene/publish/release" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
	body, _ := cap.proxy["body"].(map[string]interface{})
	if body["sceneId"].(float64) != 123 {
		t.Errorf("body = %v", cap.proxy["body"])
	}
}

func TestScenePublishWithData(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"success":true}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "publish", "--data", `{"sceneId":9,"remark":"v2"}`); err != nil {
		t.Fatalf("publish error = %v", err)
	}
	body, _ := cap.proxy["body"].(map[string]interface{})
	if body["remark"] != "v2" {
		t.Errorf("body = %v", cap.proxy["body"])
	}
}

func TestScenePublishRequiresSceneIDOrData(t *testing.T) {
	err := run(newFactory("http://localhost:1"), "publish")
	if err == nil || lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Errorf("expected validation error, got %v", err)
	}
}

func TestSceneVersionList(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"result":[]}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "version", "list", "--scene-id", "123"); err != nil {
		t.Fatalf("version list error = %v", err)
	}
	if cap.proxy["path"] != "/scene/version/list" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
	if !bytes.Contains([]byte(cap.rawURL), []byte("sceneId=123")) {
		t.Errorf("missing sceneId: %s", cap.rawURL)
	}
}

func TestSceneTriggerGet(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"result":{}}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "trigger", "get", "--scene-id", "123"); err != nil {
		t.Fatalf("trigger get error = %v", err)
	}
	if cap.proxy["path"] != "/scene/trigger/condition/get" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
}

func TestSceneTriggerSave(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"success":true}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "trigger", "save", "--data", `{"sceneId":1}`); err != nil {
		t.Fatalf("trigger save error = %v", err)
	}
	if cap.proxy["path"] != "/scene/trigger/condition/save" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
}

func TestSceneFieldMappingList(t *testing.T) {
	var cap capture
	srv := newServer(t, `{"result":[]}`, &cap)
	defer srv.Close()
	if err := run(newFactory(srv.URL), "field-mapping", "list", "--scene-id", "123"); err != nil {
		t.Fatalf("field-mapping list error = %v", err)
	}
	if cap.proxy["path"] != "/scene/field/value/list" {
		t.Errorf("path = %v", cap.proxy["path"])
	}
}
