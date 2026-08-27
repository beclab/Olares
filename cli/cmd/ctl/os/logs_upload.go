package os

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"
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
	ticketEndpointEnv     = "OLARES_TICKET_API"
	defaultTicketEndpoint = "https://ticket.olares.com"
	gzipMimeType          = "application/gzip"
	presignPath           = "/v1/olares-cli/attachments/presigned-upload"
	ticketPath            = "/v1/olares-cli/tickets"
	headerIdempotencyKey  = "Idempotency-Key"

	// JSON calls (presign / create ticket) are tiny; a long shared timeout
	// is only for the S3 PUT. Mixing them on one Client also reuses the
	// Cloudflare connection after a minutes-long upload, which surfaces as
	// "unexpected EOF" on the follow-up POST.
	jsonRequestTimeout   = 30 * time.Second
	tlsHandshakeTimeout  = 45 * time.Second
	maxTransientAttempts = 3
	defaultUploadTimeout = 30 * time.Minute
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
	options := &logUploadOptions{Timeout: defaultUploadTimeout}

	cmd := &cobra.Command{
		Use:   "upload",
		Short: "Upload collected logs to the Olares ticket platform",
		Long: `Upload an Olares log archive to the ticket platform and open a ticket.

The pairing code is obtained from the AssistHub web UI (Settings) and is the
only credential required: together with your Olares ID it authorizes this
upload, no login token is needed.

If --file is omitted, logs are collected first (requires root) and the
resulting archive is uploaded. If upload or ticket creation fails after
collection, the archive is kept; retry with --file to skip collecting
again.`,
		Run: func(cmd *cobra.Command, args []string) {
			options.Endpoint = resolveTicketEndpoint(options.Endpoint)
			if err := runLogsUpload(options); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}

	cmd.Flags().StringVar(&options.OlaresID, "olares-id", "", "Olares ID the logs belong to, e.g. alice@olares.com (required)")
	cmd.Flags().StringVar(&options.Code, "code", "", "Pairing code from the AssistHub web UI (required)")
	cmd.Flags().StringVar(&options.Endpoint, "ticket-endpoint", "", fmt.Sprintf("Ticket platform base URL (default %s, or set %s)", defaultTicketEndpoint, ticketEndpointEnv))
	cmd.Flags().StringVar(&options.File, "file", "", "Path to an existing log archive to upload; if empty, logs are collected first")
	cmd.Flags().StringVar(&options.Description, "description", "", "Optional ticket description")
	cmd.Flags().StringVar(&options.OlaresVersion, "olares-version", "", "Optional Olares version recorded on the ticket")
	cmd.Flags().DurationVar(&options.Timeout, "timeout", options.Timeout, "HTTP timeout for the archive upload to storage, raise it for large archives on slow links")

	_ = cmd.MarkFlagRequired("olares-id")
	_ = cmd.MarkFlagRequired("code")

	return cmd
}

// resolveTicketEndpoint picks the ticket API base URL: explicit flag, then
// OLARES_TICKET_API, then the production default.
func resolveTicketEndpoint(flagValue string) string {
	if endpoint := strings.TrimSpace(flagValue); endpoint != "" {
		return endpoint
	}
	if endpoint := strings.TrimSpace(os.Getenv(ticketEndpointEnv)); endpoint != "" {
		return endpoint
	}
	return defaultTicketEndpoint
}

func runLogsUpload(options *logUploadOptions) error {
	collected := false
	archivePath := options.File
	var cleanup func()
	if archivePath == "" {
		collectedPath, collectedCleanup, err := collectForUpload()
		cleanup = collectedCleanup
		if err != nil {
			if cleanup != nil {
				cleanup()
			}
			return err
		}
		archivePath = collectedPath
		collected = true
	}

	info, err := os.Stat(archivePath)
	if err != nil {
		return keepArchiveHint(archivePath, collected, fmt.Errorf("failed to stat archive %s: %v", archivePath, err))
	}
	if info.IsDir() {
		return keepArchiveHint(archivePath, collected, fmt.Errorf("archive %s is a directory, expected a file", archivePath))
	}

	endpoint := strings.TrimRight(options.Endpoint, "/")
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = defaultUploadTimeout
	}
	apiClient := newAPIClient()
	storageClient := newStorageClient(timeout)

	fmt.Fprintln(os.Stderr, "requesting upload URL...")
	presign, err := requestPresign(apiClient, endpoint, options, filepath.Base(archivePath), info.Size())
	if err != nil {
		return keepArchiveHint(archivePath, collected, err)
	}

	if err := putArchive(storageClient, presign, archivePath, info.Size()); err != nil {
		return keepArchiveHint(archivePath, collected, err)
	}

	fmt.Fprintln(os.Stderr, "creating ticket...")
	ticket, err := createTicket(apiClient, endpoint, options, presign.AttachmentID, newIdempotencyKey())
	if err != nil {
		return keepArchiveHint(archivePath, collected, err)
	}

	if cleanup != nil {
		cleanup()
	}
	fmt.Fprintf(os.Stderr, "logs uploaded, ticket created: %s (%s)\n", ticket.TicketNumber, ticket.TicketID)
	return nil
}

// keepArchiveHint leaves a collected temp archive on disk so the user can retry
// with --file instead of gathering logs again, and points them at that path.
func keepArchiveHint(archivePath string, collected bool, err error) error {
	if !collected {
		return err
	}
	fmt.Fprintf(os.Stderr, "archive kept at %s\nretry with --file %s to skip collecting again\n", archivePath, archivePath)
	return err
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
	fmt.Fprintln(os.Stderr, "collecting logs...")
	if err := collectLogs(options); err != nil {
		return "", cleanup, err
	}
	fmt.Fprintln(os.Stderr, "collection complete")

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
	if err := postJSON(client, endpoint+presignPath, reqBody, nil, &resp); err != nil {
		return nil, fmt.Errorf("request presigned upload: %w", err)
	}
	if resp.UploadURL == "" || resp.AttachmentID == "" {
		return nil, fmt.Errorf("presign response missing upload_url or attachment_id")
	}
	return &resp, nil
}

func putArchive(client *http.Client, presign *presignResponse, path string, size int64) error {
	var last error
	for attempt := 1; attempt <= maxTransientAttempts; attempt++ {
		if attempt > 1 {
			fmt.Fprintf(os.Stderr, "retrying upload (attempt %d/%d)...\n", attempt, maxTransientAttempts)
			sleep(time.Duration(attempt-1) * time.Second)
		}
		err := putArchiveOnce(client, presign, path, size)
		if err == nil {
			return nil
		}
		last = err
		if !isTransientNetErr(err) {
			return err
		}
	}
	return last
}

func putArchiveOnce(client *http.Client, presign *presignResponse, path string, size int64) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open archive: %v", err)
	}
	defer f.Close()

	method := presign.Method
	if method == "" {
		method = http.MethodPut
	}
	body, finish := uploadBody(f, size)
	req, err := http.NewRequest(method, presign.UploadURL, body)
	if err != nil {
		finish()
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
	finish()
	if err != nil {
		return fmt.Errorf("upload archive: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload archive: storage returned %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return nil
}

// uploadBody wraps the archive with a live stderr progress line when stderr is
// a terminal. Redirected stderr (scripts, CI, log capture) gets a single plain
// status line instead, so no carriage returns end up in captured output. The
// returned func must be called once the request finishes, to settle the line
// before anything else is written.
func uploadBody(f io.Reader, size int64) (io.Reader, func()) {
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		fmt.Fprintln(os.Stderr, "uploading log archive...")
		return f, func() {}
	}
	p := &progressReader{r: f, total: size, label: "Uploading", last: -1}
	return p, p.finish
}

// progressReader wraps an io.Reader and redraws a single-line stderr progress
// update (percent + bytes) without pulling in a TUI dependency. Only used when
// stderr is a terminal; see uploadBody.
type progressReader struct {
	r     io.Reader
	total int64
	read  int64
	label string
	last  int
	drawn bool
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 {
		p.read += int64(n)
		percent := 100
		if p.total > 0 {
			percent = int((p.read * 100) / p.total)
			if percent > 100 {
				percent = 100
			}
		}
		// Throttle redraws to whole-percent steps (plus the final byte).
		if percent != p.last || p.read == p.total || err == io.EOF {
			p.last = percent
			p.draw(percent)
		}
	}
	return n, err
}

func (p *progressReader) draw(percent int) {
	p.drawn = true
	// \033[K erases to end of line so a shorter update can't leave residue
	// from the previous, longer one.
	fmt.Fprintf(
		os.Stderr,
		"\r\033[K%s… %d%% (%s/%s)",
		p.label,
		percent,
		humanBytes(p.read),
		humanBytes(p.total),
	)
}

// finish closes out the progress line so later output starts clean. It is a
// no-op when nothing was drawn, e.g. an empty archive or a request that failed
// before the body was read.
func (p *progressReader) finish() {
	if p.drawn {
		fmt.Fprintln(os.Stderr)
	}
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

func createTicket(client *http.Client, endpoint string, options *logUploadOptions, attachmentID, idempotencyKey string) (*ticketResponse, error) {
	reqBody := ticketRequest{
		OlaresID:      options.OlaresID,
		Code:          options.Code,
		Description:   options.Description,
		OlaresVersion: options.OlaresVersion,
		Attachments:   []attachmentRef{{AttachmentID: attachmentID}},
	}
	headers := map[string]string{}
	if idempotencyKey != "" {
		headers[headerIdempotencyKey] = idempotencyKey
	}
	var resp ticketResponse
	if err := postJSON(client, endpoint+ticketPath, reqBody, headers, &resp); err != nil {
		return nil, fmt.Errorf("create ticket: %w", err)
	}
	return &resp, nil
}

func postJSON(client *http.Client, url string, payload any, headers map[string]string, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal request: %v", err)
	}
	var last error
	for attempt := 1; attempt <= maxTransientAttempts; attempt++ {
		if attempt > 1 {
			sleep(time.Duration(attempt-1) * time.Second)
		}
		err := postJSONOnce(client, url, body, headers, out)
		if err == nil {
			return nil
		}
		last = err
		if !isTransientNetErr(err) {
			return err
		}
	}
	return last
}

func postJSONOnce(client *http.Client, url string, body []byte, headers map[string]string, out any) error {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
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

func newAPIClient() *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.DisableKeepAlives = true
	tr.TLSHandshakeTimeout = tlsHandshakeTimeout
	return &http.Client{Timeout: jsonRequestTimeout, Transport: tr}
}

func newStorageClient(timeout time.Duration) *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSHandshakeTimeout = tlsHandshakeTimeout
	return &http.Client{Timeout: timeout, Transport: tr}
}

func newIdempotencyKey() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("cli-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

// sleep is time.Sleep so tests can skip the retry backoff.
var sleep = time.Sleep

func isTransientNetErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "tls handshake timeout"):
		return true
	case strings.Contains(msg, "unexpected eof"):
		return true
	case strings.Contains(msg, "connection reset"):
		return true
	case strings.Contains(msg, "broken pipe"):
		return true
	case strings.Contains(msg, "http2: server sent goaway"):
		return true
	case strings.Contains(msg, "connection refused"):
		return true
	default:
		return false
	}
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
