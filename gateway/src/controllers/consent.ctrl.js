const ConsentDTO = require('../dtos/consent.dto');
const coreFetch = require('../utils/coreClient');

class ConsentController {
    static async create(req, res) {
        try {
            const dto = ConsentDTO.validateCreate(req.body);
            const userId = req.user.userId; // Trusted identity from JWT

            const { response, data } = await coreFetch(req, '/internal/consents', {
                method: 'POST',
                headers: { 
                    'Content-Type': 'application/json',
                    'X-User-ID': req.user.userId
                },
                body: JSON.stringify(dto)
            });
            
            if (!response.ok) {
                return res.status(response.status).json({ error: { message: data.error || 'Internal Error' } });
            }
            res.status(response.status).json(data);
        } catch (error) {
            if (error.message.includes('required')) {
                return res.status(400).json({ error: { code: 'INVALID_REQUEST', message: error.message } });
            }
            const errorMsg = error.message === 'Core request timed out' ? 'Core request timed out' : 'Failed to communicate with Core';
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: errorMsg } });
        }
    }

    static async get(req, res) {
        try {
            const userId = req.user.userId;
            const consentId = req.params.id;

            const { response, data } = await coreFetch(req, `/internal/consents/${consentId}`, {
                method: 'GET',
                headers: {
                    'X-User-ID': req.user.userId
                }
            });
            
            if (!response.ok) {
                return res.status(response.status).json({ error: { message: data.error || 'Internal Error' } });
            }
            res.status(200).json(data);
        } catch (error) {
            const errorMsg = error.message === 'Core request timed out' ? 'Core request timed out' : 'Failed to communicate with Core';
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: errorMsg } });
        }
    }

    static async list(req, res) {
        try {
            const userId = req.user.userId;

            const { response, data } = await coreFetch(req, '/internal/consents', {
                method: 'GET',
                headers: {
                    'X-User-ID': req.user.userId
                }
            });
            
            if (!response.ok) {
                return res.status(response.status).json({ error: { message: data.error || 'Internal Error' } });
            }
            res.status(200).json(data);
        } catch (error) {
            const errorMsg = error.message === 'Core request timed out' ? 'Core request timed out' : 'Failed to communicate with Core';
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: errorMsg } });
        }
    }

    static async revoke(req, res) {
        try {
            const consentId = req.params.id;

            const { response, data } = await coreFetch(req, `/internal/consents/${consentId}/revoke`, {
                method: 'POST',
                headers: {
                    'X-User-ID': req.user.userId
                }
            });
            
            if (!response.ok) {
                return res.status(response.status).json({ error: { message: data?.error || 'Internal Error' } });
            }
            res.status(200).json(data);
        } catch (error) {
            const errorMsg = error.message === 'Core request timed out' ? 'Core request timed out' : 'Failed to communicate with Core';
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: errorMsg } });
        }
    }
}

module.exports = ConsentController;
