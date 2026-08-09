const express = require('express');
const dotenv = require('dotenv');
const path = require('path');

dotenv.config({ path: path.resolve(__dirname, '../../.env') });

const authRoutes = require('./routes/auth.routes');

const app = express();
app.use(express.json());

const PORT = process.env.GATEWAY_PORT || 3000;

app.get('/health', (req, res) => {
    res.status(200).json({ status: 'Gateway is healthy' });
});

// Mount routes
app.use('/api/v1/auth', authRoutes);

if (require.main === module) {
    app.listen(PORT, () => {
        console.log(`Gateway Service starting on port ${PORT}...`);
    });
}

module.exports = app;
