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

    it('should enforce rate limits on sensitive endpoints', async () => {
        // Our rate limiter allows 50 requests per minute per IP.
        // Let's send 51 requests.
        let statusCodes = [];
        for (let i = 0; i < 52; i++) {
            const res = await request(app).post('/api/v1/auth/login').send({ email: 'fake@example.com', password: 'fake' });
            statusCodes.push(res.statusCode);
        }

        // The first 50 should be something like 400 (validation error) or 500
        // But the 52nd request should definitely be 429
        expect(statusCodes[statusCodes.length - 1]).toBe(429);
        expect(statusCodes.includes(429)).toBe(true);
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
