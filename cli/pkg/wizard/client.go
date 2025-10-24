package wizard

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Client implementation - based on original TypeScript Client class
type Client struct {
	State  ClientState
	Sender Sender
}

func NewClient(state ClientState, sender Sender) *Client {
	return &Client{
		State:  state,
		Sender: sender,
	}
}

// Implement AppAPI interface
func (c *Client) StartAuthRequest(params StartAuthRequestParams) (*StartAuthRequestResponse, error) {
	// Build request parameters
	requestParams := []interface{}{params}
	
	// Send request
	response, err := c.call("startAuthRequest", requestParams)
	if err != nil {
		return nil, err
	}
	
	// Parse response
	var result StartAuthRequestResponse
	if err := c.parseResponse(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse StartAuthRequest response: %v", err)
	}
	
	return &result, nil
}

func (c *Client) CompleteAuthRequest(params CompleteAuthRequestParams) (*CompleteAuthRequestResponse, error) {
	// Build request parameters
	requestParams := []interface{}{params}
	
	// Send request
	response, err := c.call("completeAuthRequest", requestParams)
	if err != nil {
		return nil, err
	}
	
	// Parse response
	var result CompleteAuthRequestResponse
	if err := c.parseResponse(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse CompleteAuthRequest response: %v", err)
	}
	
	return &result, nil
}

// Generic RPC call method
func (c *Client) call(method string, params []interface{}) (*Response, error) {
	session := c.State.GetSession()
	
	// Build request
	req := &Request{
		Method: method,
		Params: params,
		Device: c.State.GetDevice(),
	}
	
	// If session exists, add authentication info
	if session != nil {
		auth, err := c.authenticateRequest(req, session)
		if err != nil {
			return nil, fmt.Errorf("failed to authenticate request: %v", err)
		}
		req.Auth = auth
		
		// Temporary debug: print full request JSON
		if reqJSON, err := json.Marshal(req); err == nil {
			log.Printf("Full request JSON: %s", string(reqJSON))
		}
	}
	
	// Send request
	response, err := c.Sender.Send(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	
	// Check response error
	if response.Error != nil {
		return nil, NewAuthError(
			ErrorCode(response.Error.Code),
			response.Error.Message,
			nil,
		)
	}
	
	// If session exists, verify response
	if session != nil {
		if err := c.verifyResponse(response, session); err != nil {
			return nil, fmt.Errorf("failed to verify response: %v", err)
		}
	}
	
	return response, nil
}

// authenticateRequest implements real HMAC session signing (ref: session.ts line 176-189)
func (c *Client) authenticateRequest(req *Request, session *Session) (*RequestAuth, error) {
	if session.Key == nil {
		return nil, fmt.Errorf("session key is nil")
	}
	
	// 1. Build timestamp
	now := time.Now()
	
	// 2. Serialize request data (ref: session.ts line 158: data = req.params)
	data := req.Params // Use entire params array directly, not just first element
	
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %v", err)
	}
	
	// 3. Build signature message (ref: session.ts line 179)
	// Format: ${session}_${time.toISOString()}_${marshal(data)}
	// Use same format as ISOTime.MarshalJSON to ensure consistency
	utcTime := now.UTC()
	timeStr := fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d.%03dZ",
		utcTime.Year(), utcTime.Month(), utcTime.Day(),
		utcTime.Hour(), utcTime.Minute(), utcTime.Second(),
		utcTime.Nanosecond()/1000000)
	message := fmt.Sprintf("%s_%s_%s", session.ID, timeStr, string(dataJSON))
	
	// 4. Use HMAC-SHA256 signing (ref: HMACParams in crypto.ts)
	mac := hmac.New(sha256.New, session.Key)
	mac.Write([]byte(message))
	signature := mac.Sum(nil)
	
	log.Printf("Session signing: sessionId=%s, message_len=%d", session.ID, len(message))
	log.Printf("Signing message: %s", message)
	log.Printf("Data JSON: %s", string(dataJSON))
	log.Printf("Session key for signing: %x", session.Key)
	log.Printf("Signature (hex): %x", signature)
	
	return &RequestAuth{
		Session:   session.ID,
		Time:      ISOTime(now),            // Convert to ISOTime type
		Signature: Base64Bytes(signature),  // Convert to Base64Bytes type
	}, nil
}

// verifyResponse (simplified version)
func (c *Client) verifyResponse(response *Response, session *Session) error {
	// In actual implementation, response signature verification is needed here
	// Real implementation for response parsing
	log.Printf("Verifying response with session: %s", session.ID)
	return nil
}

// parseResponse helper method
func (c *Client) parseResponse(result interface{}, target interface{}) error {
	// Convert result to JSON, then parse to target structure
	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %v", err)
	}
	
	if err := json.Unmarshal(jsonData, target); err != nil {
		return fmt.Errorf("failed to unmarshal to target: %v", err)
	}
	
	return nil
}
