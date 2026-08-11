const CORE_URL = process.env.CORE_URL || 'http://localhost:8080';
const logger = require('../logger');

/**
 * A wrapper around fetch for Core communication that adds:
 * - 5 second timeouts
 * - X-Request-ID header propagation
 * - Error handling for network failures
 */
async function coreFetch(req, path, options = {}) {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000); // 5s timeout

    try {
        const headers = {
            'Content-Type': 'application/json',
            'X-Request-ID': req.id,
            ...(options.headers || {})
        };

        const response = await fetch(`${CORE_URL}${path}`, {
            ...options,
            headers,
            signal: controller.signal
        });

        // Some core endpoints might return empty body on 204 or errors
        let data = null;
        try {
            data = await response.json();
        } catch (e) {
            // Ignore JSON parse errors for empty responses
        }

        return { response, data };
    } catch (err) {
        logger.error('core_fetch_failed', req.id, { path, error: err.message });
        if (err.name === 'AbortError') {
            throw new Error('Core request timed out');
        }
        throw new Error('Failed to communicate with Core');
    } finally {
        clearTimeout(timeoutId);
    }
}

module.exports = coreFetch;
