const request = require('supertest');
const app = require('../src/index');

describe('App Registration Endpoints', () => {
    let appId = '';
    let secret = '';

    it('should successfully register a new application', async () => {
        const res = await request(app)
            .post('/api/v1/apps/register')
            .send({ name: 'My First Healthcare App' });
        
        expect(res.statusCode).toBe(201);
        expect(res.body).toHaveProperty('appId');
        expect(res.body).toHaveProperty('secret');
        expect(res.body.name).toBe('My First Healthcare App');
        expect(res.body).not.toHaveProperty('secret_hash');
        
        appId = res.body.appId;
        secret = res.body.secret;
    });

    it('should generate different secrets and IDs for multiple registrations', async () => {
        const res = await request(app)
            .post('/api/v1/apps/register')
            .send({ name: 'My Second Healthcare App' });
        
        expect(res.statusCode).toBe(201);
        expect(res.body.appId).not.toBe(appId);
        expect(res.body.secret).not.toBe(secret);
    });

    it('should reject registration with missing name', async () => {
        const res = await request(app)
            .post('/api/v1/apps/register')
            .send({});
        expect(res.statusCode).toBe(400);
    });

    it('should reject registration with empty name', async () => {
        const res = await request(app)
            .post('/api/v1/apps/register')
            .send({ name: '   ' });
        expect(res.statusCode).toBe(400);
    });

    it('should reject registration with excessively long name', async () => {
        const longName = 'A'.repeat(101);
        const res = await request(app)
            .post('/api/v1/apps/register')
            .send({ name: longName });
        expect(res.statusCode).toBe(400);
    });

    it('should safely reject invalid JSON', async () => {
        const res = await request(app)
            .post('/api/v1/apps/register')
            .set('Content-Type', 'application/json')
            .send('{"name": "broken');
        expect(res.statusCode).toBe(400); // Express json parser handles this gracefully
    });
});
