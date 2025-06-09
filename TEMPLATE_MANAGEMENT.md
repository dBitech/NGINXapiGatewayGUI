# Template Management Guide

This guide explains how to use and manage external Jinja2 templates in the API Gateway GUI.

**Note this overview & all of the provided documentation were editeid by Generative AI models;  Thier content may not be completly accurate, and sources may not be properly cited. 

## Overview

The application now supports external Jinja2 templates instead of hardcoded Go templates, providing greater flexibility for nginx configuration customization without requiring application recompilation.

## Available Templates

The application comes with several pre-built templates:

### 1. `nginx.conf.j2` (Default)
- **Description**: Full-featured nginx configuration with all standard directives and optimizations
- **Use Case**: Production environments requiring comprehensive functionality
- **Features**: 
  - Complete upstream configuration
  - Advanced load balancing
  - Security headers
  - Gzip compression
  - Rate limiting
  - SSL/TLS configuration

### 2. `minimal.j2`
- **Description**: Minimal nginx configuration with only essential directives
- **Use Case**: Simple deployments or resource-constrained environments
- **Features**:
  - Basic upstream configuration
  - Simple server blocks
  - Essential proxy settings

### 3. `development.j2`
- **Description**: Development-focused configuration with debugging features
- **Use Case**: Development and testing environments
- **Features**:
  - Debug logging enabled
  - Relaxed security settings
  - Extended timeouts
  - Detailed error pages

### 4. `route.j2`
- **Description**: Modular route configuration template
- **Use Case**: Included by main templates for route definitions
- **Features**:
  - Reusable route blocks
  - Consistent routing patterns

## Configuration

### Setting Default Template

You can configure the default template in your configuration file:

```yaml
templates_path: "./templates"         # Path to template directory
default_template: "nginx.conf.j2"   # Default template file name
```

### Environment Variables

Override template settings using environment variables:

```bash
APIGATEWAY_TEMPLATES_PATH=./custom-templates
APIGATEWAY_DEFAULT_TEMPLATE=minimal.j2
```

## API Endpoints

### List Available Templates
```
GET /api/v1/templates
```

Returns a list of all available template files with metadata:

```json
[
  {
    "name": "nginx.conf.j2",
    "display_name": "Standard Configuration",
    "description": "Full-featured nginx configuration with all standard directives",
    "mod_time": "2025-06-09T10:30:00Z",
    "size": 2048
  }
]
```

### Generate Configuration with Specific Template
```
POST /api/v1/templates/{template_name}/generate
```

Generates nginx configuration using the specified template:

```bash
curl -X POST http://localhost:8080/api/v1/templates/minimal.j2/generate
```

### Preview Template Output
```
GET /api/v1/templates/{template_name}/preview
```

Preview the rendered template without applying it:

```bash
curl http://localhost:8080/api/v1/templates/development.j2/preview
```

### Validate Template Syntax
```
POST /api/v1/templates/{template_name}/validate
```

Validate template syntax without rendering:

```bash
curl -X POST http://localhost:8080/api/v1/templates/nginx.conf.j2/validate
```

## Template Variables

All templates have access to the following context variables:

### Global Configuration (`Global`)
```jinja2
worker_processes {{ Global.WorkerProcesses }};
worker_connections {{ Global.WorkerConnections }};
keepalive_timeout {{ Global.KeepAliveTimeout }};
client_max_body_size {{ Global.ClientMaxBodySize }};
server_tokens {{ 'on' if Global.ServerTokens else 'off' }};
gzip {{ 'on' if Global.Gzip else 'off' }};
```

### Backend Services (`Backends`)
```jinja2
{% for backend in Backends %}
upstream {{ backend.Name }} {
    {% for server in backend.Servers %}
    server {{ server.Host }}:{{ server.Port }}{% if server.Weight %} weight={{ server.Weight }}{% endif %};
    {% endfor %}
}
{% endfor %}
```

### Server Configurations (`Servers`)
```jinja2
{% for server in Servers %}
server {
    listen {{ server.Port }};
    server_name {{ server.ServerName }};
    
    {% for route in server.Routes %}
    location {{ route.Path }} {
        proxy_pass http://{{ route.Backend }};
        # Additional route-specific configuration
    }
    {% endfor %}
}
{% endfor %}
```

## Creating Custom Templates

### Template Structure

1. Create a new `.j2` file in the templates directory
2. Use Jinja2 syntax for variables and control structures
3. Access configuration data through the context variables

### Example Custom Template

```jinja2
# Custom nginx template example
worker_processes {{ Global.WorkerProcesses }};

events {
    worker_connections {{ Global.WorkerConnections }};
}

http {
    # Include route template
    {% include "route.j2" %}
    
    # Custom logic
    {% if Global.Gzip %}
    gzip on;
    gzip_types {{ Global.GzipTypes|join(' ') }};
    {% endif %}
    
    # Backend definitions
    {% for backend in Backends %}
    upstream {{ backend.Name }} {
        {% for server in backend.Servers %}
        server {{ server.Host }}:{{ server.Port }};
        {% endfor %}
    }
    {% endfor %}
    
    # Server blocks
    {% for server in Servers %}
    server {
        listen {{ server.Port }};
        server_name {{ server.ServerName }};
        
        {% for route in server.Routes %}
        location {{ route.Path }} {
            proxy_pass http://{{ route.Backend }};
        }
        {% endfor %}
    }
    {% endfor %}
}
```

### Template Validation

The system automatically validates templates for:
- Syntax errors
- File existence
- Template compilation

Use the validation API endpoint to check templates before deployment.

## Best Practices

1. **Backup Templates**: Keep backups of custom templates
2. **Version Control**: Store templates in version control
3. **Testing**: Always test templates in development first
4. **Validation**: Use the validation endpoint before applying templates
5. **Documentation**: Document custom template variables and features

## Troubleshooting

### Common Issues

1. **Template Not Found**
   - Verify the template file exists in the templates directory
   - Check file permissions
   - Ensure correct file extension (.j2)

2. **Syntax Errors**
   - Use the validation endpoint to check syntax
   - Verify Jinja2 syntax is correct
   - Check variable names match context data

3. **Missing Variables**
   - Ensure all required configuration is present
   - Check variable names in template match context
   - Use conditional blocks for optional variables

### Debug Mode

Enable debug logging to troubleshoot template issues:

```yaml
log_level: "debug"
```

This will provide detailed information about template loading and rendering processes.
