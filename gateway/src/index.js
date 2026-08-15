const express = require('express');
const dotenv = require('dotenv');
const path = require('path');
const logger = require('./logger');
const correlationMiddleware = require('./middlewares/correlation.mid');
const rateLimiter = require('./middlewares/rateLimit.mid');
const { metricsMiddleware, getMetrics } = require('./middlewares/metrics.mid');

dotenv.config({ path: path.resolve(__dirname, '../../.env') });

if (!process.env.JWT_SECRET) {
    logger.error('startup_failed', null, { reason: 'JWT_SECRET is missing' });
    process.exit(1);
}

const authRoutes = require('./routes/auth.routes');
const vaultRoutes = require('./routes/vault.routes');
const appRoutes = require('./routes/app.routes');
const consentRoutes = require('./routes/consent.routes');
const accessRoutes = require('./routes/access.routes');
const tokenRoutes = require('./routes/token.routes');
const notificationRoutes = require('./routes/notification.routes');
const userRoutes = require('./routes/user.routes');
const natsClient = require('./utils/natsClient');
const app = express();

// Security Headers
app.use((req, res, next) => {
    res.setHeader('X-Content-Type-Options', 'nosniff');
    res.setHeader('X-Frame-Options', 'DENY');
    res.setHeader('Strict-Transport-Security', 'max-age=31536000; includeSubDomains');
    next();
});

// Payload limits
app.use(express.json({ limit: '100kb' }));

// Correlation ID
app.use(correlationMiddleware);

// Metrics
app.use(metricsMiddleware);

// Global request logging (optional, could be noisy, keeping it minimal)
app.use((req, res, next) => {
    logger.info('request_started', req.id, { method: req.method, path: req.path });
    next();
});

const PORT = process.env.GATEWAY_PORT || 3000;

let isShuttingDown = false;

app.get('/health/live', (req, res) => {
    res.status(200).json({ status: 'Gateway is alive' });
});

app.get('/health/ready', (req, res) => {
    if (isShuttingDown) {
        return res.status(503).json({ status: 'Gateway is shutting down' });
    }
    res.status(200).json({ status: 'Gateway is ready' });
});

app.get('/metrics', getMetrics);

// Mount routes (apply rate limiting to sensitive routes)
const standardLimit = rateLimiter(60000, 50); // 50 req/min

app.use('/api/v1/auth', standardLimit, authRoutes);
app.use('/api/v1/vault', vaultRoutes);
app.use('/api/v1/apps', standardLimit, appRoutes);
app.use('/api/v1/consents', consentRoutes);
app.use('/api/v1/access', standardLimit, accessRoutes);
app.use('/api/v1/oauth', standardLimit, tokenRoutes);
app.use('/api/v1/notifications', standardLimit, notificationRoutes);
app.use('/api/v1/user', standardLimit, userRoutes);

if (require.main === module) {
    natsClient.connect().then(() => {
        const server = app.listen(PORT, () => {
            logger.info('startup_complete', null, { port: PORT });
        });

        const shutdown = () => {
            logger.info('shutdown_initiated');
            isShuttingDown = true;
            
            // Give ongoing requests 5 seconds to finish
            setTimeout(() => {
                logger.error('shutdown_forced');
                process.exit(1);
            }, 5000);

            server.close(async () => {
                await natsClient.close();
                logger.info('shutdown_complete');
                process.exit(0);
            });
        };

        process.on('SIGTERM', shutdown);
        process.on('SIGINT', shutdown);
    });
}

module.exports = app;
