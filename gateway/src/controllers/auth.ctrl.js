const jwt = require('jsonwebtoken');
const AuthDTO = require('../dtos/auth.dto');
const coreFetch = require('../utils/coreClient');
const JWT_SECRET = process.env.JWT_SECRET;

class AuthController {
    static async register(req, res) {
        try {
            const dto = AuthDTO.validateRegister(req.body);
            
            const { response, data } = await coreFetch(req, '/internal/auth/register', {
                method: 'POST',
                body: JSON.stringify(dto)
            });
            
            if (!response.ok) {
                return res.status(response.status).json({ error: { message: data.error || 'Internal Error' } });
            }
            res.status(response.status).json(data);
        } catch (error) {
            if (error.message.includes('required')) {
                return res.status(400).json({ error: { code: 'BAD_REQUEST', message: error.message } });
            }
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: 'Failed to communicate with Core' } });
        }
    }

    static async login(req, res) {
        try {
            const dto = AuthDTO.validateLogin(req.body);
            
            // Collect Device Posture data for Risk Scoring
            dto.ipAddress = req.ip || req.connection.remoteAddress;
            dto.userAgent = req.get('user-agent') || 'unknown';

            const { response, data } = await coreFetch(req, '/internal/auth/login', {
                method: 'POST',
                body: JSON.stringify(dto)
            });
            if (!response.ok) {
                return res.status(response.status).json({ error: { message: data.error || 'Internal Error' } });
            }
            
            const expiresIn = 3600;
            const token = jwt.sign({ userId: data.userId }, JWT_SECRET, { expiresIn });
            
            res.status(200).json({ token });
        } catch (error) {
            if (error.message.includes('required')) {
                return res.status(400).json({ error: { code: 'BAD_REQUEST', message: error.message } });
            }
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: 'Failed to communicate with Core' } });
        }
    }

    static async webauthnRegisterBegin(req, res) {
        const { email } = req.body;
        if (!email) return res.status(400).json({ error: { message: "email required" } });
        
        try {
            const { response, data } = await coreFetch(req, `/internal/auth/webauthn/register/begin?email=${encodeURIComponent(email)}`, { method: 'GET' });
            if (!response.ok) return res.status(response.status).json(data);
            res.json(data);
        } catch (error) {
            res.status(500).json({ error: { message: 'Failed to communicate with Core' } });
        }
    }

    static async webauthnRegisterFinish(req, res) {
        const { email, sessionId, response: webauthnResponse } = req.body;
        if (!email || !sessionId) return res.status(400).json({ error: { message: "email and sessionId required" } });
        
        try {
            const { response, data } = await coreFetch(req, `/internal/auth/webauthn/register/finish?email=${encodeURIComponent(email)}&sessionId=${encodeURIComponent(sessionId)}`, { 
                method: 'POST', 
                body: JSON.stringify(webauthnResponse || {}) 
            });
            if (!response.ok) return res.status(response.status).json(data);
            res.json(data);
        } catch (error) {
            res.status(500).json({ error: { message: 'Failed to communicate with Core' } });
        }
    }

    static async webauthnLoginBegin(req, res) {
        const { email } = req.body;
        if (!email) return res.status(400).json({ error: { message: "email required" } });
        
        try {
            const { response, data } = await coreFetch(req, `/internal/auth/webauthn/login/begin?email=${encodeURIComponent(email)}`, { method: 'GET' });
            if (!response.ok) return res.status(response.status).json(data);
            res.json(data);
        } catch (error) {
            res.status(500).json({ error: { message: 'Failed to communicate with Core' } });
        }
    }

    static async webauthnLoginFinish(req, res) {
        const { email, sessionId, response: webauthnResponse } = req.body;
        if (!email || !sessionId) return res.status(400).json({ error: { message: "email and sessionId required" } });
        
        try {
            const { response, data } = await coreFetch(req, `/internal/auth/webauthn/login/finish?email=${encodeURIComponent(email)}&sessionId=${encodeURIComponent(sessionId)}`, { 
                method: 'POST', 
                body: JSON.stringify(webauthnResponse || {}) 
            });
            if (!response.ok) return res.status(response.status).json(data);
            
            // Issue JWT upon successful WebAuthn Login
            const token = jwt.sign({ userId: data.userId }, JWT_SECRET, { expiresIn: 3600 });
            res.status(200).json({ token });
        } catch (error) {
            res.status(500).json({ error: { message: 'Failed to communicate with Core' } });
        }
    }

    static me(req, res) {
        res.status(200).json({
            userId: req.user.userId,
            expiresAt: req.user.exp
        });
    }
}

module.exports = AuthController;
