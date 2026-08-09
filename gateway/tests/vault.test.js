const request = require('supertest');
const app = require('../src/index');

describe('Vault Endpoints', () => {
    let tokenA = '';
    let tokenB = '';
    let vaultId = '';

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
    });

    it('should reject unauthenticated POST', async () => {
        const res = await request(app)
            .post('/api/v1/vault/data')
            .send({ dataType: 'test', data: 'data' });
        expect(res.statusCode).toBe(401);
    });

    it('should reject unauthenticated GET', async () => {
        const res = await request(app)
            .get('/api/v1/vault/data/some-id');
        expect(res.statusCode).toBe(401);
    });

    it('should reject missing dataType', async () => {
        const res = await request(app)
            .post('/api/v1/vault/data')
            .set('Authorization', `Bearer ${tokenA}`)
            .send({ data: 'my secret' });
        expect(res.statusCode).toBe(400);
    });

    it('should reject missing data', async () => {
        const res = await request(app)
            .post('/api/v1/vault/data')
            .set('Authorization', `Bearer ${tokenA}`)
            .send({ dataType: 'test' });
        expect(res.statusCode).toBe(400);
    });

    it('should allow User A to create vault data', async () => {
        const res = await request(app)
            .post('/api/v1/vault/data')
            .set('Authorization', `Bearer ${tokenA}`)
            .send({ dataType: 'medical_report', data: 'Patient has XYZ condition' });
        
        expect(res.statusCode).toBe(201);
        expect(res.body).toHaveProperty('id');
        expect(res.body.dataType).toBe('medical_report');
        vaultId = res.body.id;
    });

    it('should allow User A to retrieve their own vault data', async () => {
        const res = await request(app)
            .get(`/api/v1/vault/data/${vaultId}`)
            .set('Authorization', `Bearer ${tokenA}`);
        
        expect(res.statusCode).toBe(200);
        expect(res.body.dataType).toBe('medical_report');
        expect(res.body.data).toBe('Patient has XYZ condition');
    });

    it('should reject User B trying to retrieve User A data (IDOR prevention)', async () => {
        const res = await request(app)
            .get(`/api/v1/vault/data/${vaultId}`)
            .set('Authorization', `Bearer ${tokenB}`);
        
        expect(res.statusCode).toBe(404); // Returns 404 so existence is not leaked
    });

    it('should handle client trying to spoof userId in body safely', async () => {
        // Even if User B tries to specify User A's ID in body, it should be ignored by the DTO and Controller
        const res = await request(app)
            .post('/api/v1/vault/data')
            .set('Authorization', `Bearer ${tokenB}`)
            .send({ userId: 'spoofed_id', dataType: 'test', data: 'test data' });
        
        expect(res.statusCode).toBe(201);
        
        // Let's verify it belongs to B
        const spoofedId = res.body.id;
        
        const resB = await request(app)
            .get(`/api/v1/vault/data/${spoofedId}`)
            .set('Authorization', `Bearer ${tokenB}`);
            
        expect(resB.statusCode).toBe(200); // User B can access their own data
        
        const resA = await request(app)
            .get(`/api/v1/vault/data/${spoofedId}`)
            .set('Authorization', `Bearer ${tokenA}`);
            
        expect(resA.statusCode).toBe(404); // User A cannot access User B's data
    });
});
