const jwt = require('jsonwebtoken');
const AuthDTO = require('../dtos/auth.dto');
const coreFetch = require('../utils/coreClient');
const JWT_SECRET = process.env.JWT_SECRET;

class AuthController {
    static async register(req, res) {
        try {
            const dto = AuthDTO.validateRegister(req.body);
            
            const { response, data } = await coreFetch(req, '/internal/users', {
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
            const errorMsg = error.message === 'Core request timed out' ? 'Core request timed out' : 'Failed to communicate with Core';
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: errorMsg } });
        }
    }

    static async login(req, res) {
        try {
            const dto = AuthDTO.validateLogin(req.body);

            const { response, data } = await coreFetch(req, '/internal/users/verify', {
                method: 'POST',
                body: JSON.stringify(dto)
            });
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
            const errorMsg = error.message === 'Core request timed out' ? 'Core request timed out' : 'Failed to communicate with Core';
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: errorMsg } });
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
