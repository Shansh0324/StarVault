const request = require('supertest');
const app = require('../src/index');

describe('Tokenization API', () => {
    let userToken;
    let appId;
    let appSecret = 'TestSecret123!';
    let generatedToken;

    beforeAll(async () => {
        // Register user and get JWT
        const uniqueEmail = `tokenuser_${Date.now()}@test.com`;
        await request(app)
            .post('/api/v1/auth/register')
            .send({
                email: uniqueEmail,
                password: 'Password123!',
                firstName: 'Test',
                lastName: 'User'
            });

        const loginRes = await request(app)
            .post('/api/v1/auth/login')
            .send({
                email: uniqueEmail,
                password: 'Password123!'
            });
        userToken = loginRes.body.token;

        // Register App
        const appRes = await request(app)
            .post('/api/v1/apps/register')
            .set('Authorization', `Bearer ${userToken}`)
            .send({
                name: 'Token Test App'
            });
        appId = appRes.body.appId;
        appSecret = appRes.body.secret;

        // Give Consent
        await request(app)
            .post('/api/v1/consents')
            .set('Authorization', `Bearer ${userToken}`)
            .send({
                appId: appId,
                scopes: ['medical_data'],
                purpose: 'For token tests',
                expiresAt: new Date(Date.now() + 86400000).toISOString()
            });
    });

    test('Should issue a token successfully', async () => {
        const res = await request(app)
            .post('/api/v1/oauth/token')
            .set('Authorization', `Bearer ${userToken}`)
            .send({
                appId: appId,
                appSecret: appSecret,
                scope: 'medical_data'
            });

        expect(res.status).toBe(200);
        expect(res.body).toHaveProperty('token');
        expect(res.body).toHaveProperty('expiresIn');
        generatedToken = res.body.token;
    });

    test('Should reject token issuance with invalid secret', async () => {
        const res = await request(app)
            .post('/api/v1/oauth/token')
            .set('Authorization', `Bearer ${userToken}`)
            .send({
                appId: appId,
                appSecret: 'wrongsecret',
                scope: 'medical_data'
            });

        expect(res.status).toBe(401);
    });

    test('Should access vault data using short-lived access token', async () => {
        // Create vault data first
        const vaultRes = await request(app)
            .post('/api/v1/vault/data')
            .set('Authorization', `Bearer ${userToken}`)
            .send({
                dataType: 'medical_data',
                data: 'BASE64ENCSTUFF'
            });
        const vaultId = vaultRes.body.id;

        // Access with Token
        const accessRes = await request(app)
            .post('/api/v1/access/data')
            .set('Authorization', `Bearer ${generatedToken}`)
            .send({
                vaultDataId: vaultId,
                scope: 'medical_data'
            });

        console.log("accessRes.body:", accessRes.body);
        expect(accessRes.status).toBe(200);
        expect(accessRes.body).toHaveProperty('data');
    });

    test('Should successfully revoke the access token', async () => {
        const res = await request(app)
            .post('/api/v1/oauth/revoke')
            .set('Authorization', `Bearer ${userToken}`)
            .send({
                token: generatedToken
            });

        expect(res.status).toBe(204);
    });

    test('Should fail to access data with revoked token', async () => {
        const accessRes = await request(app)
            .post('/api/v1/access/data')
            .set('Authorization', `Bearer ${generatedToken}`)
            .send({
                vaultDataId: 'some-id',
                scope: 'medical_data'
            });

        expect(accessRes.status).toBe(401); // revoked token is 401 Unauthorized
    });
});
