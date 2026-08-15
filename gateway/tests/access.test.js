const apiClient = require('./helpers/api.client');
require('dotenv').config({ path: require('path').resolve(__dirname, '../../.env') });

describe('Access Gateway Endpoints', () => {
    let tokenA, tokenB;
    let appId, appSecret;
    let vaultDataId, consentId;

    beforeAll(async () => {
        const setup = await apiClient.setupStandardEnvironment();
        tokenA = setup.userA.token;
        tokenB = setup.userB.token;
        appId = setup.app.id;
        appSecret = setup.app.secret;

        // Create Vault Data for User A
        const vaultRes = await apiClient.post('/api/v1/vault/data', { dataType: 'medical_data', data: 'Secret patient record' }, tokenA);
        vaultDataId = vaultRes.body.id;

        // Create Consent for User A to App
        const consentRes = await apiClient.post('/api/v1/consents', { appId, scopes: ['medical_data', 'email'], purpose: 'Test access', expiresAt: '2050-01-01T00:00:00Z' }, tokenA);
        consentId = consentRes.body.consentId;
    });

    it('should allow valid access and return decrypted data', async () => {
        const res = await apiClient.post('/api/v1/access/data', { appId, secret: appSecret, scope: 'medical_data', vaultDataId }, tokenA);
        expect(res.statusCode).toBe(200);
        expect(res.body.data).toBe('Secret patient record');

        // Poll for audit log (NATS worker inserts to DB as PENDING_BATCH)
        const coreUrl = process.env.CORE_URL || 'http://localhost:8080';
        let audit;
        let auditFound = false;
        for (let i = 0; i < 10; i++) {
            await new Promise(resolve => setTimeout(resolve, 500));
            const auditRes = await fetch(`${coreUrl}/internal/audits/latest?appId=${appId}`);
            if (auditRes.status === 200) {
                audit = await auditRes.json();
                auditFound = true;
                break;
            }
        }
        expect(auditFound).toBe(true);
        
        expect(audit.action).toBe('ACCESS_GRANTED');
        expect(audit.appId).toBe(appId);
        // With Merkle batching, status is PENDING_BATCH until the BatchWorker runs
        expect(['PENDING_BATCH', 'COMMITTED']).toContain(audit.blockchainStatus);
        expect(audit.eventHash).not.toBe('');
    }, 10000);

    it('should reject access with invalid app secret', async () => {
        const res = await apiClient.post('/api/v1/access/data', { appId, secret: 'wrong-secret', scope: 'medical_data', vaultDataId }, tokenA);
        expect(res.statusCode).toBe(401);
    });

    it('should reject access if no consent exists', async () => {
        // User B has not granted consent
        const res = await apiClient.post('/api/v1/access/data', { appId, secret: appSecret, scope: 'medical_data', vaultDataId }, tokenB);
        expect(res.statusCode).toBe(403);
    });

    it('should reject access if scope does not match vault data type', async () => {
        const res = await apiClient.post('/api/v1/access/data', { appId, secret: appSecret, scope: 'email', vaultDataId }, tokenA);
        expect(res.statusCode).toBe(403);
    });

    it('should reject access if consent is revoked', async () => {
        await apiClient.post(`/api/v1/consents/${consentId}/revoke`, {}, tokenA);
        const res = await apiClient.post('/api/v1/access/data', { appId, secret: appSecret, scope: 'medical_data', vaultDataId }, tokenA);
        expect(res.statusCode).toBe(403);
    });
});
