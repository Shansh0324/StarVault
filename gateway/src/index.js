const express = require('express');
const dotenv = require('dotenv');
const path = require('path');

dotenv.config({ path: path.resolve(__dirname, '../../.env') });

const authRoutes = require('./routes/auth.routes');
const vaultRoutes = require('./routes/vault.routes');
const appRoutes = require('./routes/app.routes');
const consentRoutes = require('./routes/consent.routes');
const accessRoutes = require('./routes/access.routes');
const app = express();
app.use(express.json());

const PORT = process.env.GATEWAY_PORT || 3000;

app.get('/health', (req, res) => {
    res.status(200).json({ status: 'Gateway is healthy' });
});

// Mount routes
app.use('/api/v1/auth', authRoutes);
app.use('/api/v1/vault', vaultRoutes);
app.use('/api/v1/apps', appRoutes);
app.use('/api/v1/consents', consentRoutes);
app.use('/api/v1/access', accessRoutes);
if (require.main === module) {
    app.listen(PORT, () => {
        console.log(`Gateway Service starting on port ${PORT}...`);
    });
}

module.exports = app;
