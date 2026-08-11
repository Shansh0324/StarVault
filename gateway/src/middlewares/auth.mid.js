const jwt = require('jsonwebtoken');
const JWT_SECRET = process.env.JWT_SECRET || 'fallback_secret';

const requireAuth = (req, res, next) => {
    const authHeader = req.headers.authorization;
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
        return res.status(401).json({ error: { code: 'UNAUTHENTICATED', message: 'Missing or invalid token' } });
    }

    const token = authHeader.split(' ')[1];
    try {
        const decoded = jwt.verify(token, JWT_SECRET);
        req.user = decoded; // Contains userId and exp
        next();
    } catch (err) {
        return res.status(401).json({ error: { code: 'UNAUTHENTICATED', message: 'Expired or invalid token' } });
    }
};

const smartAuth = (req, res, next) => {
    const authHeader = req.headers.authorization;
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
        return res.status(401).json({ error: { code: 'UNAUTHENTICATED', message: 'Missing or invalid token' } });
    }

    const token = authHeader.split(' ')[1];
    
    // A standard JWT has exactly two dots (header.payload.signature)
    if (token.split('.').length === 3) {
        // It's a JWT (User flow)
        try {
            const decoded = jwt.verify(token, JWT_SECRET);
            req.user = decoded; 
            return next();
        } catch (err) {
            return res.status(401).json({ error: { code: 'UNAUTHENTICATED', message: 'Expired or invalid user token' } });
        }
    } else {
        // It's likely an opaque token (App token flow)
        req.isOpaqueToken = true;
        req.accessToken = token;
        return next();
    }
};

module.exports = { requireAuth, smartAuth };
