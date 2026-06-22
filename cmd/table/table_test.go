// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package table

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	lterrors "github.com/lingtong/cli/internal/errors"
	"github.com/lingtong/cli/internal/output"
)

type proxyRequest struct {
	Path   string                 `json:"path"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params,omitempty"`
	Body   map[string]interface{} `json:"body,omitempty"`
}

func newTestFactory(serverURL string) (*cmdutil.Factory, *bytes.Buffer, *bytes.Buffer) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	return &cmdutil.Factory{
		Config: &config.Config{
			Host:  serverURL,
			Token: "test-token",
		},
		IOStreams: &output.IOStreams{
			In:     io.NopCloser(strings.NewReader("")),
			Out:    out,
			ErrOut: errOut,
		},
	}, out, errOut
}

func decodeProxyRequest(t *testing.T, r *http.Request) proxyRequest {
	t.Helper()
	var req proxyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		t.Fatalf("failed to decode proxy request: %v", err)
	}
	return req
}

func writeProxyResponse(t *testing.T, w http.ResponseWriter, data interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		t.Fatalf("failed to write response: %v", err)
	}
}

// ==================== table list ====================

func TestNewCmdTableList_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if got := r.URL.RequestURI(); got != "/gw/ai/proxy" {
			t.Errorf("expected proxy URI /gw/ai/proxy, got %s", got)
		}

		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/basicdata/list" {
			t.Errorf("expected proxy path /basicdata/list, got %s", proxyReq.Path)
		}
		if proxyReq.Method != "GET" {
			t.Errorf("expected method GET, got %s", proxyReq.Method)
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result": []map[string]interface{}{
				{"id": 1, "name": "Orders"},
				{"id": 2, "name": "Customers"},
			},
		})
	}))
	defer server.Close()

	f, out, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Len() == 0 {
		t.Fatal("expected output, got empty")
	}
}

func TestNewCmdTableList_WithAppId(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if got := query.Get("appId"); got != "123" {
			t.Errorf("expected appId=123 in query, got %s", got)
		}

		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/basicdata/list" {
			t.Errorf("expected proxy path /basicdata/list, got %s", proxyReq.Path)
		}
		if proxyReq.Params != nil {
			t.Error("expected no params in proxy body for GET request")
		}

		writeProxyResponse(t, w, map[string]interface{}{"success": true, "result": []interface{}{}})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"list", "--app-id", "123"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableList_NoHostConfigured(t *testing.T) {
	f, _, errOut := newTestFactory("")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected host config error")
	}
	if !strings.Contains(err.Error(), "no host configured") {
		t.Fatalf("expected host config error, got %v", err)
	}
}

// ==================== table data query ====================

func TestNewCmdTableDataQuery_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/basicdata/record/listNew" {
			t.Errorf("expected proxy path /basicdata/record/listNew, got %s", proxyReq.Path)
		}
		if proxyReq.Method != "POST" {
			t.Errorf("expected method POST, got %s", proxyReq.Method)
		}
		if proxyReq.Body == nil {
			t.Fatal("expected body in proxy request")
		}
		if basicDataId, ok := proxyReq.Body["basicDataId"]; !ok || basicDataId != float64(1553) {
			t.Errorf("expected basicDataId=1553 in body, got %v", proxyReq.Body["basicDataId"])
		}
		// Check default pagination
		if pageNum, ok := proxyReq.Body["pageNum"]; !ok || pageNum != float64(1) {
			t.Errorf("expected default pageNum=1, got %v", proxyReq.Body["pageNum"])
		}
		if pageSize, ok := proxyReq.Body["pageSize"]; !ok || pageSize != float64(20) {
			t.Errorf("expected default pageSize=20, got %v", proxyReq.Body["pageSize"])
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  map[string]interface{}{"data": []interface{}{}},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "query", "--basic-data-id", "1553"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableDataQuery_WithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/basicdata/record/listNew" {
			t.Errorf("expected proxy path /basicdata/record/listNew, got %s", proxyReq.Path)
		}
		body := proxyReq.Body
		// text filter
		if text, ok := body["text"]; !ok || text != `{"col_abc":"hello"}` {
			t.Errorf("expected text={\"col_abc\":\"hello\"}, got %v", body["text"])
		}
		// ids filter
		if ids, ok := body["ids"]; !ok {
			t.Error("expected ids in body")
		} else {
			idList, ok := ids.([]interface{})
			if !ok {
				t.Errorf("expected ids to be a list, got %T", ids)
			} else if len(idList) != 2 {
				t.Errorf("expected 2 ids, got %d", len(idList))
			}
		}
		// pagination
		if pageNum, ok := body["pageNum"]; !ok || pageNum != float64(3) {
			t.Errorf("expected pageNum=3, got %v", body["pageNum"])
		}
		if pageSize, ok := body["pageSize"]; !ok || pageSize != float64(50) {
			t.Errorf("expected pageSize=50, got %v", body["pageSize"])
		}
		// sort
		if column, ok := body["column"]; !ok || column != "created" {
			t.Errorf("expected column=created, got %v", body["column"])
		}
		if asc, ok := body["asc"]; !ok || asc != true {
			t.Errorf("expected asc=true, got %v", body["asc"])
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  map[string]interface{}{"data": []interface{}{}},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{
		"data", "query",
		"--basic-data-id", "1553",
		"--text", `{"col_abc":"hello"}`,
		"--ids", "1001,1002",
		"--page", "3",
		"--page-size", "50",
		"--order-by", "created",
		"--order-asc",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableDataQuery_WithFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/basicdata/record/listNew" {
			t.Errorf("expected proxy path /basicdata/record/listNew, got %s", proxyReq.Path)
		}
		body := proxyReq.Body
		// --filter maps to body["text"]
		if text, ok := body["text"]; !ok || text != `{"1":"系统订单","17":"-2"}` {
			t.Errorf("expected text={\"1\":\"系统订单\",\"17\":\"-2\"}, got %v", body["text"])
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  map[string]interface{}{"data": []interface{}{}},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{
		"data", "query",
		"--basic-data-id", "1568",
		"--filter", `{"1":"系统订单","17":"-2"}`,
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableDataQuery_WithViewId(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/basicdata/record/listNew" {
			t.Errorf("expected proxy path /basicdata/record/listNew, got %s", proxyReq.Path)
		}
		body := proxyReq.Body
		// --view-id maps to body["viewId"]
		if viewId, ok := body["viewId"]; !ok || viewId != float64(456) {
			t.Errorf("expected viewId=456, got %v", body["viewId"])
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  map[string]interface{}{"data": []interface{}{}},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{
		"data", "query",
		"--basic-data-id", "1568",
		"--view-id", "456",
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableDataQuery_MissingBasicDataId(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "query"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --basic-data-id")
	}
}

// ==================== table data create ====================

func TestNewCmdTableDataCreate_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/basicdata/record/save" {
			t.Errorf("expected proxy path /basicdata/record/save, got %s", proxyReq.Path)
		}
		if proxyReq.Method != "POST" {
			t.Errorf("expected method POST, got %s", proxyReq.Method)
		}
		if proxyReq.Body == nil {
			t.Fatal("expected body in proxy request")
		}
		// Verify basicDataId
		if basicDataId, ok := proxyReq.Body["basicDataId"]; !ok {
			t.Error("expected basicDataId in body")
		} else if basicDataId != float64(1553) {
			t.Errorf("expected basicDataId=1553, got %v", basicDataId)
		}
		// Verify schemaId
		if schemaId, ok := proxyReq.Body["schemaId"]; !ok {
			t.Error("expected schemaId in body")
		} else if schemaId != float64(2189) {
			t.Errorf("expected schemaId=2189, got %v", schemaId)
		}
		// Verify data
		if _, ok := proxyReq.Body["data"]; !ok {
			t.Error("expected data in body")
		}
		// Verify version is NOT present when not specified
		if _, ok := proxyReq.Body["version"]; ok {
			t.Error("version should not be in body when not specified")
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  map[string]interface{}{"id": 500},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "create", "--basic-data-id", "1553", "--schema-id", "2189", "--data", `{"0":"test"}`})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableDataCreate_WithVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if version, ok := proxyReq.Body["version"]; !ok {
			t.Error("expected version in body when specified")
		} else if version != float64(5) {
			t.Errorf("expected version=5, got %v", version)
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  map[string]interface{}{"id": 501},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "create", "--basic-data-id", "1553", "--schema-id", "2189", "--data", `{"0":"test"}`, "--version", "5"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableDataCreate_MissingRequiredFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"missing all", []string{"data", "create"}},
		{"missing schema-id and data", []string{"data", "create", "--basic-data-id", "100"}},
		{"missing data", []string{"data", "create", "--basic-data-id", "100", "--schema-id", "1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _, errOut := newTestFactory("http://example.com")
			cmd := NewCmdTable(f)
			cmd.SetErr(errOut)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Error("expected error for missing required flags")
			}
		})
	}
}

func TestNewCmdTableDataCreate_NoHostConfigured(t *testing.T) {
	f, _, errOut := newTestFactory("")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "create", "--basic-data-id", "1553", "--schema-id", "2189", "--data", `{"0":"test"}`})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected host config error")
	}
	if !strings.Contains(err.Error(), "no host configured") {
		t.Fatalf("expected host config error, got %v", err)
	}
}

// ==================== table create ====================

func TestNewCmdTableCreate_ProxyPath(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		proxyReq := decodeProxyRequest(t, r)

		if proxyReq.Method != "POST" {
			t.Errorf("expected method POST, got %s", proxyReq.Method)
		}

		if requestCount == 1 {
			// First request: /basicdata/save
			if proxyReq.Path != "/basicdata/save" {
				t.Errorf("expected proxy path /basicdata/save, got %s", proxyReq.Path)
			}
			// Verify body
			if appId, ok := proxyReq.Body["appId"]; !ok {
				t.Error("expected appId in body")
			} else if appId != float64(165) {
				t.Errorf("expected appId=165, got %v", appId)
			}
			if name, ok := proxyReq.Body["name"]; !ok {
				t.Error("expected name in body")
			} else if name != "TestTable" {
				t.Errorf("expected name=TestTable, got %v", name)
			}

			writeProxyResponse(t, w, map[string]interface{}{
				"success": true,
				"result":  map[string]interface{}{"id": 1630, "name": "TestTable"},
			})
		} else if requestCount == 2 {
			// Second request: /basicdata/schema/save
			if proxyReq.Path != "/basicdata/schema/save" {
				t.Errorf("expected proxy path /basicdata/schema/save, got %s", proxyReq.Path)
			}
			// Verify basicDataId
			if basicDataId, ok := proxyReq.Body["basicDataId"]; !ok {
				t.Error("expected basicDataId in body")
			} else if basicDataId != float64(1630) {
				t.Errorf("expected basicDataId=1630, got %v", basicDataId)
			}
			// Verify columnsSchema
			if columns, ok := proxyReq.Body["columnsSchema"].([]interface{}); !ok {
				t.Error("expected columnsSchema in body")
			} else if len(columns) != 2 {
				t.Errorf("expected 2 columns, got %d", len(columns))
			}

			writeProxyResponse(t, w, map[string]interface{}{
				"success": true,
				"result": map[string]interface{}{
					"id":            2545,
					"basicDataId":   1630,
					"version":       1,
					"columnsSchema": proxyReq.Body["columnsSchema"],
				},
			})
		}
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"create", "--app-id", "165", "--name", "TestTable", "--columns-schema", `[{"title":"Name","key":"name","type":"text"},{"title":"Age","key":"age","type":"number"}]`})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount != 2 {
		t.Errorf("expected 2 requests, got %d", requestCount)
	}
}

func TestNewCmdTableCreate_WithoutSchema(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		proxyReq := decodeProxyRequest(t, r)

		if proxyReq.Path != "/basicdata/save" {
			t.Errorf("expected proxy path /basicdata/save, got %s", proxyReq.Path)
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  map[string]interface{}{"id": 1631, "name": "TestTable2"},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"create", "--app-id", "165", "--name", "TestTable2"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount != 1 {
		t.Errorf("expected 1 request (no schema), got %d", requestCount)
	}
}

func TestNewCmdTableCreate_AutoAssignColumnIds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)

		if proxyReq.Path == "/basicdata/save" {
			writeProxyResponse(t, w, map[string]interface{}{
				"success": true,
				"result":  map[string]interface{}{"id": 1632},
			})
		} else if proxyReq.Path == "/basicdata/schema/save" {
			// Verify auto-assigned IDs
			if columns, ok := proxyReq.Body["columnsSchema"].([]interface{}); ok {
				for i, col := range columns {
					if colMap, ok := col.(map[string]interface{}); ok {
						expectedID := float64(i + 1)
						if id, ok := colMap["id"].(float64); ok {
							if id != expectedID {
								t.Errorf("expected column %d id=%v, got %v", i, expectedID, id)
							}
						} else {
							t.Errorf("expected column %d to have id", i)
						}
					}
				}
			}
			writeProxyResponse(t, w, map[string]interface{}{
				"success": true,
				"result":  map[string]interface{}{"id": 2546},
			})
		}
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	// Columns without IDs - should auto-assign
	cmd.SetArgs([]string{"create", "--app-id", "165", "--name", "TestTable3", "--columns-schema", `[{"title":"A","key":"a","type":"text"},{"title":"B","key":"b","type":"text"}]`})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableCreate_MissingRequiredFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"missing all", []string{"create"}},
		{"missing name", []string{"create", "--app-id", "165"}},
		{"missing app-id", []string{"create", "--name", "Test"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _, errOut := newTestFactory("http://example.com")
			cmd := NewCmdTable(f)
			cmd.SetErr(errOut)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Error("expected error for missing required flags")
			}
		})
	}
}

func TestNewCmdTableCreate_NoHostConfigured(t *testing.T) {
	f, _, errOut := newTestFactory("")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"create", "--app-id", "165", "--name", "Test"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected host config error")
	}
	if !strings.Contains(err.Error(), "no host configured") {
		t.Fatalf("expected host config error, got %v", err)
	}
}

// ==================== table schema ====================

func TestNewCmdTableSchemaQuery_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if got := query.Get("businessId"); got != "1553" {
			t.Errorf("expected businessId=1553 in query, got %s", got)
		}
		if got := query.Get("businessType"); got != "2" {
			t.Errorf("expected businessType=2 in query, got %s", got)
		}

		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/basicdata/schema/get" {
			t.Errorf("expected proxy path /basicdata/schema/get, got %s", proxyReq.Path)
		}
		if proxyReq.Method != "GET" {
			t.Errorf("expected method GET, got %s", proxyReq.Method)
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result": map[string]interface{}{
				"id":           2189,
				"businessId":   1553,
				"businessType": 2,
				"version":      5,
				"columnsSchema": []interface{}{
					map[string]interface{}{"id": 0, "title": "ID", "type": "number", "key": "id"},
					map[string]interface{}{"id": 1, "title": "Name", "type": "text", "key": "name"},
				},
			},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"schema", "query", "--basic-data-id", "1553"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableSchemaQuery_MissingBasicDataId(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"schema", "query"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for missing --basic-data-id")
	}
}

func TestNewCmdTableSchemaQuery_NoHostConfigured(t *testing.T) {
	f, _, errOut := newTestFactory("")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"schema", "query", "--basic-data-id", "1553"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected host config error")
	}
	if !strings.Contains(err.Error(), "no host configured") {
		t.Fatalf("expected host config error, got %v", err)
	}
}

func TestNewCmdTableSchemaUpdate_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/basicdata/schema/update" {
			t.Errorf("expected proxy path /basicdata/schema/update, got %s", proxyReq.Path)
		}
		if proxyReq.Method != "POST" {
			t.Errorf("expected method POST, got %s", proxyReq.Method)
		}
		if proxyReq.Body == nil {
			t.Fatal("expected body in proxy request")
		}
		// Verify id is set
		if id, ok := proxyReq.Body["id"]; !ok {
			t.Error("expected id in body")
		} else if id != float64(2189) {
			t.Errorf("expected id=2189, got %v", id)
		}
		// Verify columnsSchema is present
		if _, ok := proxyReq.Body["columnsSchema"]; !ok {
			t.Error("expected columnsSchema in body")
		}
		// Verify version is present (from JSON)
		if version, ok := proxyReq.Body["version"]; !ok {
			t.Error("expected version in body")
		} else if version != float64(5) {
			t.Errorf("expected version=5, got %v", version)
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result": map[string]interface{}{
				"id":      2189,
				"version": 6,
			},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"schema", "update", "--schema-id", "2189", "--schema", `{"columnsSchema":[{"id":0,"title":"ID","type":"number","key":"id"}],"version":5}`})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableSchemaUpdate_MissingRequiredFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"missing all", []string{"schema", "update"}},
		{"missing schema", []string{"schema", "update", "--schema-id", "2189"}},
		{"missing schema-id", []string{"schema", "update", "--schema", `{"id":1}`}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _, errOut := newTestFactory("http://example.com")
			cmd := NewCmdTable(f)
			cmd.SetErr(errOut)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if err == nil {
				t.Error("expected error for missing required flags")
			}
		})
	}
}

func TestNewCmdTableSchemaUpdate_NoHostConfigured(t *testing.T) {
	f, _, errOut := newTestFactory("")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"schema", "update", "--schema-id", "2189", "--schema", `{"columnsSchema":[]}`})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected host config error")
	}
	if !strings.Contains(err.Error(), "no host configured") {
		t.Fatalf("expected host config error, got %v", err)
	}
}

// ==================== table data batch-update ====================

func TestNewCmdTableDataBatchUpdate_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}

		// Decode the proxy request
		var proxyReq struct {
			Path   string                   `json:"path"`
			Method string                   `json:"method"`
			Body   []map[string]interface{} `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&proxyReq); err != nil {
			t.Fatalf("failed to decode proxy request: %v", err)
		}

		if proxyReq.Path != "/basicdata/record/batchUpdate" {
			t.Errorf("expected proxy path /basicdata/record/batchUpdate, got %s", proxyReq.Path)
		}
		if proxyReq.Method != "POST" {
			t.Errorf("expected method POST, got %s", proxyReq.Method)
		}
		if len(proxyReq.Body) != 2 {
			t.Errorf("expected 2 records, got %d", len(proxyReq.Body))
		}

		// Verify first record
		if id, ok := proxyReq.Body[0]["id"].(float64); !ok || id != 100 {
			t.Errorf("expected first record id=100, got %v", proxyReq.Body[0]["id"])
		}
		if bdi, ok := proxyReq.Body[0]["basicDataId"].(float64); !ok || bdi != 1633 {
			t.Errorf("expected basicDataId=1633, got %v", proxyReq.Body[0]["basicDataId"])
		}
		if si, ok := proxyReq.Body[0]["schemaId"].(float64); !ok || si != 2546 {
			t.Errorf("expected schemaId=2546, got %v", proxyReq.Body[0]["schemaId"])
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  proxyReq.Body,
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-update",
		"--basic-data-id", "1633",
		"--schema-id", "2546",
		"--records", `[{"id":100,"data":{"1":"val1"}},{"id":101,"data":{"1":"val2"}}]`,
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableDataBatchUpdate_WithVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var proxyReq struct {
			Path string                   `json:"path"`
			Body []map[string]interface{} `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&proxyReq); err != nil {
			t.Fatalf("failed to decode proxy request: %v", err)
		}

		// Verify version is auto-filled
		for i, rec := range proxyReq.Body {
			if ver, ok := rec["version"].(float64); !ok || ver != 5 {
				t.Errorf("record %d: expected version=5, got %v", i, rec["version"])
			}
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  proxyReq.Body,
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-update",
		"--basic-data-id", "1633",
		"--schema-id", "2546",
		"--version", "5",
		"--records", `[{"id":100,"data":{"1":"val1"}}]`,
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableDataBatchUpdate_MissingId(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-update",
		"--basic-data-id", "1633",
		"--schema-id", "2546",
		"--records", `[{"data":{"1":"val1"}}]`,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing id")
	}
	if !strings.Contains(err.Error(), "missing required field 'id'") {
		t.Fatalf("expected missing id error, got %v", err)
	}
}

func TestNewCmdTableDataBatchUpdate_MissingData(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-update",
		"--basic-data-id", "1633",
		"--schema-id", "2546",
		"--records", `[{"id":100}]`,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing data")
	}
	if !strings.Contains(err.Error(), "missing required field 'data'") {
		t.Fatalf("expected missing data error, got %v", err)
	}
}

func TestNewCmdTableDataBatchUpdate_EmptyRecords(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-update",
		"--basic-data-id", "1633",
		"--schema-id", "2546",
		"--records", `[]`,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty records")
	}
	if !strings.Contains(err.Error(), "at least one record") {
		t.Fatalf("expected empty records error, got %v", err)
	}
}

func TestNewCmdTableDataBatchUpdate_InvalidJSON(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-update",
		"--basic-data-id", "1633",
		"--schema-id", "2546",
		"--records", `not-json`,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid --records JSON") {
		t.Fatalf("expected invalid JSON error, got %v", err)
	}
}

func TestNewCmdTableDataBatchUpdate_NoHostConfigured(t *testing.T) {
	f, _, errOut := newTestFactory("")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-update",
		"--basic-data-id", "1633",
		"--schema-id", "2546",
		"--records", `[{"id":100,"data":{"1":"val1"}}]`,
	})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected host config error")
	}
	if !strings.Contains(err.Error(), "no host configured") {
		t.Fatalf("expected host config error, got %v", err)
	}
}

// ==================== delete tests ====================

func TestNewCmdTableDataDelete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := decodeProxyRequest(t, r)
		// delete now wraps batchDelete
		if req.Path != "/basicdata/record/batchDelete" {
			t.Fatalf("expected path /basicdata/record/batchDelete, got %s", req.Path)
		}
		if req.Method != "POST" {
			t.Fatalf("expected method POST, got %s", req.Method)
		}
		// Check that schemaId is in body
		if schemaId, ok := req.Body["schemaId"].(float64); !ok || int(schemaId) != 2546 {
			t.Fatalf("expected body schemaId=2546, got %v", req.Body["schemaId"])
		}
		// Check that ids is an array with single element
		ids, ok := req.Body["ids"].([]interface{})
		if !ok || len(ids) != 1 {
			t.Fatalf("expected body ids with 1 element, got %v", req.Body["ids"])
		}
		if id, ok := ids[0].(float64); !ok || int64(id) != 33832272 {
			t.Fatalf("expected ids[0]=33832272, got %v", ids[0])
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"result":  nil,
		})
	}))
	defer server.Close()

	f, out, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "delete", "--schema-id", "2546", "--id", "33832272", "--yes"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableDataDelete_MissingID(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "delete", "--schema-id", "2546"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --id")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Fatalf("expected required error, got %v", err)
	}
}

func TestNewCmdTableDataDelete_MissingSchemaId(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "delete", "--id", "33832272"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --schema-id")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Fatalf("expected required error, got %v", err)
	}
}

func TestNewCmdTableDataDelete_NoHostConfigured(t *testing.T) {
	f, _, errOut := newTestFactory("")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "delete", "--schema-id", "2546", "--id", "33832272", "--yes"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected host config error")
	}
	if !strings.Contains(err.Error(), "no host configured") {
		t.Fatalf("expected host config error, got %v", err)
	}
}

// ==================== batch-delete tests ====================

func TestNewCmdTableDataBatchDelete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := decodeProxyRequest(t, r)
		if req.Path != "/basicdata/record/batchDelete" {
			t.Fatalf("expected path /basicdata/record/batchDelete, got %s", req.Path)
		}
		if req.Method != "POST" {
			t.Fatalf("expected method POST, got %s", req.Method)
		}
		// Check schemaId
		if schemaId, ok := req.Body["schemaId"].(float64); !ok || int(schemaId) != 2546 {
			t.Fatalf("expected body schemaId=2546, got %v", req.Body["schemaId"])
		}
		// Check ids array
		ids, ok := req.Body["ids"].([]interface{})
		if !ok || len(ids) != 3 {
			t.Fatalf("expected body ids with 3 elements, got %v", req.Body["ids"])
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"result":  nil,
		})
	}))
	defer server.Close()

	f, out, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-delete", "--schema-id", "2546", "--ids", "33832272,33832273,33832274", "--yes"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableDataBatchDelete_MissingSchemaId(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-delete", "--ids", "1,2,3"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --schema-id")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Fatalf("expected required error, got %v", err)
	}
}

func TestNewCmdTableDataBatchDelete_MissingIds(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-delete", "--schema-id", "2546"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --ids")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Fatalf("expected required error, got %v", err)
	}
}

func TestNewCmdTableDataBatchDelete_InvalidIdFormat(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-delete", "--schema-id", "2546", "--ids", "abc,def"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid ID format")
	}
	if !strings.Contains(err.Error(), "invalid ID format") {
		t.Fatalf("expected invalid ID format error, got %v", err)
	}
}

func TestNewCmdTableDataBatchDelete_NoHostConfigured(t *testing.T) {
	f, _, errOut := newTestFactory("")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-delete", "--schema-id", "2546", "--ids", "1,2,3", "--yes"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected host config error")
	}
	if !strings.Contains(err.Error(), "no host configured") {
		t.Fatalf("expected host config error, got %v", err)
	}
}

// ==================== data update tests ====================

func TestTableDataUpdate_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := decodeProxyRequest(t, r)
		if req.Path != "/basicdata/record/update" {
			t.Fatalf("expected path /basicdata/record/update, got %s", req.Path)
		}
		if req.Method != "POST" {
			t.Fatalf("expected method POST, got %s", req.Method)
		}
		if bdi, ok := req.Body["basicDataId"].(float64); !ok || int(bdi) != 123 {
			t.Fatalf("expected basicDataId=123, got %v", req.Body["basicDataId"])
		}
		if si, ok := req.Body["schemaId"].(float64); !ok || int(si) != 1 {
			t.Fatalf("expected schemaId=1, got %v", req.Body["schemaId"])
		}
		if id, ok := req.Body["id"].(float64); !ok || int(id) != 999 {
			t.Fatalf("expected id=999, got %v", req.Body["id"])
		}
		if d, ok := req.Body["data"].(map[string]interface{}); !ok || d["1"] != "x" {
			t.Fatalf("expected data={1:x}, got %v", req.Body["data"])
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	f, out, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "update", "--basic-data-id", "123", "--schema-id", "1", "--id", "999", "--data", `{"1":"x"}`})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTableDataUpdate_WithVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := decodeProxyRequest(t, r)
		if v, ok := req.Body["version"].(float64); !ok || int(v) != 5 {
			t.Fatalf("expected version=5, got %v", req.Body["version"])
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	f, out, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "update", "--basic-data-id", "123", "--schema-id", "1", "--id", "999", "--data", `{"1":"x"}`, "--version", "5"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTableDataUpdate_VersionOmitted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := decodeProxyRequest(t, r)
		if _, present := req.Body["version"]; present {
			t.Fatalf("version should be omitted when not set, got %v", req.Body["version"])
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	f, out, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "update", "--basic-data-id", "123", "--schema-id", "1", "--id", "999", "--data", `{"1":"x"}`})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTableDataUpdate_MissingRequiredFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"missing basic-data-id", []string{"data", "update", "--schema-id", "1", "--id", "1", "--data", `{}`}},
		{"missing schema-id", []string{"data", "update", "--basic-data-id", "1", "--id", "1", "--data", `{}`}},
		{"missing id", []string{"data", "update", "--basic-data-id", "1", "--schema-id", "1", "--data", `{}`}},
		{"missing data", []string{"data", "update", "--basic-data-id", "1", "--schema-id", "1", "--id", "1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, _, errOut := newTestFactory("http://example.com")
			cmd := NewCmdTable(f)
			cmd.SetErr(errOut)
			cmd.SetArgs(tt.args)
			if err := cmd.Execute(); err == nil {
				t.Fatal("expected error for missing flag")
			}
		})
	}
}

func TestTableDataUpdate_InvalidData(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "update", "--basic-data-id", "1", "--schema-id", "1", "--id", "1", "--data", "{bad"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --data JSON")
	}
	if lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestTableDataUpdate_NoHostConfigured(t *testing.T) {
	f, _, errOut := newTestFactory("")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "update", "--basic-data-id", "1", "--schema-id", "1", "--id", "1", "--data", `{"1":"x"}`})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected host config error")
	}
}

// ==================== data count tests ====================

func TestTableDataCount_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := decodeProxyRequest(t, r)
		if req.Path != "/basicdata/record/count" {
			t.Fatalf("expected path /basicdata/record/count, got %s", req.Path)
		}
		if req.Method != "GET" {
			t.Fatalf("expected method GET, got %s", req.Method)
		}
		// GET params are in URL query string
		query := r.URL.Query()
		if got := query.Get("schemaId"); got != "2546" {
			t.Fatalf("expected schemaId=2546 in query, got %s", got)
		}
		if got := query.Get("version"); got != "1" {
			t.Fatalf("expected version=1 in query, got %s", got)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"result": 42})
	}))
	defer server.Close()

	f, out, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "count", "--schema-id", "2546", "--version", "1"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTableDataCount_WithFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// GET params are in URL query string
		query := r.URL.Query()
		dataFrom := query.Get("dataFrom")
		if dataFrom == "" {
			t.Fatalf("expected dataFrom in query, got empty")
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"result": 5})
	}))
	defer server.Close()

	f, out, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "count", "--schema-id", "2546", "--version", "1", "--filter", `{"1":"系统订单"}`})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTableDataCount_WithBasicDataId(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if got := query.Get("basicDataId"); got != "123" {
			t.Fatalf("expected basicDataId=123 in query, got %s", got)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"result": 10})
	}))
	defer server.Close()

	f, out, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "count", "--schema-id", "2546", "--version", "1", "--basic-data-id", "123"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTableDataCount_MissingSchemaId(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "count", "--version", "1"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing --schema-id")
	}
}

func TestTableDataCount_MissingVersion(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "count", "--schema-id", "2546"})

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing --version")
	}
}

func TestTableDataCount_InvalidFilter(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "count", "--schema-id", "2546", "--version", "1", "--filter", "{bad"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --filter JSON")
	}
	if lterrors.CategoryOf(err) != lterrors.CategoryValidation {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestTableDataCount_NoHostConfigured(t *testing.T) {
	f, _, errOut := newTestFactory("")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "count", "--schema-id", "2546", "--version", "1"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected host config error")
	}
}

// ==================== delete confirmation tests ====================

func TestTableDataDelete_NeedsConfirmation(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "delete", "--schema-id", "2546", "--id", "33832272"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation error without --yes")
	}
	if lterrors.CategoryOf(err) != lterrors.CategoryConfirmation {
		t.Fatalf("expected confirmation error, got %v", err)
	}
	if got := output.ExitCodeOf(err); got != output.ExitConfirmationRequired {
		t.Fatalf("exit code = %d, want %d", got, output.ExitConfirmationRequired)
	}
}

func TestTableDataDelete_ConfirmedWithYes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := decodeProxyRequest(t, r)
		if req.Path != "/basicdata/record/batchDelete" {
			t.Fatalf("expected path /basicdata/record/batchDelete, got %s", req.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	f, out, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "delete", "--schema-id", "2546", "--id", "33832272", "--yes"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTableDataBatchDelete_NeedsConfirmation(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-delete", "--schema-id", "2546", "--ids", "1,2,3"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected confirmation error without --yes")
	}
	if lterrors.CategoryOf(err) != lterrors.CategoryConfirmation {
		t.Fatalf("expected confirmation error, got %v", err)
	}
	if got := output.ExitCodeOf(err); got != output.ExitConfirmationRequired {
		t.Fatalf("exit code = %d, want %d", got, output.ExitConfirmationRequired)
	}
}

func TestTableDataBatchDelete_ConfirmedWithYes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := decodeProxyRequest(t, r)
		if req.Path != "/basicdata/record/batchDelete" {
			t.Fatalf("expected path /basicdata/record/batchDelete, got %s", req.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	f, out, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"data", "batch-delete", "--schema-id", "2546", "--ids", "1,2,3", "--yes"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ==================== schema update tests ====================

func TestNewCmdTableSchemaUpdate_WithColumnsSchema(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := decodeProxyRequest(t, r)
		callCount++

		if callCount == 1 {
			// First call: auto-fetch version
			if req.Path != "/basicdata/schema/get" {
				t.Fatalf("expected path /basicdata/schema/get, got %s", req.Path)
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"result": map[string]interface{}{
					"version": 5,
				},
			})
		} else {
			// Second call: update schema
			if req.Path != "/basicdata/schema/update" {
				t.Fatalf("expected path /basicdata/schema/update, got %s", req.Path)
			}
			// Check that version was auto-filled
			if version, ok := req.Body["version"].(float64); !ok || int(version) != 5 {
				t.Fatalf("expected body version=5, got %v", req.Body["version"])
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"result":  map[string]interface{}{"id": 2546, "version": 6},
			})
		}
	}))
	defer server.Close()

	f, out, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"schema", "update", "--schema-id", "2546", "--basic-data-id", "1633",
		"--columns-schema", `[{"id":1,"title":"Name","type":"text","key":"name","primaryKey":true,"show":true}]`})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableSchemaUpdate_WithSchema(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := decodeProxyRequest(t, r)
		if req.Path != "/basicdata/schema/update" {
			t.Fatalf("expected path /basicdata/schema/update, got %s", req.Path)
		}
		// Check that version from JSON is used
		if version, ok := req.Body["version"].(float64); !ok || int(version) != 3 {
			t.Fatalf("expected body version=3, got %v", req.Body["version"])
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"result":  map[string]interface{}{"id": 2546, "version": 4},
		})
	}))
	defer server.Close()

	f, out, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"schema", "update", "--schema-id", "2546",
		"--schema", `{"columnsSchema":[{"id":1,"title":"Name","type":"text"}],"version":3}`})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableSchemaUpdate_MissingBothFlags(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"schema", "update", "--schema-id", "2546"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --schema and --columns-schema")
	}
	if !strings.Contains(err.Error(), "either --schema or --columns-schema") {
		t.Fatalf("expected either flag error, got %v", err)
	}
}

func TestNewCmdTableSchemaUpdate_AutoVersionRequiresBasicDataId(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	// No --basic-data-id and no version in JSON
	cmd.SetArgs([]string{"schema", "update", "--schema-id", "2546",
		"--columns-schema", `[{"id":1,"title":"Name","type":"text"}]`})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --basic-data-id")
	}
	if !strings.Contains(err.Error(), "--basic-data-id is required") {
		t.Fatalf("expected basic-data-id required error, got %v", err)
	}
}

// ==================== table view ====================

func TestNewCmdTableViewSave_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/tableView/save" {
			t.Errorf("expected proxy path /tableView/save, got %s", proxyReq.Path)
		}
		if proxyReq.Method != "POST" {
			t.Errorf("expected method POST, got %s", proxyReq.Method)
		}
		body := proxyReq.Body
		if schemaId, ok := body["schemaId"]; !ok || schemaId != float64(2367) {
			t.Errorf("expected schemaId=2367, got %v", body["schemaId"])
		}
		if name, ok := body["name"]; !ok || name != "测试视图" {
			t.Errorf("expected name=测试视图, got %v", body["name"])
		}
		if gm, ok := body["groupMetadata"]; !ok {
			t.Error("expected groupMetadata in body")
		} else {
			gmMap := gm.(map[string]interface{})
			if gmMap["columnId"] != "1" {
				t.Errorf("expected columnId=1, got %v", gmMap["columnId"])
			}
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  map[string]interface{}{"id": 688, "name": "测试视图"},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"view", "save", "--schema-id", "2367", "--name", "测试视图", "--group-column", "1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableViewList_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/tableView/list" {
			t.Errorf("expected proxy path /tableView/list, got %s", proxyReq.Path)
		}
		if proxyReq.Method != "GET" {
			t.Errorf("expected method GET, got %s", proxyReq.Method)
		}
		// GET params are in URL query, not in proxy body
		query := r.URL.Query()
		if got := query.Get("schemaId"); got != "2367" {
			t.Errorf("expected schemaId=2367 in query, got %s", got)
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  []interface{}{map[string]interface{}{"id": 621, "name": "表格"}},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"view", "list", "--schema-id", "2367"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableViewUpdate_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/tableView/update" {
			t.Errorf("expected proxy path /tableView/update, got %s", proxyReq.Path)
		}
		body := proxyReq.Body
		if id, ok := body["id"]; !ok || id != float64(621) {
			t.Errorf("expected id=621, got %v", body["id"])
		}
		if schemaId, ok := body["schemaId"]; !ok || schemaId != float64(2367) {
			t.Errorf("expected schemaId=2367, got %v", body["schemaId"])
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  map[string]interface{}{"id": 621},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"view", "update", "--id", "621", "--schema-id", "2367", "--name", "新名称"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableViewDelete_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/tableView/delete" {
			t.Errorf("expected proxy path /tableView/delete, got %s", proxyReq.Path)
		}
		if proxyReq.Method != "GET" {
			t.Errorf("expected method GET, got %s", proxyReq.Method)
		}
		// GET params are in URL query, not in proxy body
		query := r.URL.Query()
		if got := query.Get("id"); got != "621" {
			t.Errorf("expected id=621 in query, got %s", got)
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"view", "delete", "--id", "621"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableViewGroupData_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/tableView/getGroupData" {
			t.Errorf("expected proxy path /tableView/getGroupData, got %s", proxyReq.Path)
		}
		if proxyReq.Method != "GET" {
			t.Errorf("expected method GET, got %s", proxyReq.Method)
		}
		// GET params are in URL query, not in proxy body
		query := r.URL.Query()
		if got := query.Get("id"); got != "621" {
			t.Errorf("expected id=621 in query, got %s", got)
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  []interface{}{map[string]interface{}{"esKey": "data1568.1", "key": "测试", "count": 1}},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"view", "group-data", "--view-id", "621"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableViewMerits_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/tableView/getMerits" {
			t.Errorf("expected proxy path /tableView/getMerits, got %s", proxyReq.Path)
		}
		if proxyReq.Method != "POST" {
			t.Errorf("expected method POST, got %s", proxyReq.Method)
		}
		body := proxyReq.Body
		if schemaId, ok := body["schemaId"]; !ok || schemaId != float64(2367) {
			t.Errorf("expected schemaId=2367, got %v", body["schemaId"])
		}
		if viewId, ok := body["viewId"]; !ok || viewId != float64(621) {
			t.Errorf("expected viewId=621, got %v", body["viewId"])
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  []interface{}{map[string]interface{}{"columnId": "0", "esMerit": "count", "value": "76159"}},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"view", "merits", "--schema-id", "2367", "--view-id", "621"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableDataQuery_WithViewGroupData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/basicdata/record/listNew" {
			t.Errorf("expected proxy path /basicdata/record/listNew, got %s", proxyReq.Path)
		}
		body := proxyReq.Body
		if vgd, ok := body["viewGroupData"]; !ok {
			t.Error("expected viewGroupData in body")
		} else {
			vgdStr := vgd.(string)
			if !strings.Contains(vgdStr, "data1568.1") {
				t.Errorf("expected viewGroupData to contain data1568.1, got %s", vgdStr)
			}
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  map[string]interface{}{"data": []interface{}{}},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{
		"data", "query",
		"--basic-data-id", "1568",
		"--view-group-data", `{"esKey":"data1568.1","key":"测试","leafId":"#data1568.1@测试"}`,
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ==================== table pivot ====================

func TestNewCmdTablePivotQuery_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/api/pivot-table/query" {
			t.Errorf("expected proxy path /api/pivot-table/query, got %s", proxyReq.Path)
		}
		if proxyReq.Method != "POST" {
			t.Errorf("expected method POST, got %s", proxyReq.Method)
		}
		body := proxyReq.Body
		if tableId, ok := body["tableId"]; !ok || tableId != "1568" {
			t.Errorf("expected tableId=1568, got %v", body["tableId"])
		}
		if viewId, ok := body["viewId"]; !ok || viewId != "621" {
			t.Errorf("expected viewId=621, got %v", body["viewId"])
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  map[string]interface{}{"data": map[string]interface{}{"columns": []interface{}{}, "rows": []interface{}{}}},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"pivot", "query", "--table-id", "1568", "--view-id", "621"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTablePivotConfig_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/api/pivot-table/config" {
			t.Errorf("expected proxy path /api/pivot-table/config, got %s", proxyReq.Path)
		}
		if proxyReq.Method != "POST" {
			t.Errorf("expected method POST, got %s", proxyReq.Method)
		}
		body := proxyReq.Body
		if tableId, ok := body["tableId"]; !ok || tableId != "1568" {
			t.Errorf("expected tableId=1568, got %v", body["tableId"])
		}
		if viewId, ok := body["viewId"]; !ok || viewId != "621" {
			t.Errorf("expected viewId=621, got %v", body["viewId"])
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  map[string]interface{}{"dimensions": []interface{}{}, "measures": []interface{}{}},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"pivot", "config", "--table-id", "1568", "--view-id", "621"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTablePivotConfigSave_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		if proxyReq.Path != "/api/pivot-table/config/save" {
			t.Errorf("expected proxy path /api/pivot-table/config/save, got %s", proxyReq.Path)
		}
		if proxyReq.Method != "POST" {
			t.Errorf("expected method POST, got %s", proxyReq.Method)
		}
		body := proxyReq.Body
		if tableId, ok := body["tableId"]; !ok || tableId != "1568" {
			t.Errorf("expected tableId=1568, got %v", body["tableId"])
		}
		if viewId, ok := body["viewId"]; !ok || viewId != "621" {
			t.Errorf("expected viewId=621, got %v", body["viewId"])
		}
		if businessId, ok := body["businessId"]; !ok || businessId != float64(1568) {
			t.Errorf("expected businessId=1568, got %v", body["businessId"])
		}
		if _, ok := body["config"]; !ok {
			t.Error("expected config in body")
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  map[string]interface{}{"viewId": "621"},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{
		"pivot", "config-save",
		"--table-id", "1568",
		"--view-id", "621",
		"--business-id", "1568",
		"--config", `{"dimensions":[],"measures":[],"advanced":{"openDimensionSplit":false,"openStatistic":true,"displayMode":"normal"}}`,
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ==================== table update ====================

func TestNewCmdTableUpdate_ProxyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxyReq := decodeProxyRequest(t, r)
		// First request: GET /basicdata/get
		if proxyReq.Path == "/basicdata/get" {
			writeProxyResponse(t, w, map[string]interface{}{
				"success": true,
				"result": map[string]interface{}{
					"id":            float64(1568),
					"appId":         float64(165),
					"name":          "快麦出入库记录",
					"source":        float64(1),
					"type":          float64(1),
					"openHighMode":  float64(1),
					"openConnector": float64(1),
				},
			})
			return
		}
		// Second request: POST /basicdata/update
		if proxyReq.Path != "/basicdata/update" {
			t.Errorf("expected proxy path /basicdata/update, got %s", proxyReq.Path)
		}
		if proxyReq.Method != "POST" {
			t.Errorf("expected method POST, got %s", proxyReq.Method)
		}
		body := proxyReq.Body
		if id, ok := body["id"]; !ok || id != float64(1568) {
			t.Errorf("expected id=1568, got %v", body["id"])
		}
		if openConnector, ok := body["openConnector"]; !ok || openConnector != float64(1) {
			t.Errorf("expected openConnector=1, got %v", body["openConnector"])
		}

		writeProxyResponse(t, w, map[string]interface{}{
			"success": true,
			"result":  map[string]interface{}{"id": float64(1568)},
		})
	}))
	defer server.Close()

	f, _, errOut := newTestFactory(server.URL)
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"update", "--id", "1568", "--open-connector", "1"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewCmdTableUpdate_MissingID(t *testing.T) {
	f, _, errOut := newTestFactory("http://example.com")
	cmd := NewCmdTable(f)
	cmd.SetErr(errOut)
	cmd.SetArgs([]string{"update"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --id")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Fatalf("expected id required error, got %v", err)
	}
}
