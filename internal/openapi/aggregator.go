package openapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"go-apigateway-gui/internal/models"
)

// Aggregator handles fetching and aggregating OpenAPI specifications from multiple backends
type Aggregator struct {
	client        *http.Client
	cache         map[string]*models.OpenAPISpec
	cacheMutex    sync.RWMutex
	cacheTTL      time.Duration
	lastRefresh   time.Time
	configuration *models.Configuration
}

// NewAggregator creates a new OpenAPI aggregator
func NewAggregator(config *models.Configuration, cacheTTL time.Duration) *Aggregator {
	return &Aggregator{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache:         make(map[string]*models.OpenAPISpec),
		cacheTTL:      cacheTTL,
		configuration: config,
	}
}

// UpdateConfiguration updates the configuration used by the aggregator
func (a *Aggregator) UpdateConfiguration(config *models.Configuration) {
	a.configuration = config
}

// FetchBackendSpec fetches OpenAPI specification from a single backend
func (a *Aggregator) FetchBackendSpec(backend models.Backend) *models.OpenAPISpec {
	spec := &models.OpenAPISpec{
		BackendID: backend.ID,
		FetchedAt: time.Now(),
	}

	if backend.OpenAPI == nil || !backend.OpenAPI.Enabled {
		spec.Error = "OpenAPI not enabled for this backend"
		return spec
	}

	if len(backend.Servers) == 0 {
		spec.Error = "No servers configured for backend"
		return spec
	}

	// Try each configured OpenAPI endpoint
	for _, endpoint := range backend.OpenAPI.Endpoints {
		for _, server := range backend.Servers {
			if server.Down {
				continue // Skip down servers
			}

			baseURL := fmt.Sprintf("http://%s:%d", server.Host, server.Port)
			specURL := baseURL + endpoint
			spec.URL = specURL

			if rawSpec, err := a.fetchSpecFromURL(specURL); err == nil {
				spec.Raw = rawSpec
				if err := a.parseOpenAPISpec(spec); err == nil {
					return spec // Successfully fetched and parsed
				}
				spec.Error = fmt.Sprintf("Failed to parse OpenAPI spec: %v", err)
			} else {
				spec.Error = fmt.Sprintf("Failed to fetch from %s: %v", specURL, err)
			}
		}
	}

	return spec
}

// fetchSpecFromURL fetches the OpenAPI specification from a URL
func (a *Aggregator) fetchSpecFromURL(url string) (map[string]interface{}, error) {
	resp, err := a.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(body, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return spec, nil
}

// parseOpenAPISpec extracts relevant parts from the raw OpenAPI specification
func (a *Aggregator) parseOpenAPISpec(spec *models.OpenAPISpec) error {
	raw := spec.Raw

	// Determine OpenAPI version
	if swagger, ok := raw["swagger"].(string); ok {
		spec.Version = swagger // Swagger 2.0
	} else if openapi, ok := raw["openapi"].(string); ok {
		spec.Version = openapi // OpenAPI 3.x
	} else {
		return fmt.Errorf("unable to determine OpenAPI version")
	}

	// Extract paths
	if paths, ok := raw["paths"].(map[string]interface{}); ok {
		spec.Paths = paths
	}

	// Extract info
	if info, ok := raw["info"].(map[string]interface{}); ok {
		spec.Info = info
	}

	// Extract components (OpenAPI 3.x) or definitions (Swagger 2.0)
	if components, ok := raw["components"].(map[string]interface{}); ok {
		spec.Components = components
	} else if definitions, ok := raw["definitions"].(map[string]interface{}); ok {
		// Convert Swagger 2.0 definitions to OpenAPI 3.x components format
		spec.Components = map[string]interface{}{
			"schemas": definitions,
		}
	}

	// Extract tags
	if tags, ok := raw["tags"].([]interface{}); ok {
		spec.Tags = tags
	}

	// Extract servers
	if servers, ok := raw["servers"].([]interface{}); ok {
		spec.Servers = servers
	}

	return nil
}

// AggregateSpecs creates a unified OpenAPI specification from all enabled backends
func (a *Aggregator) AggregateSpecs() (*models.AggregatedOpenAPISpec, error) {
	a.cacheMutex.Lock()
	defer a.cacheMutex.Unlock()

	// Check if we need to refresh the cache
	if time.Since(a.lastRefresh) > a.cacheTTL {
		a.refreshCache()
	}

	aggregated := &models.AggregatedOpenAPISpec{
		OpenAPI:     "3.0.3",
		GeneratedAt: time.Now(),
		Info: map[string]interface{}{
			"title":       "API Gateway - Unified API Documentation",
			"description": "Aggregated OpenAPI documentation from all backend services",
			"version":     "1.0.0",
		},
		Paths:      make(map[string]interface{}),
		Components: make(map[string]interface{}),
		Tags:       make([]interface{}, 0),
		Backends:   make([]string, 0),
	}

	// Build servers list from gateway configuration
	if a.configuration != nil {
		servers := make([]interface{}, 0)
		for _, server := range a.configuration.Servers {
			for _, listen := range server.Listen {
				serverURL := a.buildServerURL(server, listen)
				if serverURL != "" {
					servers = append(servers, map[string]interface{}{
						"url":         serverURL,
						"description": fmt.Sprintf("%s (%s)", server.Name, listen),
					})
				}
			}
		}
		aggregated.Servers = servers
	}

	// Aggregate specs from all backends
	for _, backend := range a.configuration.Backends {
		if spec, exists := a.cache[backend.ID]; exists && spec.Error == "" {
			if err := a.mergeSpecIntoAggregated(aggregated, spec, backend); err != nil {
				// Log error but continue with other backends
				continue
			}
			aggregated.Backends = append(aggregated.Backends, backend.ID)
		}
	}

	return aggregated, nil
}

// refreshCache fetches fresh OpenAPI specifications from all backends
func (a *Aggregator) refreshCache() {
	a.lastRefresh = time.Now()

	if a.configuration == nil {
		return
	}

	// Use goroutines to fetch specs concurrently
	var wg sync.WaitGroup
	specChan := make(chan *models.OpenAPISpec, len(a.configuration.Backends))

	for _, backend := range a.configuration.Backends {
		if backend.OpenAPI != nil && backend.OpenAPI.Enabled {
			wg.Add(1)
			go func(b models.Backend) {
				defer wg.Done()
				spec := a.FetchBackendSpec(b)
				specChan <- spec
			}(backend)
		}
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(specChan)
	}()

	// Collect results
	for spec := range specChan {
		a.cache[spec.BackendID] = spec
	}
}

// mergeSpecIntoAggregated merges a backend's OpenAPI spec into the aggregated spec
func (a *Aggregator) mergeSpecIntoAggregated(aggregated *models.AggregatedOpenAPISpec, spec *models.OpenAPISpec, backend models.Backend) error {
	// Transform and merge paths
	if err := a.mergePaths(aggregated, spec, backend); err != nil {
		return fmt.Errorf("failed to merge paths for backend %s: %w", backend.ID, err)
	}

	// Merge components
	a.mergeComponents(aggregated, spec, backend)

	// Merge tags
	a.mergeTags(aggregated, spec, backend)

	return nil
}

// mergePaths transforms backend paths according to gateway routing and merges them
func (a *Aggregator) mergePaths(aggregated *models.AggregatedOpenAPISpec, spec *models.OpenAPISpec, backend models.Backend) error {
	if spec.Paths == nil {
		return nil
	}

	// Find routes that use this backend
	backendRoutes := a.findRoutesForBackend(backend.ID)

	for originalPath, pathItem := range spec.Paths {
		pathItemMap, ok := pathItem.(map[string]interface{})
		if !ok {
			continue
		}

		// Transform the path for each route that uses this backend
		for _, route := range backendRoutes {
			transformedPath := a.transformPath(originalPath, route)

			// Clone the path item to avoid conflicts
			clonedPathItem := a.clonePathItem(pathItemMap)

			// Add backend information to operation descriptions
			a.addBackendInfo(clonedPathItem, backend, route)

			// Merge into aggregated paths
			if existing, exists := aggregated.Paths[transformedPath]; exists {
				// Merge operations from different backends
				if existingMap, ok := existing.(map[string]interface{}); ok {
					a.mergePathOperations(existingMap, clonedPathItem)
				}
			} else {
				aggregated.Paths[transformedPath] = clonedPathItem
			}
		}
	}

	return nil
}

// findRoutesForBackend finds all routes that point to a specific backend
func (a *Aggregator) findRoutesForBackend(backendID string) []models.Route {
	var routes []models.Route

	if a.configuration == nil {
		return routes
	}

	for _, server := range a.configuration.Servers {
		for _, route := range server.Routes {
			if route.BackendID == backendID {
				routes = append(routes, route)
			}
		}
	}

	return routes
}

// transformPath transforms a backend path according to gateway routing rules
func (a *Aggregator) transformPath(backendPath string, route models.Route) string {
	routePath := route.Path

	// Remove trailing slash from route path for consistency
	routePath = strings.TrimSuffix(routePath, "/")

	if route.StripPath {
		// If strip_path is true, the route path replaces the backend base
		// Backend path: /users/{id}
		// Route path: /api/v1/users
		// Result: /api/v1/users/{id}

		// Remove leading slash from backend path
		backendPath = strings.TrimPrefix(backendPath, "/")

		if backendPath == "" {
			return routePath
		}

		return routePath + "/" + backendPath
	} else {
		// If strip_path is false, append backend path to route path
		// Backend path: /users/{id}
		// Route path: /api/v1
		// Result: /api/v1/users/{id}

		return routePath + backendPath
	}
}

// Helper methods for spec merging and transformation
func (a *Aggregator) clonePathItem(pathItem map[string]interface{}) map[string]interface{} {
	clone := make(map[string]interface{})
	for k, v := range pathItem {
		clone[k] = v
	}
	return clone
}

func (a *Aggregator) addBackendInfo(pathItem map[string]interface{}, backend models.Backend, route models.Route) {
	// Add backend information to each HTTP method in the path
	httpMethods := []string{"get", "post", "put", "delete", "patch", "head", "options", "trace"}

	for _, method := range httpMethods {
		if operation, exists := pathItem[method]; exists {
			if operationMap, ok := operation.(map[string]interface{}); ok {
				// Add backend info to description
				if desc, hasDesc := operationMap["description"].(string); hasDesc {
					operationMap["description"] = fmt.Sprintf("%s\n\n**Backend:** %s (%s)",
						desc, backend.Name, backend.ID)
				} else {
					operationMap["description"] = fmt.Sprintf("**Backend:** %s (%s)",
						backend.Name, backend.ID)
				}

				// Add custom extension with backend info
				operationMap["x-backend-info"] = map[string]interface{}{
					"backend_id":   backend.ID,
					"backend_name": backend.Name,
					"route_path":   route.Path,
					"strip_path":   route.StripPath,
				}
			}
		}
	}
}

func (a *Aggregator) mergePathOperations(target, source map[string]interface{}) {
	for key, value := range source {
		if _, exists := target[key]; !exists {
			target[key] = value
		}
		// TODO: Handle operation conflicts (same method on same path from different backends)
	}
}

func (a *Aggregator) mergeComponents(aggregated *models.AggregatedOpenAPISpec, spec *models.OpenAPISpec, backend models.Backend) {
	if spec.Components == nil {
		return
	}

	if aggregated.Components == nil {
		aggregated.Components = make(map[string]interface{})
	}

	// Merge each component type (schemas, parameters, responses, etc.)
	for componentType, components := range spec.Components {
		if componentsMap, ok := components.(map[string]interface{}); ok {
			if _, exists := aggregated.Components[componentType]; !exists {
				aggregated.Components[componentType] = make(map[string]interface{})
			}

			if targetComponents, ok := aggregated.Components[componentType].(map[string]interface{}); ok {
				for name, component := range componentsMap {
					// Prefix component names with backend ID to avoid conflicts
					prefixedName := fmt.Sprintf("%s_%s", backend.ID, name)
					targetComponents[prefixedName] = component
				}
			}
		}
	}
}

func (a *Aggregator) mergeTags(aggregated *models.AggregatedOpenAPISpec, spec *models.OpenAPISpec, backend models.Backend) {
	if spec.Tags == nil {
		return
	}

	// Add backend-specific tags
	for _, tag := range spec.Tags {
		if tagMap, ok := tag.(map[string]interface{}); ok {
			// Prefix tag name with backend name to avoid conflicts
			if name, hasName := tagMap["name"].(string); hasName {
				newTag := make(map[string]interface{})
				for k, v := range tagMap {
					newTag[k] = v
				}
				newTag["name"] = fmt.Sprintf("%s - %s", backend.Name, name)
				newTag["x-backend-id"] = backend.ID

				aggregated.Tags = append(aggregated.Tags, newTag)
			}
		}
	}
}

func (a *Aggregator) buildServerURL(server models.Server, listen string) string {
	// Parse listen directive
	parts := strings.Fields(listen)
	if len(parts) == 0 {
		return ""
	}

	port := parts[0]
	isSSL := len(parts) > 1 && strings.Contains(strings.Join(parts[1:], " "), "ssl")

	// Use first server name or default
	hostname := "localhost"
	if len(server.ServerName) > 0 && server.ServerName[0] != "" {
		hostname = server.ServerName[0]
	}

	scheme := "http"
	if isSSL {
		scheme = "https"
	}

	// Handle default ports
	if (scheme == "http" && port == "80") || (scheme == "https" && port == "443") {
		return fmt.Sprintf("%s://%s", scheme, hostname)
	}

	return fmt.Sprintf("%s://%s:%s", scheme, hostname, port)
}

// GetCacheStatus returns the current cache status
func (a *Aggregator) GetCacheStatus() map[string]interface{} {
	a.cacheMutex.RLock()
	defer a.cacheMutex.RUnlock()

	status := map[string]interface{}{
		"last_refresh":    a.lastRefresh,
		"cache_ttl":       a.cacheTTL,
		"cached_backends": make(map[string]interface{}),
	}

	for backendID, spec := range a.cache {
		backendStatus := map[string]interface{}{
			"fetched_at": spec.FetchedAt,
			"url":        spec.URL,
			"version":    spec.Version,
			"has_error":  spec.Error != "",
		}

		if spec.Error != "" {
			backendStatus["error"] = spec.Error
		} else {
			pathCount := 0
			if spec.Paths != nil {
				pathCount = len(spec.Paths)
			}
			backendStatus["path_count"] = pathCount
		}

		status["cached_backends"].(map[string]interface{})[backendID] = backendStatus
	}

	return status
}

// ForceRefresh forces a cache refresh
func (a *Aggregator) ForceRefresh() {
	a.cacheMutex.Lock()
	defer a.cacheMutex.Unlock()
	a.refreshCache()
}

// GetAggregatedSpec returns the current aggregated OpenAPI specification
func (a *Aggregator) GetAggregatedSpec() (*models.AggregatedOpenAPISpec, error) {
	return a.AggregateSpecs()
}

// RefreshCache forces a refresh of the cache and returns any error
func (a *Aggregator) RefreshCache() error {
	a.ForceRefresh()
	return nil
}

// GetCachedSpec returns a cached spec for a specific backend
func (a *Aggregator) GetCachedSpec(backendID string) *models.OpenAPISpec {
	a.cacheMutex.RLock()
	defer a.cacheMutex.RUnlock()
	return a.cache[backendID]
}

// GetStats returns aggregation statistics
func (a *Aggregator) GetStats() map[string]interface{} {
	a.cacheMutex.RLock()
	defer a.cacheMutex.RUnlock()

	successfulFetches := 0
	failedFetches := 0

	for _, spec := range a.cache {
		if spec.Error == "" {
			successfulFetches++
		} else {
			failedFetches++
		}
	}

	return map[string]interface{}{
		"LastRefresh":       a.lastRefresh,
		"CachedSpecs":       len(a.cache),
		"SuccessfulFetches": successfulFetches,
		"FailedFetches":     failedFetches,
	}
}

// GetCacheTTL returns the cache TTL duration
func (a *Aggregator) GetCacheTTL() time.Duration {
	return a.cacheTTL
}
