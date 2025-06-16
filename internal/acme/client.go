package acme

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"

	"go-apigateway-gui/internal/models"
)

// Manager handles ACME certificate operations
type Manager struct {
	config     *models.ACMEConfig
	client     *lego.Client
	user       *User
	httpServer *http01.ProviderServer
	certDir    string
}

// User represents an ACME user account
type User struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

// GetEmail returns the user's email address
func (u *User) GetEmail() string {
	return u.Email
}

// GetRegistration returns the user's registration resource
func (u *User) GetRegistration() *registration.Resource {
	return u.Registration
}

// GetPrivateKey returns the user's private key
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// NewManager creates a new ACME manager
func NewManager(config *models.ACMEConfig, certDir string) (*Manager, error) {
	if config == nil || !config.Enabled {
		return nil, fmt.Errorf("ACME configuration is not enabled")
	}

	if config.Email == "" {
		return nil, fmt.Errorf("ACME email is required")
	}

	if len(config.Domains) == 0 {
		return nil, fmt.Errorf("ACME domains are required")
	}

	// Ensure certificate directory exists
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create certificate directory: %v", err)
	}

	manager := &Manager{
		config:  config,
		certDir: certDir,
	}

	if err := manager.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize ACME manager: %v", err)
	}

	return manager, nil
}

// initialize sets up the ACME client
func (m *Manager) initialize() error {
	// Create or load user account
	user, err := m.getOrCreateUser()
	if err != nil {
		return fmt.Errorf("failed to get or create user: %v", err)
	}
	m.user = user

	// Determine ACME directory URL based on provider and environment
	directoryURL := m.getDirectoryURL()

	// Create lego config
	config := lego.NewConfig(user)
	config.CADirURL = directoryURL

	// Set certificate type based on key type
	switch m.config.KeyType {
	case "EC256":
		config.Certificate.KeyType = certcrypto.EC256
	case "EC384":
		config.Certificate.KeyType = certcrypto.EC384
	case "RSA4096":
		config.Certificate.KeyType = certcrypto.RSA4096
	default:
		config.Certificate.KeyType = certcrypto.RSA2048
	}

	// Create ACME client
	client, err := lego.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create ACME client: %v", err)
	}
	m.client = client

	// Set up challenge solver based on challenge type
	if err := m.setupChallenge(); err != nil {
		return fmt.Errorf("failed to setup challenge: %v", err)
	}

	// Register user if not already registered
	if user.Registration == nil {
		reg, err := client.Registration.Register(registration.RegisterOptions{
			TermsOfServiceAgreed: true,
		})
		if err != nil {
			return fmt.Errorf("failed to register user: %v", err)
		}
		user.Registration = reg

		// Save user registration
		if err := m.saveUser(user); err != nil {
			log.Printf("Warning: failed to save user registration: %v", err)
		}
	}

	return nil
}

// getDirectoryURL returns the ACME directory URL based on provider and environment
func (m *Manager) getDirectoryURL() string {
	provider := m.config.Provider
	environment := m.config.Environment

	// Default to Let's Encrypt if not specified
	if provider == "" {
		provider = "letsencrypt"
	}
	if environment == "" {
		environment = "production"
	}

	switch provider {
	case "letsencrypt":
		if environment == "staging" {
			return lego.LEDirectoryStaging
		}
		return lego.LEDirectoryProduction
	case "buypass":
		if environment == "staging" {
			return "https://api.test4.buypass.no/acme/directory"
		}
		return "https://api.buypass.com/acme/directory"
	case "zerossl":
		return "https://acme.zerossl.com/v2/DV90"
	default:
		// Fallback to Let's Encrypt production
		return lego.LEDirectoryProduction
	}
}

// setupChallenge configures the challenge solver
func (m *Manager) setupChallenge() error {
	switch m.config.ChallengeType {
	case "http-01", "":
		// HTTP-01 challenge (default)
		port := m.config.HTTPChallengePort
		if port == 0 {
			port = 80 // Default HTTP port
		}

		httpProvider := http01.NewProviderServer("", fmt.Sprintf("%d", port))
		if err := m.client.Challenge.SetHTTP01Provider(httpProvider); err != nil {
			return fmt.Errorf("failed to set HTTP-01 provider: %v", err)
		}
		m.httpServer = httpProvider

	case "dns-01":
		// DNS-01 challenge
		if m.config.DNSProvider == "" {
			return fmt.Errorf("DNS provider is required for DNS-01 challenge")
		}

		// Note: DNS provider setup would require specific provider implementations
		// This is a simplified version - real implementation would need specific DNS providers
		return fmt.Errorf("DNS-01 challenge not yet implemented")

	case "tls-alpn-01":
		// TLS-ALPN-01 challenge
		if err := m.client.Challenge.SetTLSALPN01Provider(http01.NewProviderServer("", "443")); err != nil {
			return fmt.Errorf("failed to set TLS-ALPN-01 provider: %v", err)
		}

	default:
		return fmt.Errorf("unsupported challenge type: %s", m.config.ChallengeType)
	}

	return nil
}

// getOrCreateUser loads or creates an ACME user account
func (m *Manager) getOrCreateUser() (*User, error) {
	userKeyPath := filepath.Join(m.certDir, "user.key")
	userDataPath := filepath.Join(m.certDir, "user.json")

	// Try to load existing user
	if user, err := m.loadUser(userKeyPath, userDataPath); err == nil {
		return user, nil
	}

	// Create new user
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	user := &User{
		Email: m.config.Email,
		key:   privateKey,
	}

	// Save user private key
	if err := m.savePrivateKey(privateKey, userKeyPath); err != nil {
		return nil, fmt.Errorf("failed to save user private key: %v", err)
	}

	return user, nil
}

// loadUser loads an existing user account
func (m *Manager) loadUser(keyPath, dataPath string) (*User, error) {
	// Load private key
	keyData, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	user := &User{
		Email: m.config.Email,
		key:   privateKey,
	}

	return user, nil
}

// saveUser saves user account information
func (m *Manager) saveUser(user *User) error {
	// In a real implementation, you might save registration details to JSON
	// For now, we just ensure the key is saved
	userKeyPath := filepath.Join(m.certDir, "user.key")
	return m.savePrivateKey(user.key.(*rsa.PrivateKey), userKeyPath)
}

// savePrivateKey saves a private key to file
func (m *Manager) savePrivateKey(key *rsa.PrivateKey, path string) error {
	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	})

	return ioutil.WriteFile(path, keyPEM, 0600)
}

// ObtainCertificate requests a new certificate from the ACME provider
func (m *Manager) ObtainCertificate() (*certificate.Resource, error) {
	if m.client == nil {
		return nil, fmt.Errorf("ACME client not initialized")
	}

	request := certificate.ObtainRequest{
		Domains: m.config.Domains,
		Bundle:  true,
	}

	if m.config.MustStaple {
		request.MustStaple = true
	}

	cert, err := m.client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain certificate: %v", err)
	}

	// Save certificate to configured paths
	if err := m.saveCertificate(cert); err != nil {
		log.Printf("Warning: failed to save certificate: %v", err)
	}

	return cert, nil
}

// RenewCertificate renews an existing certificate
func (m *Manager) RenewCertificate() (*certificate.Resource, error) {
	if m.client == nil {
		return nil, fmt.Errorf("ACME client not initialized")
	}

	// Load existing certificate
	cert, err := m.loadCertificate()
	if err != nil {
		return nil, fmt.Errorf("failed to load existing certificate: %v", err)
	}

	renewedCert, err := m.client.Certificate.Renew(*cert, true, false, "")
	if err != nil {
		return nil, fmt.Errorf("failed to renew certificate: %v", err)
	}

	// Save renewed certificate
	if err := m.saveCertificate(renewedCert); err != nil {
		log.Printf("Warning: failed to save renewed certificate: %v", err)
	}

	return renewedCert, nil
}

// NeedsRenewal checks if the certificate needs renewal
func (m *Manager) NeedsRenewal() (bool, error) {
	certPath := m.config.CertPath
	if certPath == "" {
		certPath = filepath.Join(m.certDir, "cert.pem")
	}

	certData, err := ioutil.ReadFile(certPath)
	if err != nil {
		return true, nil // Certificate doesn't exist, needs to be obtained
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return true, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return true, fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Check if certificate expires within the renewal window
	renewDays := m.config.RenewDays
	if renewDays == 0 {
		renewDays = 30 // Default to 30 days
	}

	renewalTime := time.Now().AddDate(0, 0, renewDays)
	return cert.NotAfter.Before(renewalTime), nil
}

// saveCertificate saves the certificate and private key to files
func (m *Manager) saveCertificate(cert *certificate.Resource) error {
	certPath := m.config.CertPath
	if certPath == "" {
		certPath = filepath.Join(m.certDir, "cert.pem")
	}

	keyPath := m.config.KeyPath
	if keyPath == "" {
		keyPath = filepath.Join(m.certDir, "key.pem")
	}

	// Ensure directories exist
	if err := os.MkdirAll(filepath.Dir(certPath), 0755); err != nil {
		return fmt.Errorf("failed to create cert directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(keyPath), 0755); err != nil {
		return fmt.Errorf("failed to create key directory: %v", err)
	}

	// Save certificate
	if err := ioutil.WriteFile(certPath, cert.Certificate, 0644); err != nil {
		return fmt.Errorf("failed to save certificate: %v", err)
	}

	// Save private key
	if err := ioutil.WriteFile(keyPath, cert.PrivateKey, 0600); err != nil {
		return fmt.Errorf("failed to save private key: %v", err)
	}

	log.Printf("Certificate saved to %s, private key saved to %s", certPath, keyPath)
	return nil
}

// loadCertificate loads an existing certificate resource
func (m *Manager) loadCertificate() (*certificate.Resource, error) {
	certPath := m.config.CertPath
	if certPath == "" {
		certPath = filepath.Join(m.certDir, "cert.pem")
	}

	keyPath := m.config.KeyPath
	if keyPath == "" {
		keyPath = filepath.Join(m.certDir, "key.pem")
	}

	certData, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %v", err)
	}

	keyData, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %v", err)
	}

	return &certificate.Resource{
		Domain:      m.config.Domains[0],
		Certificate: certData,
		PrivateKey:  keyData,
	}, nil
}

// StartHTTPChallenge starts the HTTP challenge server if using HTTP-01
func (m *Manager) StartHTTPChallenge() error {
	if m.config.ChallengeType == "http-01" && m.httpServer != nil {
		port := m.config.HTTPChallengePort
		if port == 0 {
			port = 80
		}

		log.Printf("HTTP challenge server configured for port %d", port)
		// Note: The actual HTTP challenge server is started automatically by lego
		// when ObtainCertificate() or RenewCertificate() is called
	}
	return nil
}

// StopHTTPChallenge stops the HTTP challenge server
func (m *Manager) StopHTTPChallenge() error {
	// Note: The HTTP challenge provider in lego is stateless and doesn't require cleanup
	// The server is automatically stopped when the challenge is completed
	return nil
}

// GetCertificatePaths returns the paths where certificates are stored
func (m *Manager) GetCertificatePaths() (certPath, keyPath string) {
	certPath = m.config.CertPath
	if certPath == "" {
		certPath = filepath.Join(m.certDir, "cert.pem")
	}

	keyPath = m.config.KeyPath
	if keyPath == "" {
		keyPath = filepath.Join(m.certDir, "key.pem")
	}

	return certPath, keyPath
}
