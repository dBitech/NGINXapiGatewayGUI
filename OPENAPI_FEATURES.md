# OpenAPI Aggregation Features

## Overview

The nginx API Gateway GUI now includes comprehensive OpenAPI/Swagger aggregation capabilities, allowing you to:

- Automatically discover and aggregate OpenAPI specifications from multiple backend services
- Provide a unified API documentation experience
- Configure per-backend OpenAPI settings with authentication support
- Monitor OpenAPI specification fetch status and errors
- View aggregated API documentation through an integrated Swagger UI

**Note this overview & all of the provided documentation were editeid by Generative AI models;  Thier content may not be completly accurate, and sources may not be properly cited. 

## Features

### 1. Backend OpenAPI Configuration

Each backend can be configured with OpenAPI integration:

- **Enable/Disable**: Toggle OpenAPI aggregation per backend
- **Multiple Endpoints**: Support for multiple OpenAPI specification endpoints per backend
- **Authentication**: Support for Bearer tokens, Basic auth, and API keys
- **Custom Headers**: Add custom headers for OpenAPI spec requests
- **Timeout Configuration**: Configurable request timeouts

### 2. OpenAPI Aggregation Service

The aggregation service provides:

- **Unified Specification**: Combines OpenAPI specs from all enabled backends
- **Path Transformation**: Automatically prefixes paths with backend routing
- **Component Merging**: Merges schemas, responses, and parameters with backend prefixes
- **Tag Management**: Preserves and organizes tags from different backends
- **Caching**: Configurable cache TTL to reduce backend load
- **Error Handling**: Graceful handling of unavailable or malformed specs

### 3. Web UI Integration

#### Navigation
- New "OpenAPI" section in the main navigation sidebar
- Real-time status monitoring and backend health display

#### Backend Configuration
- OpenAPI settings integrated into backend creation/editing forms
- Support for multiple endpoint URLs per backend
- Authentication configuration (Bearer, Basic, API Key)
- Custom headers and timeout settings

#### OpenAPI Dashboard
- Service status overview with cache metrics
- Backend-specific OpenAPI status and error monitoring
- One-click cache refresh functionality
- Direct link to unified Swagger UI

#### Swagger UI Integration
- Standalone Swagger UI page for aggregated documentation
- Modern Swagger UI 4.x with full feature support
- Deep linking and filtering capabilities
- Responsive design for desktop and mobile

## API Endpoints

### GET /api/v1/openapi/status
Returns overall OpenAPI service status:
```json
{
  "service_status": "running",
  "total_backends": 2,
  "enabled_backends": 2,
  "cached_specs": 2,
  "successful_fetches": 2,
  "failed_fetches": 0,
  "cache_ttl_minutes": 5,
  "last_refresh": "2025-06-09T11:49:04-07:00"
}
```

### GET /api/v1/openapi/backends
Returns per-backend OpenAPI status:
```json
{
  "backends": [
    {
      "id": "petstore-api",
      "name": "Petstore API",
      "openapi_status": "enabled",
      "spec_url": "http://petstore.swagger.io/v2/swagger.json",
      "status": "success",
      "last_fetch": "2025-06-09T11:49:04-07:00",
      "paths_count": 20
    }
  ],
  "total_backends": 2,
  "enabled_backends": 2
}
```

### POST /api/v1/openapi/refresh
Forces refresh of all cached OpenAPI specifications:
```json
{
  "message": "OpenAPI cache refreshed successfully",
  "refreshed_at": "2025-06-09T11:49:04-07:00"
}
```

### GET /api/v1/openapi/aggregated
Returns the unified OpenAPI 3.0.3 specification with all backend APIs merged.

### GET /api/v1/openapi/swagger-ui
Serves a standalone Swagger UI page displaying the aggregated API documentation.

## Configuration Example

```yaml
backends:
  - id: petstore-api
    name: Petstore API
    description: Swagger Petstore API for testing
    servers:
      - host: petstore.swagger.io
        port: 80
    openapi:
      enabled: true
      endpoints:
        - "/v2/swagger.json"
      timeout: "30s"
      headers:
        User-Agent: "nginx-gateway/1.0"
      auth:
        type: "bearer"
        token: "your-api-token"
```

## Advanced Features

### Authentication Support
- **Bearer Token**: `Authorization: Bearer <token>`
- **Basic Auth**: Standard HTTP Basic authentication
- **API Key**: Custom header or query parameter authentication

### Error Handling
- Graceful handling of network timeouts
- Invalid JSON/YAML specification handling
- Backend unavailability fallback
- Detailed error reporting in web UI

### Caching Strategy
- Configurable cache TTL (default: 5 minutes)
- Force refresh capability
- Individual backend cache invalidation
- Memory-efficient caching with cleanup

### Path Transformation
- Automatic path prefixing based on backend routing rules
- Preservation of original path parameters
- Server URL rewriting for gateway context

## Testing

The system has been tested with:
- **Petstore API**: Official Swagger Petstore (petstore.swagger.io)
- **HTTPBin API**: HTTP testing service (httpbin.org)
- **Multiple Endpoints**: Backends with multiple OpenAPI specification URLs
- **Error Scenarios**: Network failures, malformed specs, authentication errors

## Browser Support

The Swagger UI integration supports:
- Chrome 60+
- Firefox 55+
- Safari 11+
- Edge 79+

## Security Considerations

- API keys and tokens are stored securely in backend configuration
- HTTPS support for backend OpenAPI endpoints
- Input validation for all OpenAPI configuration fields
- Rate limiting protection for OpenAPI spec fetching
