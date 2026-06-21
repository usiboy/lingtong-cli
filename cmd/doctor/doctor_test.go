// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package doctor

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	"github.com/lingtong/cli/internal/output"
)

func newTestFactory(host string) *cmdutil.Factory {
	return &cmdutil.Factory{
		Config: &config.Config{Host: host},
		IOStreams: &output.IOStreams{
			In:     &bytes.Buffer{},
			Out:    &bytes.Buffer{},
			ErrOut: &bytes.Buffer{},
		},
	}
}

func TestDoctorReportOutput(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": {}}`))
	}))
	defer server.Close()

	f := newTestFactory(server.URL)
	var buf bytes.Buffer
	f.IOStreams.Out = &buf

	results := runChecks(f)
	printReport(&buf, results)

	output := buf.String()

	// Should contain report header
	if !strings.Contains(output, "Doctor Report") {
		t.Errorf("output should contain 'Doctor Report': %s", output)
	}

	// Should contain check results
	if !strings.Contains(output, "CLI Version") {
		t.Error("output should contain CLI Version check")
	}
	if !strings.Contains(output, "Config File") {
		t.Error("output should contain Config File check")
	}
	if !strings.Contains(output, "Host") {
		t.Error("output should contain Host check")
	}
	if !strings.Contains(output, "Result:") {
		t.Error("output should contain Result summary")
	}
}

func TestDoctorNoHost(t *testing.T) {
	f := newTestFactory("")
	var buf bytes.Buffer
	f.IOStreams.Out = &buf

	results := runChecks(f)
	printReport(&buf, results)

	output := buf.String()
	if !strings.Contains(output, "not configured") {
		t.Errorf("output should mention 'not configured': %s", output)
	}
}

func TestDoctorAPIUnreachable(t *testing.T) {
	// Use a server that immediately closes connections
	f := newTestFactory("http://127.0.0.1:1")
	var buf bytes.Buffer
	f.IOStreams.Out = &buf

	results := runChecks(f)
	printReport(&buf, results)

	output := buf.String()
	// Should show API as fail or warn
	if !strings.Contains(output, "API") {
		t.Error("output should contain API check")
	}
}

func TestCheckVersion(t *testing.T) {
	result := checkVersion()
	if result.Name != "CLI Version" {
		t.Errorf("Name = %q, want %q", result.Name, "CLI Version")
	}
	// In test environment, version is usually "dev"
	if result.Status != statusOK && result.Status != statusWarn {
		t.Errorf("Status = %v, want ok or warn", result.Status)
	}
}

func TestCheckHostEmpty(t *testing.T) {
	f := newTestFactory("")
	result := checkHost(f)
	if result.Status != statusFail {
		t.Errorf("Status = %v, want fail for empty host", result.Status)
	}
}

func TestCheckHostValid(t *testing.T) {
	f := newTestFactory("https://example.com")
	result := checkHost(f)
	if result.Status != statusOK {
		t.Errorf("Status = %v, want ok for valid host", result.Status)
	}
	if !strings.Contains(result.Message, "https://example.com") {
		t.Errorf("Message = %q, should contain host URL", result.Message)
	}
}

func TestCheckHostNoScheme(t *testing.T) {
	f := newTestFactory("example.com")
	result := checkHost(f)
	if result.Status != statusWarn {
		t.Errorf("Status = %v, want warn for host without scheme", result.Status)
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "****"},
		{"short", "****"},
		{"apk-12345678", "****"},
		{"apk-1234567890abcdef", "apk-1234...bcdef"},
	}
	for _, tt := range tests {
		got := maskToken(tt.input)
		if got != tt.want {
			t.Errorf("maskToken(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestPrintReportFormat(t *testing.T) {
	results := []checkResult{
		{"Test1", statusOK, "all good"},
		{"Test2", statusFail, "something wrong"},
		{"Test3", statusWarn, "be careful"},
		{"Test4", statusSkip, "not applicable"},
	}
	var buf bytes.Buffer
	printReport(&buf, results)

	output := buf.String()
	if !strings.Contains(output, "[OK]") {
		t.Error("missing [OK] marker")
	}
	if !strings.Contains(output, "[XX]") {
		t.Error("missing [XX] marker")
	}
	if !strings.Contains(output, "[!!]") {
		t.Error("missing [!!] marker")
	}
	if !strings.Contains(output, "[--]") {
		t.Error("missing [--] marker")
	}
	if !strings.Contains(output, "3/4 checks passed") {
		t.Errorf("expected 3/4 passed, got: %s", output)
	}
}

func TestCheckServiceSpec_Embedded(t *testing.T) {
	t.Setenv("LINGTONG_OPENAPI", "")
	r := checkServiceSpec()
	if r.Status != statusOK {
		t.Errorf("Status = %v, want ok (embedded spec)", r.Status)
	}
	if !strings.Contains(r.Message, "embedded") {
		t.Errorf("Message = %q, want it to mention embedded", r.Message)
	}
	if !strings.Contains(r.Message, "operations") {
		t.Errorf("Message = %q, want operation count", r.Message)
	}
}

func TestCheckServiceSpec_BadOverride(t *testing.T) {
	t.Setenv("LINGTONG_OPENAPI", "/nonexistent/openapi.json")
	r := checkServiceSpec()
	if r.Status != statusWarn {
		t.Errorf("Status = %v, want warn for unreadable override", r.Status)
	}
}

func TestCheckServiceSpec_GoodOverride(t *testing.T) {
	spec := "../../internal/openapi/spec/openapi.json"
	if _, err := os.Stat(spec); err != nil {
		t.Skip("spec file not present")
	}
	t.Setenv("LINGTONG_OPENAPI", spec)
	r := checkServiceSpec()
	if r.Status != statusOK {
		t.Errorf("Status = %v, want ok for valid override", r.Status)
	}
	if !strings.Contains(r.Message, "override") {
		t.Errorf("Message = %q, want it to mention override", r.Message)
	}
}
