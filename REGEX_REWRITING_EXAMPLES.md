# Regex Pattern Matching and URL Rewriting Examples

This document provides examples of how to use the regex pattern matching and URL rewriting functionality in the Nginx Configuration Manager GUI.

**Note this overview & all of the provided documentation were editeid by Generative AI models;  Thier content may not be completly accurate, and sources may not be properly cited. 

## Features Overview

The enhanced API Gateway now supports:

1. **Regex Pattern Matching** - Advanced location matching using regular expressions
2. **URL Rewriting** - Multiple rewrite rules per route with various flags
3. **Location Modifiers** - Support for nginx location modifiers (`~`, `~*`, `^~`, `=`)

## Configuration Examples

### 1. Basic Regex Pattern Matching

**Use Case**: Match API versioned endpoints like `/api/v1/users`, `/api/v2/users`, etc.

```json
{
  "name": "Versioned API Route",
  "path": "/api/v[0-9]+/users",
  "regex_config": {
    "enabled": true,
    "pattern": "^/api/v([0-9]+)/users$",
    "case_sensitive": false,
    "modifier": "~"
  },
  "methods": ["GET", "POST"],
  "backend_id": "api-backend"
}
```

**Generated Nginx Location**:
```nginx
location ~^/api/v([0-9]+)/users$ {
    proxy_pass http://api-backend;
}
```

### 2. Case-Insensitive Pattern Matching

**Use Case**: Match file extensions regardless of case

```json
{
  "name": "Image Files",
  "path": "/images/",
  "regex_config": {
    "enabled": true,
    "pattern": "\\.(jpe?g|png|gif|bmp)$",
    "case_sensitive": false,
    "modifier": "~*"
  },
  "methods": ["GET"],
  "backend_id": "static-backend"
}
```

**Generated Nginx Location**:
```nginx
location ~*\.(jpe?g|png|gif|bmp)$ {
    proxy_pass http://static-backend;
}
```

### 3. URL Rewriting with Capture Groups

**Use Case**: Rewrite user profile URLs from `/profile/123` to `/users/123/profile`

```json
{
  "name": "Profile Rewrite",
  "path": "/profile/",
  "regex_config": {
    "enabled": true,
    "pattern": "^/profile/([0-9]+)$",
    "case_sensitive": false,
    "modifier": "~"
  },
  "rewrite_rules": [
    {
      "enabled": true,
      "pattern": "^/profile/([0-9]+)$",
      "replacement": "/users/$1/profile",
      "flag": "break",
      "break_chain": false
    }
  ],
  "methods": ["GET"],
  "backend_id": "user-backend"
}
```

**Generated Nginx Location**:
```nginx
location ~^/profile/([0-9]+)$ {
    rewrite ^/profile/([0-9]+)$ /users/$1/profile break;
    proxy_pass http://user-backend;
}
```

### 4. Multiple Rewrite Rules

**Use Case**: Complex API routing with multiple transformation patterns

```json
{
  "name": "Complex API Route",
  "path": "/api/",
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
      "flag": "last",
      "break_chain": false
    },
    {
      "enabled": true,
      "pattern": "^/api/v([0-9]+)/posts/([0-9]+)/actions/([a-z]+)$",
      "replacement": "/v$1/post/$2/action/$3",
      "flag": "break",
      "break_chain": true
    }
  ],
  "methods": ["GET", "POST", "PUT", "DELETE"],
  "backend_id": "api-backend"
}
```

**Generated Nginx Location**:
```nginx
location ~*^/api/v([0-9]+)/(users|posts|comments)/([0-9]+)/actions/([a-z]+)$ {
    rewrite ^/api/v([0-9]+)/users/([0-9]+)/actions/([a-z]+)$ /v$1/user/$2/action/$3 last;
    rewrite ^/api/v([0-9]+)/posts/([0-9]+)/actions/([a-z]+)$ /v$1/post/$2/action/$3 break;
    proxy_pass http://api-backend;
}
```

## Location Modifiers

| Modifier | Description | Example |
|----------|-------------|---------|
| `~` | Case-sensitive regex | `location ~ ^/api/v[0-9]+/` |
| `~*` | Case-insensitive regex | `location ~* \.(jpg|png)$` |
| `^~` | Prefix match with priority | `location ^~ /api/` |
| `=` | Exact match | `location = /health` |

## Rewrite Flags

| Flag | Description | Behavior |
|------|-------------|----------|
| `last` | Stop processing, start new location search | Most common for internal rewrites |
| `break` | Stop processing in current location | Continue with current location |
| `redirect` | Return 302 redirect | Client sees new URL |
| `permanent` | Return 301 redirect | SEO-friendly permanent redirect |

## API Usage Examples

### Creating a Route via REST API

```bash
curl -X POST http://localhost:8080/api/v1/routes/server/{server-id} \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API Route with Regex",
    "path": "/api/v[0-9]+/users",
    "regex_config": {
      "enabled": true,
      "pattern": "^/api/v([0-9]+)/users/([0-9]+)$",
      "case_sensitive": false,
      "modifier": "~"
    },
    "rewrite_rules": [
      {
        "enabled": true,
        "pattern": "^/api/v([0-9]+)/users/([0-9]+)$",
        "replacement": "/users/$2",
        "flag": "break",
        "break_chain": false
      }
    ],
    "methods": ["GET", "POST"],
    "backend_id": "user-api-backend"
  }'
```

### Updating Regex Configuration

```bash
curl -X PUT http://localhost:8080/api/v1/routes/{route-id} \
  -H "Content-Type: application/json" \
  -d '{
    "regex_config": {
      "enabled": true,
      "pattern": "^/api/v([0-9]+)/users/([0-9]+)/profile$",
      "case_sensitive": true,
      "modifier": "~"
    }
  }'
```

## Testing and Validation

1. **Generate Configuration**: `POST /api/v1/config/generate`
2. **Test Configuration**: `POST /api/v1/config/test` (requires nginx installed)
3. **View Generated Config**: Check the nginx.conf file in your configured output directory

## Best Practices

1. **Use Specific Patterns**: Make regex patterns as specific as possible to avoid unintended matches
2. **Test Thoroughly**: Always test regex patterns with various input scenarios
3. **Order Matters**: More specific location blocks should come before general ones
4. **Performance**: Simple prefix matches perform better than complex regex patterns
5. **Escape Characters**: Properly escape special regex characters in patterns
6. **Capture Groups**: Use numbered capture groups ($1, $2, etc.) for rewrite replacements

## Common Use Cases

- **API Versioning**: Route different API versions to different backends
- **Legacy URL Support**: Rewrite old URLs to new endpoints
- **File Type Routing**: Route different file types to appropriate handlers
- **Geographic Routing**: Route based on URL patterns indicating regions
- **A/B Testing**: Route traffic based on URL patterns for testing
