const AccessDto = require('../dtos/access.dto');
const coreFetch = require('../utils/coreClient');

const accessData = async (req, res) => {
    try {
        AccessDto.validate(req);

        const headers = {};
        if (req.isOpaqueToken) {
            headers['X-Access-Token'] = req.accessToken;
        } else {
            headers['X-User-ID'] = req.user.userId;
        }

        const { response, data } = await coreFetch(req, '/internal/access/data', {
            method: 'POST',
            headers: headers,
            body: JSON.stringify(req.body)
        });

        if (!response.ok) {
            return res.status(response.status).json(data);
        }

        res.status(200).json(data);
    } catch (err) {
        if (err.message.includes('required') || err.message.includes('must be')) {
            return res.status(400).json({ error: { code: 'VALIDATION_ERROR', message: err.message } });
        }
        
        const errorMsg = err.message === 'Core request timed out' ? 'Core request timed out' : 'Internal Server Error';
        res.status(500).json({ error: { code: 'INTERNAL_ERROR', message: errorMsg } });
    }
};

module.exports = { accessData };
