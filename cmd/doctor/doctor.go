// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package doctor

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/lingtong/cli/internal/auth"
	"github.com/lingtong/cli/internal/build"
	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	"github.com/lingtong/cli/internal/openapi"
	"github.com/spf13/cobra"
)

type checkStatus string

const (
	statusOK   checkStatus = "ok"
	statusWarn checkStatus = "warn"
	statusFail checkStatus = "fail"
	statusSkip checkStatus = "skip"
)

type checkResult struct {
	Name    string
	Status  checkStatus
	Message string
}

// NewCmdDoctor creates the doctor command for diagnosing CLI health.
func NewCmdDoctor(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose CLI health and configuration",
		Long: `Check CLI version, configuration, authentication, and API connectivity.

EXAMPLES:
    lingtong-cli doctor`,
		RunE: func(cmd *cobra.Command, args []string) error {
			results := runChecks(f)
			printReport(f.IOStreams.Out, results)
			return nil
		},
	}
	return cmd
}

func runChecks(f *cmdutil.Factory) []checkResult {
	var results []checkResult
	results = append(results, checkVersion())
	results = append(results, checkConfigFile())
	results = append(results, checkHost(f))
	results = append(results, checkAuth(f))
	results = append(results, checkAPI(f))
	results = append(results, checkEnvVars())
	results = append(results, checkServiceSpec())
	return results
}

func checkServiceSpec() checkResult {
	// Surface a misconfigured explicit override as a diagnostic warning, since
	// the user clearly intended to use it.
	if env := os.Getenv("LINGTONG_OPENAPI"); env != "" {
		if spec, err := openapi.Parse(env); err == nil {
			return checkResult{"Service Spec", statusOK, fmt.Sprintf("override %s (%d operations)", env, countOperations(spec))}
		}
		return checkResult{"Service Spec", statusWarn, fmt.Sprintf("LINGTONG_OPENAPI set but unreadable: %s", env)}
	}
	// Otherwise report the same source `service` and `schema` resolve to.
	if override := openapi.OverridePath(); override != "" {
		if spec, err := openapi.Parse(override); err == nil {
			return checkResult{"Service Spec", statusOK, fmt.Sprintf("override %s (%d operations)", override, countOperations(spec))}
		}
		return checkResult{"Service Spec", statusWarn, fmt.Sprintf("override set but unreadable: %s", override)}
	}
	if openapi.HasEmbeddedSpec() {
		if spec, err := openapi.EmbeddedSpec(); err == nil {
			return checkResult{"Service Spec", statusOK, fmt.Sprintf("embedded (%d operations)", countOperations(spec))}
		}
		return checkResult{"Service Spec", statusWarn, "embedded spec failed to parse"}
	}
	return checkResult{"Service Spec", statusWarn, "no spec available; 'service' commands disabled"}
}

func countOperations(spec *openapi.Spec) int {
	count := 0
	for _, ops := range spec.GroupByPrefix() {
		count += len(ops)
	}
	return count
}

func checkVersion() checkResult {
	v := build.Version
	if v == "" || v == "dev" {
		return checkResult{"CLI Version", statusWarn, fmt.Sprintf("%s (development build)", v)}
	}
	date := build.Date
	if date != "" && date != "unknown" {
		return checkResult{"CLI Version", statusOK, fmt.Sprintf("%s (%s)", v, date)}
	}
	return checkResult{"CLI Version", statusOK, v}
}

func checkConfigFile() checkResult {
	path, err := config.DefaultConfigPath()
	if err != nil {
		return checkResult{"Config File", statusFail, fmt.Sprintf("cannot determine path: %v", err)}
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return checkResult{"Config File", statusFail, fmt.Sprintf("not found at %s", path)}
		}
		return checkResult{"Config File", statusFail, fmt.Sprintf("error: %v", err)}
	}
	if info.IsDir() {
		return checkResult{"Config File", statusFail, fmt.Sprintf("path is a directory: %s", path)}
	}
	return checkResult{"Config File", statusOK, path}
}

func checkHost(f *cmdutil.Factory) checkResult {
	host := strings.TrimSpace(f.Config.Host)
	if host == "" {
		return checkResult{"Host", statusFail, "not configured"}
	}
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		return checkResult{"Host", statusWarn, fmt.Sprintf("%s (missing scheme)", host)}
	}
	return checkResult{"Host", statusOK, host}
}

func checkAuth(f *cmdutil.Factory) checkResult {
	token, err := auth.GetToken()
	if err != nil {
		return checkResult{"Auth", statusFail, "not authenticated"}
	}
	if token == "" {
		return checkResult{"Auth", statusFail, "not authenticated"}
	}
	masked := maskToken(token)
	return checkResult{"Auth", statusOK, fmt.Sprintf("authenticated (%s)", masked)}
}

func checkAPI(f *cmdutil.Factory) checkResult {
	host := strings.TrimSpace(f.Config.Host)
	if host == "" {
		return checkResult{"API", statusSkip, "no host configured"}
	}

	start := time.Now()
	c := client.NewClientWithTimeout(host, f.Config.Token, 5*time.Second)
	_, err := c.Get("/gw/ai/connector/info", map[string]interface{}{
		"connector": "kmerp",
		"env":       "test",
	})
	latency := time.Since(start)

	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "401") || strings.Contains(errStr, "无身份信息") {
			return checkResult{"API", statusWarn, fmt.Sprintf("reachable but auth failed (latency: %v)", latency.Round(time.Millisecond))}
		}
		return checkResult{"API", statusFail, fmt.Sprintf("unreachable: %v", err)}
	}
	return checkResult{"API", statusOK, fmt.Sprintf("reachable (latency: %v)", latency.Round(time.Millisecond))}
}

func checkEnvVars() checkResult {
	var set []string
	var unset []string

	envVars := []struct {
		name     string
		required bool
	}{
		{"LINGTONG_TOKEN", false},
		{"LINGTONG_API_TOKEN", false},
		{"LINGTONG_HOST", false},
	}

	for _, ev := range envVars {
		if v := os.Getenv(ev.name); v != "" {
			set = append(set, ev.name)
		} else {
			unset = append(unset, ev.name)
		}
	}

	if len(set) > 0 {
		return checkResult{"Env Vars", statusOK, fmt.Sprintf("set: %s", strings.Join(set, ", "))}
	}
	return checkResult{"Env Vars", statusSkip, "none set (optional)"}
}

func printReport(w io.Writer, results []checkResult) {
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "lingtong-cli Doctor Report")
	fmt.Fprintln(w, "==========================")
	fmt.Fprintln(w, "")

	passed := 0
	total := len(results)

	for _, r := range results {
		var icon string
		switch r.Status {
		case statusOK:
			icon = "[OK]"
			passed++
		case statusWarn:
			icon = "[!!]"
			passed++ // warnings still count as passed
		case statusFail:
			icon = "[XX]"
		case statusSkip:
			icon = "[--]"
			passed++ // skipped checks don't count as failures
		}
		fmt.Fprintf(w, "  %s %s: %s\n", icon, r.Name, r.Message)
	}

	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "Result: %d/%d checks passed\n", passed, total)
	fmt.Fprintln(w, "")
}

func maskToken(token string) string {
	if len(token) <= 12 {
		return "****"
	}
	suffixLen := min(5, len(token)-10)
	return token[:8] + "..." + token[len(token)-suffixLen:]
}
