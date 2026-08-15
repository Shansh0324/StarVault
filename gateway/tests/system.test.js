const request = require('supertest');
const app = require('../src/index');

describe('System Hardening Endpoints', () => {
    it('should return 200 on /health/live', async () => {
        const res = await request(app).get('/health/live');
        expect(res.statusCode).toBe(200);
        expect(res.body.status).toBe('Gateway is alive');
    });

    it('should return 200 on /health/ready', async () => {
        const res = await request(app).get('/health/ready');
        expect(res.statusCode).toBe(200);
        expect(res.body.status).toBe('Gateway is ready');
    });

    it('should fail open on rate limits when Redis is unavailable (test env)', async () => {
        // With Redis-backed rate limiting, the limiter fails open when Redis
        // is unreachable (which it is in the unit test environment).
        // This validates the fail-open safety behavior.
        let statusCodes = [];
        for (let i = 0; i < 52; i++) {
            const res = await request(app).post('/api/v1/auth/login').send({ email: 'fake@example.com', password: 'fake' });
            statusCodes.push(res.statusCode);
        }

        // Without Redis, no request should be rate-limited (fail-open)
        expect(statusCodes.includes(429)).toBe(false);
    });

    it('should assign a correlation ID if missing', async () => {
        const res = await request(app).get('/health/live');
        expect(res.headers['x-request-id']).toBeDefined();
        expect(res.headers['x-request-id'].length).toBeGreaterThan(10);
    });

    it('should propagate correlation ID if provided', async () => {
        const customId = 'my-custom-id-123';
        const res = await request(app).get('/health/live').set('x-request-id', customId);
        expect(res.headers['x-request-id']).toBe(customId);
    });
});
