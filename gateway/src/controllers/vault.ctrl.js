const VaultDTO = require('../dtos/vault.dto');

const CORE_URL = process.env.CORE_URL || 'http://localhost:8080';

class VaultController {
    static async create(req, res) {
        try {
            const dto = VaultDTO.validateCreate(req.body);
            
            // req.user comes from JWT middleware requireAuth
            const userId = req.user.userId;

            const response = await fetch(`${CORE_URL}/internal/vault/data`, {
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
            if (error.message.includes('required') || error.message.includes('exceeds')) {
                return res.status(400).json({ error: { code: 'INVALID_REQUEST', message: error.message } });
            }
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: 'Failed to communicate with Core' } });
        }
    }

    static async get(req, res) {
        try {
            const userId = req.user.userId;
            const vaultId = req.params.id;

            if (!vaultId) {
                return res.status(400).json({ error: { code: 'INVALID_REQUEST', message: 'Vault ID is required' } });
            }

            const response = await fetch(`${CORE_URL}/internal/vault/data/${vaultId}`, {
                method: 'GET',
                headers: { 
                    'X-User-ID': userId 
                }
            });
            
            const data = await response.json();
            if (!response.ok) {
                // If it's a 404 from core due to IDOR, we return it as is.
                return res.status(response.status).json({ error: { message: data.error || 'Internal Error' } });
            }
            
            res.status(200).json(data);
        } catch (error) {
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: 'Failed to communicate with Core' } });
        }
    }
}

module.exports = VaultController;
