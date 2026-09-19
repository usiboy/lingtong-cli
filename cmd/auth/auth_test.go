// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/lingtong/cli/internal/cmdutil"
	"github.com/lingtong/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func overrideAuthDependencies(t *testing.T, store func(string) error, get func() (string, error), delete func() error, verify func(string, string) error) {
	t.Helper()

	originalStoreToken := storeToken
	originalGetToken := getToken
	originalDeleteToken := deleteToken
	originalVerifyTokenFunc := verifyTokenFunc

	if store != nil {
		storeToken = store
	}
	if get != nil {
		getToken = get
	}
	if delete != nil {
		deleteToken = delete
	}
	if verify != nil {
		verifyTokenFunc = verify
	}

	t.Cleanup(func() {
		storeToken = originalStoreToken
		getToken = originalGetToken
		deleteToken = originalDeleteToken
		verifyTokenFunc = originalVerifyTokenFunc
	})
}

// TestNewCmdAuth tests the creation of the auth command
func TestNewCmdAuth(t *testing.T) {
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host:  "https://test.example.com",
			Brand: "lingtong",
		},
	}

	cmd := NewCmdAuth(f)

	assert.Equal(t, "auth", cmd.Use)
	assert.Contains(t, cmd.Short, "authentication")
	assert.NotNil(t, cmd)

	// Verify subcommands are registered
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 6)

	cmdNames := make([]string, len(subcommands))
	for i, c := range subcommands {
		cmdNames[i] = c.Name()
	}
	assert.Contains(t, cmdNames, "login")
	assert.Contains(t, cmdNames, "status")
	assert.Contains(t, cmdNames, "logout")
	assert.Contains(t, cmdNames, "list")
	assert.Contains(t, cmdNames, "use")
	assert.Contains(t, cmdNames, "remove")
}

// TestNewCmdAuthLogin tests the login command structure
func TestNewCmdAuthLogin_Structure(t *testing.T) {
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host: "https://test.example.com",
		},
	}

	cmd := NewCmdAuth(f)
	loginCmd, _, err := cmd.Find([]string{"login"})
	require.NoError(t, err)

	assert.Equal(t, "login", loginCmd.Use)
	assert.Contains(t, loginCmd.Short, "Login")

	// Verify flags exist
	tokenFlag := loginCmd.Flags().Lookup("token")
	require.NotNil(t, tokenFlag)
	assert.Equal(t, "", tokenFlag.DefValue)

	fromEnvFlag := loginCmd.Flags().Lookup("from-env")
	require.NotNil(t, fromEnvFlag)
	assert.Equal(t, "false", fromEnvFlag.DefValue)
}

// TestMaskToken tests the maskToken function with various inputs
func TestMaskToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "empty token",
			token:    "",
			expected: "****",
		},
		{
			name:     "short token (8 chars)",
			token:    "apk-1234",
			expected: "****",
		},
		{
			name:     "exact 12 chars token",
			token:    "apk-12345678",
			expected: "****",
		},
		{
			name:     "13 chars token (minimum to show masked)",
			token:    "apk-123456789",
			expected: "apk-1234...789",
		},
		{
			name:     "standard length token",
			token:    "apk-Gx6vDOEmALY7iJRcLZcD4nWF",
			expected: "apk-Gx6v...D4nWF",
		},
		{
			name:     "very long token",
			token:    "apk-Gx6vDOEmALY7iJRcLZcD4nWFabcdef1234567890",
			expected: "apk-Gx6v...67890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskToken(tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestVerifyToken tests the verifyToken function
func TestVerifyToken(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		statusCode  int
		response    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid token - successful response",
			token:       "apk-valid-token",
			statusCode:  http.StatusOK,
			response:    `{"success": true}`,
			expectError: false,
		},
		{
			name:        "invalid token - 401 unauthorized",
			token:       "apk-invalid-token",
			statusCode:  http.StatusUnauthorized,
			response:    `{"error": "无身份信息"}`,
			expectError: true,
			errorMsg:    "invalid token",
		},
		{
			name:        "invalid token - 401 without Chinese message",
			token:       "apk-invalid-token",
			statusCode:  http.StatusUnauthorized,
			response:    `{"error": "unauthorized"}`,
			expectError: true,
			errorMsg:    "invalid token",
		},
		{
			name:        "server error - 500 (not critical)",
			token:       "apk-valid-token",
			statusCode:  http.StatusInternalServerError,
			response:    `{"error": "internal error"}`,
			expectError: false,
		},
		{
			name:        "not found - 404 (not critical)",
			token:       "apk-valid-token",
			statusCode:  http.StatusNotFound,
			response:    `{"error": "not found"}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request is proxied correctly
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/gw/ai/proxy", r.URL.Path)
				assert.Contains(t, r.Header.Get("Authorization"), "Bearer "+tt.token)

				// Verify proxy body contains the expected path
				body, _ := io.ReadAll(r.Body)
				var proxyReq map[string]interface{}
				err := json.Unmarshal(body, &proxyReq)
				require.NoError(t, err)
				assert.Equal(t, "/gw/ai/connector/info", proxyReq["path"])

				// Return mock response
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			err := verifyToken(server.URL, tt.token)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestVerifyToken_InvalidHost tests verifyToken with invalid host
func TestVerifyToken_InvalidHost(t *testing.T) {
	// This should not return an error for network failures (only 401 is critical)
	err := verifyToken("http://localhost:99999", "apk-test-token")
	// Network error is not critical, so should return nil
	assert.NoError(t, err)
}

// TestNewCmdAuthLogin_NoHost tests login without configured host
func TestNewCmdAuthLogin_NoHost(t *testing.T) {
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host: "",
		},
	}

	cmd := NewCmdAuth(f)
	loginCmd, _, _ := cmd.Find([]string{"login"})

	err := loginCmd.RunE(loginCmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no host configured")
}

// TestNewCmdAuthStatus_NoToken tests status when not authenticated
func TestNewCmdAuthStatus_NoToken(t *testing.T) {
	overrideAuthDependencies(t, nil, func() (string, error) {
		return "", fmt.Errorf("not authenticated")
	}, nil, nil)

	f := &cmdutil.Factory{
		Config: &config.Config{
			Host: "https://test.example.com",
		},
	}

	cmd := NewCmdAuth(f)
	statusCmd, _, _ := cmd.Find([]string{"status"})

	// Capture output
	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = statusCmd.RunE(statusCmd, []string{})

	w.Close()
	os.Stdout = oldStdout
	io.Copy(&buf, r)

	output := buf.String()
	assert.Contains(t, output, "Not authenticated")
}

// TestNewCmdAuthStatus_WithToken tests status when authenticated
func TestNewCmdAuthStatus_WithToken(t *testing.T) {
	token := "apk-status-test-token"
	overrideAuthDependencies(t, nil, func() (string, error) {
		return token, nil
	}, nil, func(host, token string) error {
		assert.Equal(t, "https://test.example.com", host)
		assert.Equal(t, "apk-status-test-token", token)
		return nil
	})

	f := &cmdutil.Factory{
		Config: &config.Config{
			Host: "https://test.example.com",
		},
	}

	cmd := NewCmdAuth(f)
	statusCmd, _, _ := cmd.Find([]string{"status"})

	var buf bytes.Buffer
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err := statusCmd.RunE(statusCmd, []string{})

	w.Close()
	os.Stdout = oldStdout
	io.Copy(&buf, r)

	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Authenticated: Yes")
	assert.Contains(t, output, maskToken(token))
	assert.Contains(t, output, "https://test.example.com")
	assert.Contains(t, output, "Valid")
}

// TestNewCmdLogout tests logout command
func TestNewCmdLogout(t *testing.T) {
	f := &cmdutil.Factory{
		Config: &config.Config{
			Host: "https://test.example.com",
		},
	}

	cmd := NewCmdAuth(f)
	logoutCmd, _, _ := cmd.Find([]string{"logout"})

	assert.Equal(t, "logout", logoutCmd.Use)
	assert.Contains(t, logoutCmd.Short, "Logout")
}

// TestMaskToken_EdgeCases tests edge cases for maskToken
func TestMaskToken_EdgeCases(t *testing.T) {
	// Test with exactly 12 characters
	assert.Equal(t, "****", maskToken("123456789012"))

	// Test with 11 characters
	assert.Equal(t, "****", maskToken("12345678901"))

	// Test special characters in token
	specialToken := "apk-!@#$%^&*()_+-=[]{}|;':\",./<>?abcdef"
	result := maskToken(specialToken)
	assert.True(t, strings.HasSuffix(result, "bcdef"))
	assert.True(t, strings.HasPrefix(result, "apk-!@#$"))
}

// TestVerifyToken_EmptyToken tests verifyToken with empty token
func TestVerifyToken_EmptyToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should still make the request even with empty token
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "无身份信息"}`))
	}))
	defer server.Close()

	err := verifyToken(server.URL, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")
}

// TestVerifyToken_ResponseContains401 tests when response body contains "401"
func TestVerifyToken_ResponseContains401(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Response contains "401" in body but status is 200
		w.Write([]byte(`{"code": 401, "message": "session expired"}`))
	}))
	defer server.Close()

	err := verifyToken(server.URL, "apk-test")
	// Should not error since HTTP status is 200
	assert.NoError(t, err)
}

// TestNewCmdAuthLogin_TokenFromEnv tests login with --from-env flag
func TestNewCmdAuthLogin_TokenFromEnv(t *testing.T) {
	t.Setenv("LINGTONG_API_TOKEN", "")

	f := &cmdutil.Factory{
		Config: &config.Config{
			Host: "https://test.example.com",
		},
	}

	cmd := NewCmdAuth(f)
	loginCmd, _, _ := cmd.Find([]string{"login"})

	// Set flags manually
	loginCmd.Flags().Set("from-env", "true")

	// Since token is empty and from-env is true but no env var, should error
	err := loginCmd.RunE(loginCmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LINGTONG_API_TOKEN environment variable not set")
}

// TestNewCmdAuthLogin_TokenFromEnv_WithValue tests login with --from-env and valid env var
func TestNewCmdAuthLogin_TokenFromEnv_WithValue(t *testing.T) {
	token := "apk-env-test-token"
	var storedToken string
	var verifiedToken string
	originalStoreTokenForAuth := storeTokenForAuth
	storeTokenForAuth = func(_, _ string, token string) error {
		storedToken = token
		return nil
	}
	t.Cleanup(func() { storeTokenForAuth = originalStoreTokenForAuth })
	t.Setenv("HOME", t.TempDir())
	overrideAuthDependencies(t, func(token string) error {
		return nil
	}, nil, nil, func(host, token string) error {
		assert.Equal(t, "https://test.example.com", host)
		verifiedToken = token
		return nil
	})
	t.Setenv("LINGTONG_API_TOKEN", token)

	f := &cmdutil.Factory{
		Config: &config.Config{
			Host: "https://test.example.com",
		},
	}

	cmd := NewCmdAuth(f)
	loginCmd, _, _ := cmd.Find([]string{"login"})
	loginCmd.Flags().Set("from-env", "true")

	err := loginCmd.RunE(loginCmd, []string{})
	require.NoError(t, err)
	assert.Equal(t, token, storedToken)
	assert.Equal(t, token, verifiedToken)
}

// TestVerifyToken_ChineseErrorMessage tests Chinese error message detection
func TestVerifyToken_ChineseErrorMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "无身份信息", "code": 401}`))
	}))
	defer server.Close()

	err := verifyToken(server.URL, "apk-test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")
	assert.Contains(t, err.Error(), "authentication failed")
}

// TestNewCmdAuth_LongDescription tests the long description
func TestNewCmdAuth_LongDescription(t *testing.T) {
	f := &cmdutil.Factory{
		Config: &config.Config{Host: "https://test.com"},
	}

	cmd := NewCmdAuth(f)
	assert.Contains(t, cmd.Long, "API token")
	assert.Contains(t, cmd.Long, "apk-xxx")
}

// TestNewCmdAuthLogin_LongDescription tests login long description
func TestNewCmdAuthLogin_LongDescription(t *testing.T) {
	f := &cmdutil.Factory{
		Config: &config.Config{Host: "https://test.com"},
	}

	cmd := NewCmdAuth(f)
	loginCmd, _, _ := cmd.Find([]string{"login"})
	assert.Contains(t, loginCmd.Long, "--token")
	assert.Contains(t, loginCmd.Long, "--from-env")
	assert.Contains(t, loginCmd.Long, "apk-xxx")
}

// TestMaskToken_FuzzLike tests various token lengths
func TestMaskToken_FuzzLike(t *testing.T) {
	tests := []struct {
		length   int
		expected string
	}{
		{0, "****"},
		{1, "****"},
		{5, "****"},
		{10, "****"},
		{12, "****"},
		{13, ""}, // Will be calculated
		{20, ""},
		{50, ""},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("length_%d", tt.length), func(t *testing.T) {
			token := strings.Repeat("a", tt.length)
			result := maskToken(token)

			if tt.length <= 12 {
				assert.Equal(t, "****", result)
			} else {
				suffixLen := min(5, tt.length-10)
				expected := strings.Repeat("a", 8) + "..." + strings.Repeat("a", suffixLen)
				assert.Equal(t, expected, result)
				assert.Len(t, result, 8+3+suffixLen)
			}
		})
	}
}
