const jwt = require('jsonwebtoken');
const AuthDTO = require('../dtos/auth.dto');

const CORE_URL = process.env.CORE_URL || 'http://localhost:8080';
const JWT_SECRET = process.env.JWT_SECRET || 'fallback_secret';

class AuthController {
    static async register(req, res) {
        try {
            const dto = AuthDTO.validateRegister(req.body);
            
            const response = await fetch(`${CORE_URL}/internal/users`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(dto)
            });
            const data = await response.json();
            
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

            const response = await fetch(`${CORE_URL}/internal/users/verify`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(dto)
            });
            
            const data = await response.json();
            if (!response.ok) {
                return res.status(response.status).json({ error: { message: data.error || 'Internal Error' } });
            }
            
            // Issue JWT
            const expiresIn = 3600; // 1 hour
            const token = jwt.sign({ userId: data.userId }, JWT_SECRET, { expiresIn });
            
            res.status(200).json({ token });
        } catch (error) {
            if (error.message.includes('required')) {
                return res.status(400).json({ error: { code: 'BAD_REQUEST', message: error.message } });
            }
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: 'Failed to communicate with Core' } });
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
