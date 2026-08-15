const apiClient = require('./helpers/api.client');
const crypto = require('crypto');

describe('BYOK / User-Controlled Key Management', () => {
    let userToken;
    let userId;
    let vaultDataId;

    const testEmail = `byoktest_${Date.now()}@example.com`;
    const testPassword = 'StrongPassword123!';
    const customKey = crypto.randomBytes(32).toString('hex'); // 64-char hex string

    beforeAll(async () => {
        // Register and login user
        await apiClient.post('/api/v1/auth/register', { email: testEmail, password: testPassword });
        const loginRes = await apiClient.post('/api/v1/auth/login', { email: testEmail, password: testPassword });
        userToken = loginRes.body.token;
        userId = loginRes.body.userId;
    });

    test('Should upload custom encryption key successfully', async () => {
        const res = await apiClient.put('/api/v1/user/key', { key: customKey }, userToken);
        expect(res.status).toBe(200);
        expect(res.body.message).toBe('Key updated successfully');
    });

    test('Should encrypt data with the custom key', async () => {
        const payload = {
            dataType: 'medical_data',
            data: 'This is top secret data encrypted with BYOK'
        };

        const res = await apiClient.post('/api/v1/vault/data', payload, userToken);
        expect(res.status).toBe(201);
        expect(res.body.id).toBeDefined();
        
        vaultDataId = res.body.id;
    });

    test('Should read data back using the custom key (transparently)', async () => {
        // Create an app, get token, etc. to test decryption.
        // Actually, for a quick test, we can just read the data via vault if we have a vault read endpoint?
        // Wait, there is no direct read endpoint for users (only apps via access).
        // Let's create an app, consent, and token to read it.
        const appRes = await apiClient.post('/api/v1/apps/register', { name: 'BYOK Reader App' }, userToken);
        const appId = appRes.body.appId;
        const appSecret = appRes.body.secret;

        const consentRes = await apiClient.post('/api/v1/consents', {
            appId,
            scopes: ['medical_data'],
            purpose: 'Testing BYOK',
            expiresAt: '2050-01-01T00:00:00Z'
        }, userToken);
        const consentId = consentRes.body.consentId;

        const tokenRes = await apiClient.post('/api/v1/oauth/token', {
            appId,
            appSecret,
            scope: 'medical_data'
        }, userToken);
        const appToken = tokenRes.body.token;

        const accessRes = await apiClient.post('/api/v1/access/data', {
            appId,
            secret: appSecret,
            scope: 'medical_data',
            vaultDataId
        }, appToken);

        expect(accessRes.status).toBe(200);
        expect(accessRes.body.data).toBe('This is top secret data encrypted with BYOK');
    });

    test('Should revoke custom encryption key (crypto-shred)', async () => {
        const res = await apiClient.delete('/api/v1/user/key', userToken);
        expect(res.status).toBe(200);
        expect(res.body.message).toContain('crypto-shredded');
    });

    test('Should fail to read data after key is revoked (decryption error)', async () => {
        // App tries to read it again
        const appRes = await apiClient.post('/api/v1/apps/register', { name: 'BYOK Failed Reader App' }, userToken);
        const appId = appRes.body.appId;
        const appSecret = appRes.body.secret;

        const consentRes = await apiClient.post('/api/v1/consents', {
            appId,
            scopes: ['medical_data'],
            purpose: 'Testing BYOK shred',
            expiresAt: '2050-01-01T00:00:00Z'
        }, userToken);
        
        const tokenRes = await apiClient.post('/api/v1/oauth/token', {
            appId,
            appSecret,
            scope: 'medical_data'
        }, userToken);
        const appToken = tokenRes.body.token;

        const accessRes = await apiClient.post('/api/v1/access/data', {
            appId,
            secret: appSecret,
            scope: 'medical_data',
            vaultDataId
        }, appToken);

        // Core maps decryption failure to a 403 Forbidden via AccessService
        expect(accessRes.status).toBe(403); 
    });
});
