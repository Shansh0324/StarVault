const CORE_URL = process.env.CORE_URL || 'http://localhost:8080';

const issueToken = async (req, res) => {
    try {
        const userId = req.user.userId;
        const response = await fetch(`${CORE_URL}/internal/oauth/token`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-User-ID': userId,
                'X-Request-ID': req.headers['x-request-id'] || ''
            },
            body: JSON.stringify(req.body)
        });

        const data = await response.json();
        
        if (!response.ok) {
            return res.status(response.status).json(data);
        }
        
        return res.status(200).json(data);
    } catch (err) {
        return res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: 'Failed to communicate with Core service' } });
    }
};

const revokeToken = async (req, res) => {
    try {
        const userId = req.user.userId;
        const response = await fetch(`${CORE_URL}/internal/oauth/revoke`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-User-ID': userId,
                'X-Request-ID': req.headers['x-request-id'] || ''
            },
            body: JSON.stringify(req.body)
        });

        if (response.status === 204) {
            return res.status(204).send();
        }

        let data;
        try {
            data = await response.json();
        } catch(e) {
            data = { error: { message: "unknown error" }};
        }
        
        return res.status(response.status).json(data);
    } catch (err) {
        return res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: 'Failed to communicate with Core service' } });
    }
};

module.exports = {
    issueToken,
    revokeToken
};
