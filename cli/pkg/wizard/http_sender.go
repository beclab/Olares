package wizard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPSender implements HTTP-based Sender interface
type HTTPSender struct {
	BaseURL string
	Client  *http.Client
}

// NewHTTPSender creates new HTTP Sender
func NewHTTPSender(baseURL string) *HTTPSender {
	return &HTTPSender{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Send implements Sender interface, sends HTTP request
func (h *HTTPSender) Send(req *Request) (*Response, error) {
	// Build request URL
	url := fmt.Sprintf("%s/api/rpc", h.BaseURL)
	
	// Serialize request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create HTTP request
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	// Set request headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	
	// Authentication info already included in JSON request body's auth field, no extra HTTP headers needed
	
	// Send request
	httpResp, err := h.Client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()
	
	// Read response
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	
	// Check HTTP status code
	if httpResp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error %d: %s", httpResp.StatusCode, string(respBody))
	}
	
	// Parse response
	var response Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	
	return &response, nil
}
