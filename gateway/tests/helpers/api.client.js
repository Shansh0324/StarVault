const request = require('supertest');
const app = require('../../src/index');

class ApiClient {
    constructor() {
        this.app = app;
    }

    async post(path, body = {}, token = null) {
        const req = request(this.app).post(path);
        if (token) {
            req.set('Authorization', `Bearer ${token}`);
        }
        return req.send(body);
    }

    async get(path, token = null) {
        const req = request(this.app).get(path);
        if (token) {
            req.set('Authorization', `Bearer ${token}`);
        }
        return req.send();
    }

    async put(path, body = {}, token = null) {
        const req = request(this.app).put(path);
        if (token) {
            req.set('Authorization', `Bearer ${token}`);
        }
        return req.send(body);
    }

    async delete(path, token = null) {
        const req = request(this.app).delete(path);
        if (token) {
            req.set('Authorization', `Bearer ${token}`);
        }
        return req.send();
    }

    // Helper for common setup across all tests
    async setupStandardEnvironment() {
        const emailA = `userA_${Date.now()}@example.com`;
        const emailB = `userB_${Date.now()}@example.com`;
        const password = 'Password123';

        // 1. Create Users
        await this.post('/api/v1/auth/register', { email: emailA, password });
        const resA = await this.post('/api/v1/auth/login', { email: emailA, password });
        
        await this.post('/api/v1/auth/register', { email: emailB, password });
        const resB = await this.post('/api/v1/auth/login', { email: emailB, password });

        // 2. Create App
        const appRes = await this.post('/api/v1/apps/register', { name: `Test App ${Date.now()}` });

        return {
            userA: { email: emailA, token: resA.body.token },
            userB: { email: emailB, token: resB.body.token },
            app: {
                id: appRes.body.appId,
                secret: appRes.body.secret
            },
            password
        };
    }
}

module.exports = new ApiClient();
