package wizard

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// WizardConfig contains activation wizard configuration
type WizardConfig struct {
	System   SystemConfig   `json:"system"`
	Password PasswordConfig `json:"password"`
}

// SystemConfig system configuration
type SystemConfig struct {
	Location string     `json:"location"`      // Timezone location, e.g. "Asia/Shanghai"
	Language string     `json:"language"`      // Language, e.g. "zh-CN" or "en-US"
	Theme    string     `json:"theme"`         // Theme, e.g. "dark" or "light"
	FRP      *FRPConfig `json:"frp,omitempty"` // Optional FRP configuration
}

type FRPConfig struct {
	Host string `json:"host"`
	Jws  string `json:"jws"`
}

// PasswordConfig password configuration
type PasswordConfig struct {
	CurrentPassword string `json:"current_password"` // Current password (from wizard settings)
	NewPassword     string `json:"new_password"`     // New password (for reset)
}

// TerminusInfo Terminus information response
type TerminusInfo struct {
	WizardStatus string `json:"wizardStatus"`
	OlaresId     string `json:"olaresId"`
	// Other fields...
}

// ActivationWizard activation wizard
type ActivationWizard struct {
	BaseURL      string
	Config       WizardConfig
	AccessToken  string
	MaxRetries   int
	PollInterval time.Duration
}

// NewActivationWizard creates a new activation wizard
func NewActivationWizard(baseURL, accessToken string, config WizardConfig) *ActivationWizard {
	return &ActivationWizard{
		BaseURL:      baseURL,
		Config:       config,
		AccessToken:  accessToken,
		MaxRetries:   10,
		PollInterval: 2 * time.Second,
	}
}

// RunWizard runs the complete activation wizard process (ref: ActivateWizard.vue updateInfo)
func (w *ActivationWizard) RunWizard() error {
	log.Println("=== Starting Terminus Activation Wizard ===")

	// Initialize state tracking variables (ref: ActivateWizard.vue)
	var getHostTerminusCount int = 0
	var updateTerminusInfoInProgress bool = false

	// 1. Get initial status
	status, err := w.updateTerminusInfo()
	if err != nil {
		return fmt.Errorf("failed to get initial status: %v", err)
	}

	log.Printf("Initial wizard status: %s", status)

	// 2. State machine loop processing (ref: ActivateWizard.vue updateInfo function)
	for {
		// Check failure status (ref: updateInfo line 230-236)
		if status == "vault_activate_failed" || status == "system_activate_failed" || status == "network_activate_failed" {
			return fmt.Errorf("activation failed with status: %s", status)
		}

		// Check in-progress status (ref: updateInfo line 238-244)
		if status == "vault_activating" || status == "system_activating" || status == "network_activating" || status == "wait_activate_network" {
			log.Printf("⏳ System is %s, waiting...", status)
		} else {
			// Handle specific status (ref: updateInfo line 246-284)
			switch status {
			case "completed":
				log.Println("✅ Activation completed successfully!")
				return nil

			case "wait_activate_system":
				log.Println("📋 Configuring system...")
				if err := w.configSystem(); err != nil {
					return fmt.Errorf("system configuration failed: %v", err)
				}

			// case "wait_activate_network":
			// 	log.Println("🌐 Configuring network...")
			// 	return nil

			case "wait_reset_password":
				log.Println("🔐 Resetting password...")
				status, err := w.authRequestTerminusInfo()
				if err != nil {
					return fmt.Errorf("failed to get terminus info by authurl: %v", err)
				}

				if status == "wait_reset_password" {
					// Directly perform password reset, no need for complex DNS waiting logic
					if err := w.performPasswordReset(); err != nil {
						return fmt.Errorf("password reset failed: %v", err)
					}
					log.Println("✅ Password reset completed")
				}

			default:
				log.Printf("⏳ Unknown status: %s, waiting...", status)
			}
		}

		// Wait and update status (ref: ActivateWizard.vue setInterval 2 seconds)
		time.Sleep(w.PollInterval)

		// Update status, prevent concurrency (ref: updateInfo line 225-228)
		if !updateTerminusInfoInProgress {
			updateTerminusInfoInProgress = true
			newStatus, err := w.updateTerminusInfo()
			updateTerminusInfoInProgress = false

			if err != nil {
				log.Printf("Warning: Failed to update status: %v", err)
				getHostTerminusCount++
				if getHostTerminusCount >= 10 {
					return fmt.Errorf("too many failed attempts to get terminus info")
				}
				continue
			}

			if newStatus != status {
				log.Printf("Status changed: %s → %s", status, newStatus)
				status = newStatus

				// Reset error count
				getHostTerminusCount = 0
			}
		}
	}
}

// updateTerminusInfo updates Terminus information
func (w *ActivationWizard) updateTerminusInfo() (string, error) {
	url := fmt.Sprintf("%s/bfl/info/v1/olares-info?t=%d", w.BaseURL, time.Now().UnixMilli())

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if w.AccessToken != "" {
		req.Header.Set("X-Authorization", w.AccessToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		// If main URL fails, try backup URL
		return w.authRequestTerminusInfo()
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data TerminusInfo `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	return response.Data.WizardStatus, nil
}

// authRequestTerminusInfo backup Terminus information request
func (w *ActivationWizard) authRequestTerminusInfo() (string, error) {
	// Use globalUserStore to generate correct terminus_url

	var terminusURL = globalUserStore.GetAuthURL()

	// Build backup URL (usually terminus_url + '/api/olares-info')
	url := fmt.Sprintf("%s/bfl/info/v1/olares-info?t=%d", terminusURL, time.Now().UnixMilli())

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Data TerminusInfo `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	return response.Data.WizardStatus, nil
}

// performPasswordReset performs password reset - simplified version
func (w *ActivationWizard) performPasswordReset() error {
	log.Printf("🔐 Performing password reset...")

	// In CLI environment, we need to get necessary information from global storage
	if globalUserStore == nil {
		return fmt.Errorf("global user store not initialized")
	}

	terminusName := globalUserStore.GetTerminusName()
	localName := globalUserStore.GetLocalName()
	authURL := globalUserStore.GetAuthURL()

	// If local environment (127.0.0.1), use bflURL instead of stored authURL
	if strings.Contains(w.BaseURL, "127.0.0.1") {
		authURL = w.BaseURL
		log.Printf("Detected local environment, using bflURL: %s", authURL)
	}

	// Get passwords from wizard configuration
	currentPassword := w.getCurrentPassword()
	newPassword := w.generateNewPassword()

	log.Printf("Resetting password for user: %s", localName)

	// 1. First login to get access token
	token, err := LoginTerminus(authURL, terminusName, localName, currentPassword, false)
	if err != nil {
		return fmt.Errorf("failed to login before password reset: %v", err)
	}

	log.Printf("Login successful, proceeding with password reset...")

	// 2. Perform password reset
	err = ResetPassword(authURL, localName, currentPassword, newPassword, token.AccessToken)
	if err != nil {
		return fmt.Errorf("password reset failed: %v", err)
	}

	log.Printf("🎉 Password reset completed successfully!")

	return nil
}

// getCurrentPassword gets current password (from configuration)
func (w *ActivationWizard) getCurrentPassword() string {
	if w.Config.Password.CurrentPassword != "" {
		return w.Config.Password.CurrentPassword
	} else {
		panic("Current password not set in wizard config")
	}
}

// generateNewPassword generates new password (from configuration or generate)
func (w *ActivationWizard) generateNewPassword() string {
	if w.Config.Password.NewPassword != "" {
		return w.Config.Password.NewPassword
	} else {
		panic("New password not set in wizard config")
	}
}

// configSystem configures system
func (w *ActivationWizard) configSystem() error {
	log.Printf("Configuring system with location: %s, language: %s",
		w.Config.System.Location, w.Config.System.Language)

	url := fmt.Sprintf("%s/bfl/settings/v1alpha1/activate", w.BaseURL)

	jsonData, err := json.Marshal(w.Config.System)
	if err != nil {
		return fmt.Errorf("failed to marshal system config: %v", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second, // System configuration may take longer
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if w.AccessToken != "" {
		req.Header.Set("X-Authorization", w.AccessToken)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	log.Printf("✅ System configuration completed")
	return nil
}

// CustomWizardConfig creates custom wizard configuration
func CustomWizardConfig(location, language string, enableTunnel bool, host, jws, currentPassword, newPassword string) WizardConfig {
	config := WizardConfig{
		System: SystemConfig{
			Location: location,
			Language: language,
		},
		Password: PasswordConfig{
			CurrentPassword: currentPassword, // Need to set at runtime
			NewPassword:     newPassword,     // Need to set at runtime
		},
	}

	// If tunnel is enabled, initialize FRP configuration
	if enableTunnel {
		config.System.FRP = &FRPConfig{
			Host: host,
			Jws:  jws,
		}
	}

	return config
}

// RunActivationWizard convenient function to run activation wizard
func RunActivationWizard(baseURL, accessToken string, config WizardConfig) error {
	wizard := NewActivationWizard(baseURL, accessToken, config)
	return wizard.RunWizard()
}
