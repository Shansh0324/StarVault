const AppDTO = require('../dtos/app.dto');
const coreFetch = require('../utils/coreClient');

class AppController {
    static async register(req, res) {
        try {
            const dto = AppDTO.validateCreate(req.body);

            const { response, data } = await coreFetch(req, '/internal/apps', {
                method: 'POST',
                body: JSON.stringify(dto)
            });
            
            if (!response.ok) {
                return res.status(response.status).json({ error: { message: data.error || 'Internal Error' } });
            }
            res.status(response.status).json(data);
        } catch (error) {
            if (error.message.includes('required') || error.message.includes('too long')) {
                return res.status(400).json({ error: { code: 'INVALID_REQUEST', message: error.message } });
            }
            const errorMsg = error.message === 'Core request timed out' ? 'Core request timed out' : 'Failed to communicate with Core';
            res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: errorMsg } });
        }
    }
}

module.exports = AppController;
