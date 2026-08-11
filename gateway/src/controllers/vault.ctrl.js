const VaultDTO = require('../dtos/vault.dto');
const coreFetch = require('../utils/coreClient');

class VaultController {
    static async create(req, res) {
        try {
            const dto = VaultDTO.validateCreate(req.body);
            
            // req.user comes from JWT middleware requireAuth
            const userId = req.user.userId;

            const { response, data } = await coreFetch(req, '/internal/vault/data', {
                method: 'POST',
                headers: { 
                    'X-User-ID': userId 
                },
                body: JSON.stringify(dto)
            });
            
            if (!response.ok) {
                return res.status(response.status).json({ error: { message: data.error || 'Internal Error' } });
            }
            res.status(response.status).json(data);
        } catch (error) {
            if (error.message.includes('required') || error.message.includes('exceeds')) {
                return res.status(400).json({ error: { code: 'INVALID_REQUEST', message: error.message } });
            }
            const errorMsg = error.message === 'Core request timed out' ? 'Core request timed out' : 'Failed to communicate with Core';
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: errorMsg } });
        }
    }

    static async get(req, res) {
        try {
            const userId = req.user.userId;
            const vaultId = req.params.id;

            if (!vaultId) {
                return res.status(400).json({ error: { code: 'INVALID_REQUEST', message: 'Vault ID is required' } });
            }

            const { response, data } = await coreFetch(req, `/internal/vault/data/${vaultId}`, {
                method: 'GET',
                headers: { 'X-User-ID': userId }
            });
            if (!response.ok) {
                // If it's a 404 from core due to IDOR, we return it as is.
                return res.status(response.status).json({ error: { message: data.error || 'Internal Error' } });
            }
            
            res.status(200).json(data);
        } catch (error) {
            const errorMsg = error.message === 'Core request timed out' ? 'Core request timed out' : 'Failed to communicate with Core';
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: errorMsg } });
        }
    }
}

module.exports = VaultController;
