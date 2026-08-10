class AppDTO {
    static validateCreate(body) {
        if (!body.name || typeof body.name !== 'string' || body.name.trim() === '') {
            throw new Error('name is required and must be a non-empty string');
        }
        if (body.name.length > 100) {
            throw new Error('name is too long');
        }
        return {
            name: body.name.trim()
        };
    }
}

module.exports = AppDTO;
