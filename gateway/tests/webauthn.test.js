const request = require('supertest');
const app = require('../src/index');
const apiClient = require('./helpers/api.client');

describe('WebAuthn API Endpoints', () => {
    const testEmail = `webauthn_${Date.now()}@example.com`;

    beforeAll(async () => {
        // Register the user normally first to ensure they exist
        await apiClient.post('/api/v1/auth/register', { email: testEmail, password: 'StrongPassword123!' });
    });

    let sessionId;

    it('should begin WebAuthn registration', async () => {
        const res = await request(app)
            .post('/api/v1/auth/webauthn/register/begin')
            .send({ email: testEmail });
        
        expect(res.status).toBe(200);
        expect(res.body).toHaveProperty('options');
        expect(res.body).toHaveProperty('sessionId');
        sessionId = res.body.sessionId;
    });

    it('should finish WebAuthn registration (mock)', async () => {
        const res = await request(app)
            .post('/api/v1/auth/webauthn/register/finish')
            .send({ 
                email: testEmail, 
                sessionId: sessionId,
                response: { "mock": "data" }
            });
        
        expect(res.status).toBe(200);
        expect(res.body).toHaveProperty('status', 'success');
    });

    it('should begin WebAuthn login', async () => {
        const res = await request(app)
            .post('/api/v1/auth/webauthn/login/begin')
            .send({ email: testEmail });
        
        expect(res.status).toBe(200);
        expect(res.body).toHaveProperty('options');
        expect(res.body).toHaveProperty('sessionId');
        sessionId = res.body.sessionId;
    });

    it('should finish WebAuthn login and return a token (mock)', async () => {
        const res = await request(app)
            .post('/api/v1/auth/webauthn/login/finish')
            .send({ 
                email: testEmail, 
                sessionId: sessionId,
                response: { "mock": "data" }
            });
        
        expect(res.status).toBe(200);
        expect(res.body).toHaveProperty('token');
    });
});
