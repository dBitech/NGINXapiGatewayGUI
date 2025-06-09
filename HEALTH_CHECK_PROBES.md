# Active Health Check Configuration Examples

This document shows how to configure specific health check probes with different request types and URIs.

**Note this overview & all of the provided documentation were editeid by Generative AI models;  Thier content may not be completly accurate, and sources may not be properly cited. 

## Current Configuration Support

Your system now supports enhanced health check configuration with the following options:

### Enhanced Backend Configuration

```yaml
backends:
  - id: api-backend-with-health-check
    name: API Backend with Active Health Check
    description: Backend with custom health probe
    servers:
      - host: api1.example.com
        port: 8080
        weight: 1
      - host: api2.example.com
        port: 8080
        weight: 1
    load_balance: round_robin
    max_fails: 2                    # Passive health check
    fail_timeout: 30s               # Passive health check
    health_check:                   # Active health check (requires nginx-plus or third-party module)
      enabled: true
      path: "/health"               # ✅ Specific URI
      method: "GET"                 # ✅ Request method (GET, POST, HEAD)
      headers:                      # ✅ Custom headers
        "User-Agent": "nginx-health-check"
        "X-Health-Check": "true"
      body: ""                      # ✅ Request body for POST
      interval: "10s"               # Check every 10 seconds
      timeout: "5s"                 # 5 second timeout
      rises: 2                      # 2 successes to mark healthy
      falls: 3                      # 3 failures to mark unhealthy
      match:                        # Response validation
        status_codes: [200, 201]    # Expected status codes
        body_regex: "healthy|ok"    # Body content validation
        header_check: ""            # Header validation
```

## Health Check Examples

### Example 1: Simple GET Health Check

```yaml
health_check:
  enabled: true
  path: "/health"
  method: "GET"
  interval: "30s"
  timeout: "10s"
  rises: 2
  falls: 3
  match:
    status_codes: [200]
```

**Generated nginx config (third-party module):**
```nginx
upstream api-backend {
    server api1.example.com:8080;
    server api2.example.com:8080;
    check interval=30000 rise=2 fall=3 timeout=10000 type=http;
    check_http_send "GET /health HTTP/1.0\r\n\r\n";
    check_http_expect_alive http_200;
}
```

### Example 2: POST Health Check with Body

```yaml
health_check:
  enabled: true
  path: "/api/v1/health"
  method: "POST"
  headers:
    "Content-Type": "application/json"
    "Authorization": "Bearer health-token"
  body: '{"service": "api", "component": "all"}'
  interval: "15s"
  timeout: "8s"
  rises: 2
  falls: 2
  match:
    status_codes: [200, 201]
    body_regex: '"status":"healthy"'
```

**Generated nginx config (third-party module):**
```nginx
upstream api-backend {
    server api1.example.com:8080;
    server api2.example.com:8080;
    check interval=15000 rise=2 fall=2 timeout=8000 type=http;
    check_http_send "POST /api/v1/health HTTP/1.0\r\nContent-Type: application/json\r\nAuthorization: Bearer health-token\r\n\r\n{\"service\": \"api\", \"component\": \"all\"}";
    check_http_expect_alive http_200 http_201;
}
```

### Example 3: HEAD Health Check (Lightweight)

```yaml
health_check:
  enabled: true
  path: "/ping"
  method: "HEAD"
  interval: "5s"
  timeout: "3s"
  rises: 1
  falls: 2
  match:
    status_codes: [200, 204]
```

**Generated nginx config (third-party module):**
```nginx
upstream api-backend {
    server api1.example.com:8080;
    server api2.example.com:8080;
    check interval=5000 rise=1 fall=2 timeout=3000 type=http;
    check_http_send "HEAD /ping HTTP/1.0\r\n\r\n";
    check_http_expect_alive http_200 http_204;
}
```

### Example 4: Database Health Check

```yaml
health_check:
  enabled: true
  path: "/health/database"
  method: "GET"
  headers:
    "X-Health-Component": "database"
  interval: "60s"
  timeout: "30s"
  rises: 1
  falls: 5
  match:
    status_codes: [200]
    body_regex: '"database":"connected"'
```

## Requirements

### For Active Health Checks, you need either:

1. **Nginx Plus (Commercial)**
   - Built-in active health checking
   - Supports all features natively
   - Best performance and reliability

2. **Third-Party Modules (Open Source)**
   - nginx-upstream-check-module
   - nginx-upstream-fair
   - Requires compilation with nginx

### Backend Application Requirements

Your backend applications should implement health check endpoints that:

1. **Return appropriate HTTP status codes:**
   - `200 OK` - Service is healthy
   - `503 Service Unavailable` - Service is unhealthy
   - `204 No Content` - For HEAD requests

2. **Provide meaningful response bodies:**
   ```json
   {
     "status": "healthy",
     "timestamp": "2025-06-09T15:30:00Z",
     "services": {
       "database": "connected",
       "cache": "connected",
       "external_api": "available"
     }
   }
   ```

3. **Be fast and lightweight:**
   - Respond within 1-5 seconds
   - Minimal resource usage
   - Don't perform heavy operations

## API Usage

### Create Backend with Health Check

```bash
POST /api/v1/backends
Content-Type: application/json

{
  "name": "api-backend",
  "description": "API servers with health monitoring",
  "servers": [
    {"host": "api1.example.com", "port": 8080, "weight": 1},
    {"host": "api2.example.com", "port": 8080, "weight": 1}
  ],
  "load_balance": "round_robin",
  "max_fails": 3,
  "fail_timeout": "30s",
  "health_check": {
    "enabled": true,
    "path": "/health",
    "method": "GET",
    "interval": "30s",
    "timeout": "10s",
    "rises": 2,
    "falls": 3,
    "match": {
      "status_codes": [200]
    }
  }
}
```

### Update Health Check Configuration

```bash
PUT /api/v1/backends/{id}
Content-Type: application/json

{
  "health_check": {
    "enabled": true,
    "path": "/api/v1/status",
    "method": "POST",
    "headers": {
      "Content-Type": "application/json"
    },
    "body": "{\"check\": \"all\"}",
    "interval": "20s",
    "timeout": "8s",
    "rises": 2,
    "falls": 3,
    "match": {
      "status_codes": [200, 201],
      "body_regex": "healthy"
    }
  }
}
```

## Template Support

All nginx templates now support health check configuration:

- **nginx.conf.j2** - Main template with health check generation
- **development.j2** - Includes health check endpoints
- **minimal.j2** - Basic health check support
- **route.j2** - Route-level health monitoring

The system automatically generates the appropriate configuration based on your nginx version and available modules.
