const { connect, StringCodec } = require('nats');
const logger = require('../logger');

class NatsClient {
    constructor() {
        this.nc = null;
        this.js = null;
        this.sc = StringCodec();
    }

    async connect() {
        try {
            const url = process.env.NATS_URL || 'nats://localhost:4222';
            this.nc = await connect({ servers: url });
            logger.info('nats_connected', null, { url });
            
            // Initialize JetStream
            this.js = this.nc.jetstream();
        } catch (err) {
            logger.error('nats_connection_failed', null, { error: err.message });
        }
    }

    async close() {
        if (this.nc) {
            await this.nc.close();
            logger.info('nats_closed');
        }
    }
}

const natsClient = new NatsClient();
module.exports = natsClient;
