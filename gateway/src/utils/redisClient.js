const Redis = require('ioredis');
const logger = require('../logger');

let client = null;

function getClient() {
    if (client) return client;

    const url = process.env.REDIS_URL || 'redis://localhost:6379';
    client = new Redis(url, {
        maxRetriesPerRequest: 1,
        enableReadyCheck: false,
        lazyConnect: true,
        retryStrategy(times) {
            if (times > 3) return null; // stop retrying after 3 attempts
            return Math.min(times * 200, 2000);
        }
    });

    client.on('error', (err) => {
        logger.error('redis_error', null, { error: err.message });
    });

    client.on('connect', () => {
        logger.info('redis_connected', null, { url });
    });

    client.connect().catch(() => {
        // Connection failure is non-fatal; rate limiter will fail open.
    });

    return client;
}

module.exports = { getClient };
