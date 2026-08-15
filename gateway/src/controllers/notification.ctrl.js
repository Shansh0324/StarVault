const natsClient = require('../utils/natsClient');
const logger = require('../logger');

class NotificationController {
    static async streamAccessEvents(req, res) {
        const userId = req.user.userId;

        // Set headers for SSE
        res.setHeader('Content-Type', 'text/event-stream');
        res.setHeader('Cache-Control', 'no-cache');
        res.setHeader('Connection', 'keep-alive');
        // Tell the client to retry every 10 seconds if disconnected
        res.write('retry: 10000\n\n');

        if (!natsClient.js) {
            logger.error('sse_nats_unavailable', req.id, { userId });
            res.write(`data: ${JSON.stringify({ error: 'Notification service unavailable' })}\n\n`);
            return res.end();
        }

        const subject = `notifications.access.${userId}`;
        logger.info('sse_client_connected', req.id, { userId, subject });

        let consumer;
        let iterator;

        try {
            // Ephemeral consumer on JetStream for this specific user
            consumer = await natsClient.js.consumers.get('NOTIFICATIONS', {
                deliver_policy: 'new', // Only get new messages
                filter_subject: subject,
                ack_policy: 'none' // Fire and forget for SSE
            });
            
            iterator = await consumer.consume();
            
            // Handle client disconnect
            req.on('close', () => {
                logger.info('sse_client_disconnected', req.id, { userId });
                if (iterator) {
                    iterator.stop();
                }
            });

            for await (const msg of iterator) {
                const dataStr = natsClient.sc.decode(msg.data);
                res.write(`data: ${dataStr}\n\n`);
                
                // Flush to ensure it's sent immediately
                if (res.flush) {
                    res.flush();
                }
            }
        } catch (error) {
            // Consumer might not exist if no stream is configured, fallback to core NATS sub
            logger.warn('sse_jetstream_error_fallback_to_core', req.id, { error: error.message });
            
            try {
                const sub = natsClient.nc.subscribe(subject, {
                    callback: (err, msg) => {
                        if (err) {
                            logger.error('sse_nats_sub_error', req.id, { error: err.message });
                            return;
                        }
                        const dataStr = natsClient.sc.decode(msg.data);
                        res.write(`data: ${dataStr}\n\n`);
                    }
                });

                req.on('close', () => {
                    logger.info('sse_client_disconnected', req.id, { userId });
                    sub.unsubscribe();
                });
            } catch (fallbackError) {
                 logger.error('sse_fallback_error', req.id, { error: fallbackError.message });
                 res.write(`data: ${JSON.stringify({ error: 'Notification stream failed' })}\n\n`);
                 res.end();
            }
        }
    }
}

module.exports = NotificationController;
