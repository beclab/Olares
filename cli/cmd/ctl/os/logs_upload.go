package os

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// logUploadOptions holds inputs for `olares-cli logs upload`. The pairing code
// (issued to a Space-token-verified user in the AssistHub web UI) is the only
// credential on this path: the ticket platform re-derives it from olares_id to
// trust the uploaded logs, so no Space token is needed here.
type logUploadOptions struct {
	OlaresID      string
	Code          string
	Endpoint      string
	File          string
	Description   string
	OlaresVersion string
	Timeout       time.Duration
}

const (
	ticketEndpointEnv = "OLARES_TICKET_API"
	gzipMimeType      = "application/gzip"
	presignPath       = "/v1/olares-cli/attachments/presigned-upload"
	ticketPath        = "/v1/olares-cli/tickets"
)

type presignRequest struct {
	OlaresID   string `json:"olares_id"`
	Code       string `json:"code"`
	Filename   string `json:"filename"`
	MimeType   string `json:"mime_type"`
	SizeBytes  int64  `json:"size_bytes"`
	IsLargeLog bool   `json:"is_large_log"`
}

type presignResponse struct {
	AttachmentID string            `json:"attachment_id"`
	UploadURL    string            `json:"upload_url"`
	Method       string            `json:"method"`
	Headers      map[string]string `json:"headers"`
	ExpiresAt    string            `json:"expires_at"`
}

type attachmentRef struct {
	AttachmentID string `json:"attachment_id"`
}

type ticketRequest struct {
	OlaresID      string          `json:"olares_id"`
	Code          string          `json:"code"`
	Description   string          `json:"description,omitempty"`
	OlaresVersion string          `json:"olares_version,omitempty"`
	Attachments   []attachmentRef `json:"attachments"`
}

type ticketResponse struct {
	TicketID     string `json:"ticket_id"`
	TicketNumber string `json:"ticket_number"`
}

func newCmdLogsUpload() *cobra.Command {
	options := &logUploadOptions{Timeout: 30 * time.Minute}

	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload collected logs to the Olares ticket platform",
		Long: `Upload an Olares log archive to the ticket platform and open a ticket.

The pairing code is obtained from the AssistHub web UI (Settings) and is the
only credential required: together with your Olares ID it authorizes this
upload, no login token is needed.

If --file is omitted, logs are collected first (requires root) and the
resulting archive is uploaded.`,
		Run: func(cmd *cobra.Command, args []string) {
			if options.Endpoint == "" {
				options.Endpoint = os.Getenv(ticketEndpointEnv)
			}
			if options.Endpoint == "" {
				log.Fatalf("error: ticket endpoint is required, set --ticket-endpoint or %s", ticketEndpointEnv)
			}
			if err := runLogsUpload(options); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}

	cmd.Flags().StringVar(&options.OlaresID, "olares-id", "", "Olares ID the logs belong to, e.g. alice@olares.com (required)")
	cmd.Flags().StringVar(&options.Code, "code", "", "Pairing code from the AssistHub web UI (required)")
	cmd.Flags().StringVar(&options.Endpoint, "ticket-endpoint", "", fmt.Sprintf("Ticket platform base URL, e.g. https://api.olares.com (or set %s)", ticketEndpointEnv))
	cmd.Flags().StringVar(&options.File, "file", "", "Path to an existing log archive to upload; if empty, logs are collected first")
	cmd.Flags().StringVar(&options.Description, "description", "", "Optional ticket description")
	cmd.Flags().StringVar(&options.OlaresVersion, "olares-version", "", "Optional Olares version recorded on the ticket")
	cmd.Flags().DurationVar(&options.Timeout, "timeout", options.Timeout, "HTTP timeout for each upload/API call, raise it for large archives on slow links")

	_ = cmd.MarkFlagRequired("olares-id")
	_ = cmd.MarkFlagRequired("code")

	return cmd
}

func runLogsUpload(options *logUploadOptions) error {
	archivePath := options.File
	if archivePath == "" {
		collected, cleanup, err := collectForUpload()
		if cleanup != nil {
			defer cleanup()
		}
		if err != nil {
			return err
		}
		archivePath = collected
	}

	info, err := os.Stat(archivePath)
	if err != nil {
		return fmt.Errorf("failed to stat archive %s: %v", archivePath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("archive %s is a directory, expected a file", archivePath)
	}

	endpoint := strings.TrimRight(options.Endpoint, "/")
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	client := &http.Client{Timeout: timeout}

	fmt.Println("requesting upload URL...")
	presign, err := requestPresign(client, endpoint, options, filepath.Base(archivePath), info.Size())
	if err != nil {
		return err
	}

	fmt.Println("uploading log archive...")
	if err := putArchive(client, presign, archivePath, info.Size()); err != nil {
		return err
	}

	fmt.Println("creating ticket...")
	ticket, err := createTicket(client, endpoint, options, presign.AttachmentID)
	if err != nil {
		return err
	}

	fmt.Printf("logs uploaded, ticket created: %s (%s)\n", ticket.TicketNumber, ticket.TicketID)
	return nil
}

// collectForUpload runs a full local collection into a temp directory and
// returns the produced archive path plus a cleanup func.
func collectForUpload() (string, func(), error) {
	tempDir, err := os.MkdirTemp("", "olares-logs-upload-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %v", err)
	}
	cleanup := func() { os.RemoveAll(tempDir) }

	options := &LogCollectOptions{
		Since:            "7d",
		MaxLines:         20000,
		OutputDir:        tempDir,
		IgnoreKubeErrors: true,
	}
	if err := collectLogs(options); err != nil {
		return "", cleanup, err
	}

	matches, err := filepath.Glob(filepath.Join(tempDir, "olares-logs-*.tar.gz"))
	if err != nil || len(matches) == 0 {
		return "", cleanup, fmt.Errorf("no log archive produced under %s", tempDir)
	}
	return matches[0], cleanup, nil
}

func requestPresign(client *http.Client, endpoint string, options *logUploadOptions, filename string, size int64) (*presignResponse, error) {
	reqBody := presignRequest{
		OlaresID:   options.OlaresID,
		Code:       options.Code,
		Filename:   filename,
		MimeType:   gzipMimeType,
		SizeBytes:  size,
		IsLargeLog: true,
	}
	var resp presignResponse
	if err := postJSON(client, endpoint+presignPath, reqBody, &resp); err != nil {
		return nil, fmt.Errorf("request presigned upload: %w", err)
	}
	if resp.UploadURL == "" || resp.AttachmentID == "" {
		return nil, fmt.Errorf("presign response missing upload_url or attachment_id")
	}
	return &resp, nil
}

func putArchive(client *http.Client, presign *presignResponse, path string, size int64) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open archive: %v", err)
	}
	defer f.Close()

	method := presign.Method
	if method == "" {
		method = http.MethodPut
	}
	req, err := http.NewRequest(method, presign.UploadURL, f)
	if err != nil {
		return fmt.Errorf("build upload request: %v", err)
	}
	req.ContentLength = size
	hasContentType := false
	for k, v := range presign.Headers {
		req.Header.Set(k, v)
		if strings.EqualFold(k, "Content-Type") {
			hasContentType = true
		}
	}
	if !hasContentType {
		req.Header.Set("Content-Type", gzipMimeType)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("upload archive: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload archive: storage returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func createTicket(client *http.Client, endpoint string, options *logUploadOptions, attachmentID string) (*ticketResponse, error) {
	reqBody := ticketRequest{
		OlaresID:      options.OlaresID,
		Code:          options.Code,
		Description:   options.Description,
		OlaresVersion: options.OlaresVersion,
		Attachments:   []attachmentRef{{AttachmentID: attachmentID}},
	}
	var resp ticketResponse
	if err := postJSON(client, endpoint+ticketPath, reqBody, &resp); err != nil {
		return nil, fmt.Errorf("create ticket: %w", err)
	}
	return &resp, nil
}

func postJSON(client *http.Client, url string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal request: %v", err)
	}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return apiError(resp.StatusCode, data)
	}
	if out != nil {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decode response: %v", err)
		}
	}
	return nil
}

// apiError turns a non-2xx ticket API response into a friendly message, adding
// a hint for the documented error statuses.
func apiError(status int, body []byte) error {
	detail := strings.TrimSpace(string(body))
	var parsed struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &parsed) == nil {
		if parsed.Message != "" {
			detail = parsed.Message
		}
		if parsed.Code != "" {
			detail = fmt.Sprintf("%s (%s)", detail, parsed.Code)
		}
	}

	var hint string
	switch status {
	case http.StatusUnauthorized:
		hint = "pairing code is invalid or expired; get a fresh one from the AssistHub web UI"
	case http.StatusForbidden:
		hint = "attachment does not belong to this Olares ID"
	case http.StatusRequestEntityTooLarge:
		hint = "log archive exceeds the server size limit"
	case http.StatusUnsupportedMediaType:
		hint = "log archive type is not allowed"
	case http.StatusTooManyRequests:
		hint = "rate limit exceeded, retry later"
	}
	if hint != "" {
		return fmt.Errorf("server returned %d: %s (%s)", status, detail, hint)
	}
	return fmt.Errorf("server returned %d: %s", status, detail)
}
