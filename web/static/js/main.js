// Main application JavaScript
let currentConfig = {};
let currentSection = 'dashboard';

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    initializeApp();
    loadStatus();
    loadConfiguration();
    
    // Set up navigation
    document.querySelectorAll('.nav-link[data-section]').forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            const section = this.getAttribute('data-section');
            showSection(section);
        });
    });
    
    // Set up health check toggle (delegated event listener)
    document.addEventListener('change', function(e) {
        if (e.target && e.target.id === 'health-check-enabled') {
            toggleHealthCheckPanel(e.target.checked);
        }
    });
    
    // Auto-refresh status every 30 seconds
    setInterval(loadStatus, 30000);
});

function initializeApp() {
    showSection('dashboard');
}

function showSection(section) {
    // Hide all sections
    document.querySelectorAll('.section').forEach(s => s.style.display = 'none');
    
    // Show selected section
    document.getElementById(section + '-section').style.display = 'block';
    
    // Update navigation
    document.querySelectorAll('.nav-link[data-section]').forEach(link => {
        link.classList.remove('active');
    });
    document.querySelector(`[data-section="${section}"]`).classList.add('active');
    
    currentSection = section;
    
    // Load section-specific data
    switch(section) {
        case 'backends':
            loadBackends();
            break;
        case 'servers':
            loadServers();
            break;
        case 'routes':
            loadServersForRoutes();
            break;
        case 'config':
            loadGlobalConfig();
            break;
    }
}

// API calls
async function apiCall(method, endpoint, data = null) {
    const options = {
        method: method,
        headers: {
            'Content-Type': 'application/json',
        }
    };
    
    if (data) {
        options.body = JSON.stringify(data);
    }
    
    try {
        const response = await fetch(`/api/v1${endpoint}`, options);
        const result = await response.json();
        
        if (!response.ok) {
            throw new Error(result.error || 'Request failed');
        }
        
        return result;
    } catch (error) {
        showAlert('error', error.message);
        throw error;
    }
}

// Status functions
async function loadStatus() {
    try {
        const status = await apiCall('GET', '/status');
        
        // Update status badges
        const statusElement = document.getElementById('nginx-status');
        const statusCardElement = document.getElementById('nginx-status-card');
        
        statusElement.textContent = status.nginx_status;
        statusCardElement.textContent = status.nginx_status;
        
        // Update status colors
        statusElement.className = `badge bg-${status.nginx_status === 'running' ? 'success' : 'danger'}`;
        
        // Update counts
        document.getElementById('backends-count').textContent = status.backends_count || 0;
        document.getElementById('servers-count').textContent = status.servers_count || 0;
        document.getElementById('routes-count').textContent = status.routes_count || 0;
        
    } catch (error) {
        console.error('Failed to load status:', error);
    }
}

async function loadConfiguration() {
    try {
        const config = await apiCall('GET', '/config');
        currentConfig = config;
    } catch (error) {
        console.error('Failed to load configuration:', error);
    }
}

// Backend functions
async function loadBackends() {
    try {
        const backends = await apiCall('GET', '/backends');
        renderBackends(backends);
    } catch (error) {
        console.error('Failed to load backends:', error);
    }
}

function renderBackends(backends) {
    const container = document.getElementById('backends-list');
    
    if (!backends || backends.length === 0) {
        container.innerHTML = '<div class="alert alert-info">No backends configured yet.</div>';
        return;
    }
    
    let html = '<div class="row">';
    
    backends.forEach(backend => {
        html += `
            <div class="col-md-6 mb-3">
                <div class="card">
                    <div class="card-header d-flex justify-content-between align-items-center">
                        <h6 class="mb-0">${backend.name}</h6>
                        <div class="btn-group">
                            <button class="btn btn-sm btn-outline-primary" onclick="editBackend('${backend.id}')">
                                <i class="fas fa-edit"></i>
                            </button>
                            <button class="btn btn-sm btn-outline-danger" onclick="deleteBackend('${backend.id}')">
                                <i class="fas fa-trash"></i>
                            </button>
                        </div>
                    </div>
                    <div class="card-body">
                        <p class="text-muted">${backend.description || 'No description'}</p>
                        <div class="row">
                            <div class="col-6">
                                <small class="text-muted">Load Balance:</small><br>
                                <span class="badge bg-info">${backend.load_balance || 'round_robin'}</span>
                            </div>
                            <div class="col-6">
                                <small class="text-muted">Servers:</small><br>
                                <span class="badge bg-secondary">${backend.servers ? backend.servers.length : 0}</span>
                            </div>
                        </div>
                        <div class="row mt-2">
                            <div class="col-6">
                                <small class="text-muted">Health Check:</small><br>
                                ${backend.health_check && backend.health_check.enabled ? 
                                    `<span class="badge bg-success">Enabled</span>` : 
                                    `<span class="badge bg-secondary">Disabled</span>`
                                }
                            </div>
                            <div class="col-6">
                                <small class="text-muted">Max Fails:</small><br>
                                <span class="badge bg-warning">${backend.max_fails || 3}</span>
                            </div>
                        </div>
                        ${backend.servers && backend.servers.length > 0 ? 
                            `<div class="mt-2">
                                <small class="text-muted">Server List:</small><br>
                                ${backend.servers.map(s => `<small>${s.host}:${s.port}</small>`).join('<br>')}
                            </div>` : ''
                        }
                        ${backend.health_check && backend.health_check.enabled ? 
                            `<div class="mt-2">
                                <small class="text-muted">Health Check:</small><br>
                                <small>${backend.health_check.method || 'GET'} ${backend.health_check.path || '/health'}</small><br>
                                <small>Interval: ${backend.health_check.interval || '30s'}</small>
                            </div>` : ''
                        }
                    </div>
                </div>
            </div>
        `;
    });
    
    html += '</div>';
    container.innerHTML = html;
}

function showBackendModal(backendId = null) {
    const modal = new bootstrap.Modal(document.getElementById('backendModal'));
    
    // Reset form
    document.getElementById('backend-form').reset();
    document.getElementById('backend-id').value = '';
    
    // Reset servers container
    document.getElementById('backend-servers').innerHTML = `
        <div class="backend-server-item">
            <div class="row">
                <div class="col-md-4">
                    <input type="text" class="form-control" placeholder="Host" name="server-host">
                </div>
                <div class="col-md-2">
                    <input type="number" class="form-control" placeholder="Port" name="server-port">
                </div>
                <div class="col-md-2">
                    <input type="number" class="form-control" placeholder="Weight" name="server-weight" value="1">
                </div>
                <div class="col-md-2">
                    <div class="form-check">
                        <input class="form-check-input" type="checkbox" name="server-backup">
                        <label class="form-check-label">Backup</label>
                    </div>
                </div>
                <div class="col-md-2">
                    <button type="button" class="btn btn-danger btn-sm" onclick="removeBackendServer(this)">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        </div>
    `;
    
    if (backendId) {
        // Load existing backend data
        loadBackendForEdit(backendId);
    }
    
    modal.show();
}

async function loadBackendForEdit(backendId) {
    try {
        const backend = await apiCall('GET', `/backends/${backendId}`);
        
        document.getElementById('backend-id').value = backend.id;
        document.getElementById('backend-name').value = backend.name;
        document.getElementById('backend-description').value = backend.description || '';
        document.getElementById('backend-load-balance').value = backend.load_balance || 'round_robin';
        document.getElementById('backend-max-fails').value = backend.max_fails || 3;
        document.getElementById('backend-fail-timeout').value = backend.fail_timeout || '30s';
        
        // Load servers
        if (backend.servers && backend.servers.length > 0) {
            const container = document.getElementById('backend-servers');
            container.innerHTML = '';
            
            backend.servers.forEach(server => {
                const serverHtml = `
                    <div class="backend-server-item">
                        <div class="row">
                            <div class="col-md-4">
                                <input type="text" class="form-control" placeholder="Host" name="server-host" value="${server.host}">
                            </div>
                            <div class="col-md-2">
                                <input type="number" class="form-control" placeholder="Port" name="server-port" value="${server.port}">
                            </div>
                            <div class="col-md-2">
                                <input type="number" class="form-control" placeholder="Weight" name="server-weight" value="${server.weight || 1}">
                            </div>
                            <div class="col-md-2">
                                <div class="form-check">
                                    <input class="form-check-input" type="checkbox" name="server-backup" ${server.backup ? 'checked' : ''}>
                                    <label class="form-check-label">Backup</label>
                                </div>
                            </div>
                            <div class="col-md-2">
                                <button type="button" class="btn btn-danger btn-sm" onclick="removeBackendServer(this)">
                                    <i class="fas fa-trash"></i>
                                </button>
                            </div>
                        </div>
                    </div>
                `;
                container.insertAdjacentHTML('beforeend', serverHtml);
            });
        }
        
        // Load health check configuration
        if (backend.health_check) {
            document.getElementById('health-check-enabled').checked = backend.health_check.enabled || false;
            document.getElementById('health-check-path').value = backend.health_check.path || '/health';
            document.getElementById('health-check-method').value = backend.health_check.method || 'GET';
            document.getElementById('health-check-interval').value = backend.health_check.interval || '30s';
            document.getElementById('health-check-timeout').value = backend.health_check.timeout || '10s';
            document.getElementById('health-check-rises').value = backend.health_check.rises || 2;
            document.getElementById('health-check-falls').value = backend.health_check.falls || 3;
            document.getElementById('health-check-body').value = backend.health_check.body || '';
            
            // Load headers
            if (backend.health_check.headers && Object.keys(backend.health_check.headers).length > 0) {
                document.getElementById('health-check-headers').value = JSON.stringify(backend.health_check.headers, null, 2);
            }
            
            // Load match configuration
            if (backend.health_check.match) {
                if (backend.health_check.match.status_codes && backend.health_check.match.status_codes.length > 0) {
                    document.getElementById('health-check-status-codes').value = backend.health_check.match.status_codes.join(', ');
                }
                document.getElementById('health-check-body-regex').value = backend.health_check.match.body_regex || '';
            }
            
            // Toggle health check panel visibility
            toggleHealthCheckPanel(backend.health_check.enabled);
        } else {
            // Reset health check form
            document.getElementById('health-check-enabled').checked = false;
            toggleHealthCheckPanel(false);
        }
        
    } catch (error) {
        console.error('Failed to load backend for edit:', error);
    }
}

function addBackendServer() {
    const container = document.getElementById('backend-servers');
    const serverHtml = `
        <div class="backend-server-item">
            <div class="row">
                <div class="col-md-4">
                    <input type="text" class="form-control" placeholder="Host" name="server-host">
                </div>
                <div class="col-md-2">
                    <input type="number" class="form-control" placeholder="Port" name="server-port">
                </div>
                <div class="col-md-2">
                    <input type="number" class="form-control" placeholder="Weight" name="server-weight" value="1">
                </div>
                <div class="col-md-2">
                    <div class="form-check">
                        <input class="form-check-input" type="checkbox" name="server-backup">
                        <label class="form-check-label">Backup</label>
                    </div>
                </div>
                <div class="col-md-2">
                    <button type="button" class="btn btn-danger btn-sm" onclick="removeBackendServer(this)">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        </div>
    `;
    container.insertAdjacentHTML('beforeend', serverHtml);
}

function removeBackendServer(button) {
    const container = document.getElementById('backend-servers');
    if (container.children.length > 1) {
        button.closest('.backend-server-item').remove();
    } else {
        showAlert('warning', 'At least one server is required.');
    }
}

async function saveBackend() {
    const form = document.getElementById('backend-form');
    const formData = new FormData(form);
    
    const backend = {
        name: formData.get('backend-name') || document.getElementById('backend-name').value,
        description: formData.get('backend-description') || document.getElementById('backend-description').value,
        load_balance: formData.get('backend-load-balance') || document.getElementById('backend-load-balance').value,
        max_fails: parseInt(formData.get('backend-max-fails') || document.getElementById('backend-max-fails').value),
        fail_timeout: formData.get('backend-fail-timeout') || document.getElementById('backend-fail-timeout').value,
        servers: []
    };
    
    // Collect servers
    const serverItems = document.querySelectorAll('.backend-server-item');
    serverItems.forEach(item => {
        const host = item.querySelector('[name="server-host"]').value;
        const port = parseInt(item.querySelector('[name="server-port"]').value);
        const weight = parseInt(item.querySelector('[name="server-weight"]').value) || 1;
        const backup = item.querySelector('[name="server-backup"]').checked;
        
        if (host && port) {
            backend.servers.push({
                id: `server_${Date.now()}_${Math.random()}`,
                host: host,
                port: port,
                weight: weight,
                backup: backup,
                down: false
            });
        }
    });
    
    // Collect health check configuration
    const healthCheckEnabled = document.getElementById('health-check-enabled').checked;
    if (healthCheckEnabled) {
        const statusCodesInput = document.getElementById('health-check-status-codes').value;
        const headersInput = document.getElementById('health-check-headers').value;
        
        let statusCodes = [];
        if (statusCodesInput) {
            statusCodes = statusCodesInput.split(',').map(code => parseInt(code.trim())).filter(code => !isNaN(code));
        }
        
        let headers = {};
        if (headersInput) {
            try {
                headers = JSON.parse(headersInput);
            } catch (e) {
                console.warn('Invalid JSON in health check headers, ignoring');
            }
        }
        
        backend.health_check = {
            enabled: true,
            path: document.getElementById('health-check-path').value || '/health',
            method: document.getElementById('health-check-method').value || 'GET',
            interval: document.getElementById('health-check-interval').value || '30s',
            timeout: document.getElementById('health-check-timeout').value || '10s',
            rises: parseInt(document.getElementById('health-check-rises').value) || 2,
            falls: parseInt(document.getElementById('health-check-falls').value) || 3,
            headers: headers,
            body: document.getElementById('health-check-body').value || '',
            match: {
                status_codes: statusCodes.length > 0 ? statusCodes : [200],
                body_regex: document.getElementById('health-check-body-regex').value || ''
            }
        };
    }
    
    if (backend.servers.length === 0) {
        showAlert('error', 'At least one server is required.');
        return;
    }
    
    try {
        const backendId = document.getElementById('backend-id').value;
        
        if (backendId) {
            await apiCall('PUT', `/backends/${backendId}`, backend);
            showAlert('success', 'Backend updated successfully!');
        } else {
            await apiCall('POST', '/backends', backend);
            showAlert('success', 'Backend created successfully!');
        }
        
        const modal = bootstrap.Modal.getInstance(document.getElementById('backendModal'));
        modal.hide();
        
        loadBackends();
        loadStatus();
        
    } catch (error) {
        console.error('Failed to save backend:', error);
    }
}

async function editBackend(backendId) {
    showBackendModal(backendId);
}

async function deleteBackend(backendId) {
    if (!confirm('Are you sure you want to delete this backend?')) {
        return;
    }
    
    try {
        await apiCall('DELETE', `/backends/${backendId}`);
        showAlert('success', 'Backend deleted successfully!');
        loadBackends();
        loadStatus();
    } catch (error) {
        console.error('Failed to delete backend:', error);
    }
}

// Server functions
async function loadServers() {
    try {
        const servers = await apiCall('GET', '/servers');
        renderServers(servers);
    } catch (error) {
        console.error('Failed to load servers:', error);
    }
}

function renderServers(servers) {
    const container = document.getElementById('servers-list');
    
    if (!servers || servers.length === 0) {
        container.innerHTML = '<div class="alert alert-info">No servers configured yet.</div>';
        return;
    }
    
    let html = '<div class="row">';
    
    servers.forEach(server => {
        html += `
            <div class="col-md-6 mb-3">
                <div class="card">
                    <div class="card-header d-flex justify-content-between align-items-center">
                        <h6 class="mb-0">${server.name}</h6>
                        <div class="btn-group">
                            <button class="btn btn-sm btn-outline-primary" onclick="editServer('${server.id}')">
                                <i class="fas fa-edit"></i>
                            </button>
                            <button class="btn btn-sm btn-outline-danger" onclick="deleteServer('${server.id}')">
                                <i class="fas fa-trash"></i>
                            </button>
                        </div>
                    </div>
                    <div class="card-body">
                        <div class="row">
                            <div class="col-6">
                                <small class="text-muted">Listen:</small><br>
                                ${server.listen ? server.listen.map(l => `<span class="badge bg-info me-1">${l}</span>`).join('') : 'Not configured'}
                            </div>
                            <div class="col-6">
                                <small class="text-muted">Routes:</small><br>
                                <span class="badge bg-secondary">${server.routes ? server.routes.length : 0}</span>
                            </div>
                        </div>
                        ${server.server_name && server.server_name.length > 0 ? 
                            `<div class="mt-2">
                                <small class="text-muted">Server Names:</small><br>
                                ${server.server_name.map(name => `<small>${name}</small>`).join(', ')}
                            </div>` : ''
                        }
                    </div>
                </div>
            </div>
        `;
    });
    
    html += '</div>';
    container.innerHTML = html;
}

function showServerModal(serverId = null) {
    const modal = new bootstrap.Modal(document.getElementById('serverModal'));
    
    // Reset form
    document.getElementById('server-form').reset();
    document.getElementById('server-id').value = '';
    
    if (serverId) {
        loadServerForEdit(serverId);
    }
    
    modal.show();
}

async function loadServerForEdit(serverId) {
    try {
        const server = await apiCall('GET', `/servers/${serverId}`);
        
        document.getElementById('server-id').value = server.id;
        document.getElementById('server-name').value = server.name;
        document.getElementById('server-listen').value = server.listen ? server.listen.join('\n') : '';
        document.getElementById('server-names').value = server.server_name ? server.server_name.join('\n') : '';
        document.getElementById('server-client-max-body-size').value = server.client_max_body_size || '';
        document.getElementById('server-access-log').value = server.access_log || '';
        document.getElementById('server-error-log').value = server.error_log || '';
        
    } catch (error) {
        console.error('Failed to load server for edit:', error);
    }
}

async function saveServer() {
    const form = document.getElementById('server-form');
    
    const server = {
        name: document.getElementById('server-name').value,
        listen: document.getElementById('server-listen').value.split('\n').filter(l => l.trim()),
        server_name: document.getElementById('server-names').value.split('\n').filter(n => n.trim()),
        client_max_body_size: document.getElementById('server-client-max-body-size').value || undefined,
        access_log: document.getElementById('server-access-log').value || undefined,
        error_log: document.getElementById('server-error-log').value || undefined,
        routes: []
    };
    
    try {
        const serverId = document.getElementById('server-id').value;
        
        if (serverId) {
            await apiCall('PUT', `/servers/${serverId}`, server);
            showAlert('success', 'Server updated successfully!');
        } else {
            await apiCall('POST', '/servers', server);
            showAlert('success', 'Server created successfully!');
        }
        
        const modal = bootstrap.Modal.getInstance(document.getElementById('serverModal'));
        modal.hide();
        
        loadServers();
        loadStatus();
        
    } catch (error) {
        console.error('Failed to save server:', error);
    }
}

async function editServer(serverId) {
    showServerModal(serverId);
}

async function deleteServer(serverId) {
    if (!confirm('Are you sure you want to delete this server?')) {
        return;
    }
    
    try {
        await apiCall('DELETE', `/servers/${serverId}`);
        showAlert('success', 'Server deleted successfully!');
        loadServers();
        loadStatus();
    } catch (error) {
        console.error('Failed to delete server:', error);
    }
}

// Route functions
async function loadServersForRoutes() {
    try {
        const servers = await apiCall('GET', '/servers');
        const select = document.getElementById('server-select');
        
        select.innerHTML = '<option value="">Select a server</option>';
        
        servers.forEach(server => {
            const option = document.createElement('option');
            option.value = server.id;
            option.textContent = server.name;
            select.appendChild(option);
        });
        
    } catch (error) {
        console.error('Failed to load servers for routes:', error);
    }
}

async function loadRoutes() {
    const serverId = document.getElementById('server-select').value;
    
    if (!serverId) {
        document.getElementById('routes-list').innerHTML = '<div class="alert alert-info">Please select a server to view routes.</div>';
        return;
    }
    
    try {
        const routes = await apiCall('GET', `/routes/server/${serverId}`);
        renderRoutes(routes, serverId);
    } catch (error) {
        console.error('Failed to load routes:', error);
    }
}

function renderRoutes(routes, serverId) {
    const container = document.getElementById('routes-list');
    
    if (!routes || routes.length === 0) {
        container.innerHTML = '<div class="alert alert-info">No routes configured for this server.</div>';
        return;
    }
    
    let html = '<div class="table-responsive"><table class="table table-striped">';
    html += '<thead><tr><th>Name</th><th>Path</th><th>Methods</th><th>Backend</th><th>Options</th><th>Actions</th></tr></thead><tbody>';
    
    routes.forEach(route => {
        const methods = route.methods ? route.methods.map(m => `<span class="badge bg-primary method-badge">${m}</span>`).join('') : '';
        const options = [];
        
        if (route.strip_path) options.push('<span class="badge bg-info">Strip Path</span>');
        if (route.preserve_host) options.push('<span class="badge bg-info">Preserve Host</span>');
        if (route.rate_limiting && route.rate_limiting.enabled) options.push('<span class="badge bg-warning">Rate Limited</span>');
        if (route.caching && route.caching.enabled) options.push('<span class="badge bg-success">Cached</span>');
        if (route.cors && route.cors.enabled) options.push('<span class="badge bg-secondary">CORS</span>');
        if (route.regex_config && route.regex_config.enabled) options.push('<span class="badge bg-primary">Regex</span>');
        if (route.rewrite_rules && route.rewrite_rules.length > 0) options.push('<span class="badge bg-dark">Rewrite</span>');
        
        html += `
            <tr>
                <td>${route.name}</td>
                <td><code>${route.path}</code></td>
                <td>${methods}</td>
                <td>${route.backend_id || 'Not configured'}</td>
                <td>${options.join(' ')}</td>
                <td>
                    <div class="btn-group">
                        <button class="btn btn-sm btn-outline-primary" onclick="editRoute('${serverId}', '${route.id}')">
                            <i class="fas fa-edit"></i>
                        </button>
                        <button class="btn btn-sm btn-outline-danger" onclick="deleteRoute('${serverId}', '${route.id}')">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </td>
            </tr>
        `;
    });
    
    html += '</tbody></table></div>';
    container.innerHTML = html;
}

function showRouteModal(serverId = null, routeId = null) {
    try {
        const currentServerId = serverId || document.getElementById('server-select').value;
        
        if (!currentServerId) {
            showAlert('error', 'Please select a server first.');
            return;
        }
        
        const modalElement = document.getElementById('routeModal');
        if (!modalElement) {
            console.error('Modal element not found!');
            return;
        }
        
        if (typeof bootstrap === 'undefined') {
            console.error('Bootstrap not loaded!');
            return;
        }
        
        const modal = new bootstrap.Modal(modalElement);
        
        // Reset form
        document.getElementById('route-form').reset();
        document.getElementById('route-id').value = '';
        document.getElementById('route-server-id').value = currentServerId;
        
        // Clear rewrite rules
        clearRewriteRules();
        
        // Clear cache rules
        clearCacheRules();
        
        // Load backends for selection
        loadBackendsForRoute();
        
        if (routeId) {
            loadRouteForEdit(currentServerId, routeId);
        }
        
        modal.show();
    } catch (error) {
        console.error('Error in showRouteModal:', error);
    }
}

async function loadBackendsForRoute() {
    try {
        const backends = await apiCall('GET', '/backends');
        const select = document.getElementById('route-backend');
        
        select.innerHTML = '<option value="">Select Backend</option>';
        
        backends.forEach(backend => {
            const option = document.createElement('option');
            option.value = backend.id;
            option.textContent = backend.name;
            select.appendChild(option);
        });
        
    } catch (error) {
        console.error('Failed to load backends for route:', error);
    }
}

async function loadRouteForEdit(serverId, routeId) {
    try {
        const route = await apiCall('GET', `/routes/${routeId}`);
        
        document.getElementById('route-id').value = route.id;
        document.getElementById('route-name').value = route.name;
        document.getElementById('route-path').value = route.path;
        document.getElementById('route-backend').value = route.backend_id;
        document.getElementById('route-strip-path').checked = route.strip_path || false;
        document.getElementById('route-preserve-host').checked = route.preserve_host || false;
        
        // Set methods
        if (route.methods) {
            route.methods.forEach(method => {
                const checkbox = document.getElementById(`method-${method.toLowerCase()}`);
                if (checkbox) checkbox.checked = true;
            });
        }
        
        // Load advanced configurations
        if (route.rate_limiting) {
            document.getElementById('rate-limiting-enabled').checked = route.rate_limiting.enabled || false;
            document.getElementById('rate-limiting-rate').value = route.rate_limiting.rate || '';
            document.getElementById('rate-limiting-burst').value = route.rate_limiting.burst || '';
            document.getElementById('rate-limiting-status').value = route.rate_limiting.status_code || 429;
        }
        
        if (route.caching) {
            document.getElementById('caching-enabled').checked = route.caching.enabled || false;
            document.getElementById('caching-ttl').value = route.caching.ttl || '';
            document.getElementById('caching-key').value = route.caching.key || '';
            
            // Load advanced caching options
            if (route.caching.methods) {
                const methodsEl = document.getElementById('cache-methods');
                if (methodsEl) methodsEl.value = route.caching.methods.join(', ');
            }
            
            if (route.caching.status_codes) {
                const statusCodesEl = document.getElementById('cache-status-codes');
                if (statusCodesEl) statusCodesEl.value = route.caching.status_codes.join(', ');
            }
            
            if (route.caching.ignore_headers) {
                const ignoreHeadersEl = document.getElementById('cache-ignore-headers');
                if (ignoreHeadersEl) ignoreHeadersEl.value = route.caching.ignore_headers.join(', ');
            }
            
            if (route.caching.bypass) {
                const bypassEl = document.getElementById('cache-bypass');
                if (bypassEl) bypassEl.value = route.caching.bypass.join('\n');
            }
            
            if (route.caching.min_uses) {
                const minUsesEl = document.getElementById('cache-min-uses');
                if (minUsesEl) minUsesEl.value = route.caching.min_uses;
            }
            
            if (route.caching.inactive) {
                const inactiveEl = document.getElementById('cache-inactive');
                if (inactiveEl) inactiveEl.value = route.caching.inactive;
            }
            
            if (route.caching.max_size) {
                const maxSizeEl = document.getElementById('cache-max-size');
                if (maxSizeEl) maxSizeEl.value = route.caching.max_size;
            }
            
            // Load cache URI rules
            clearCacheRules();
            if (route.caching.uri_rules && route.caching.uri_rules.length > 0) {
                route.caching.uri_rules.forEach(rule => {
                    addCacheRule(
                        rule.pattern || '',
                        rule.type || 'prefix',
                        rule.ttl || '',
                        rule.methods ? rule.methods.join(', ') : '',
                        rule.status_codes ? rule.status_codes.join(', ') : '',
                        rule.no_cache || false,
                        rule.description || ''
                    );
                });
            }
            
            // Load cache validity rules
            const validRulesContainer = document.getElementById('cache-valid-rules');
            if (validRulesContainer) {
                validRulesContainer.innerHTML = '';
                if (route.caching.valid && route.caching.valid.length > 0) {
                    route.caching.valid.forEach(validRule => {
                        addCacheValidRule(
                            validRule.status_codes ? validRule.status_codes.join(', ') : '',
                            validRule.ttl || ''
                        );
                    });
                }
            }
        }
        
        if (route.cors) {
            document.getElementById('cors-enabled').checked = route.cors.enabled || false;
            document.getElementById('cors-origins').value = route.cors.allow_origins ? route.cors.allow_origins.join('\n') : '';
            document.getElementById('cors-methods').value = route.cors.allow_methods ? route.cors.allow_methods.join('\n') : '';
        }
        
        if (route.ssl) {
            document.getElementById('ssl-enabled').checked = route.ssl.enabled || false;
            document.getElementById('ssl-certificate').value = route.ssl.certificate || '';
            document.getElementById('ssl-private-key').value = route.ssl.private_key || '';
        }
        
        // Load regex configuration
        if (route.regex_config) {
            document.getElementById('regex-enabled').checked = route.regex_config.enabled || false;
            document.getElementById('regex-pattern').value = route.regex_config.pattern || '';
            document.getElementById('regex-modifier').value = route.regex_config.modifier || '';
            document.getElementById('regex-case-insensitive').checked = route.regex_config.case_insensitive || false;
        }
        
        // Load rewrite rules
        clearRewriteRules();
        if (route.rewrite_rules && route.rewrite_rules.length > 0) {
            route.rewrite_rules.forEach(rule => {
                addRewriteRule(rule.pattern || '', rule.replacement || '', rule.flag || '');
            });
        }
        
    } catch (error) {
        console.error('Failed to load route for edit:', error);
    }
}

async function saveRoute() {
    const serverId = document.getElementById('route-server-id').value;
    
    // Collect selected methods
    const methods = [];
    ['get', 'post', 'put', 'delete'].forEach(method => {
        const checkbox = document.getElementById(`method-${method}`);
        if (checkbox && checkbox.checked) {
            methods.push(method.toUpperCase());
        }
    });
    
    const route = {
        name: document.getElementById('route-name').value,
        path: document.getElementById('route-path').value,
        backend_id: document.getElementById('route-backend').value,
        methods: methods,
        strip_path: document.getElementById('route-strip-path').checked,
        preserve_host: document.getElementById('route-preserve-host').checked
    };
    
    // Add regex configuration if enabled
    if (document.getElementById('regex-enabled').checked) {
        route.regex_config = {
            enabled: true,
            pattern: document.getElementById('regex-pattern').value.trim() || null,
            modifier: document.getElementById('regex-modifier').value || '',
            case_insensitive: document.getElementById('regex-case-insensitive').checked
        };
    }
    
    // Add rewrite rules
    const rewriteRules = collectRewriteRules();
    if (rewriteRules.length > 0) {
        route.rewrite_rules = rewriteRules;
    }
    
    // Add rate limiting if enabled
    if (document.getElementById('rate-limiting-enabled').checked) {
        route.rate_limiting = {
            enabled: true,
            rate: document.getElementById('rate-limiting-rate').value,
            burst: parseInt(document.getElementById('rate-limiting-burst').value) || 5,
            status_code: parseInt(document.getElementById('rate-limiting-status').value) || 429
        };
    }
    
    // Add caching if enabled
    if (document.getElementById('caching-enabled').checked) {
        const cacheConfig = {
            enabled: true,
            ttl: document.getElementById('caching-ttl').value,
            key: document.getElementById('caching-key').value || '$scheme$request_method$host$request_uri'
        };
        
        // Advanced caching options
        const methodsEl = document.getElementById('cache-methods');
        if (methodsEl && methodsEl.value.trim()) {
            cacheConfig.methods = methodsEl.value.split(',').map(m => m.trim()).filter(m => m);
        }
        
        const statusCodesEl = document.getElementById('cache-status-codes');
        if (statusCodesEl && statusCodesEl.value.trim()) {
            cacheConfig.status_codes = statusCodesEl.value.split(',').map(s => parseInt(s.trim())).filter(s => !isNaN(s));
        }
        
        const ignoreHeadersEl = document.getElementById('cache-ignore-headers');
        if (ignoreHeadersEl && ignoreHeadersEl.value.trim()) {
            cacheConfig.ignore_headers = ignoreHeadersEl.value.split(',').map(h => h.trim()).filter(h => h);
        }
        
        const bypassEl = document.getElementById('cache-bypass');
        if (bypassEl && bypassEl.value.trim()) {
            cacheConfig.bypass = bypassEl.value.split('\n').map(b => b.trim()).filter(b => b);
        }
        
        const minUsesEl = document.getElementById('cache-min-uses');
        if (minUsesEl && minUsesEl.value) {
            cacheConfig.min_uses = parseInt(minUsesEl.value);
        }
        
        const inactiveEl = document.getElementById('cache-inactive');
        if (inactiveEl && inactiveEl.value.trim()) {
            cacheConfig.inactive = inactiveEl.value;
        }
        
        const maxSizeEl = document.getElementById('cache-max-size');
        if (maxSizeEl && maxSizeEl.value.trim()) {
            cacheConfig.max_size = maxSizeEl.value;
        }
        
        // Collect cache URI rules
        const uriRules = collectCacheRules();
        if (uriRules.length > 0) {
            cacheConfig.uri_rules = uriRules;
        }
        
        // Collect cache validity rules
        const validRules = [];
        document.querySelectorAll('#cache-valid-rules .cache-valid-rule').forEach(ruleEl => {
            const statusCodes = ruleEl.querySelector('.cache-valid-status-codes').value;
            const ttl = ruleEl.querySelector('.cache-valid-ttl').value;
            if (statusCodes && ttl) {
                validRules.push({
                    status_codes: statusCodes.split(',').map(s => parseInt(s.trim())).filter(s => !isNaN(s)),
                    ttl: ttl
                });
            }
        });
        if (validRules.length > 0) {
            cacheConfig.valid = validRules;
        }
        
        route.caching = cacheConfig;
    }
    
    // Add CORS if enabled
    if (document.getElementById('cors-enabled').checked) {
        route.cors = {
            enabled: true,
            allow_origins: document.getElementById('cors-origins').value.split('\n').filter(o => o.trim()),
            allow_methods: document.getElementById('cors-methods').value.split('\n').filter(m => m.trim())
        };
    }
    
    // Add SSL if enabled
    if (document.getElementById('ssl-enabled').checked) {
        route.ssl = {
            enabled: true,
            certificate: document.getElementById('ssl-certificate').value,
            private_key: document.getElementById('ssl-private-key').value
        };
    }
    
    try {
        const routeId = document.getElementById('route-id').value;
        
        if (routeId) {
            await apiCall('PUT', `/routes/${routeId}`, route);
            showAlert('success', 'Route updated successfully!');
        } else {
            await apiCall('POST', `/routes/server/${serverId}`, route);
            showAlert('success', 'Route created successfully!');
        }
        
        const modal = bootstrap.Modal.getInstance(document.getElementById('routeModal'));
        modal.hide();
        
        loadRoutes();
        loadStatus();
        
    } catch (error) {
        console.error('Failed to save route:', error);
    }
}

async function editRoute(serverId, routeId) {
    showRouteModal(serverId, routeId);
}

async function deleteRoute(serverId, routeId) {
    if (!confirm('Are you sure you want to delete this route?')) {
        return;
    }
    
    try {
        await apiCall('DELETE', `/routes/${routeId}`);
        showAlert('success', 'Route deleted successfully!');
        loadRoutes();
        loadStatus();
    } catch (error) {
        console.error('Failed to delete route:', error);
    }
}

// Configuration functions
async function loadGlobalConfig() {
    try {
        const config = await apiCall('GET', '/config');
        
        if (config.global) {
            document.getElementById('worker-processes').value = config.global.worker_processes || 'auto';
            document.getElementById('worker-connections').value = config.global.worker_connections || 1024;
            document.getElementById('keepalive-timeout').value = config.global.keepalive_timeout || '65s';
            document.getElementById('client-max-body-size').value = config.global.client_max_body_size || '1m';
            document.getElementById('server-tokens').checked = config.global.server_tokens || false;
            document.getElementById('gzip').checked = config.global.gzip !== false;
        }
        
    } catch (error) {
        console.error('Failed to load global config:', error);
    }
}

// Quick action functions
async function generateConfig() {
    try {
        await apiCall('POST', '/config/generate');
        showAlert('success', 'Nginx configuration generated successfully!');
    } catch (error) {
        console.error('Failed to generate config:', error);
    }
}

async function testConfig() {
    try {
        await apiCall('POST', '/config/test');
        showAlert('success', 'Nginx configuration test passed!');
    } catch (error) {
        console.error('Failed to test config:', error);
    }
}

async function reloadNginx() {
    try {
        await apiCall('POST', '/config/reload');
        showAlert('success', 'Nginx reloaded successfully!');
        loadStatus();
    } catch (error) {
        console.error('Failed to reload nginx:', error);
    }
}

// Utility functions
function showAlert(type, message) {
    const alertHtml = `
        <div class="alert alert-${type === 'error' ? 'danger' : type} alert-dismissible fade show" role="alert">
            ${message}
            <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
        </div>
    `;
    
    // Find the current section and prepend the alert
    const currentSectionElement = document.querySelector('.section[style*="block"]');
    if (currentSectionElement) {
        currentSectionElement.insertAdjacentHTML('afterbegin', alertHtml);
        
        // Auto-dismiss success alerts after 5 seconds
        if (type === 'success') {
            setTimeout(() => {
                const alert = currentSectionElement.querySelector('.alert');
                if (alert) {
                    const bsAlert = new bootstrap.Alert(alert);
                    bsAlert.close();
                }
            }, 5000);
        }
    }
}

// Health check toggle function
function toggleHealthCheckPanel(enabled) {
    const panel = document.getElementById('health-check-config');
    if (panel) {
        if (enabled) {
            panel.style.display = 'block';
        } else {
            panel.style.display = 'none';
        }
    }
}

// Rewrite rule management functions
let rewriteRuleCounter = 0;

function addRewriteRule(pattern = '', replacement = '', flag = '') {
    rewriteRuleCounter++;
    const container = document.getElementById('rewrite-rules-container');
    
    const ruleDiv = document.createElement('div');
    ruleDiv.className = 'border rounded p-3 mb-3 rewrite-rule-item';
    ruleDiv.setAttribute('data-rule-id', rewriteRuleCounter);
    
    ruleDiv.innerHTML = `
        <div class="row">
            <div class="col-md-4">
                <div class="mb-3">
                    <label class="form-label">Pattern</label>
                    <input type="text" class="form-control rewrite-pattern" value="${pattern}" placeholder="^/old-path/(.*)$">
                    <div class="form-text">Regex pattern to match</div>
                </div>
            </div>
            <div class="col-md-4">
                <div class="mb-3">
                    <label class="form-label">Replacement</label>
                    <input type="text" class="form-control rewrite-replacement" value="${replacement}" placeholder="/new-path/$1">
                    <div class="form-text">Replacement URL (use $1, $2 for captures)</div>
                </div>
            </div>
            <div class="col-md-3">
                <div class="mb-3">
                    <label class="form-label">Flag</label>
                    <select class="form-select rewrite-flag">
                        <option value="">None</option>
                        <option value="last" ${flag === 'last' ? 'selected' : ''}>last</option>
                        <option value="break" ${flag === 'break' ? 'selected' : ''}>break</option>
                        <option value="redirect" ${flag === 'redirect' ? 'selected' : ''}>redirect</option>
                        <option value="permanent" ${flag === 'permanent' ? 'selected' : ''}>permanent</option>
                    </select>
                </div>
            </div>
            <div class="col-md-1">
                <div class="mb-3">
                    <label class="form-label">&nbsp;</label>
                    <div>
                        <button type="button" class="btn btn-danger btn-sm" onclick="removeRewriteRule(${rewriteRuleCounter})">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </div>
            </div>
        </div>
    `;
    
    container.appendChild(ruleDiv);
}

function removeRewriteRule(ruleId) {
    const ruleElement = document.querySelector(`[data-rule-id="${ruleId}"]`);
    if (ruleElement) {
        ruleElement.remove();
    }
}

function collectRewriteRules() {
    const rules = [];
    const ruleElements = document.querySelectorAll('.rewrite-rule-item');
    
    ruleElements.forEach(element => {
        const pattern = element.querySelector('.rewrite-pattern').value.trim();
        const replacement = element.querySelector('.rewrite-replacement').value.trim();
        const flag = element.querySelector('.rewrite-flag').value;
        
        if (pattern && replacement) {
            const rule = { pattern, replacement };
            if (flag) {
                rule.flag = flag;
            }
            rules.push(rule);
        }
    });
    
    return rules;
}

function clearRewriteRules() {
    const container = document.getElementById('rewrite-rules-container');
    if (container) {
        container.innerHTML = '';
    }
}

// Cache rule management functions
let cacheRuleCounter = 0;

function addCacheRule(pattern = '', type = 'prefix', ttl = '', methods = '', statusCodes = '', noCache = false, description = '') {
    cacheRuleCounter++;
    const container = document.getElementById('cache-rules-container');
    
    const ruleDiv = document.createElement('div');
    ruleDiv.className = 'cache-rule-item border rounded p-3 mb-3';
    ruleDiv.setAttribute('data-rule-id', cacheRuleCounter);
    
    ruleDiv.innerHTML = `
        <div class="row">
            <div class="col-md-3">
                <div class="mb-3">
                    <label class="form-label">URI Pattern</label>
                    <input type="text" class="form-control cache-pattern" value="${pattern}" placeholder="/api/*">
                    <div class="form-text">URI pattern to match</div>
                </div>
            </div>
            <div class="col-md-2">
                <div class="mb-3">
                    <label class="form-label">Type</label>
                    <select class="form-select cache-type">
                        <option value="prefix" ${type === 'prefix' ? 'selected' : ''}>Prefix</option>
                        <option value="exact" ${type === 'exact' ? 'selected' : ''}>Exact</option>
                        <option value="regex" ${type === 'regex' ? 'selected' : ''}>Regex</option>
                    </select>
                </div>
            </div>
            <div class="col-md-2">
                <div class="mb-3">
                    <label class="form-label">TTL</label>
                    <input type="text" class="form-control cache-ttl" value="${ttl}" placeholder="30m">
                    <div class="form-text">Cache duration</div>
                </div>
            </div>
            <div class="col-md-2">
                <div class="mb-3">
                    <label class="form-label">Methods</label>
                    <input type="text" class="form-control cache-methods" value="${methods}" placeholder="GET,POST">
                    <div class="form-text">HTTP methods</div>
                </div>
            </div>
            <div class="col-md-2">
                <div class="mb-3">
                    <label class="form-label">Status Codes</label>
                    <input type="text" class="form-control cache-status-codes" value="${statusCodes}" placeholder="200,201">
                    <div class="form-text">Response codes</div>
                </div>
            </div>
            <div class="col-md-1">
                <div class="mb-3">
                    <label class="form-label">&nbsp;</label>
                    <div>
                        <button type="button" class="btn btn-danger btn-sm" onclick="removeCacheRule(${cacheRuleCounter})">
                            <i class="fas fa-trash"></i>
                        </button>
                    </div>
                </div>
            </div>
        </div>
        <div class="row">
            <div class="col-md-4">
                <div class="form-check">
                    <input class="form-check-input cache-no-cache" type="checkbox" ${noCache ? 'checked' : ''}>
                    <label class="form-check-label">Disable Cache</label>
                </div>
            </div>
            <div class="col-md-8">
                <div class="mb-3">
                    <label class="form-label">Description</label>
                    <input type="text" class="form-control cache-description" value="${description}" placeholder="Optional description for this rule">
                </div>
            </div>
        </div>
    `;
    
    container.appendChild(ruleDiv);
}

function removeCacheRule(ruleId) {
    const ruleElement = document.querySelector(`[data-rule-id="${ruleId}"]`);
    if (ruleElement) {
        ruleElement.remove();
    }
}

function clearCacheRules() {
    const container = document.getElementById('cache-rules-container');
    if (container) {
        container.innerHTML = '';
    }
    cacheRuleCounter = 0;
}

function collectCacheRules() {
    const rules = [];
    const ruleElements = document.querySelectorAll('.cache-rule-item');
    
    ruleElements.forEach(element => {
        const pattern = element.querySelector('.cache-pattern').value.trim();
        const type = element.querySelector('.cache-type').value;
        const ttl = element.querySelector('.cache-ttl').value.trim();
        const methodsStr = element.querySelector('.cache-methods').value.trim();
        const statusCodesStr = element.querySelector('.cache-status-codes').value.trim();
        const noCache = element.querySelector('.cache-no-cache').checked;
        const description = element.querySelector('.cache-description').value.trim();
        
        if (pattern) {
            const rule = {
                pattern: pattern,
                type: type,
                no_cache: noCache
            };
            
            if (ttl) rule.ttl = ttl;
            if (description) rule.description = description;
            
            if (methodsStr) {
                rule.methods = methodsStr.split(',').map(m => m.trim()).filter(m => m);
            }
            
            if (statusCodesStr) {
                rule.status_codes = statusCodesStr.split(',').map(c => parseInt(c.trim())).filter(c => !isNaN(c));
            }
            
            rules.push(rule);
        }
    });
    
    return rules;
}

// Add cache validity rule
function addCacheValidRule(statusCodes = '', ttl = '') {
    const container = document.getElementById('cache-valid-rules');
    if (!container) return;
    
    const ruleDiv = document.createElement('div');
    ruleDiv.className = 'cache-valid-rule row mb-2';
    
    ruleDiv.innerHTML = `
        <div class="col-md-5">
            <input type="text" class="form-control cache-valid-status-codes" 
                   placeholder="200,201,300-399" value="${statusCodes}"
                   title="HTTP status codes (e.g., 200,201,300-399)">
        </div>
        <div class="col-md-5">
            <input type="text" class="form-control cache-valid-ttl" 
                   placeholder="1h" value="${ttl}"
                   title="TTL for these status codes">
        </div>
        <div class="col-md-2">
            <button type="button" class="btn btn-outline-danger btn-sm" 
                    onclick="this.closest('.cache-valid-rule').remove()">
                <i class="fas fa-trash"></i>
            </button>
        </div>
    `;
    
    container.appendChild(ruleDiv);
}
