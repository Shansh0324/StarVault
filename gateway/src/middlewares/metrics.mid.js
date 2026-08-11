const metrics = {
    http_requests_total: new Map(),
    http_request_duration_ms_sum: new Map(),
    http_request_duration_ms_count: new Map()
};

function getLabelsString(labels) {
    const pairs = Object.entries(labels).map(([k, v]) => `${k}="${v}"`);
    return pairs.length > 0 ? `{${pairs.join(',')}}` : '';
}

function recordRequest(method, path, status, duration) {
    const labels = { method, status: String(status) };
    const labelStr = getLabelsString(labels);
    
    // Increment total requests
    metrics.http_requests_total.set(labelStr, (metrics.http_requests_total.get(labelStr) || 0) + 1);
    
    // Record duration sum and count for average
    metrics.http_request_duration_ms_sum.set(labelStr, (metrics.http_request_duration_ms_sum.get(labelStr) || 0) + duration);
    metrics.http_request_duration_ms_count.set(labelStr, (metrics.http_request_duration_ms_count.get(labelStr) || 0) + 1);
}

function metricsMiddleware(req, res, next) {
    if (req.path === '/metrics' || req.path.startsWith('/health')) {
        return next();
    }
    
    const start = Date.now();
    res.on('finish', () => {
        const duration = Date.now() - start;
        recordRequest(req.method, req.path, res.statusCode, duration);
    });
    next();
}

function getMetrics(req, res) {
    let output = '';
    
    output += '# HELP http_requests_total Total HTTP requests\n';
    output += '# TYPE http_requests_total counter\n';
    for (const [labels, count] of metrics.http_requests_total.entries()) {
        output += `http_requests_total${labels} ${count}\n`;
    }
    
    output += '# HELP http_request_duration_ms_sum Total duration of HTTP requests in ms\n';
    output += '# TYPE http_request_duration_ms_sum counter\n';
    for (const [labels, sum] of metrics.http_request_duration_ms_sum.entries()) {
        output += `http_request_duration_ms_sum${labels} ${sum}\n`;
    }

    output += '# HELP http_request_duration_ms_count Count of HTTP requests for duration average\n';
    output += '# TYPE http_request_duration_ms_count counter\n';
    for (const [labels, count] of metrics.http_request_duration_ms_count.entries()) {
        output += `http_request_duration_ms_count${labels} ${count}\n`;
    }

    res.set('Content-Type', 'text/plain');
    res.send(output);
}

module.exports = { metricsMiddleware, getMetrics };
