package api

import (
	"net/http"
	"time"

	"go-apigateway-gui/internal/models"
	"go-apigateway-gui/internal/nginx"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, nginxMgr *nginx.Manager) {
	// Load configuration on startup
	if err := nginxMgr.LoadConfiguration(); err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	api := r.Group("/api/v1")
	{
		// Configuration endpoints
		api.GET("/config", getConfiguration(nginxMgr))
		api.POST("/config/generate", generateConfig(nginxMgr))
		api.POST("/config/test", testConfig(nginxMgr))
		api.POST("/config/reload", reloadConfig(nginxMgr))

		// Template endpoints
		api.GET("/templates", getTemplates(nginxMgr))
		api.POST("/templates/:name/generate", generateConfigWithTemplate(nginxMgr))
		api.GET("/templates/:name/preview", previewTemplate(nginxMgr))
		api.POST("/templates/:name/validate", validateTemplate(nginxMgr))

		// Backend endpoints
		api.GET("/backends", getBackends(nginxMgr))
		api.POST("/backends", createBackend(nginxMgr))
		api.GET("/backends/:id", getBackend(nginxMgr))
		api.PUT("/backends/:id", updateBackend(nginxMgr))
		api.DELETE("/backends/:id", deleteBackend(nginxMgr))

		// Server endpoints
		api.GET("/servers", getServers(nginxMgr))
		api.POST("/servers", createServer(nginxMgr))
		api.GET("/servers/:id", getServer(nginxMgr))
		api.PUT("/servers/:id", updateServer(nginxMgr))
		api.DELETE("/servers/:id", deleteServer(nginxMgr))

		// Route endpoints - moved to separate group to avoid conflicts
		routes := api.Group("/routes")
		{
			routes.GET("/server/:serverid", getRoutes(nginxMgr))
			routes.POST("/server/:serverid", createRoute(nginxMgr))
			routes.GET("/:id", getRoute(nginxMgr))
			routes.PUT("/:id", updateRoute(nginxMgr))
			routes.DELETE("/:id", deleteRoute(nginxMgr))
		}

		// Nginx cache management endpoints
		cache := api.Group("/cache")
		{
			cache.DELETE("/purge", purgeNginxCache(nginxMgr))
			cache.DELETE("/purge/url", purgeNginxCacheURL(nginxMgr))
			cache.DELETE("/purge/backend/:id", purgeNginxCacheBackend(nginxMgr))
			cache.DELETE("/purge/zone/:zone", purgeNginxCacheZone(nginxMgr))
			cache.GET("/stats", getNginxCacheStats(nginxMgr))
			cache.POST("/warm", warmNginxCache(nginxMgr))
		}

		// Certificate management endpoints
		certificates := api.Group("/certificates")
		{
			certificates.GET("", listCertificates(nginxMgr))
			certificates.POST("/obtain/:serverid", obtainCertificate(nginxMgr))
			certificates.POST("/renew/:serverid", renewCertificate(nginxMgr))
			certificates.GET("/:serverid", getCertificateInfo(nginxMgr))
			certificates.DELETE("/:serverid", removeCertificateConfig(nginxMgr))
			certificates.GET("/:serverid/status", getCertificateStatus(nginxMgr))
		}

		// OpenAPI aggregation endpoints
		openapi := api.Group("/openapi")
		{
			openapi.GET("/aggregated", getAggregatedOpenAPI(nginxMgr))
			openapi.POST("/refresh", refreshOpenAPICache(nginxMgr))
			openapi.GET("/backends", getOpenAPIBackends(nginxMgr))
			openapi.GET("/status", getOpenAPIStatus(nginxMgr))
			openapi.GET("/swagger-ui", serveSwaggerUI(nginxMgr))
		}

		// Status endpoints
		api.GET("/status", getStatus(nginxMgr))
	}
}

// Configuration handlers
func getConfiguration(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		config := nginxMgr.GetConfiguration()
		c.JSON(http.StatusOK, config)
	}
}

func generateConfig(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := nginxMgr.GenerateNginxConfig(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Configuration generated successfully"})
	}
}

func testConfig(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := nginxMgr.TestConfiguration(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Configuration test passed"})
	}
}

func reloadConfig(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := nginxMgr.ReloadNginx(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Nginx reloaded successfully"})
	}
}

func getTotalRoutesCount(servers []models.Server) int {
	total := 0
	for _, server := range servers {
		total += len(server.Routes)
	}
	return total
}

// Template handlers
func getTemplates(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		templates, err := nginxMgr.ListAvailableTemplates()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, templates)
	}
}

func generateConfigWithTemplate(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		templateName := c.Param("name")

		if err := nginxMgr.GenerateNginxConfigWithTemplate(templateName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Configuration generated successfully",
			"template": templateName,
		})
	}
}

func previewTemplate(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		templateName := c.Param("name")

		preview, err := nginxMgr.PreviewTemplate(templateName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"template": templateName,
			"preview":  preview,
		})
	}
}

func validateTemplate(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		templateName := c.Param("name")

		if err := nginxMgr.ValidateTemplate(templateName); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"valid":    false,
				"error":    err.Error(),
				"template": templateName,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"valid":    true,
			"message":  "Template is valid",
			"template": templateName,
		})
	}
}

// Backend handlers
func getBackends(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		config := nginxMgr.GetConfiguration()
		c.JSON(http.StatusOK, config.Backends)
	}
}

func createBackend(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var backend models.Backend
		if err := c.ShouldBindJSON(&backend); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := nginxMgr.AddBackend(backend); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, backend)
	}
}

func getBackend(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		backend, err := nginxMgr.GetBackend(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, backend)
	}
}

func updateBackend(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var backend models.Backend
		if err := c.ShouldBindJSON(&backend); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := nginxMgr.UpdateBackend(id, backend); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, backend)
	}
}

func deleteBackend(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := nginxMgr.DeleteBackend(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Backend deleted successfully"})
	}
}

// Server handlers
func getServers(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		config := nginxMgr.GetConfiguration()
		c.JSON(http.StatusOK, config.Servers)
	}
}

func createServer(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var server models.Server
		if err := c.ShouldBindJSON(&server); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := nginxMgr.AddServer(server); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, server)
	}
}

func getServer(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		server, err := nginxMgr.GetServer(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, server)
	}
}

func updateServer(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var server models.Server
		if err := c.ShouldBindJSON(&server); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := nginxMgr.UpdateServer(id, server); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, server)
	}
}

func deleteServer(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := nginxMgr.DeleteServer(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Server deleted successfully"})
	}
}

// Route handlers
func getRoutes(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverID := c.Param("serverid")
		server, err := nginxMgr.GetServer(serverID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, server.Routes)
	}
}

func createRoute(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverID := c.Param("serverid")
		var route models.Route
		if err := c.ShouldBindJSON(&route); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := nginxMgr.AddRoute(serverID, route); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, route)
	}
}

func getRoute(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		routeID := c.Param("id")

		// Search all servers for the route
		config := nginxMgr.GetConfiguration()
		for _, server := range config.Servers {
			for _, route := range server.Routes {
				if route.ID == routeID {
					c.JSON(http.StatusOK, route)
					return
				}
			}
		}

		c.JSON(http.StatusNotFound, gin.H{"error": "Route not found"})
	}
}

func updateRoute(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		routeID := c.Param("id")
		var route models.Route
		if err := c.ShouldBindJSON(&route); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := nginxMgr.UpdateRouteByID(routeID, route); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, route)
	}
}

func deleteRoute(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		routeID := c.Param("id")

		if err := nginxMgr.DeleteRouteByID(routeID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Route deleted successfully"})
	}
}

// Status handlers
func getStatus(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		status, err := nginxMgr.GetNginxStatus()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		config := nginxMgr.GetConfiguration()
		response := gin.H{
			"nginx_status":   status,
			"backends_count": len(config.Backends),
			"servers_count":  len(config.Servers),
			"routes_count":   getTotalRoutesCount(config.Servers),
		}

		c.JSON(http.StatusOK, response)
	}
}

// purgeNginxCacheHandler handles cache purge requests
func purgeNginxCache(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		cacheManager := nginxMgr.GetCacheManager()
		if err := cacheManager.PurgeAll(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "Nginx cache purged successfully",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// purgeNginxCacheURLHandler handles cache purge by URL
func purgeNginxCacheURL(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			URL string `json:"url" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cacheManager := nginxMgr.GetCacheManager()
		if err := cacheManager.PurgeURL(req.URL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "Cache purged for URL: " + req.URL,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// purgeNginxCacheBackendHandler handles cache purge by backend ID
func purgeNginxCacheBackend(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		backendID := c.Param("id")

		cacheManager := nginxMgr.GetCacheManager()
		if err := cacheManager.PurgeBackend(backendID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "Cache purged for backend ID: " + backendID,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// purgeNginxCacheZoneHandler handles cache purge by zone
func purgeNginxCacheZone(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		zone := c.Param("zone")

		cacheManager := nginxMgr.GetCacheManager()
		if err := cacheManager.PurgeZone(zone); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "Cache purged for zone: " + zone,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// getNginxCacheStatsHandler returns cache statistics
func getNginxCacheStats(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		cacheManager := nginxMgr.GetCacheManager()
		stats, err := cacheManager.GetStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, stats)
	}
}

// warmNginxCacheHandler warms up the cache with specified URLs
func warmNginxCache(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			URLs []string `json:"urls" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		cacheManager := nginxMgr.GetCacheManager()
		if err := cacheManager.WarmCache(req.URLs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "Cache warming initiated",
			"urls":      req.URLs,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

// Certificate Management Handlers

// Certificate request structure
type ObtainCertificateRequest struct {
	Email         string   `json:"email" binding:"required"`
	Domains       []string `json:"domains" binding:"required"`
	Provider      string   `json:"provider"`       // "letsencrypt", "buypass", "zerossl"
	Environment   string   `json:"environment"`    // "staging", "production"
	ChallengeType string   `json:"challenge_type"` // "http-01", "dns-01", "tls-alpn-01"
}

// Certificate information response
type CertificateResponse struct {
	ServerID      string     `json:"server_id"`
	Domains       []string   `json:"domains"`
	Provider      string     `json:"provider"`
	Environment   string     `json:"environment"`
	NeedsRenewal  bool       `json:"needs_renewal"`
	ChallengeType string     `json:"challenge_type"`
	IssuedAt      *time.Time `json:"issued_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
}

// listCertificates returns all managed certificates
func listCertificates(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		certificates, err := nginxMgr.ListManagedCertificates()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"certificates": certificates})
	}
}

// obtainCertificate requests a new certificate for a server
func obtainCertificate(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverID := c.Param("serverid")
		if serverID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Server ID is required"})
			return
		}

		var req ObtainCertificateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create ACME configuration
		acmeConfig := &models.ACMEConfig{
			Enabled:       true,
			Provider:      req.Provider,
			Environment:   req.Environment,
			Email:         req.Email,
			Domains:       req.Domains,
			ChallengeType: req.ChallengeType,
			RenewDays:     30,
			CheckInterval: "24h",
		}

		if err := nginxMgr.EnableACMEForServer(serverID, acmeConfig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Certificate obtained successfully"})
	}
}

// renewCertificate renews an existing certificate
func renewCertificate(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverID := c.Param("serverid")
		if serverID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Server ID is required"})
			return
		}

		certMgr := nginxMgr.GetCertificateManager()
		if err := certMgr.RenewCertificate(serverID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Certificate renewed successfully"})
	}
}

// getCertificateInfo returns information about a certificate
func getCertificateInfo(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverID := c.Param("serverid")
		if serverID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Server ID is required"})
			return
		}

		certInfo, err := nginxMgr.GetCertificateInfo(serverID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, certInfo)
	}
}

// removeCertificateConfig removes ACME configuration for a server
func removeCertificateConfig(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverID := c.Param("serverid")
		if serverID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Server ID is required"})
			return
		}

		if err := nginxMgr.DisableACMEForServer(serverID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Certificate configuration removed successfully"})
	}
}

// getCertificateStatus returns the status of a certificate
func getCertificateStatus(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		serverID := c.Param("serverid")
		if serverID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Server ID is required"})
			return
		}

		certMgr := nginxMgr.GetCertificateManager()
		certInfo, err := certMgr.GetCertificateInfo(serverID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		status := gin.H{
			"server_id":      certInfo.ServerID,
			"domains":        certInfo.Domains,
			"needs_renewal":  certInfo.NeedsRenewal,
			"provider":       certInfo.Provider,
			"environment":    certInfo.Environment,
			"challenge_type": certInfo.ChallengeType,
		}

		c.JSON(http.StatusOK, status)
	}
}

// OpenAPI handlers (temporarily disabled - not implemented)

// getAggregatedOpenAPI returns the aggregated OpenAPI specification from all enabled backends
func getAggregatedOpenAPI(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "OpenAPI aggregation not yet implemented"})
	}
}

// refreshOpenAPICache forces a refresh of all cached OpenAPI specifications
func refreshOpenAPICache(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "OpenAPI cache refresh not yet implemented"})
	}
}

// getOpenAPIBackends returns a list of backends with their OpenAPI configuration status
func getOpenAPIBackends(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "OpenAPI backends listing not yet implemented"})
	}
}

// getOpenAPIStatus returns the overall status of OpenAPI aggregation
func getOpenAPIStatus(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "OpenAPI status not yet implemented"})
	}
}

// serveSwaggerUI serves a standalone Swagger UI page for the aggregated OpenAPI specification
func serveSwaggerUI(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		swaggerHTML := `<!DOCTYPE html>
<html>
<head>
	<title>API Gateway - Swagger UI</title>
	<link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
	<style>
		html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
		*, *:before, *:after { box-sizing: inherit; }
		body { margin:0; background: #fafafa; }
		.swagger-ui .topbar { display: none; }
	</style>
</head>
<body>
	<div id="swagger-ui"></div>
	<script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
	<script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
	<script>
		window.onload = function() {
			SwaggerUIBundle({
				url: '/api/v1/openapi/aggregated',
				dom_id: '#swagger-ui',
				deepLinking: true,
				presets: [
					SwaggerUIBundle.presets.apis,
					SwaggerUIStandalonePreset
				],
				plugins: [
					SwaggerUIBundle.plugins.DownloadUrl
				],
				layout: "StandaloneLayout",
				validatorUrl: null,
				docExpansion: "list",
				filter: true,
				showRequestHeaders: true,
				showExtensions: true,
				showCommonExtensions: true
			});
		};
	</script>
</body>
</html>`

		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, swaggerHTML)
	}
}
