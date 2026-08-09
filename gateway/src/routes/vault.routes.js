const express = require('express');
const VaultController = require('../controllers/vault.ctrl');
const { requireAuth } = require('../middlewares/auth.mid');

const router = express.Router();

router.post('/data', requireAuth, VaultController.create);
router.get('/data/:id', requireAuth, VaultController.get);

module.exports = router;
