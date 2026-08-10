const ConsentDTO = require('../dtos/consent.dto');

const CORE_URL = process.env.CORE_URL || 'http://localhost:8080';

class ConsentController {
    static async create(req, res) {
        try {
            const dto = ConsentDTO.validateCreate(req.body);
            const userId = req.user.userId; // Trusted identity from JWT

            const response = await fetch(`${CORE_URL}/internal/consents`, {
                method: 'POST',
                headers: { 
                    'Content-Type': 'application/json',
                    'X-User-ID': userId
                },
                body: JSON.stringify(dto)
            });
            
            const data = await response.json();
            if (!response.ok) {
                return res.status(response.status).json({ error: { message: data.error || 'Internal Error' } });
            }
            res.status(response.status).json(data);
        } catch (error) {
            if (error.message.includes('required')) {
                return res.status(400).json({ error: { code: 'INVALID_REQUEST', message: error.message } });
            }
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: 'Failed to communicate with Core' } });
        }
    }

    static async get(req, res) {
        try {
            const userId = req.user.userId;
            const consentId = req.params.id;

            const response = await fetch(`${CORE_URL}/internal/consents/${consentId}`, {
                method: 'GET',
                headers: { 'X-User-ID': userId }
            });
            
            const data = await response.json();
            if (!response.ok) {
                return res.status(response.status).json({ error: { message: data.error || 'Internal Error' } });
            }
            res.status(200).json(data);
        } catch (error) {
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: 'Failed to communicate with Core' } });
        }
    }

    static async list(req, res) {
        try {
            const userId = req.user.userId;

            const response = await fetch(`${CORE_URL}/internal/consents`, {
                method: 'GET',
                headers: { 'X-User-ID': userId }
            });
            
            const data = await response.json();
            if (!response.ok) {
                return res.status(response.status).json({ error: { message: data.error || 'Internal Error' } });
            }
            res.status(200).json(data);
        } catch (error) {
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: 'Failed to communicate with Core' } });
        }
    }

    static async revoke(req, res) {
        try {
            const userId = req.user.userId;
            const consentId = req.params.id;

            const response = await fetch(`${CORE_URL}/internal/consents/${consentId}/revoke`, {
                method: 'POST',
                headers: { 'X-User-ID': userId }
            });
            
            const data = await response.json();
            if (!response.ok) {
                return res.status(response.status).json({ error: { message: data.error || 'Internal Error' } });
            }
            res.status(200).json(data);
        } catch (error) {
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: 'Failed to communicate with Core' } });
        }
    }
}

module.exports = ConsentController;
