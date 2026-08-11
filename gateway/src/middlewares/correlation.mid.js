const crypto = require('crypto');

function correlationMiddleware(req, res, next) {
    let reqId = req.headers['x-request-id'];
    
    // Validate or generate UUID
    if (!reqId || typeof reqId !== 'string' || reqId.length > 50) {
        reqId = crypto.randomUUID();
    }
    
    req.id = reqId;
    res.setHeader('X-Request-ID', reqId);
    next();
}

module.exports = correlationMiddleware;
