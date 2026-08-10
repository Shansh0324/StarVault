const AccessDto = require('../dtos/access.dto');

const CORE_URL = process.env.CORE_URL || 'http://localhost:8080';

const accessData = async (req, res) => {
    try {
        AccessDto.validate(req);

        // Core handles the request
        const response = await fetch(`${CORE_URL}/internal/access/data`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-User-ID': req.user.userId
            },
            body: JSON.stringify(req.body)
        });

        const data = await response.json();

        if (!response.ok) {
            return res.status(response.status).json(data);
        }

        res.status(200).json(data);
    } catch (err) {
        if (err.message.includes('required') || err.message.includes('must be')) {
            return res.status(400).json({ error: { code: 'VALIDATION_ERROR', message: err.message } });
        }
        
        console.error('Core Error:', err.message);
        res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: 'Internal Server Error' } });
    }
};

module.exports = { accessData };
