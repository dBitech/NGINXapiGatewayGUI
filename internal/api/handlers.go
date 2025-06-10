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

// OpenAPI handlers

// getAggregatedOpenAPI returns the aggregated OpenAPI specification from all enabled backends
func getAggregatedOpenAPI(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		aggregator := nginxMgr.GetOpenAPIAggregator()

		// Get aggregated spec
		aggregatedSpec, err := aggregator.GetAggregatedSpec()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to aggregate OpenAPI specifications",
				"details": err.Error(),
			})
			return
		}

		// Check if any specs were found
		if aggregatedSpec == nil || len(aggregatedSpec.Paths) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "No OpenAPI specifications found",
				"message": "Ensure backends have OpenAPI enabled and accessible",
			})
			return
		}

		// Set content type for OpenAPI spec
		c.Header("Content-Type", "application/json")
		c.JSON(http.StatusOK, aggregatedSpec)
	}
}

// refreshOpenAPICache forces a refresh of all cached OpenAPI specifications
func refreshOpenAPICache(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		aggregator := nginxMgr.GetOpenAPIAggregator()

		// Clear cache and force refresh
		err := aggregator.RefreshCache()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to refresh OpenAPI cache",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "OpenAPI cache refreshed successfully",
			"refreshed_at": time.Now().Format(time.RFC3339),
		})
	}
}

// getOpenAPIBackends returns a list of backends with their OpenAPI configuration status
func getOpenAPIBackends(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		config := nginxMgr.GetConfiguration()
		aggregator := nginxMgr.GetOpenAPIAggregator()

		type BackendStatus struct {
			models.Backend
			OpenAPIStatus string     `json:"openapi_status"`
			LastFetched   *time.Time `json:"last_fetched,omitempty"`
			Error         string     `json:"error,omitempty"`
		}

		var backendStatuses []BackendStatus

		for _, backend := range config.Backends {
			status := BackendStatus{
				Backend:       backend,
				OpenAPIStatus: "disabled",
			}

			if backend.OpenAPI != nil && backend.OpenAPI.Enabled {
				status.OpenAPIStatus = "enabled"

				// Get cached spec info
				if spec := aggregator.GetCachedSpec(backend.ID); spec != nil {
					status.LastFetched = &spec.FetchedAt
					if spec.Error != "" {
						status.Error = spec.Error
						status.OpenAPIStatus = "error"
					} else if spec.Raw != nil {
						status.OpenAPIStatus = "available"
					}
				}
			}

			backendStatuses = append(backendStatuses, status)
		}

		c.JSON(http.StatusOK, gin.H{
			"backends":         backendStatuses,
			"total_backends":   len(config.Backends),
			"enabled_backends": countEnabledOpenAPIBackends(config.Backends),
		})
	}
}

// getOpenAPIStatus returns the current status of the OpenAPI aggregation service
func getOpenAPIStatus(nginxMgr *nginx.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		config := nginxMgr.GetConfiguration()
		aggregator := nginxMgr.GetOpenAPIAggregator()

		// Count backends by status
		totalBackends := len(config.Backends)
		enabledBackends := countEnabledOpenAPIBackends(config.Backends)

		// Get aggregation status
		stats := aggregator.GetStats()

		response := gin.H{
			"service_status":     "running",
			"cache_ttl_minutes":  int(aggregator.GetCacheTTL().Minutes()),
			"last_refresh":       stats["LastRefresh"],
			"total_backends":     totalBackends,
			"enabled_backends":   enabledBackends,
			"cached_specs":       stats["CachedSpecs"],
			"successful_fetches": stats["SuccessfulFetches"],
			"failed_fetches":     stats["FailedFetches"],
		}

		c.JSON(http.StatusOK, response)
	}
}

// Helper functions

func countEnabledOpenAPIBackends(backends []models.Backend) int {
	count := 0
	for _, backend := range backends {
		if backend.OpenAPI != nil && backend.OpenAPI.Enabled {
			count++
		}
	}
	return count
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
