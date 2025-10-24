package wizard

import (
	"fmt"
)

// WebPlatform implementation - based on original TypeScript WebPlatform
type WebPlatform struct {
	SupportedAuthTypes []AuthType
	App                AppAPI // App interface, currently only interface definition
	Mnemonic           string // Mnemonic for real JWS signing
	DID                string // DID for user identification
}

func NewWebPlatform(app AppAPI) *WebPlatform {
	return &WebPlatform{
		SupportedAuthTypes: []AuthType{AuthTypeSSI},
		App:                app,
	}
}

// NewWebPlatformWithMnemonic creates WebPlatform with mnemonic
func NewWebPlatformWithMnemonic(app AppAPI, mnemonic, did string) *WebPlatform {
	return &WebPlatform{
		SupportedAuthTypes: []AuthType{AuthTypeSSI},
		App:                app,
		Mnemonic:           mnemonic,
		DID:                did,
	}
}

func (p *WebPlatform) getAuthClient(authType AuthType) (AuthClient, error) {
	// Only support SSI authentication type
	if authType != AuthTypeSSI {
		return nil, fmt.Errorf("authentication type not supported: %s", authType)
	}
	
	// Check if global storage is initialized
	if globalUserStore == nil {
		return nil, fmt.Errorf("global stores not initialized, call InitializeGlobalStores first")
	}
	
	// Use global storage
	return &SSIAuthClient{
		UserStore: globalUserStore,
		// JWSSigner removed as UserStore.SignJWS() is actually used
	}, nil
}

func (p *WebPlatform) prepareCompleteAuthRequest(req *StartAuthRequestResponse) (map[string]any, error) {
	client, err := p.getAuthClient(req.Type)
	if err != nil {
		return nil, NewAuthError(ErrorCodeAuthenticationFailed, "Authentication type not supported!", err)
	}

	return client.PrepareAuthentication(req.Data)
}

func (p *WebPlatform) StartAuthRequest(opts StartAuthRequestOptions) (*StartAuthRequestResponse, error) {
	params := StartAuthRequestParams{
		DID:                *opts.DID,
		Type:               opts.Type,
		SupportedTypes:     p.SupportedAuthTypes,
		Purpose:            opts.Purpose,
		AuthenticatorID:    opts.AuthenticatorID,
		AuthenticatorIndex: opts.AuthenticatorIndex,
	}

	return p.App.StartAuthRequest(params)
}

func (p *WebPlatform) CompleteAuthRequest(req *StartAuthRequestResponse) (*AuthenticateResponse, error) {
	// If request already verified, return directly
	if req.RequestStatus == AuthRequestStatusVerified {
		return &AuthenticateResponse{
			DID:           req.DID,
			Token:         req.Token,
			DeviceTrusted: req.DeviceTrusted,
			AccountStatus: *req.AccountStatus,
			Provisioning:  *req.Provisioning,
		}, nil
	}

	// Prepare authentication data
	data, err := p.prepareCompleteAuthRequest(req)
	if err != nil {
		return nil, NewAuthError(ErrorCodeAuthenticationFailed, "The request was canceled.", err)
	}

	if data == nil {
		return nil, NewAuthError(ErrorCodeAuthenticationFailed, "The request was canceled.", nil)
	}

	// Only support SSI authentication type, no need to handle other types

	// Call App API to complete authentication
	params := CompleteAuthRequestParams{
		ID:   req.ID,
		Data: data,
		DID:  req.DID,
	}

	response, err := p.App.CompleteAuthRequest(params)
	if err != nil {
		return nil, err
	}

	return &AuthenticateResponse{
		DID:           req.DID,
		Token:         req.Token,
		DeviceTrusted: response.DeviceTrusted,
		AccountStatus: response.AccountStatus,
		Provisioning:  response.Provisioning,
	}, nil
}

// Global variables
var platform Platform
var globalUserStore *UserStore
// globalJWSSigner removed as UserStore.SignJWS() is actually used

func SetPlatform(p Platform) {
	platform = p
}

// InitializeGlobalStores initializes global storage
func InitializeGlobalStores(mnemonic, terminusName string) error {
	// Create UserStore (contains all necessary JWS signing functionality)
	userStore, err := NewUserStore(mnemonic, terminusName)
	if err != nil {
		return fmt.Errorf("failed to create UserStore: %w", err)
	}
	
	// Set global variables
	globalUserStore = userStore
	
	return nil
}
