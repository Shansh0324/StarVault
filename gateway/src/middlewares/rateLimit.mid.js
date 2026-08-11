const logger = require('../logger');

const rateLimitMap = new Map();

// Configurable limit (e.g. 50 requests per windowMs)
function rateLimiter(windowMs = 60000, maxRequests = 50) {
    return (req, res, next) => {
        const ip = req.ip || req.connection.remoteAddress;
        const now = Date.now();
        
        let record = rateLimitMap.get(ip);
        if (!record || now > record.resetTime) {
            record = { count: 1, resetTime: now + windowMs };
            rateLimitMap.set(ip, record);
        } else {
            record.count++;
        }

        if (record.count > maxRequests) {
            logger.warn('rate_limit_exceeded', req.id, { ip });
            return res.status(429).json({ error: { code: 'TOO_MANY_REQUESTS', message: 'Rate limit exceeded' } });
        }

        // Clean up expired entries periodically (simplistic cleanup for MVP)
        if (Math.random() < 0.01) {
            for (let [key, val] of rateLimitMap.entries()) {
                if (Date.now() > val.resetTime) {
                    rateLimitMap.delete(key);
                }
            }
        }

        next();
    };
}

module.exports = rateLimiter;
