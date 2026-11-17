// Copyright (c) 2026 Lingtong
// SPDX-License-Identifier: MIT

package table

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/lingtong/cli/internal/client"
	"github.com/lingtong/cli/internal/cmdutil"
	lterrors "github.com/lingtong/cli/internal/errors"
	"github.com/lingtong/cli/internal/output"
	"github.com/spf13/cobra"
)

type importOptions struct {
	file, name         string
	basicDataID, appID int64
	wait, dryRun       bool
	timeout, interval  time.Duration
}

type importProgress struct {
	Status  *int   `json:"status,omitempty"`
	Current *int   `json:"current,omitempty"`
	Total   *int   `json:"total,omitempty"`
	Message string `json:"message,omitempty"`
}

type importResult struct {
	BasicDataID int64  `json:"basicDataId,omitempty"`
	Name        string `json:"name,omitempty"`
	Created     bool   `json:"created"`
	UUID        string `json:"uuid,omitempty"`
	Completed   bool   `json:"completed"`
	importProgress
}

type importUploader func(context.Context, *client.Client, *os.File) (string, error)

func newCmdTableImport(f *cmdutil.Factory, existingOnly bool, upload importUploader) *cobra.Command {
	var o importOptions
	if upload == nil {
		upload = func(ctx context.Context, c *client.Client, file *os.File) (string, error) {
			return c.UploadTableFile(ctx, file)
		}
	}
	cmd := &cobra.Command{
		Use: "import", Short: "Import a local Excel/CSV file and wait for completion",
		Long: `Upload a local .xlsx, .xls or .csv file (maximum 50 MB) and import its data.
Use --basic-data-id for an existing table, or --app-id to create a table from
the file's headers. New table names default to the filename without extension.
Existing tables use their import-template headers and the server's primary-key
matching rules; matching records may be updated.

Waits for completion by default. --timeout bounds the entire operation.
--wait=false returns the task UUID after submission. A timeout or interruption
does not cancel the server task: resume with table import-progress --uuid UUID.
Progress goes to stderr; stdout contains only the final structured result.

EXAMPLES:
    lingtong-cli table import --app-id 165 --file ./orders.xlsx --name "订单"
    lingtong-cli table import --basic-data-id 1568 --file ./orders.xlsx
    lingtong-cli table data import --basic-data-id 1568 --file ./orders.csv --timeout 20m
    lingtong-cli table import --app-id 165 --file ./orders.xlsx --dry-run`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateImportOptions(cmd, o, existingOnly); err != nil {
				return err
			}
			file, err := openImportFile(o.file)
			if err != nil {
				return err
			}
			defer file.Close()
			if o.basicDataID == 0 && o.name == "" {
				o.name = strings.TrimSpace(strings.TrimSuffix(filepath.Base(o.file), filepath.Ext(o.file)))
				if o.name == "" {
					return invalidImport("--name must be nonempty")
				}
			}
			if o.dryRun {
				return f.NewWriter(output.Format(cmd.Flag("format").Value.String()), "table.import").Write(map[string]interface{}{
					"dryRun": true, "file": o.file, "basicDataId": o.basicDataID, "appId": o.appID,
					"name": o.name, "createTable": o.basicDataID == 0, "wait": o.wait,
					"timeout": o.timeout.String(), "pollInterval": o.interval.String(),
				})
			}
			if err := requireImportAuth(f); err != nil {
				return err
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			ctx, cancel := context.WithTimeout(ctx, o.timeout)
			defer cancel()
			c := client.NewClientWithTimeout(f.Config.Host, f.Config.Token, o.timeout)
			result := &importResult{BasicDataID: o.basicDataID, Name: o.name}
			if err := runTableImport(ctx, c, f, o, file, upload, result); err != nil {
				return importOperationError(err, result)
			}
			return writeImportResult(cmd, f, result, "table.import")
		},
	}
	cmd.Flags().StringVar(&o.file, "file", "", "Local .xlsx, .xls or .csv file (required, maximum 50 MB)")
	cmd.Flags().Int64Var(&o.basicDataID, "basic-data-id", 0, "Existing table ID")
	if !existingOnly {
		cmd.Flags().Int64Var(&o.appID, "app-id", 0, "Application ID for creating a table from the file")
		cmd.Flags().StringVar(&o.name, "name", "", "New table name (defaults to the filename without extension)")
	}
	cmd.Flags().BoolVar(&o.dryRun, "dry-run", false, "Validate the local file and show the plan without network requests")
	addImportWaitFlags(cmd, &o)
	return cmd
}

func addImportWaitFlags(cmd *cobra.Command, o *importOptions) {
	cmd.Flags().BoolVar(&o.wait, "wait", true, "Wait for the import result")
	cmd.Flags().DurationVar(&o.timeout, "timeout", 10*time.Minute, "Maximum total duration, including upload and polling")
	cmd.Flags().DurationVar(&o.interval, "poll-interval", 2*time.Second, "Interval between progress queries")
}

func invalidImport(message string) error {
	return lterrors.NewValidationError(lterrors.SubtypeInvalidArgument, "%s", message)
}

func validateImportWaitOptions(cmd *cobra.Command, o importOptions) error {
	if o.timeout <= 0 || o.interval <= 0 {
		return invalidImport("--timeout and --poll-interval must be positive")
	}
	format := cmd.Flag("format").Value.String()
	if format != "json" && format != "pretty" && format != "table" {
		return invalidImport("--format must be json, pretty or table")
	}
	return nil
}

func validateImportOptions(cmd *cobra.Command, o importOptions, existingOnly bool) error {
	if err := validateImportWaitOptions(cmd, o); err != nil {
		return err
	}
	if strings.TrimSpace(o.file) == "" {
		return invalidImport("--file is required")
	}
	if o.basicDataID < 0 || o.appID < 0 {
		return invalidImport("table and application IDs must be positive")
	}
	if existingOnly && o.basicDataID == 0 {
		return invalidImport("--basic-data-id is required")
	}
	if (o.basicDataID == 0) == (o.appID == 0) {
		return invalidImport("specify exactly one of --basic-data-id or --app-id")
	}
	if cmd.Flags().Changed("name") && (o.basicDataID != 0 || strings.TrimSpace(o.name) == "") {
		return invalidImport("--name requires --app-id and must be nonempty")
	}
	return nil
}

func openImportFile(path string) (*os.File, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".xlsx" && ext != ".xls" && ext != ".csv" {
		return nil, invalidImport("--file must have .xlsx, .xls or .csv extension")
	}
	// Stat before opening rejects directories/devices/FIFOs without blocking.
	info, err := os.Stat(path)
	if err != nil {
		return nil, invalidImport(fmt.Sprintf("cannot read --file: %v", err))
	}
	if !info.Mode().IsRegular() || info.Size() == 0 || info.Size() > 50*1024*1024 {
		return nil, invalidImport("--file must be a nonempty regular file of at most 50 MB")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, invalidImport(fmt.Sprintf("cannot open --file: %v", err))
	}
	return file, nil
}

func requireImportAuth(f *cmdutil.Factory) error {
	if err := cmdutil.RequireHostConfigured(f.Config.Host); err != nil {
		return err
	}
	if f.Config.Token == "" {
		return lterrors.NewAuthenticationError(lterrors.SubtypeTokenMissing, "not authenticated").WithHint("Run 'lingtong-cli auth login' first")
	}
	return nil
}

func runTableImport(ctx context.Context, c *client.Client, f *cmdutil.Factory, o importOptions, file *os.File, upload importUploader, result *importResult) error {
	fmt.Fprintln(f.IOStreams.ErrOut, "Uploading import file...")
	fileURL, err := upload(ctx, c, file)
	if err != nil {
		return err
	}
	if result.BasicDataID == 0 {
		var parsed struct {
			Columns json.RawMessage `json:"columnsSchema"`
		}
		if err := importAPI(ctx, c, http.MethodPost, "/basicdata/record/import/columnsSchema", map[string]interface{}{"fileUrl": fileURL}, nil, &parsed); err != nil {
			return err
		}
		var columns []json.RawMessage
		if err := json.Unmarshal(parsed.Columns, &columns); err != nil || len(columns) == 0 {
			return lterrors.NewAPIError(lterrors.SubtypeServerError, "file did not produce a nonempty columnsSchema")
		}
		var created struct {
			ID json.Number `json:"id"`
		}
		if err := importAPI(ctx, c, http.MethodPost, "/basicdata/save", nil, map[string]interface{}{
			"appId": o.appID, "name": o.name, "source": 1, "type": 1, "openHighMode": 0, "openConnector": 0,
		}, &created); err != nil {
			return err
		}
		result.BasicDataID, err = created.ID.Int64()
		if err != nil || result.BasicDataID <= 0 {
			return lterrors.NewAPIError(lterrors.SubtypeServerError, "table creation did not return a valid ID")
		}
		result.Created = true
		fmt.Fprintf(f.IOStreams.ErrOut, "Created table %d; saving inferred columns...\n", result.BasicDataID)
		// Preserve the backend-inferred schema exactly, including IDs and types.
		if err := importAPI(ctx, c, http.MethodPost, "/basicdata/schema/save", nil, map[string]interface{}{
			"basicDataId": result.BasicDataID, "columnsSchema": parsed.Columns,
		}, nil); err != nil {
			return err
		}
	}
	// These are Spring request parameters, not a JSON body. The proxy preserves
	// outer URL parameters even for POST; its POST wrapper ignores proxy.params.
	if err := importAPI(ctx, c, http.MethodPost, "/basicdata/record/import", map[string]interface{}{
		"basicDataId": result.BasicDataID, "fileUrl": fileURL,
	}, nil, &result.UUID); err != nil {
		return err
	}
	if strings.TrimSpace(result.UUID) == "" {
		return lterrors.NewAPIError(lterrors.SubtypeServerError, "import did not return a task UUID; check the table before resubmitting")
	}
	fmt.Fprintf(f.IOStreams.ErrOut, "Import submitted: table=%d uuid=%s\n", result.BasicDataID, result.UUID)
	if !o.wait {
		return nil
	}
	return waitForTableImport(ctx, c, f, o, result)
}

func importAPI(ctx context.Context, c *client.Client, method, path string, params map[string]interface{}, body, result interface{}) error {
	var raw []byte
	var err error
	if params != nil {
		raw, err = c.DoQueryContext(ctx, method, path, params)
	} else {
		raw, err = c.DoContext(ctx, method, path, nil, body)
	}
	if err != nil {
		return err
	}
	var response struct {
		Success *bool           `json:"success"`
		Result  json.RawMessage `json:"result"`
		Message string          `json:"message"`
		Msg     string          `json:"msg"`
	}
	if json.Unmarshal(raw, &response) != nil || response.Success == nil {
		return lterrors.NewAPIError(lterrors.SubtypeServerError, "%s: invalid API response", path)
	}
	if !*response.Success {
		message := response.Message
		if message == "" {
			message = response.Msg
		}
		return lterrors.NewAPIError(lterrors.SubtypeServerError, "%s: %s", path, message)
	}
	if result != nil {
		if len(response.Result) == 0 {
			return lterrors.NewAPIError(lterrors.SubtypeServerError, "%s: missing result", path)
		}
		if json.Unmarshal(response.Result, result) != nil {
			return lterrors.NewAPIError(lterrors.SubtypeServerError, "%s: invalid result", path)
		}
	}
	return nil
}

func waitForTableImport(ctx context.Context, c *client.Client, f *cmdutil.Factory, o importOptions, result *importResult) error {
	lastProgress := ""
	for {
		var progress *importProgress
		if err := importAPI(ctx, c, http.MethodGet, "/basicdata/record/getImportProgress", map[string]interface{}{"uuid": result.UUID}, nil, &progress); err != nil {
			return err
		}
		// The worker may not have published its first status yet (result:null).
		if progress != nil {
			if progress.Status == nil {
				return lterrors.NewAPIError(lterrors.SubtypeServerError, "import progress is missing status")
			}
			result.importProgress = *progress
			encoded, _ := json.Marshal(progress)
			if string(encoded) != lastProgress {
				fmt.Fprintf(f.IOStreams.ErrOut, "Import progress: %s\n", encoded)
				lastProgress = string(encoded)
			}
			switch *progress.Status {
			case 0, 1:
			case 2:
				result.Completed = true
				return nil
			case 3:
				return lterrors.NewAPIError(lterrors.SubtypeServerError, "import failed: %s", progress.Message)
			default:
				return lterrors.NewAPIError(lterrors.SubtypeServerError, "unknown import status: %d", *progress.Status)
			}
		}
		if !o.wait {
			return nil
		}
		timer := time.NewTimer(o.interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func importOperationError(err error, result *importResult) error {
	problem := lterrors.Problem{Category: lterrors.CategoryAPI, Subtype: lterrors.SubtypeServerError}
	if p, ok := lterrors.ProblemOf(err); ok {
		problem = *p
	}
	problem.Message = err.Error()
	if errors.Is(err, context.DeadlineExceeded) {
		problem.Category, problem.Subtype, problem.Message = lterrors.CategoryNetwork, lterrors.SubtypeNetworkTimeout, "import operation timed out"
	}
	if errors.Is(err, context.Canceled) {
		problem.Message = "import operation interrupted"
	}
	problem.Retryable = false // Retrying a submitted import may update records twice.
	if result.BasicDataID > 0 {
		problem.Message += fmt.Sprintf(" (basicDataId=%d)", result.BasicDataID)
	}
	if result.UUID != "" {
		problem.Message += " (uuid=" + result.UUID + ")"
		problem.Hint = "The server task is not cancelled. Inspect it without resubmitting: lingtong-cli table import-progress --uuid " + result.UUID
	} else if result.Created {
		problem.Hint = fmt.Sprintf("Table %d was created and has been retained. Inspect its schema/data before retrying; do not create another table blindly.", result.BasicDataID)
	} else if result.BasicDataID > 0 {
		problem.Hint = "Inspect the table before resubmitting if the submission response was lost."
	}
	return &problem
}

func writeImportResult(cmd *cobra.Command, f *cmdutil.Factory, result *importResult, identity string) error {
	return f.NewWriter(output.Format(cmd.Flag("format").Value.String()), identity).Write(map[string]interface{}{"success": true, "result": result})
}

func newCmdTableImportProgress(f *cmdutil.Factory) *cobra.Command {
	var uuid string
	var o importOptions
	cmd := &cobra.Command{
		Use: "import-progress", Short: "Resume waiting for an existing import task without resubmitting",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(uuid) == "" {
				return invalidImport("--uuid is required")
			}
			if err := validateImportWaitOptions(cmd, o); err != nil {
				return err
			}
			if err := requireImportAuth(f); err != nil {
				return err
			}
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			ctx, cancel := context.WithTimeout(ctx, o.timeout)
			defer cancel()
			result := &importResult{UUID: uuid}
			c := client.NewClientWithTimeout(f.Config.Host, f.Config.Token, o.timeout)
			if err := waitForTableImport(ctx, c, f, o, result); err != nil {
				return importOperationError(err, result)
			}
			return writeImportResult(cmd, f, result, "table.import-progress")
		},
	}
	cmd.Flags().StringVar(&uuid, "uuid", "", "Import task UUID (required)")
	addImportWaitFlags(cmd, &o)
	return cmd
}
