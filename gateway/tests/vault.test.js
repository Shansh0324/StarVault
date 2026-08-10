const apiClient = require('./helpers/api.client');

describe('Vault Endpoints', () => {
    let tokenA = '';
    let tokenB = '';
    let vaultId = '';

    beforeAll(async () => {
        const setup = await apiClient.setupStandardEnvironment();
        tokenA = setup.userA.token;
        tokenB = setup.userB.token;
    });

    it('should reject unauthenticated POST', async () => {
        const res = await apiClient.post('/api/v1/vault/data', { dataType: 'test', data: 'data' });
        expect(res.statusCode).toBe(401);
    });

    it('should reject unauthenticated GET', async () => {
        const res = await apiClient.get('/api/v1/vault/data/some-id');
        expect(res.statusCode).toBe(401);
    });

    it('should reject missing dataType', async () => {
        const res = await apiClient.post('/api/v1/vault/data', { data: 'my secret' }, tokenA);
        expect(res.statusCode).toBe(400);
    });

    it('should reject missing data', async () => {
        const res = await apiClient.post('/api/v1/vault/data', { dataType: 'test' }, tokenA);
        expect(res.statusCode).toBe(400);
    });

    it('should allow User A to create vault data', async () => {
        const res = await apiClient.post('/api/v1/vault/data', { dataType: 'medical_report', data: 'Patient has XYZ condition' }, tokenA);
        
        expect(res.statusCode).toBe(201);
        expect(res.body).toHaveProperty('id');
        expect(res.body.dataType).toBe('medical_report');
        vaultId = res.body.id;
    });

    it('should allow User A to retrieve their own vault data', async () => {
        const res = await apiClient.get(`/api/v1/vault/data/${vaultId}`, tokenA);
        
        expect(res.statusCode).toBe(200);
        expect(res.body.dataType).toBe('medical_report');
        expect(res.body.data).toBe('Patient has XYZ condition');
    });

    it('should reject User B trying to retrieve User A data (IDOR prevention)', async () => {
        const res = await apiClient.get(`/api/v1/vault/data/${vaultId}`, tokenB);
        
        expect(res.statusCode).toBe(404); // Returns 404 so existence is not leaked
    });

    it('should handle client trying to spoof userId in body safely', async () => {
        // Even if User B tries to specify User A's ID in body, it should be ignored by the DTO and Controller
        const res = await apiClient.post('/api/v1/vault/data', { userId: 'spoofed_id', dataType: 'test', data: 'test data' }, tokenB);
        
        expect(res.statusCode).toBe(201);
        
        // Let's verify it belongs to B
        const spoofedId = res.body.id;
        
        const resB = await apiClient.get(`/api/v1/vault/data/${spoofedId}`, tokenB);
        expect(resB.statusCode).toBe(200); // User B can access their own data
        
        const resA = await apiClient.get(`/api/v1/vault/data/${spoofedId}`, tokenA);
        expect(resA.statusCode).toBe(404); // User A cannot access User B's data
    });
});
