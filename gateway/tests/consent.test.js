const request = require('supertest');
const app = require('../src/index');

describe('Consent Manager Endpoints', () => {
    let tokenA = '';
    let tokenB = '';
    let appId = '';
    let consentId = '';

    beforeAll(async () => {
        // Register User A
        const emailA = `usera_${Date.now()}@example.com`;
        await request(app).post('/api/v1/auth/register').send({ email: emailA, password: 'password123' });
        const resA = await request(app).post('/api/v1/auth/login').send({ email: emailA, password: 'password123' });
        tokenA = resA.body.token;

        // Register User B
        const emailB = `userb_${Date.now()}@example.com`;
        await request(app).post('/api/v1/auth/register').send({ email: emailB, password: 'password123' });
        const resB = await request(app).post('/api/v1/auth/login').send({ email: emailB, password: 'password123' });
        tokenB = resB.body.token;

        // Register App
        const resApp = await request(app).post('/api/v1/apps/register').send({ name: 'Consent Test App' });
        appId = resApp.body.appId;
    });

    it('should reject unauthenticated consent creation', async () => {
        const res = await request(app).post('/api/v1/consents').send({});
        expect(res.statusCode).toBe(401);
    });

    it('should reject missing appId', async () => {
        const res = await request(app)
            .post('/api/v1/consents')
            .set('Authorization', `Bearer ${tokenA}`)
            .send({ scopes: ['email'], purpose: 'Test', expiresAt: '2050-01-01T00:00:00Z' });
        expect(res.statusCode).toBe(400);
    });

    it('should reject empty scopes', async () => {
        const res = await request(app)
            .post('/api/v1/consents')
            .set('Authorization', `Bearer ${tokenA}`)
            .send({ appId, scopes: [], purpose: 'Test', expiresAt: '2050-01-01T00:00:00Z' });
        expect(res.statusCode).toBe(400);
    });

    it('should reject unsupported scopes', async () => {
        const res = await request(app)
            .post('/api/v1/consents')
            .set('Authorization', `Bearer ${tokenA}`)
            .send({ appId, scopes: ['email', 'magic_scope'], purpose: 'Test', expiresAt: '2050-01-01T00:00:00Z' });
        expect(res.statusCode).toBe(400);
    });

    it('should reject past expiration', async () => {
        const res = await request(app)
            .post('/api/v1/consents')
            .set('Authorization', `Bearer ${tokenA}`)
            .send({ appId, scopes: ['email'], purpose: 'Test', expiresAt: '2020-01-01T00:00:00Z' });
        expect(res.statusCode).toBe(400);
    });

    it('should allow User A to create consent', async () => {
        const res = await request(app)
            .post('/api/v1/consents')
            .set('Authorization', `Bearer ${tokenA}`)
            .send({ appId, scopes: ['email', 'profile'], purpose: 'Test Personalization', expiresAt: '2050-01-01T00:00:00Z' });
        
        expect(res.statusCode).toBe(201);
        expect(res.body).toHaveProperty('consentId');
        expect(res.body.scopes).toContain('email');
        expect(res.body.status).toBe('ACTIVE');
        consentId = res.body.consentId;
    });

    it('should allow User A to retrieve their own consent', async () => {
        const res = await request(app)
            .get(`/api/v1/consents/${consentId}`)
            .set('Authorization', `Bearer ${tokenA}`);
        
        expect(res.statusCode).toBe(200);
        expect(res.body.status).toBe('ACTIVE');
    });

    it('should reject User B trying to retrieve User A consent (IDOR)', async () => {
        const res = await request(app)
            .get(`/api/v1/consents/${consentId}`)
            .set('Authorization', `Bearer ${tokenB}`);
        
        expect(res.statusCode).toBe(404); // Not Found to prevent leaking existence
    });

    it('should allow User A to list their consents', async () => {
        const res = await request(app)
            .get('/api/v1/consents')
            .set('Authorization', `Bearer ${tokenA}`);
        
        expect(res.statusCode).toBe(200);
        expect(Array.isArray(res.body)).toBe(true);
        expect(res.body.length).toBeGreaterThan(0);
        expect(res.body[0].consentId).toBe(consentId);
    });

    it('should reject User B trying to revoke User A consent (IDOR)', async () => {
        const res = await request(app)
            .post(`/api/v1/consents/${consentId}/revoke`)
            .set('Authorization', `Bearer ${tokenB}`);
        
        expect(res.statusCode).toBe(404);
    });

    it('should allow User A to revoke their own consent', async () => {
        const res = await request(app)
            .post(`/api/v1/consents/${consentId}/revoke`)
            .set('Authorization', `Bearer ${tokenA}`);
        
        expect(res.statusCode).toBe(200);

        // Fetch it again to check status
        const fetchRes = await request(app)
            .get(`/api/v1/consents/${consentId}`)
            .set('Authorization', `Bearer ${tokenA}`);
        expect(fetchRes.body.status).toBe('REVOKED');
        expect(fetchRes.body.revokedAt).toBeDefined();
    });
});
