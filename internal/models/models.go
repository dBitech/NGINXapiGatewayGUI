package models

import "time"

// Backend represents an upstream server
type Backend struct {
	ID          string            `json:"id" yaml:"id"`
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Servers     []BackendServer   `json:"servers" yaml:"servers"`
	HealthCheck *HealthCheck      `json:"health_check,omitempty" yaml:"health_check,omitempty"`
	OpenAPI     *OpenAPIConfig    `json:"openapi,omitempty" yaml:"openapi,omitempty"`
	LoadBalance string            `json:"load_balance" yaml:"load_balance"` // round_robin, least_conn, ip_hash
	MaxFails    int               `json:"max_fails" yaml:"max_fails"`
	FailTimeout string            `json:"fail_timeout" yaml:"fail_timeout"`
	Headers     map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
}

// BackendServer represents a single server in a backend
type BackendServer struct {
	ID     string `json:"id" yaml:"id"`
	Host   string `json:"host" yaml:"host"`
	Port   int    `json:"port" yaml:"port"`
	Weight int    `json:"weight" yaml:"weight"`
	Backup bool   `json:"backup" yaml:"backup"`
	Down   bool   `json:"down" yaml:"down"`
}

// HealthCheck configuration for backend health monitoring
type HealthCheck struct {
	Enabled  bool              `json:"enabled" yaml:"enabled"`
	Path     string            `json:"path" yaml:"path"`                           // Health check endpoint (e.g., "/health", "/status")
	Method   string            `json:"method" yaml:"method"`                       // HTTP method: GET, POST, HEAD (default: GET)
	Headers  map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"` // Custom headers for health check
	Body     string            `json:"body,omitempty" yaml:"body,omitempty"`       // Request body for POST/PUT
	Interval string            `json:"interval" yaml:"interval"`                   // Check interval (e.g., "10s", "30s")
	Timeout  string            `json:"timeout" yaml:"timeout"`                     // Request timeout (e.g., "5s")
	Rises    int               `json:"rises" yaml:"rises"`                         // Successes to mark healthy (default: 2)
	Falls    int               `json:"falls" yaml:"falls"`                         // Failures to mark unhealthy (default: 3)
	Match    *HealthCheckMatch `json:"match,omitempty" yaml:"match,omitempty"`     // Response validation
}

// HealthCheckMatch defines what constitutes a healthy response
type HealthCheckMatch struct {
	StatusCodes []int  `json:"status_codes,omitempty" yaml:"status_codes,omitempty"` // Expected HTTP status codes (default: 200-299)
	BodyRegex   string `json:"body_regex,omitempty" yaml:"body_regex,omitempty"`     // Regex pattern to match response body
	HeaderCheck string `json:"header_check,omitempty" yaml:"header_check,omitempty"` // Header validation
}

// RewriteRule represents a URL rewrite configuration
type RewriteRule struct {
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	Pattern     string `json:"pattern" yaml:"pattern"`         // regex pattern to match
	Replacement string `json:"replacement" yaml:"replacement"` // replacement string with $1, $2 etc for captures
	Flag        string `json:"flag" yaml:"flag"`               // nginx rewrite flags: last, break, redirect, permanent
	BreakChain  bool   `json:"break_chain" yaml:"break_chain"` // stop processing more rewrites
}

// RegexConfig for advanced path matching
type RegexConfig struct {
	Enabled       bool   `json:"enabled" yaml:"enabled"`
	Pattern       string `json:"pattern" yaml:"pattern"`               // regex pattern for location matching
	CaseSensitive bool   `json:"case_sensitive" yaml:"case_sensitive"` // case sensitive matching
	Modifier      string `json:"modifier" yaml:"modifier"`             // nginx location modifier: ~, ~*, ^~
}

// Route represents a frontend route configuration
type Route struct {
	ID             string            `json:"id" yaml:"id"`
	Name           string            `json:"name" yaml:"name"`
	Description    string            `json:"description" yaml:"description"`
	Path           string            `json:"path" yaml:"path"`
	RegexConfig    *RegexConfig      `json:"regex_config,omitempty" yaml:"regex_config,omitempty"`   // regex pattern matching
	RewriteRules   []RewriteRule     `json:"rewrite_rules,omitempty" yaml:"rewrite_rules,omitempty"` // URL rewriting rules
	Methods        []string          `json:"methods" yaml:"methods"`                                 // GET, POST, PUT, DELETE, etc.
	BackendID      string            `json:"backend_id" yaml:"backend_id"`
	StripPath      bool              `json:"strip_path" yaml:"strip_path"`
	PreserveHost   bool              `json:"preserve_host" yaml:"preserve_host"`
	Timeout        string            `json:"timeout" yaml:"timeout"`
	Headers        map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	RateLimiting   *RateLimit        `json:"rate_limiting,omitempty" yaml:"rate_limiting,omitempty"`
	Authentication *Auth             `json:"authentication,omitempty" yaml:"authentication,omitempty"`
	Caching        *CacheConfig      `json:"caching,omitempty" yaml:"caching,omitempty"`
	SSL            *SSLConfig        `json:"ssl,omitempty" yaml:"ssl,omitempty"`
	CORS           *CORSConfig       `json:"cors,omitempty" yaml:"cors,omitempty"`
	CreatedAt      time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at" yaml:"updated_at"`
}

// RateLimit configuration
type RateLimit struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	Rate       string `json:"rate" yaml:"rate"`               // e.g., "10r/s", "100r/m"
	Burst      int    `json:"burst" yaml:"burst"`             // burst size
	Nodelay    bool   `json:"nodelay" yaml:"nodelay"`         // nodelay option
	StatusCode int    `json:"status_code" yaml:"status_code"` // HTTP status for rate limit exceeded
}

// Auth configuration
type Auth struct {
	Type     string            `json:"type" yaml:"type"` // basic, jwt, api_key
	Config   map[string]string `json:"config" yaml:"config"`
	Required bool              `json:"required" yaml:"required"`
}

// CacheConfig for response caching with advanced TTL controls
type CacheConfig struct {
	Enabled       bool             `json:"enabled" yaml:"enabled"`
	TTL           string           `json:"ttl" yaml:"ttl"`                                 // Default TTL (e.g., "1h", "30m", "600s")
	Key           string           `json:"key" yaml:"key"`                                 // Cache key pattern
	Methods       []string         `json:"methods" yaml:"methods"`                         // Cacheable HTTP methods
	StatusCodes   []int            `json:"status_codes" yaml:"status_codes"`               // Cacheable status codes
	IgnoreHeaders []string         `json:"ignore_headers" yaml:"ignore_headers"`           // Headers to ignore in cache key
	URIRules      []CacheURIRule   `json:"uri_rules,omitempty" yaml:"uri_rules,omitempty"` // Per-URI caching rules
	Bypass        []string         `json:"bypass,omitempty" yaml:"bypass,omitempty"`       // Conditions to bypass cache
	Valid         []CacheValidRule `json:"valid,omitempty" yaml:"valid,omitempty"`         // Cache validity rules
	MinUses       int              `json:"min_uses,omitempty" yaml:"min_uses,omitempty"`   // Minimum uses before caching
	Inactive      string           `json:"inactive,omitempty" yaml:"inactive,omitempty"`   // Inactive time before removal
	MaxSize       string           `json:"max_size,omitempty" yaml:"max_size,omitempty"`   // Maximum cache size
}

// CacheURIRule defines caching behavior for specific URI patterns
type CacheURIRule struct {
	Pattern     string   `json:"pattern" yaml:"pattern"`                               // URI pattern (regex or exact match)
	Type        string   `json:"type" yaml:"type"`                                     // "exact", "prefix", "regex"
	TTL         string   `json:"ttl" yaml:"ttl"`                                       // TTL for this pattern
	Methods     []string `json:"methods,omitempty" yaml:"methods,omitempty"`           // HTTP methods for this rule
	StatusCodes []int    `json:"status_codes,omitempty" yaml:"status_codes,omitempty"` // Status codes for this rule
	NoCache     bool     `json:"no_cache,omitempty" yaml:"no_cache,omitempty"`         // Disable caching for this pattern
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`   // Rule description
}

// CacheValidRule defines cache validity conditions
type CacheValidRule struct {
	StatusCodes []int  `json:"status_codes" yaml:"status_codes"` // Status codes
	TTL         string `json:"ttl" yaml:"ttl"`                   // TTL for these status codes
}

// SSLConfig for SSL/TLS configuration
type SSLConfig struct {
	Enabled     bool     `json:"enabled" yaml:"enabled"`
	Certificate string   `json:"certificate" yaml:"certificate"`
	PrivateKey  string   `json:"private_key" yaml:"private_key"`
	Protocols   []string `json:"protocols" yaml:"protocols"`
	Ciphers     string   `json:"ciphers" yaml:"ciphers"`
	Redirect    bool     `json:"redirect" yaml:"redirect"` // redirect HTTP to HTTPS
}

// CORSConfig for Cross-Origin Resource Sharing
type CORSConfig struct {
	Enabled          bool     `json:"enabled" yaml:"enabled"`
	AllowOrigins     []string `json:"allow_origins" yaml:"allow_origins"`
	AllowMethods     []string `json:"allow_methods" yaml:"allow_methods"`
	AllowHeaders     []string `json:"allow_headers" yaml:"allow_headers"`
	ExposeHeaders    []string `json:"expose_headers" yaml:"expose_headers"`
	AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials"`
	MaxAge           int      `json:"max_age" yaml:"max_age"`
}

// Server represents the main server configuration
type Server struct {
	ID                string            `json:"id" yaml:"id"`
	Name              string            `json:"name" yaml:"name"`
	Listen            []string          `json:"listen" yaml:"listen"` // ports and interfaces
	ServerName        []string          `json:"server_name" yaml:"server_name"`
	Routes            []Route           `json:"routes" yaml:"routes"`
	SSL               *SSLConfig        `json:"ssl,omitempty" yaml:"ssl,omitempty"`
	AccessLog         string            `json:"access_log" yaml:"access_log"`
	ErrorLog          string            `json:"error_log" yaml:"error_log"`
	ClientMaxBodySize string            `json:"client_max_body_size" yaml:"client_max_body_size"`
	Headers           map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	CreatedAt         time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" yaml:"updated_at"`
}

// Configuration represents the complete nginx configuration
type Configuration struct {
	Backends []Backend    `json:"backends" yaml:"backends"`
	Servers  []Server     `json:"servers" yaml:"servers"`
	Global   GlobalConfig `json:"global" yaml:"global"`
}

// GlobalConfig represents global nginx settings
type GlobalConfig struct {
	WorkerProcesses   string   `json:"worker_processes" yaml:"worker_processes"`
	WorkerConnections int      `json:"worker_connections" yaml:"worker_connections"`
	KeepAliveTimeout  string   `json:"keepalive_timeout" yaml:"keepalive_timeout"`
	ClientMaxBodySize string   `json:"client_max_body_size" yaml:"client_max_body_size"`
	ServerTokens      bool     `json:"server_tokens" yaml:"server_tokens"`
	Gzip              bool     `json:"gzip" yaml:"gzip"`
	GzipTypes         []string `json:"gzip_types" yaml:"gzip_types"`
}

// OpenAPIConfig represents OpenAPI/Swagger configuration for a backend
type OpenAPIConfig struct {
	Enabled     bool     `json:"enabled" yaml:"enabled"`
	Endpoints   []string `json:"endpoints" yaml:"endpoints"`                         // e.g., ["/swagger.json", "/openapi.yaml"]
	BasePath    string   `json:"base_path,omitempty" yaml:"base_path,omitempty"`     // Override base path
	Title       string   `json:"title,omitempty" yaml:"title,omitempty"`             // Override service title
	Description string   `json:"description,omitempty" yaml:"description,omitempty"` // Override description
	Version     string   `json:"version,omitempty" yaml:"version,omitempty"`         // Override version
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`               // Custom tags to apply
}

// OpenAPISpec represents a fetched and parsed OpenAPI specification
type OpenAPISpec struct {
	BackendID  string                 `json:"backend_id"`
	URL        string                 `json:"url"`
	Version    string                 `json:"version"`         // "2.0" or "3.0.x"
	Raw        map[string]interface{} `json:"raw"`             // Raw OpenAPI document
	Paths      map[string]interface{} `json:"paths"`           // Extracted paths
	Info       map[string]interface{} `json:"info"`            // API info
	Components map[string]interface{} `json:"components"`      // Components/definitions
	Tags       []interface{}          `json:"tags"`            // API tags
	Servers    []interface{}          `json:"servers"`         // Server definitions
	FetchedAt  time.Time              `json:"fetched_at"`      // When this spec was fetched
	Error      string                 `json:"error,omitempty"` // Error message if fetch failed
}

// AggregatedOpenAPISpec represents the final unified OpenAPI specification
type AggregatedOpenAPISpec struct {
	OpenAPI     string                 `json:"openapi"`
	Info        map[string]interface{} `json:"info"`
	Servers     []interface{}          `json:"servers"`
	Paths       map[string]interface{} `json:"paths"`
	Components  map[string]interface{} `json:"components,omitempty"`
	Tags        []interface{}          `json:"tags,omitempty"`
	GeneratedAt time.Time              `json:"generated_at"`
	Backends    []string               `json:"backends"` // List of backend IDs included
}
