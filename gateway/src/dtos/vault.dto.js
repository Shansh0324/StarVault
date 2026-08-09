class VaultDTO {
    static validateCreate(body) {
        if (!body.dataType || typeof body.dataType !== 'string' || body.dataType.trim() === '') {
            throw new Error('dataType is required and must be a non-empty string');
        }
        if (!body.data || typeof body.data !== 'string' || body.data.trim() === '') {
            throw new Error('data is required and must be a non-empty string');
        }
        if (body.data.length > 50000) {
            throw new Error('data payload exceeds maximum allowed size');
        }
        return {
            dataType: body.dataType.trim(),
            data: body.data.trim()
        };
    }
}

module.exports = VaultDTO;
