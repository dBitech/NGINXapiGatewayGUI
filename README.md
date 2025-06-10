# Nginx Configuration Manager

A web-based GUI for managing nginx configurations with a focus on API gateway functionality. This application provides an intuitive interface for creating backends, servers, and routes while generating production-ready nginx configuration files.

**Note this overview & all of the provided documentation were editeid by Generative AI models;  Thier content may not be completly accurate, and sources may not be properly cited. 

## Features

### Core Functionality
- **Backend Management**: Create and manage upstream server pools with load balancing
- **Server Configuration**: Configure virtual hosts with SSL, logging, and custom settings
- **Route Management**: Define URI paths with advanced features like rate limiting, caching, and CORS
- **Real-time Configuration**: Generate, test, and reload nginx configurations on-the-fly

### Advanced Features
- **Load Balancing**: Round-robin, least connections, and IP hash algorithms
- **Rate Limiting**: Configurable request rate limits per route
- **Response Caching**: Intelligent caching with custom TTL and cache keys
- **Cache Management**: Advanced cache invalidation and warming capabilities
- **CORS Support**: Cross-origin resource sharing configuration
- **SSL/TLS**: SSL certificate management and HTTPS redirection
- **Health Monitoring**: Backend server health checks (nginx-plus required)
- **Authentication**: Basic auth, JWT, and API key authentication options
- **Regex Pattern Matching**: Advanced location matching using regular expressions with nginx modifiers
- **URL Rewriting**: Multiple rewrite rules per route with capture groups and various flags
- **OpenAPI Aggregation**: Dynamic Swagger/OpenAPI specification aggregation from backend services
- **Unified API Documentation**: Centralized Swagger UI displaying all backend APIs in one interface
- **API Discovery**: Automatic detection and aggregation of OpenAPI specifications from multiple endpoints

### Cache Management Features
- **Cache Invalidation**: Purge entire cache, specific URL patterns, or backend-specific caches
- **Cache Zones**: Support for multiple nginx cache zones with zone-specific purging
- **Cache Statistics**: Real-time cache metrics including size, file count, and hit/miss ratios
- **Cache Warming**: Pre-populate cache with frequently accessed URLs
- **Backend Cache Integration**: Automatically purge cache when backend configurations change
- **nginx-cache-purge Module**: Native integration with nginx cache purge module for optimal performance
- **Web UI Integration**: Intuitive cache management interface with one-click purge operations
- **API Endpoints**: RESTful API for programmatic cache management and automation

### Template System
- **External Jinja2 Templates**: Use external Jinja2 templates instead of hardcoded configurations
- **Multiple Template Options**: Choose from Standard, Minimal, or Development configurations
- **Template Management**: List, validate, preview, and generate configurations with specific templates
- **Custom Templates**: Create and use custom templates for specialized configurations
- **Template API**: Full REST API for template operations and management

### GUI Features
- **Dashboard**: Overview of system status and configuration metrics
- **Responsive Design**: Works on desktop and mobile devices
- **Live Status**: Real-time nginx process monitoring
- **Configuration Backup**: Automatic backup before changes
- **Validation**: Built-in nginx configuration testing

## Quick Start

### Prerequisites
- Go 1.21 or later
- nginx installed and accessible
- Write permissions to nginx configuration directory

### Installation

1. Clone the repository:
```bash
git clone https://github.com/dBiTech/go-apigateway-gui.git
cd go-apigateway-gui
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o apigateway-gui .
```

4. Create configuration file (optional):
```bash
# Copy the example configuration for your platform
cp config.yaml my-config.yaml          # Windows
# or
cp config-unix.yaml my-config.yaml     # Linux/Unix

# Edit the configuration as needed
```

5. Run the application:
```bash
# Using default configuration
./apigateway-gui

# Using custom configuration file
./apigateway-gui --config my-config.yaml

# Using environment variables
APIGATEWAY_PORT=9000 ./apigateway-gui

# With custom port via command line
./apigateway-gui --port 8080
```

## Configuration

The application supports flexible configuration through multiple methods:

### Configuration File

Create a `config.yaml` file in your working directory or use the `--config` flag:

```yaml
# Server configuration
port: "8080"
log_level: "info"

# Nginx paths
nginx_config_path: "/etc/nginx/nginx.conf"
nginx_executable_path: "nginx"

# Application settings
apigateway_config_path: "./apigateway-config.yaml"
backup_path: "./backups"
web_assets_path: "./web"
```

### Environment Variables

Override any configuration setting using the `APIGATEWAY_` prefix:

```bash
export APIGATEWAY_PORT=9000
export APIGATEWAY_NGINX_CONFIG_PATH="/usr/local/nginx/conf/nginx.conf"
export APIGATEWAY_LOG_LEVEL="debug"
```

### Command Line Flags

```bash
# View all available options
./apigateway-gui --help

# Common options
./apigateway-gui --port 8080 --verbose
./apigateway-gui --config /path/to/config.yaml
```

### Configuration Precedence

Configuration is loaded in the following order (later values override earlier ones):
1. Default values
2. Configuration file
3. Environment variables
4. Command line flags

## CLI Commands

### Server Management
```bash
# Start the web server (default command)
./apigateway-gui

# Start with custom port
./apigateway-gui --port 9000
```

### Configuration Management
```bash
# Validate configuration
./apigateway-gui config validate

# Show current configuration
./apigateway-gui config show

# Validate specific config file
./apigateway-gui config validate --config ./my-config.yaml
```

### Utility Commands
```bash
# Show version
./apigateway-gui version

# Show help
./apigateway-gui help
```
export NGINX_EXECUTABLE="nginx"
export PORT="8080"
export BACKUP_PATH="./backups"
```

4. Run the application:
```bash
go run main.go
```

5. Open your browser and navigate to `http://localhost:8080`

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Web server port |
| `NGINX_CONFIG_PATH` | `/etc/nginx/nginx.conf` | Path to nginx configuration file |
| `NGINX_EXECUTABLE` | `nginx` | nginx executable path or command |
| `BACKUP_PATH` | `./backups` | Directory for configuration backups |

### Windows Setup

For Windows users, you may need to adjust the nginx executable path:

```powershell
$env:NGINX_EXECUTABLE = "C:\nginx\nginx.exe"
$env:NGINX_CONFIG_PATH = "C:\nginx\conf\nginx.conf"
```

## Usage Guide

### 1. Creating Backends

Backends represent upstream server pools that handle your actual application traffic.

1. Navigate to the "Backends" section
2. Click "Add Backend"
3. Configure:
   - **Name**: Unique identifier for the backend
   - **Load Balance Method**: Choose from round-robin, least connections, or IP hash
   - **Servers**: Add one or more backend servers with host, port, and weight
   - **Health Checks**: Configure health monitoring (nginx-plus required)

### 2. Configuring Servers

Servers define virtual hosts that listen for incoming requests.

1. Navigate to the "Servers" section
2. Click "Add Server"
3. Configure:
   - **Listen Ports**: Specify ports and interfaces (e.g., "80", "443 ssl")
   - **Server Names**: Domain names for this virtual host
   - **SSL**: Configure SSL certificates and HTTPS redirection
   - **Logging**: Set custom access and error log paths

### 3. Creating Routes

Routes define how incoming requests are matched and forwarded to backends.

1. Navigate to the "Routes" section
2. Select a server from the dropdown
3. Click "Add Route"
4. Configure basic settings:
   - **Path**: URI pattern to match (e.g., "/api/v1", "/admin")
   - **Methods**: HTTP methods to allow (GET, POST, PUT, DELETE)
   - **Backend**: Target backend for this route
   - **Path Options**: Strip path prefix, preserve host header

5. Configure advanced features:
   - **Rate Limiting**: Protect against abuse with request rate limits
   - **Caching**: Cache responses to improve performance
   - **CORS**: Enable cross-origin requests with custom headers
   - **SSL**: Route-specific SSL settings
   - **Regex Matching**: Use regular expressions for advanced path matching
   - **URL Rewriting**: Define rewrite rules for request modification

### 4. Configuration Management

Use the "Configuration" section to:
- Adjust global nginx settings
- Generate nginx configuration files
- Test configuration syntax
- Reload nginx with new settings

### 5. OpenAPI Integration

Use the "OpenAPI" section to:
- View aggregated API documentation status
- Configure OpenAPI endpoints for backends
- Access unified Swagger UI interface
- Manage authentication for protected OpenAPI endpoints
- Monitor OpenAPI cache and refresh data

## OpenAPI Integration

The Nginx Configuration Manager now includes powerful OpenAPI/Swagger aggregation capabilities, allowing you to create a unified API documentation interface from multiple backend services.

### OpenAPI Features

- **Dynamic Aggregation**: Automatically fetches and combines OpenAPI specifications from multiple backend services
- **Unified Documentation**: Single Swagger UI interface displaying all backend APIs
- **Authentication Support**: Bearer tokens, Basic auth, and API key authentication for protected OpenAPI endpoints
- **Cache Management**: Intelligent caching with manual refresh capabilities
- **Real-time Status**: Monitor OpenAPI endpoint health and last update times
- **Path Transformation**: Automatic path prefixing and conflict resolution

### Configuration

Add OpenAPI configuration to your backends:

```yaml
backends:
  - id: petstore-api
    name: "Petstore API"
    servers:
      - host: "petstore.swagger.io"
        port: 443
    openapi:
      enabled: true
      endpoints:
        - url: "https://petstore.swagger.io/v2/swagger.json"
          auth:
            type: "none"
```

### Web Interface

The OpenAPI section in the web interface provides:

1. **Dashboard**: Overview of OpenAPI status and metrics
2. **Backend Management**: Configure OpenAPI endpoints per backend
3. **Swagger UI**: Unified documentation interface at `/api/v1/openapi/swagger-ui`
4. **Cache Control**: Manual refresh and status monitoring

### API Endpoints

```bash
# Get aggregated OpenAPI specification
GET /api/v1/openapi/spec

# Refresh OpenAPI cache
POST /api/v1/openapi/refresh

# Get OpenAPI service status
GET /api/v1/openapi/status

# Access Swagger UI
GET /api/v1/openapi/swagger-ui
```

For detailed OpenAPI configuration examples and advanced features, see `OPENAPI_FEATURES.md`.

## Advanced Features: Regex Pattern Matching and URL Rewriting

The Nginx Configuration Manager now supports advanced routing capabilities through regex pattern matching and URL rewriting. These features enable sophisticated traffic routing scenarios.

### Regex Pattern Matching

Regex patterns allow you to match complex URL structures that go beyond simple prefix matching.

#### Location Modifiers

- `~` - Case-sensitive regex matching
- `~*` - Case-insensitive regex matching  
- `^~` - Prefix matching with priority over regex
- `=` - Exact match

#### Example: API Versioning

```json
{
  "regex_config": {
    "enabled": true,
    "pattern": "^/api/v([0-9]+)/users/([0-9]+)$",
    "case_sensitive": false,
    "modifier": "~"
  }
}
```

This matches URLs like `/api/v1/users/123`, `/api/v2/users/456`, etc.

### URL Rewriting

URL rewriting allows you to transform incoming request URLs before they reach your backend services.

#### Rewrite Flags

- `last` - Stop processing current rewrite set, search for new location
- `break` - Stop processing in current location block
- `redirect` - Return HTTP 302 redirect
- `permanent` - Return HTTP 301 redirect

#### Example: Profile URL Rewriting

```json
{
  "rewrite_rules": [
    {
      "enabled": true,
      "pattern": "^/profile/([0-9]+)$",
      "replacement": "/users/$1/profile",
      "flag": "break"
    }
  ]
}
```

This rewrites `/profile/123` to `/users/123/profile` internally.

### Complex Example: Multi-Resource API

For handling multiple resource types with different transformations:

```json
{
  "name": "Multi-Resource API",
  "regex_config": {
    "enabled": true,
    "pattern": "^/api/v([0-9]+)/(users|posts|comments)/([0-9]+)/actions/([a-z]+)$",
    "case_sensitive": true,
    "modifier": "~*"
  },
  "rewrite_rules": [
    {
      "enabled": true,
      "pattern": "^/api/v([0-9]+)/users/([0-9]+)/actions/([a-z]+)$",
      "replacement": "/v$1/user/$2/action/$3",
      "flag": "last"
    },
    {
      "enabled": true,
      "pattern": "^/api/v([0-9]+)/posts/([0-9]+)/actions/([a-z]+)$",
      "replacement": "/v$1/post/$2/action/$3",
      "flag": "break"
    }
  ]
}
```

See `REGEX_REWRITING_EXAMPLES.md` for more detailed examples and use cases.

## API Reference

The application provides a REST API for programmatic configuration management.

### Backends

```bash
# List all backends
GET /api/v1/backends

# Create a new backend
POST /api/v1/backends
{
  "name": "api-backend",
  "description": "Main API servers",
  "load_balance": "round_robin",
  "servers": [
    {"host": "10.0.1.10", "port": 8080, "weight": 1},
    {"host": "10.0.1.11", "port": 8080, "weight": 1}
  ]
}

# Get a specific backend
GET /api/v1/backends/{id}

# Update a backend
PUT /api/v1/backends/{id}

# Delete a backend
DELETE /api/v1/backends/{id}
```

### Servers

```bash
# List all servers
GET /api/v1/servers

# Create a new server
POST /api/v1/servers
{
  "name": "api-server",
  "listen": ["80", "443 ssl"],
  "server_name": ["api.example.com", "www.api.example.com"]
}
```

### Routes

```bash
# List routes for a server
GET /api/v1/servers/{server_id}/routes

# Create a new route
POST /api/v1/servers/{server_id}/routes
{
  "name": "API v1 routes",
  "path": "/api/v1",
  "methods": ["GET", "POST"],
  "backend_id": "backend_123",
  "rate_limiting": {
    "enabled": true,
    "rate": "100r/m",
    "burst": 20
  }
}
```

### Configuration Management

```bash
# Get current configuration
GET /api/v1/config

# Generate nginx configuration
POST /api/v1/config/generate

# Test configuration
POST /api/v1/config/test

# Reload nginx
POST /api/v1/config/reload

# Get system status
GET /api/v1/status
```

### OpenAPI Management

```bash
# Get aggregated OpenAPI specification
GET /api/v1/openapi/spec

# Refresh OpenAPI cache from all backends
POST /api/v1/openapi/refresh

# Get OpenAPI service status and metrics
GET /api/v1/openapi/status

# Access unified Swagger UI interface
GET /api/v1/openapi/swagger-ui

# Get OpenAPI status for specific backend
GET /api/v1/backends/{id}/openapi/status
```

## Configuration Examples

### Example 1: Simple API Gateway

```yaml
backends:
  - id: api-backend
    name: API Backend
    load_balance: round_robin
    servers:
      - host: 10.0.1.10
        port: 8080
        weight: 1
      - host: 10.0.1.11
        port: 8080
        weight: 1

servers:
  - id: api-server
    name: API Server
    listen: ["80", "443 ssl"]
    server_name: ["api.example.com"]
    routes:
      - id: api-v1
        name: API v1
        path: "/api/v1"
        methods: ["GET", "POST", "PUT", "DELETE"]
        backend_id: api-backend
        rate_limiting:
          enabled: true
          rate: "100r/m"
          burst: 20
        cors:
          enabled: true
          allow_origins: ["https://app.example.com"]
          allow_methods: ["GET", "POST", "PUT", "DELETE"]
```

### Example 2: Multi-Service Gateway with SSL

```yaml
backends:
  - id: auth-service
    name: Authentication Service
    servers:
      - host: auth.internal
        port: 3000
  
  - id: user-service
    name: User Service
    servers:
      - host: users.internal
        port: 3001
  
  - id: order-service
    name: Order Service
    load_balance: least_conn
    servers:
      - host: orders1.internal
        port: 3002
      - host: orders2.internal
        port: 3002

servers:
  - id: main-gateway
    name: Main Gateway
    listen: ["80", "443 ssl"]
    server_name: ["gateway.example.com"]
    ssl:
      enabled: true
      certificate: "/etc/ssl/certs/gateway.crt"
      private_key: "/etc/ssl/private/gateway.key"
      redirect: true
    routes:
      - name: Authentication
        path: "/auth"
        methods: ["POST"]
        backend_id: auth-service
        strip_path: true
        rate_limiting:
          enabled: true
          rate: "10r/s"
          burst: 5
      
      - name: User API
        path: "/users"
        methods: ["GET", "POST", "PUT", "DELETE"]
        backend_id: user-service
        strip_path: true
        caching:
          enabled: true
          ttl: "5m"
          methods: ["GET"]
      
      - name: Orders API
        path: "/orders"
        methods: ["GET", "POST", "PUT", "DELETE"]
        backend_id: order-service
        strip_path: true
        authentication:
          type: "basic"
          required: true
```

### Example 3: API Gateway with OpenAPI Aggregation

```yaml
backends:
  - id: petstore-api
    name: Petstore API
    servers:
      - host: petstore.swagger.io
        port: 443
    openapi:
      enabled: true
      endpoints:
        - url: "https://petstore.swagger.io/v2/swagger.json"
          auth:
            type: "none"
  
  - id: user-service
    name: User Service
    servers:
      - host: users.internal
        port: 3000
    openapi:
      enabled: true
      endpoints:
        - url: "http://users.internal:3000/api-docs"
          auth:
            type: "bearer"
            token: "your-api-token"

servers:
  - id: api-gateway
    name: API Gateway with OpenAPI
    listen: ["80", "443 ssl"]
    server_name: ["api.example.com"]
    routes:
      - name: Petstore API
        path: "/petstore"
        methods: ["GET", "POST", "PUT", "DELETE"]
        backend_id: petstore-api
        strip_path: true
      
      - name: User Service
        path: "/users"
        methods: ["GET", "POST", "PUT", "DELETE"]
        backend_id: user-service
        strip_path: true
        
      - name: Unified API Documentation
        path: "/docs"
        proxy_pass: "http://localhost:8080/api/v1/openapi/swagger-ui"
```

This configuration enables:
- Two backend services with OpenAPI endpoints
- Automatic API documentation aggregation
- Unified Swagger UI accessible at `/docs`
- Different authentication methods for OpenAPI endpoints

The application follows a clean architecture pattern:

```
├── main.go                 # Application entry point
├── internal/
│   ├── api/               # REST API handlers
│   ├── config/            # Configuration management
│   ├── models/            # Data models
│   └── nginx/             # nginx configuration generation
├── web/
│   ├── static/            # CSS, JS, images
│   └── templates/         # HTML templates
└── backups/               # Configuration backups
```

## Development

### Building from Source

```bash
go build -o nginx-config-manager main.go
```

### Running Tests

```bash
go test ./...
```

### Docker Support

Create a Dockerfile:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o nginx-config-manager main.go

FROM alpine:latest
RUN apk --no-cache add nginx
WORKDIR /root/
COPY --from=builder /app/nginx-config-manager .
COPY --from=builder /app/web ./web
EXPOSE 8080
CMD ["./nginx-config-manager"]
```

Build and run:

```bash
docker build -t nginx-config-manager .
docker run -p 8080:8080 -v /etc/nginx:/etc/nginx nginx-config-manager
```

## Security Considerations

- **File Permissions**: Ensure the application has appropriate permissions to read/write nginx configuration files
- **Authentication**: Consider adding authentication for production deployments
- **Network Security**: Run behind a reverse proxy or firewall in production
- **Configuration Validation**: Always test configurations before applying them
- **Backup Strategy**: Regular backups are created, but consider additional backup strategies

## Troubleshooting

### Common Issues

1. **Permission Denied**: Ensure the application has write access to the nginx configuration directory
2. **nginx Command Not Found**: Set the correct `NGINX_EXECUTABLE` environment variable
3. **Configuration Test Fails**: Check the generated nginx configuration for syntax errors
4. **Cannot Connect to Backend**: Verify backend server connectivity and firewall rules

### Logs

The application logs important events to stdout. Use standard logging tools to capture and analyze logs:

```bash
# Run with logging
./nginx-config-manager 2>&1 | tee app.log

# Or with systemd
journalctl -u nginx-config-manager -f
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions:
- Create an issue on GitHub
- Check the troubleshooting section
- Review the configuration examples

## Roadmap

- [x] OpenAPI/Swagger aggregation and unified documentation
- [x] Dynamic API specification discovery
- [x] Unified Swagger UI interface
- [ ] Import existing nginx configurations
- [ ] Configuration templates
- [ ] Metrics and monitoring integration
- [ ] Multi-instance management
- [ ] Kubernetes integration
- [ ] Advanced authentication methods
- [ ] Configuration versioning and rollback
