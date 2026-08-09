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

module.exports = { requireAuth };
