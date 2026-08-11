const validateTokenRequest = (req, res, next) => {
    const { appId, appSecret, scope } = req.body;
    
    if (!appId || typeof appId !== 'string') {
        return res.status(400).json({ error: { code: 'BAD_REQUEST', message: 'appId is required and must be a string' } });
    }
    if (!appSecret || typeof appSecret !== 'string') {
        return res.status(400).json({ error: { code: 'BAD_REQUEST', message: 'appSecret is required and must be a string' } });
    }
    if (!scope || typeof scope !== 'string') {
        return res.status(400).json({ error: { code: 'BAD_REQUEST', message: 'scope is required and must be a string' } });
    }
    
    next();
};

const validateRevokeRequest = (req, res, next) => {
    const { token } = req.body;
    
    if (!token || typeof token !== 'string') {
        return res.status(400).json({ error: { code: 'BAD_REQUEST', message: 'token is required and must be a string' } });
    }
    
    next();
};

module.exports = {
    validateTokenRequest,
    validateRevokeRequest
};
