const coreFetch = require('../utils/coreClient');

class UserController {
    static async updateKey(req, res) {
        try {
            const { key } = req.body;
            
            // Basic validation for 64-char hex string
            if (!key || typeof key !== 'string' || key.length !== 64 || !/^[0-9a-fA-F]+$/.test(key)) {
                return res.status(400).json({ error: { message: "key must be a 64-character hex string" } });
            }

            const { response, data } = await coreFetch(req, '/internal/users/key', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                    'X-User-ID': req.user.userId
                },
                body: JSON.stringify({ key })
            });

            if (!response.ok) {
                return res.status(response.status).json({ error: { message: data?.error || 'Failed to update key' } });
            }

            res.status(200).json({ message: "Key updated successfully" });
        } catch (error) {
            console.error("UserController.updateKey error:", error);
            res.status(500).json({ error: { message: 'Internal Server Error' } });
        }
    }

    static async revokeKey(req, res) {
        try {
            const { response, data } = await coreFetch(req, '/internal/users/key', {
                method: 'DELETE',
                headers: {
                    'X-User-ID': req.user.userId
                }
            });

            if (!response.ok) {
                return res.status(response.status).json({ error: { message: data?.error || 'Failed to revoke key' } });
            }

            res.status(200).json({ message: "Key revoked successfully. Data is now crypto-shredded." });
        } catch (error) {
            console.error("UserController.revokeKey error:", error);
            res.status(500).json({ error: { message: 'Internal Server Error' } });
        }
    }
}

module.exports = UserController;
