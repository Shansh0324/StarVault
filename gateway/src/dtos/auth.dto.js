class AuthDTO {
    static validateRegister(body) {
        if (!body.email || typeof body.email !== 'string') {
            throw new Error('Email is required and must be a string');
        }
        if (!body.password || typeof body.password !== 'string') {
            throw new Error('Password is required and must be a string');
        }
        return {
            email: body.email.trim(),
            password: body.password
        };
    }

    static validateLogin(body) {
        // For MVP, login structure is same as register
        return this.validateRegister(body);
    }
}

module.exports = AuthDTO;
