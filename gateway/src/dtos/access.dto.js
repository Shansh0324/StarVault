class AccessDto {
    static validate(req) {
        const { appId, secret, scope, vaultDataId } = req.body;

        if (!req.isOpaqueToken) {
            if (!appId || typeof appId !== 'string') {
                throw new Error('appId is required and must be a string');
            }
            if (!secret || typeof secret !== 'string') {
                throw new Error('secret is required and must be a string');
            }
        }
        if (!scope || typeof scope !== 'string') {
            throw new Error('scope is required and must be a string');
        }
        if (!vaultDataId || typeof vaultDataId !== 'string') {
            throw new Error('vaultDataId is required and must be a string');
        }
    }
}

module.exports = AccessDto;
