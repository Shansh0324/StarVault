const { getClient } = require('../utils/redisClient');
const logger = require('../logger');

// Atomic INCR + EXPIRE via Lua to avoid race between the two commands.
const LUA_SCRIPT = `
local count = redis.call('INCR', KEYS[1])
if count == 1 then
    redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return count
`;

function rateLimiter(windowMs = 60000, maxRequests = 50) {
    return async (req, res, next) => {
        const ip = req.ip || req.connection.remoteAddress;
        const key = `rl:${ip}`;

        try {
            const redis = getClient();
            const count = await redis.eval(LUA_SCRIPT, 1, key, windowMs);

            if (count > maxRequests) {
                logger.warn('rate_limit_exceeded', req.id, { ip });
                return res.status(429).json({ error: { code: 'TOO_MANY_REQUESTS', message: 'Rate limit exceeded' } });
            }
        } catch (err) {
            // Fail open: Redis down should not block requests.
            logger.error('rate_limit_redis_fallback', req.id, { error: err.message, ip });
        }

        next();
    };
}

module.exports = rateLimiter;
