class ConsentDTO {
    static validateCreate(body) {
        if (!body.appId || typeof body.appId !== 'string' || body.appId.trim() === '') {
            throw new Error('appId is required');
        }
        if (!body.scopes || !Array.isArray(body.scopes) || body.scopes.length === 0) {
            throw new Error('scopes array is required and must not be empty');
        }
        if (!body.purpose || typeof body.purpose !== 'string' || body.purpose.trim() === '') {
            throw new Error('purpose is required');
        }
        if (!body.expiresAt || typeof body.expiresAt !== 'string' || body.expiresAt.trim() === '') {
            throw new Error('expiresAt is required');
        }

        // Just basic string checks, Go Core handles RFC3339 datetime validation
        return {
            appId: body.appId.trim(),
            scopes: body.scopes,
            purpose: body.purpose.trim(),
            expiresAt: body.expiresAt.trim()
        };
    }
}

module.exports = ConsentDTO;
