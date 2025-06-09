# Nginx Backend Health Checking Guide

## Overview

Nginx determines backend server health through several mechanisms, ranging from basic passive checks to advanced active health monitoring. This guide explains how nginx health checks work in different scenarios and how they're configured in this API Gateway GUI.

**Note this overview & all of the provided documentation were editeid by Generative AI models;  Thier content may not be completly accurate, and sources may not be properly cited. 

## How Nginx Determines Backend Health

### 1. **Passive Health Checks (Open Source Nginx)**

Nginx open source uses passive health checking by default. This means it only marks servers as unhealthy when actual requests fail.

#### Configuration Parameters:
- **`max_fails`**: Number of failed attempts before marking server as unhealthy (default: 1)
- **`fail_timeout`**: Time period for failed attempts and recovery time (default: 10s)

#### How it Works:
1. **Request Processing**: Nginx forwards requests to backend servers in the upstream pool
2. **Failure Detection**: If a server fails to respond or returns an error, it's counted as a failure
3. **Threshold Reached**: When `max_fails` is reached within `fail_timeout` period, server is marked as unhealthy
4. **Recovery**: After `fail_timeout` expires, nginx tries the server again
5. **Health Restoration**: One successful request restores the server to healthy status

#### Example Configuration:
```nginx
upstream backend {
    server 192.168.1.10:8080 max_fails=3 fail_timeout=30s;
    server 192.168.1.11:8080 max_fails=3 fail_timeout=30s;
}
```

### 2. **Active Health Checks (Nginx Plus)**

Nginx Plus provides active health checking, which proactively monitors backend health.

#### Configuration Parameters:
- **`path`**: Health check endpoint URL
- **`interval`**: Time between health checks
- **`rises`**: Successful checks needed to mark server healthy
- **`falls`**: Failed checks needed to mark server unhealthy
- **`timeout`**: Timeout for health check requests

#### Example Configuration:
```nginx
upstream backend {
    zone backend 64k;
    server 192.168.1.10:8080;
    server 192.168.1.11:8080;
}

server {
    location / {
        proxy_pass http://backend;
        health_check uri=/health interval=10s fails=3 passes=2;
    }
}
```

## Health Check Behavior in Our Configuration

### Current Implementation

In our nginx configuration templates, we use passive health checks with configurable parameters:

```jinja2
upstream {{ backend.ID }} {
    {% for server in backend.Servers %}
    server {{ server.Host }}:{{ server.Port }}
        {% if server.Weight %} weight={{ server.Weight }}{% endif %}
        {% if server.Backup %} backup{% endif %}
        {% if server.Down %} down{% endif %}
        {% if backend.MaxFails %} max_fails={{ backend.MaxFails }}{% endif %}
        {% if backend.FailTimeout %} fail_timeout={{ backend.FailTimeout }}{% endif %};
    {% endfor %}
}
```

### Server States

#### Healthy Server
- Responds successfully to requests
- Below `max_fails` threshold
- Available for load balancing

#### Unhealthy Server
- Has exceeded `max_fails` within `fail_timeout` period
- Temporarily removed from load balancing rotation
- Will be retried after `fail_timeout` expires

#### Backup Server
- Only used when all primary servers are unhealthy
- Configured with `backup` parameter
- Always considered for health checking

#### Down Server
- Manually marked as unavailable with `down` parameter
- Not used for requests but still health checked
- Can be brought back online by removing `down` parameter

## Health Check Configuration Options

### Backend-Level Configuration

Our API Gateway GUI supports the following health check settings:

```go
type HealthCheck struct {
    Enabled  bool   `json:"enabled"`  // Enable/disable health checking
    Path     string `json:"path"`     // Health check endpoint (Nginx Plus)
    Interval string `json:"interval"` // Check interval (Nginx Plus)
    Timeout  string `json:"timeout"`  // Request timeout (Nginx Plus)
    Rises    int    `json:"rises"`    // Successes to mark healthy (Nginx Plus)
    Falls    int    `json:"falls"`    // Failures to mark unhealthy (Nginx Plus)
}
```

### Server-Level Configuration

Each backend server can be configured with:

```go
type BackendServer struct {
    Host   string `json:"host"`   // Server hostname/IP
    Port   int    `json:"port"`   // Server port
    Weight int    `json:"weight"` // Load balancing weight
    Backup bool   `json:"backup"` // Backup server flag
    Down   bool   `json:"down"`   // Manually down flag
}
```

## Health Check Examples

### Example 1: Basic Passive Health Check

```yaml
backends:
  - id: api-backend
    name: API Backend
    max_fails: 3
    fail_timeout: 30s
    servers:
      - host: api1.example.com
        port: 8080
        weight: 1
      - host: api2.example.com
        port: 8080
        weight: 1
```

Generated nginx config:
```nginx
upstream api-backend {
    server api1.example.com:8080 weight=1 max_fails=3 fail_timeout=30s;
    server api2.example.com:8080 weight=1 max_fails=3 fail_timeout=30s;
}
```

### Example 2: Backend with Backup Server

```yaml
backends:
  - id: web-backend
    name: Web Backend
    max_fails: 2
    fail_timeout: 20s
    servers:
      - host: web1.example.com
        port: 80
        weight: 3
      - host: web2.example.com
        port: 80
        weight: 2
      - host: backup.example.com
        port: 80
        weight: 1
        backup: true
```

Generated nginx config:
```nginx
upstream web-backend {
    server web1.example.com:80 weight=3 max_fails=2 fail_timeout=20s;
    server web2.example.com:80 weight=2 max_fails=2 fail_timeout=20s;
    server backup.example.com:80 weight=1 backup max_fails=2 fail_timeout=20s;
}
```

### Example 3: Health Check Endpoint (Development Template)

Our development template includes a health check endpoint:

```nginx
# Health check endpoint
location /health {
    access_log off;
    return 200 "healthy\n";
    add_header Content-Type text/plain;
}
```

## Health Check Best Practices

### 1. **Appropriate Thresholds**
- Set `max_fails` to 2-3 for most applications
- Use longer `fail_timeout` for databases (60s+)
- Use shorter `fail_timeout` for web services (10-30s)

### 2. **Health Check Endpoints**
- Create dedicated health check endpoints (`/health`, `/status`)
- Return HTTP 200 for healthy, 503 for unhealthy
- Keep health checks lightweight and fast
- Include dependency checks (database, cache, etc.)

### 3. **Load Balancing Strategy**
- Use `backup` servers for failover scenarios
- Set appropriate `weight` values based on server capacity
- Consider using `least_conn` for long-lived connections

### 4. **Monitoring and Alerting**
- Monitor nginx error logs for backend failures
- Set up alerts for repeated health check failures
- Track response times and error rates

## Troubleshooting Health Check Issues

### Common Problems

#### 1. **Server Marked as Down Frequently**
- **Cause**: `max_fails` threshold too low or `fail_timeout` too short
- **Solution**: Increase `max_fails` or `fail_timeout` values
- **Check**: Backend server response times and error rates

#### 2. **Health Checks Not Working**
- **Cause**: Health check endpoint not responding correctly
- **Solution**: Verify endpoint returns HTTP 200 for healthy status
- **Check**: Test health endpoint directly: `curl http://backend/health`

#### 3. **Backup Server Always Used**
- **Cause**: All primary servers marked as unhealthy
- **Solution**: Check primary server health and network connectivity
- **Check**: Nginx error logs for upstream connection errors

#### 4. **Load Not Distributed Evenly**
- **Cause**: Incorrect weight configuration or server capacity differences
- **Solution**: Adjust `weight` values based on server capabilities
- **Check**: Monitor request distribution across servers

### Debugging Commands

```bash
# Check nginx error logs for upstream issues
tail -f /var/log/nginx/error.log | grep upstream

# Test backend connectivity
curl -v http://backend-server:port/

# Check nginx configuration syntax
nginx -t

# Reload nginx configuration
nginx -s reload

# Check nginx status (if status module enabled)
curl http://localhost/nginx_status
```

## API Gateway GUI Integration

### Configuration via Web Interface

1. **Backend Management**:
   - Navigate to "Backends" section
   - Configure `max_fails` and `fail_timeout` values
   - Set server weights and backup flags

2. **Health Check Monitoring**:
   - View backend status in dashboard
   - Monitor server health through nginx logs
   - Use built-in configuration testing

3. **Template Selection**:
   - Use "Development" template for health check endpoints
   - Use "Standard" template for production deployments
   - Use "Minimal" template for simple setups

### API Endpoints

```bash
# Get backend health status
GET /api/v1/backends/{id}

# Update backend health configuration
PUT /api/v1/backends/{id}
{
  "max_fails": 3,
  "fail_timeout": "30s",
  "health_check": {
    "enabled": true,
    "path": "/health"
  }
}

# Test nginx configuration
POST /api/v1/config/test

# Generate configuration with health checks
POST /api/v1/config/generate
```

This comprehensive health checking system ensures your nginx API gateway maintains high availability and automatically routes traffic away from unhealthy backend servers.
