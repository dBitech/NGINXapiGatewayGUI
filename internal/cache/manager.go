package cache

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-apigateway-gui/internal/models"
)

// Manager handles nginx cache operations
type Manager struct {
	nginxBaseURL  string // Base URL for nginx cache purge endpoints
	cachePath     string // Physical path to nginx cache directory
	cacheZones    []string
	client        *http.Client
	configuration *models.Configuration
}

// CacheStats represents cache statistics
type CacheStats struct {
	TotalSize    int64                         `json:"total_size"`
	FileCount    int                           `json:"file_count"`
	LastModified time.Time                     `json:"last_modified"`
	Zones        map[string]*CacheZoneStats    `json:"zones"`
	BackendStats map[string]*BackendCacheStats `json:"backend_stats"`
}

// CacheZoneStats represents statistics for a specific cache zone
type CacheZoneStats struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	FileCount    int       `json:"file_count"`
	LastAccessed time.Time `json:"last_accessed"`
}

// BackendCacheStats represents cache statistics for a specific backend
type BackendCacheStats struct {
	BackendID   string    `json:"backend_id"`
	BackendName string    `json:"backend_name"`
	CachedPaths []string  `json:"cached_paths"`
	HitCount    int       `json:"hit_count"`
	MissCount   int       `json:"miss_count"`
	LastHit     time.Time `json:"last_hit"`
}

// NewManager creates a new cache manager
func NewManager(nginxBaseURL, cachePath string, zones []string, config *models.Configuration) *Manager {
	return &Manager{
		nginxBaseURL:  nginxBaseURL,
		cachePath:     cachePath,
		cacheZones:    zones,
		configuration: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// UpdateConfiguration updates the configuration reference
func (cm *Manager) UpdateConfiguration(config *models.Configuration) {
	cm.configuration = config
}

// PurgeAll purges the entire nginx cache
func (cm *Manager) PurgeAll() error {
	// Method 1: Try nginx cache purge wildcard (if supported)
	if err := cm.purgeURLPattern("/*"); err == nil {
		return nil
	}

	// Method 2: Fallback to filesystem purge
	return cm.purgeFilesystem("")
}

// PurgeURL purges cache for a specific URL pattern
func (cm *Manager) PurgeURL(urlPattern string) error {
	// Ensure URL pattern starts with /
	if !strings.HasPrefix(urlPattern, "/") {
		urlPattern = "/" + urlPattern
	}

	return cm.purgeURLPattern(urlPattern)
}

// PurgeBackend purges cache for all routes associated with a specific backend
func (cm *Manager) PurgeBackend(backendID string) error {
	if cm.configuration == nil {
		return fmt.Errorf("configuration not available")
	}

	// Find all routes that use this backend
	var routePaths []string
	for _, server := range cm.configuration.Servers {
		for _, route := range server.Routes {
			if route.BackendID == backendID {
				routePaths = append(routePaths, route.Path+"*")
			}
		}
	}

	// Purge each route path
	var lastErr error
	for _, path := range routePaths {
		if err := cm.PurgeURL(path); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// PurgeZone purges a specific cache zone
func (cm *Manager) PurgeZone(zoneName string) error {
	// Check if zone exists
	zoneExists := false
	for _, zone := range cm.cacheZones {
		if zone == zoneName {
			zoneExists = true
			break
		}
	}

	if !zoneExists {
		return fmt.Errorf("cache zone '%s' not found", zoneName)
	}

	// Try zone-specific purge endpoint
	purgeURL := fmt.Sprintf("%s/api/v1/cache/purge/zone/%s", cm.nginxBaseURL, zoneName)
	resp, err := cm.client.Post(purgeURL, "application/json", nil)
	if err != nil {
		// Fallback to filesystem purge for this zone
		zonePath := filepath.Join(cm.cachePath, zoneName)
		return cm.purgeFilesystem(zonePath)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("nginx cache purge failed with status: %s", resp.Status)
	}

	return nil
}

// GetStats returns comprehensive cache statistics
func (cm *Manager) GetStats() (*CacheStats, error) {
	stats := &CacheStats{
		Zones:        make(map[string]*CacheZoneStats),
		BackendStats: make(map[string]*BackendCacheStats),
	}

	// Get filesystem statistics
	if err := cm.collectFilesystemStats(stats); err != nil {
		return nil, fmt.Errorf("failed to collect filesystem stats: %w", err)
	}

	// Get backend-specific statistics
	if err := cm.collectBackendStats(stats); err != nil {
		return nil, fmt.Errorf("failed to collect backend stats: %w", err)
	}

	return stats, nil
}

// WarmCache pre-populates cache by making requests to specified URLs
func (cm *Manager) WarmCache(urls []string) error {
	var lastErr error

	for _, urlStr := range urls {
		// Parse and validate URL
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			lastErr = fmt.Errorf("invalid URL %s: %w", urlStr, err)
			continue
		}

		// Make request to warm cache
		fullURL := fmt.Sprintf("%s%s", cm.nginxBaseURL, parsedURL.Path)
		if parsedURL.RawQuery != "" {
			fullURL += "?" + parsedURL.RawQuery
		}

		resp, err := cm.client.Get(fullURL)
		if err != nil {
			lastErr = fmt.Errorf("failed to warm cache for %s: %w", urlStr, err)
			continue
		}
		resp.Body.Close()
	}

	return lastErr
}

// Private helper methods

// purgeURLPattern purges cache using nginx cache purge module
func (cm *Manager) purgeURLPattern(pattern string) error {
	// Construct purge URL - nginx-cache-purge module endpoint
	purgeURL := fmt.Sprintf("%s/api/v1/cache/purge%s", cm.nginxBaseURL, pattern)

	resp, err := cm.client.Get(purgeURL)
	if err != nil {
		return fmt.Errorf("failed to connect to nginx purge endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("nginx cache purge failed with status %s: %s", resp.Status, string(body))
	}

	return nil
}

// purgeFilesystem removes cache files from filesystem (fallback method)
func (cm *Manager) purgeFilesystem(targetPath string) error {
	if targetPath == "" {
		targetPath = cm.cachePath
	}

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return nil // Nothing to purge
	}

	return os.RemoveAll(targetPath)
}

// collectFilesystemStats gathers cache statistics from filesystem
func (cm *Manager) collectFilesystemStats(stats *CacheStats) error {
	if _, err := os.Stat(cm.cachePath); os.IsNotExist(err) {
		return nil // Cache directory doesn't exist
	}

	return filepath.Walk(cm.cachePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		stats.FileCount++
		stats.TotalSize += info.Size()

		if info.ModTime().After(stats.LastModified) {
			stats.LastModified = info.ModTime()
		}

		// Collect zone-specific stats
		relPath, _ := filepath.Rel(cm.cachePath, path)
		pathParts := strings.Split(relPath, string(filepath.Separator))
		if len(pathParts) > 0 && pathParts[0] != "" {
			zoneName := pathParts[0]
			if _, exists := stats.Zones[zoneName]; !exists {
				stats.Zones[zoneName] = &CacheZoneStats{
					Name: zoneName,
				}
			}
			zoneStats := stats.Zones[zoneName]
			zoneStats.FileCount++
			zoneStats.Size += info.Size()
			if info.ModTime().After(zoneStats.LastAccessed) {
				zoneStats.LastAccessed = info.ModTime()
			}
		}

		return nil
	})
}

// collectBackendStats gathers backend-specific cache statistics
func (cm *Manager) collectBackendStats(stats *CacheStats) error {
	if cm.configuration == nil {
		return nil
	}

	// Initialize backend stats
	for _, backend := range cm.configuration.Backends {
		stats.BackendStats[backend.ID] = &BackendCacheStats{
			BackendID:   backend.ID,
			BackendName: backend.Name,
			CachedPaths: []string{},
		}
	}

	// TODO: Implement nginx cache inspection to get actual hit/miss statistics
	// This would require parsing nginx access logs or using nginx Plus API
	// For now, we'll provide filesystem-based approximation

	return nil
}

// GetCacheKey generates a cache key for a given URL (matches nginx cache key generation)
func (cm *Manager) GetCacheKey(scheme, method, host, uri string) string {
	// Default nginx cache key: $scheme$request_method$host$request_uri
	return scheme + method + host + uri
}

// IsHealthy checks if the cache subsystem is healthy
func (cm *Manager) IsHealthy() bool {
	// Check if cache directory is accessible
	if _, err := os.Stat(cm.cachePath); os.IsNotExist(err) {
		return false
	}

	// Check if nginx purge endpoint is accessible
	testURL := fmt.Sprintf("%s/api/v1/cache/purge/health", cm.nginxBaseURL)
	resp, err := cm.client.Get(testURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound
}
