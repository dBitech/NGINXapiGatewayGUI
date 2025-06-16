package acme

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"go-apigateway-gui/internal/models"
)

// CertificateManager handles automatic certificate management
type CertificateManager struct {
	acmeManagers map[string]*Manager
	configs      map[string]*models.ACMEConfig
	certDir      string
	mu           sync.RWMutex
	stopChan     chan struct{}
	started      bool
}

// NewCertificateManager creates a new certificate manager
func NewCertificateManager(certDir string) *CertificateManager {
	return &CertificateManager{
		acmeManagers: make(map[string]*Manager),
		configs:      make(map[string]*models.ACMEConfig),
		certDir:      certDir,
		stopChan:     make(chan struct{}),
	}
}

// AddACMEConfig adds or updates an ACME configuration
func (cm *CertificateManager) AddACMEConfig(serverID string, config *models.ACMEConfig) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if config == nil || !config.Enabled {
		// Remove existing configuration if disabled
		delete(cm.acmeManagers, serverID)
		delete(cm.configs, serverID)
		return nil
	}

	// Create ACME manager for this configuration
	manager, err := NewManager(config, cm.certDir)
	if err != nil {
		return fmt.Errorf("failed to create ACME manager for server %s: %v", serverID, err)
	}

	cm.acmeManagers[serverID] = manager
	cm.configs[serverID] = config

	log.Printf("Added ACME configuration for server %s with domains: %v", serverID, config.Domains)
	return nil
}

// RemoveACMEConfig removes an ACME configuration
func (cm *CertificateManager) RemoveACMEConfig(serverID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if manager, exists := cm.acmeManagers[serverID]; exists {
		manager.StopHTTPChallenge()
		delete(cm.acmeManagers, serverID)
		delete(cm.configs, serverID)
		log.Printf("Removed ACME configuration for server %s", serverID)
	}
}

// ObtainCertificate obtains a certificate for the specified server
func (cm *CertificateManager) ObtainCertificate(serverID string) error {
	cm.mu.RLock()
	manager, exists := cm.acmeManagers[serverID]
	cm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no ACME configuration found for server %s", serverID)
	}

	log.Printf("Obtaining certificate for server %s", serverID)

	// Start HTTP challenge server if needed
	if err := manager.StartHTTPChallenge(); err != nil {
		return fmt.Errorf("failed to start HTTP challenge: %v", err)
	}
	defer manager.StopHTTPChallenge()

	cert, err := manager.ObtainCertificate()
	if err != nil {
		return fmt.Errorf("failed to obtain certificate: %v", err)
	}

	log.Printf("Successfully obtained certificate for server %s, domains: %v", serverID, cert.Domain)
	return nil
}

// RenewCertificate renews a certificate for the specified server
func (cm *CertificateManager) RenewCertificate(serverID string) error {
	cm.mu.RLock()
	manager, exists := cm.acmeManagers[serverID]
	cm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no ACME configuration found for server %s", serverID)
	}

	log.Printf("Renewing certificate for server %s", serverID)

	// Start HTTP challenge server if needed
	if err := manager.StartHTTPChallenge(); err != nil {
		return fmt.Errorf("failed to start HTTP challenge: %v", err)
	}
	defer manager.StopHTTPChallenge()

	cert, err := manager.RenewCertificate()
	if err != nil {
		return fmt.Errorf("failed to renew certificate: %v", err)
	}

	log.Printf("Successfully renewed certificate for server %s, domains: %v", serverID, cert.Domain)
	return nil
}

// CheckRenewal checks if any certificates need renewal
func (cm *CertificateManager) CheckRenewal() map[string]bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	renewalNeeded := make(map[string]bool)

	for serverID, manager := range cm.acmeManagers {
		needs, err := manager.NeedsRenewal()
		if err != nil {
			log.Printf("Error checking renewal for server %s: %v", serverID, err)
			renewalNeeded[serverID] = true // Assume renewal needed on error
		} else {
			renewalNeeded[serverID] = needs
		}
	}

	return renewalNeeded
}

// GetCertificatePaths returns certificate and key paths for a server
func (cm *CertificateManager) GetCertificatePaths(serverID string) (certPath, keyPath string, err error) {
	cm.mu.RLock()
	manager, exists := cm.acmeManagers[serverID]
	cm.mu.RUnlock()

	if !exists {
		return "", "", fmt.Errorf("no ACME configuration found for server %s", serverID)
	}

	certPath, keyPath = manager.GetCertificatePaths()
	return certPath, keyPath, nil
}

// Start begins the automatic renewal process
func (cm *CertificateManager) Start(ctx context.Context) error {
	cm.mu.Lock()
	if cm.started {
		cm.mu.Unlock()
		return fmt.Errorf("certificate manager already started")
	}
	cm.started = true
	cm.mu.Unlock()

	log.Println("Starting certificate manager with automatic renewal")

	// Start renewal checker goroutine
	go cm.renewalChecker(ctx)

	return nil
}

// Stop stops the certificate manager
func (cm *CertificateManager) Stop() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.started {
		return
	}

	log.Println("Stopping certificate manager")
	close(cm.stopChan)
	cm.started = false

	// Stop all HTTP challenge servers
	for _, manager := range cm.acmeManagers {
		manager.StopHTTPChallenge()
	}
}

// renewalChecker periodically checks for certificates that need renewal
func (cm *CertificateManager) renewalChecker(ctx context.Context) {
	// Default check interval
	defaultInterval := 24 * time.Hour

	ticker := time.NewTicker(defaultInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Certificate renewal checker stopping due to context cancellation")
			return
		case <-cm.stopChan:
			log.Println("Certificate renewal checker stopping")
			return
		case <-ticker.C:
			cm.performRenewalCheck()
		}
	}
}

// performRenewalCheck checks all certificates and renews those that need it
func (cm *CertificateManager) performRenewalCheck() {
	log.Println("Performing certificate renewal check")

	renewalNeeded := cm.CheckRenewal()

	for serverID, needs := range renewalNeeded {
		if needs {
			log.Printf("Certificate for server %s needs renewal", serverID)

			if err := cm.RenewCertificate(serverID); err != nil {
				log.Printf("Failed to renew certificate for server %s: %v", serverID, err)
				continue
			}

			log.Printf("Successfully renewed certificate for server %s", serverID)

			// TODO: Trigger nginx reload after successful renewal
			// This would require integration with the nginx manager
		}
	}
}

// GetServerConfigurations returns all ACME configurations
func (cm *CertificateManager) GetServerConfigurations() map[string]*models.ACMEConfig {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	configs := make(map[string]*models.ACMEConfig)
	for serverID, config := range cm.configs {
		configs[serverID] = config
	}

	return configs
}

// GetCertificateInfo returns information about a certificate
func (cm *CertificateManager) GetCertificateInfo(serverID string) (*CertificateInfo, error) {
	cm.mu.RLock()
	manager, exists := cm.acmeManagers[serverID]
	config := cm.configs[serverID]
	cm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no ACME configuration found for server %s", serverID)
	}

	needsRenewal, err := manager.NeedsRenewal()
	if err != nil {
		return nil, fmt.Errorf("failed to check renewal status: %v", err)
	}

	certPath, keyPath := manager.GetCertificatePaths()

	return &CertificateInfo{
		ServerID:      serverID,
		Domains:       config.Domains,
		Provider:      config.Provider,
		Environment:   config.Environment,
		CertPath:      certPath,
		KeyPath:       keyPath,
		NeedsRenewal:  needsRenewal,
		ChallengeType: config.ChallengeType,
	}, nil
}

// CertificateInfo contains information about a managed certificate
type CertificateInfo struct {
	ServerID      string   `json:"server_id"`
	Domains       []string `json:"domains"`
	Provider      string   `json:"provider"`
	Environment   string   `json:"environment"`
	CertPath      string   `json:"cert_path"`
	KeyPath       string   `json:"key_path"`
	NeedsRenewal  bool     `json:"needs_renewal"`
	ChallengeType string   `json:"challenge_type"`
}

// ListCertificates returns information about all managed certificates
func (cm *CertificateManager) ListCertificates() ([]*CertificateInfo, error) {
	cm.mu.RLock()
	serverIDs := make([]string, 0, len(cm.acmeManagers))
	for serverID := range cm.acmeManagers {
		serverIDs = append(serverIDs, serverID)
	}
	cm.mu.RUnlock()

	var certificates []*CertificateInfo
	for _, serverID := range serverIDs {
		certInfo, err := cm.GetCertificateInfo(serverID)
		if err != nil {
			log.Printf("Failed to get certificate info for server %s: %v", serverID, err)
			continue
		}
		certificates = append(certificates, certInfo)
	}

	return certificates, nil
}
