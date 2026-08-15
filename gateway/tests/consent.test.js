const apiClient = require('./helpers/api.client');

describe('Consent Manager Endpoints', () => {
    let tokenA = '';
    let tokenB = '';
    let appId = '';
    let consentId = '';

    beforeAll(async () => {
        const setup = await apiClient.setupStandardEnvironment();
        tokenA = setup.userA.token;
        tokenB = setup.userB.token;
        appId = setup.app.id;
    });

    it('should reject unauthenticated consent creation', async () => {
        const res = await apiClient.post('/api/v1/consents', {});
        expect(res.statusCode).toBe(401);
    });

    it('should reject missing appId', async () => {
        const res = await apiClient.post('/api/v1/consents', { scopes: ['email'], purpose: 'Test', expiresAt: '2050-01-01T00:00:00Z' }, tokenA);
        expect(res.statusCode).toBe(400);
    });

    it('should reject empty scopes', async () => {
        const res = await apiClient.post('/api/v1/consents', { appId, scopes: [], purpose: 'Test', expiresAt: '2050-01-01T00:00:00Z' }, tokenA);
        expect(res.statusCode).toBe(400);
    });

    it('should reject unsupported scopes', async () => {
        const res = await apiClient.post('/api/v1/consents', { appId, scopes: ['email', 'magic_scope'], purpose: 'Test', expiresAt: '2050-01-01T00:00:00Z' }, tokenA);
        expect(res.statusCode).toBe(400);
    });

    it('should reject past expiration', async () => {
        const res = await apiClient.post('/api/v1/consents', { appId, scopes: ['email'], purpose: 'Test', expiresAt: '2020-01-01T00:00:00Z' }, tokenA);
        expect(res.statusCode).toBe(400);
    });

    it('should allow User A to create consent with policies', async () => {
        const res = await apiClient.post('/api/v1/consents', { 
            appId, 
            scopes: ['email', 'profile'], 
            purpose: 'Test Personalization', 
            expiresAt: '2050-01-01T00:00:00Z',
            policies: { "time_of_day": "00:00-23:59" } 
        }, tokenA);
        
        expect(res.statusCode).toBe(201);
        expect(res.body).toHaveProperty('consentId');
        expect(res.body.scopes).toContain('email');
        expect(res.body.status).toBe('ACTIVE');
        expect(res.body.policies).toBeDefined();
        expect(res.body.policies.time_of_day).toBe('00:00-23:59');
        consentId = res.body.consentId;
    });

    it('should allow User A to retrieve their own consent', async () => {
        const res = await apiClient.get(`/api/v1/consents/${consentId}`, tokenA);
        
        expect(res.statusCode).toBe(200);
        expect(res.body.status).toBe('ACTIVE');
    });

    it('should reject User B trying to retrieve User A consent (IDOR)', async () => {
        const res = await apiClient.get(`/api/v1/consents/${consentId}`, tokenB);
        
        expect(res.statusCode).toBe(404); // Not Found to prevent leaking existence
    });

    it('should allow User A to list their consents', async () => {
        const res = await apiClient.get('/api/v1/consents', tokenA);
        
        expect(res.statusCode).toBe(200);
        expect(Array.isArray(res.body)).toBe(true);
        expect(res.body.length).toBeGreaterThan(0);
        expect(res.body[0].consentId).toBe(consentId);
    });

    it('should reject User B trying to revoke User A consent (IDOR)', async () => {
        const res = await apiClient.post(`/api/v1/consents/${consentId}/revoke`, {}, tokenB);
        
        expect(res.statusCode).toBe(404);
    });

    it('should allow User A to revoke their own consent', async () => {
        const res = await apiClient.post(`/api/v1/consents/${consentId}/revoke`, {}, tokenA);
        expect(res.statusCode).toBe(200);

        // Fetch it again to check status
        const fetchRes = await apiClient.get(`/api/v1/consents/${consentId}`, tokenA);
        expect(fetchRes.body.status).toBe('REVOKED');
        expect(fetchRes.body.revokedAt).toBeDefined();
    });
});
