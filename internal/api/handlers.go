package api

import (
	"net/http"

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
