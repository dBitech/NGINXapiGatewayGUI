package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"go-apigateway-gui/internal/api"
	"go-apigateway-gui/internal/config"
	"go-apigateway-gui/internal/nginx"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var (
	configFile string
	port       string
	verbose    bool
	version    = "1.0.0" // Version will be set during build
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "apigateway-gui",
	Short: "A web-based GUI for managing nginx API gateway configurations",
	Long: `A comprehensive web-based interface for managing nginx configurations 
as an API gateway with support for load balancing, SSL termination, 
rate limiting, caching, and more.`,
	Run: runServer,
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Nginx API Gateway GUI v%s\n", version)
	},
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
}

// configValidateCmd validates the configuration
var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Configuration validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Configuration is valid!")
		fmt.Printf("Config file: %s\n", cfg.GetConfigPath())
		fmt.Printf("Port: %s\n", cfg.Port)
		fmt.Printf("Nginx config: %s\n", cfg.NginxConfigPath)
	},
}

// configShowCmd shows the current configuration
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Current Configuration:")
		fmt.Printf("  Port: %s\n", cfg.Port)
		fmt.Printf("  Log Level: %s\n", cfg.LogLevel)
		fmt.Printf("  Nginx Config Path: %s\n", cfg.NginxConfigPath)
		fmt.Printf("  Nginx Executable: %s\n", cfg.NginxExecutablePath)
		fmt.Printf("  API Gateway Config: %s\n", cfg.GetConfigPath())
		fmt.Printf("  Backup Path: %s\n", cfg.BackupPath)
		fmt.Printf("  Web Assets Path: %s\n", cfg.WebAssetsPath)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Root command flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default searches for config.yaml in current dir, ./config, ~/.apigateway, /etc/apigateway)")
	rootCmd.PersistentFlags().StringVar(&port, "port", "", "port to run server on (overrides config file)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Add subcommands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)

	// Add config subcommands
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configShowCmd)
}

func runServer(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override port if specified via flag
	if port != "" {
		cfg.Port = port
	}

	// Set log level
	if verbose {
		cfg.LogLevel = "debug"
	}

	// Set Gin mode based on log level
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize nginx manager with the resolved config path
	nginxMgr := nginx.NewManager(cfg.NginxConfigPath, cfg.NginxExecutablePath, cfg.TemplatesPath, cfg.GetConfigPath())

	// Setup Gin router
	r := gin.Default()

	// Serve static files
	r.Static("/static", cfg.WebAssetsPath+"/static")
	r.LoadHTMLGlob(cfg.WebAssetsPath + "/templates/*")

	// Setup API routes
	api.SetupRoutes(r, nginxMgr)

	// Setup web routes
	setupWebRoutes(r)

	log.Printf("Starting Nginx API Gateway GUI")
	log.Printf("Server listening on port %s", cfg.Port)
	log.Printf("Nginx config path: %s", cfg.NginxConfigPath)
	log.Printf("API Gateway config path: %s", cfg.GetConfigPath())
	log.Printf("Web interface: http://localhost:%s", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func setupWebRoutes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Nginx API Gateway Manager",
		})
	})
}
