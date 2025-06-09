package nginx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"go-apigateway-gui/internal/models"

	"github.com/flosch/pongo2/v6"
	"gopkg.in/yaml.v3"
)

type Manager struct {
	configPath           string
	executablePath       string
	backupPath           string
	templatesPath        string
	apiGatewayConfigPath string
	config               *models.Configuration
}

func NewManager(configPath, executablePath, templatesPath, apiGatewayConfigPath string) *Manager {
	return &Manager{
		configPath:           configPath,
		executablePath:       executablePath,
		backupPath:           "./backups",
		templatesPath:        templatesPath,
		apiGatewayConfigPath: apiGatewayConfigPath,
		config:               &models.Configuration{},
	}
}

// LoadConfiguration loads the current configuration from YAML file
func (m *Manager) LoadConfiguration() error {
	if _, err := os.Stat(m.apiGatewayConfigPath); os.IsNotExist(err) {
		// Create default configuration if it doesn't exist
		m.config = &models.Configuration{
			Backends: []models.Backend{},
			Servers:  []models.Server{},
			Global: models.GlobalConfig{
				WorkerProcesses:   "auto",
				WorkerConnections: 1024,
				KeepAliveTimeout:  "65s",
				ClientMaxBodySize: "1m",
				ServerTokens:      false,
				Gzip:              true,
				GzipTypes:         []string{"text/plain", "text/css", "application/json", "application/javascript"},
			},
		}
		return m.SaveConfiguration()
	}

	data, err := os.ReadFile(m.apiGatewayConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, m.config)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// SaveConfiguration saves the current configuration to YAML file
func (m *Manager) SaveConfiguration() error {
	data, err := yaml.Marshal(m.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(m.apiGatewayConfigPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GenerateNginxConfig generates nginx configuration from the current setup using the default template
func (m *Manager) GenerateNginxConfig() error {
	return m.GenerateNginxConfigWithTemplate("nginx.conf.j2")
}

// GenerateNginxConfigWithTemplate generates nginx configuration using a specific template
func (m *Manager) GenerateNginxConfigWithTemplate(templateName string) error {
	// Validate template exists
	if err := m.ValidateTemplate(templateName); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	// Create backup first
	if err := m.CreateBackup(); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Set up template loader for the templates directory
	templateSet := pongo2.NewSet("nginx", pongo2.MustNewLocalFileSystemLoader(m.templatesPath))

	// Load the specified template
	tmpl, err := templateSet.FromFile(templateName)
	if err != nil {
		return fmt.Errorf("failed to load template '%s': %w", templateName, err)
	}

	// Create context for template rendering
	ctx := pongo2.Context{
		"Global":   m.config.Global,
		"Backends": m.config.Backends,
		"Servers":  m.config.Servers,
	}

	// Render the template
	output, err := tmpl.Execute(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute template '%s': %w", templateName, err)
	}

	// Write the rendered output to the nginx config file
	err = os.WriteFile(m.configPath, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// TestConfiguration tests the nginx configuration
func (m *Manager) TestConfiguration() error {
	cmd := exec.Command(m.executablePath, "-t", "-c", m.configPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx config test failed: %s", string(output))
	}
	return nil
}

// ReloadNginx reloads nginx with the new configuration
func (m *Manager) ReloadNginx() error {
	if err := m.TestConfiguration(); err != nil {
		return err
	}

	cmd := exec.Command(m.executablePath, "-s", "reload", "-c", m.configPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nginx reload failed: %s", string(output))
	}
	return nil
}

// CreateBackup creates a backup of the current nginx configuration
func (m *Manager) CreateBackup() error {
	if err := os.MkdirAll(m.backupPath, 0755); err != nil {
		return err
	}

	timestamp := time.Now().Format("20060102-150405")
	backupFile := filepath.Join(m.backupPath, fmt.Sprintf("nginx-config-%s.conf", timestamp))

	input, err := os.ReadFile(m.configPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if len(input) > 0 {
		return os.WriteFile(backupFile, input, 0644)
	}

	return nil
}

// GetConfiguration returns the current configuration
func (m *Manager) GetConfiguration() *models.Configuration {
	return m.config
}

// AddBackend adds a new backend
func (m *Manager) AddBackend(backend models.Backend) error {
	backend.CreatedAt = time.Now()
	backend.UpdatedAt = time.Now()

	// Generate ID if not provided
	if backend.ID == "" {
		backend.ID = generateID("backend")
	}

	// Check if backend with this ID already exists
	for _, existing := range m.config.Backends {
		if existing.ID == backend.ID {
			return fmt.Errorf("backend with ID %s already exists", backend.ID)
		}
	}

	m.config.Backends = append(m.config.Backends, backend)
	return m.SaveConfiguration()
}

// UpdateBackend updates an existing backend
func (m *Manager) UpdateBackend(id string, backend models.Backend) error {
	for i, existing := range m.config.Backends {
		if existing.ID == id {
			backend.ID = id
			backend.CreatedAt = existing.CreatedAt
			backend.UpdatedAt = time.Now()
			m.config.Backends[i] = backend
			return m.SaveConfiguration()
		}
	}
	return fmt.Errorf("backend with ID %s not found", id)
}

// DeleteBackend removes a backend
func (m *Manager) DeleteBackend(id string) error {
	for i, backend := range m.config.Backends {
		if backend.ID == id {
			m.config.Backends = append(m.config.Backends[:i], m.config.Backends[i+1:]...)
			return m.SaveConfiguration()
		}
	}
	return fmt.Errorf("backend with ID %s not found", id)
}

// AddRoute adds a new route
func (m *Manager) AddRoute(serverID string, route models.Route) error {
	route.CreatedAt = time.Now()
	route.UpdatedAt = time.Now()

	// Generate ID if not provided
	if route.ID == "" {
		route.ID = generateID("route")
	}

	for i, server := range m.config.Servers {
		if server.ID == serverID {
			m.config.Servers[i].Routes = append(m.config.Servers[i].Routes, route)
			m.config.Servers[i].UpdatedAt = time.Now()
			return m.SaveConfiguration()
		}
	}
	return fmt.Errorf("server with ID %s not found", serverID)
}

// UpdateRoute updates an existing route
func (m *Manager) UpdateRoute(serverID, routeID string, route models.Route) error {
	for i, server := range m.config.Servers {
		if server.ID == serverID {
			for j, existing := range server.Routes {
				if existing.ID == routeID {
					route.ID = routeID
					route.CreatedAt = existing.CreatedAt
					route.UpdatedAt = time.Now()
					m.config.Servers[i].Routes[j] = route
					m.config.Servers[i].UpdatedAt = time.Now()
					return m.SaveConfiguration()
				}
			}
		}
	}
	return fmt.Errorf("route with ID %s not found in server %s", routeID, serverID)
}

// DeleteRoute removes a route
func (m *Manager) DeleteRoute(serverID, routeID string) error {
	for i, server := range m.config.Servers {
		if server.ID == serverID {
			for j, route := range server.Routes {
				if route.ID == routeID {
					m.config.Servers[i].Routes = append(server.Routes[:j], server.Routes[j+1:]...)
					m.config.Servers[i].UpdatedAt = time.Now()
					return m.SaveConfiguration()
				}
			}
		}
	}
	return fmt.Errorf("route with ID %s not found in server %s", routeID, serverID)
}

// UpdateRouteByID updates a route by finding it across all servers
func (m *Manager) UpdateRouteByID(routeID string, route models.Route) error {
	for i, server := range m.config.Servers {
		for j, existing := range server.Routes {
			if existing.ID == routeID {
				route.ID = routeID
				route.CreatedAt = existing.CreatedAt
				route.UpdatedAt = time.Now()
				m.config.Servers[i].Routes[j] = route
				m.config.Servers[i].UpdatedAt = time.Now()
				return m.SaveConfiguration()
			}
		}
	}
	return fmt.Errorf("route with ID %s not found", routeID)
}

// DeleteRouteByID removes a route by finding it across all servers
func (m *Manager) DeleteRouteByID(routeID string) error {
	for i, server := range m.config.Servers {
		for j, route := range server.Routes {
			if route.ID == routeID {
				m.config.Servers[i].Routes = append(server.Routes[:j], server.Routes[j+1:]...)
				m.config.Servers[i].UpdatedAt = time.Now()
				return m.SaveConfiguration()
			}
		}
	}
	return fmt.Errorf("route with ID %s not found", routeID)
}

// AddServer adds a new server
func (m *Manager) AddServer(server models.Server) error {
	server.CreatedAt = time.Now()
	server.UpdatedAt = time.Now()

	// Generate ID if not provided
	if server.ID == "" {
		server.ID = generateID("server")
	}

	// Check if server with this ID already exists
	for _, existing := range m.config.Servers {
		if existing.ID == server.ID {
			return fmt.Errorf("server with ID %s already exists", server.ID)
		}
	}

	m.config.Servers = append(m.config.Servers, server)
	return m.SaveConfiguration()
}

// UpdateServer updates an existing server
func (m *Manager) UpdateServer(id string, server models.Server) error {
	for i, existing := range m.config.Servers {
		if existing.ID == id {
			server.ID = id
			server.CreatedAt = existing.CreatedAt
			server.UpdatedAt = time.Now()
			// Preserve existing routes if not provided
			if len(server.Routes) == 0 {
				server.Routes = existing.Routes
			}
			m.config.Servers[i] = server
			return m.SaveConfiguration()
		}
	}
	return fmt.Errorf("server with ID %s not found", id)
}

// DeleteServer removes a server
func (m *Manager) DeleteServer(id string) error {
	for i, server := range m.config.Servers {
		if server.ID == id {
			m.config.Servers = append(m.config.Servers[:i], m.config.Servers[i+1:]...)
			return m.SaveConfiguration()
		}
	}
	return fmt.Errorf("server with ID %s not found", id)
}

// GetBackend returns a backend by ID
func (m *Manager) GetBackend(id string) (*models.Backend, error) {
	for _, backend := range m.config.Backends {
		if backend.ID == id {
			return &backend, nil
		}
	}
	return nil, fmt.Errorf("backend with ID %s not found", id)
}

// GetServer returns a server by ID
func (m *Manager) GetServer(id string) (*models.Server, error) {
	for _, server := range m.config.Servers {
		if server.ID == id {
			return &server, nil
		}
	}
	return nil, fmt.Errorf("server with ID %s not found", id)
}

// generateID generates a unique ID
func generateID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

// GetNginxStatus returns the status of nginx process
func (m *Manager) GetNginxStatus() (string, error) {
	// Check if nginx process is running
	cmd := exec.Command("pgrep", "nginx")
	if err := cmd.Run(); err != nil {
		return "stopped", nil
	}
	return "running", nil
}

// ListAvailableTemplates returns a list of available template files
func (m *Manager) ListAvailableTemplates() ([]TemplateInfo, error) {
	entries, err := os.ReadDir(m.templatesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
	}

	var templates []TemplateInfo
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".j2" {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			templateInfo := TemplateInfo{
				Name:        entry.Name(),
				DisplayName: getDisplayName(entry.Name()),
				Description: getTemplateDescription(entry.Name()),
				ModTime:     info.ModTime(),
				Size:        info.Size(),
			}
			templates = append(templates, templateInfo)
		}
	}

	return templates, nil
}

// ValidateTemplate checks if a template exists and has valid syntax
func (m *Manager) ValidateTemplate(templateName string) error {
	templatePath := filepath.Join(m.templatesPath, templateName)

	// Check if file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template '%s' does not exist", templateName)
	}

	// Try to parse the template to validate syntax
	templateSet := pongo2.NewSet("validation", pongo2.MustNewLocalFileSystemLoader(m.templatesPath))
	_, err := templateSet.FromFile(templateName)
	if err != nil {
		return fmt.Errorf("template syntax error in '%s': %w", templateName, err)
	}

	return nil
}

// PreviewTemplate renders a template with current configuration but doesn't write to file
func (m *Manager) PreviewTemplate(templateName string) (string, error) {
	// Validate template exists
	if err := m.ValidateTemplate(templateName); err != nil {
		return "", fmt.Errorf("template validation failed: %w", err)
	}

	// Set up template loader for the templates directory
	templateSet := pongo2.NewSet("preview", pongo2.MustNewLocalFileSystemLoader(m.templatesPath))

	// Load the specified template
	tmpl, err := templateSet.FromFile(templateName)
	if err != nil {
		return "", fmt.Errorf("failed to load template '%s': %w", templateName, err)
	}

	// Create context for template rendering
	ctx := pongo2.Context{
		"Global":   m.config.Global,
		"Backends": m.config.Backends,
		"Servers":  m.config.Servers,
	}

	// Render the template
	output, err := tmpl.Execute(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to execute template '%s': %w", templateName, err)
	}

	return output, nil
}

// TemplateInfo represents information about a template file
type TemplateInfo struct {
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	ModTime     time.Time `json:"mod_time"`
	Size        int64     `json:"size"`
}

// Helper function to get display name from template filename
func getDisplayName(filename string) string {
	name := filepath.Base(filename)
	ext := filepath.Ext(name)
	baseName := name[:len(name)-len(ext)]

	switch baseName {
	case "nginx.conf":
		return "Standard Configuration"
	case "minimal":
		return "Minimal Configuration"
	case "development":
		return "Development Configuration"
	default:
		return baseName
	}
}

// Helper function to get template description
func getTemplateDescription(filename string) string {
	name := filepath.Base(filename)
	ext := filepath.Ext(name)
	baseName := name[:len(name)-len(ext)]

	switch baseName {
	case "nginx.conf":
		return "Full-featured nginx configuration with all standard directives and optimizations"
	case "minimal":
		return "Minimal nginx configuration with only essential directives for basic functionality"
	case "development":
		return "Development-focused configuration with debugging features and relaxed security"
	default:
		return "Custom nginx configuration template"
	}
}
