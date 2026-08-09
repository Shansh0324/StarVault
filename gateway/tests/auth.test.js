const request = require('supertest');
const app = require('../src/index');

describe('Auth Endpoints', () => {
    const testEmail = `testuser_${Date.now()}@example.com`;
    const testPassword = 'StrongPassword123';
    let userToken = '';

    it('should successfully register a new user', async () => {
        const res = await request(app)
            .post('/api/v1/auth/register')
            .send({ email: testEmail, password: testPassword });
        
        expect(res.statusCode).toBe(201);
        expect(res.body).toHaveProperty('userId');
    });

    it('should return 409 for duplicate email registration', async () => {
        const res = await request(app)
            .post('/api/v1/auth/register')
            .send({ email: testEmail, password: testPassword });
        
        expect(res.statusCode).toBe(409);
        expect(res.body.error.message).toContain('Email already exists');
    });

    it('should successfully login and return a JWT', async () => {
        const res = await request(app)
            .post('/api/v1/auth/login')
            .send({ email: testEmail, password: testPassword });
        
        expect(res.statusCode).toBe(200);
        expect(res.body).toHaveProperty('token');
        userToken = res.body.token;
    });

    it('should return 401 for wrong password', async () => {
        const res = await request(app)
            .post('/api/v1/auth/login')
            .send({ email: testEmail, password: 'WrongPassword!' });
        
        expect(res.statusCode).toBe(401);
        expect(res.body.error.message).toBe('Invalid email or password');
    });

    it('should return 401 for nonexistent user', async () => {
        const res = await request(app)
            .post('/api/v1/auth/login')
            .send({ email: 'nobody@example.com', password: 'Password123' });
        
        expect(res.statusCode).toBe(401);
        expect(res.body.error.message).toBe('Invalid email or password');
    });

    it('should fetch /auth/me using identity from valid JWT only', async () => {
        const res = await request(app)
            .get('/api/v1/auth/me')
            .set('Authorization', `Bearer ${userToken}`);
        
        expect(res.statusCode).toBe(200);
        expect(res.body).toHaveProperty('userId');
        expect(res.body).toHaveProperty('expiresAt');
    });

    it('should return 401 for expired or invalid JWT', async () => {
        const res = await request(app)
            .get('/api/v1/auth/me')
            .set('Authorization', `Bearer invalid.jwt.token`);
        
        expect(res.statusCode).toBe(401);
        expect(res.body.error.code).toBe('UNAUTHENTICATED');
    });
});
