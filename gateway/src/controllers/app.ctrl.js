const AppDTO = require('../dtos/app.dto');

const CORE_URL = process.env.CORE_URL || 'http://localhost:8080';

class AppController {
    static async register(req, res) {
        try {
            const dto = AppDTO.validateCreate(req.body);

            const response = await fetch(`${CORE_URL}/internal/apps`, {
                method: 'POST',
                headers: { 
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(dto)
            });
            
            const data = await response.json();
            
            if (!response.ok) {
                return res.status(response.status).json({ error: { message: data.error || 'Internal Error' } });
            }
            res.status(response.status).json(data);
        } catch (error) {
            if (error.message.includes('required') || error.message.includes('too long')) {
                return res.status(400).json({ error: { code: 'INVALID_REQUEST', message: error.message } });
            }
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: 'Failed to communicate with Core' } });
        }
    }
}

module.exports = AppController;
