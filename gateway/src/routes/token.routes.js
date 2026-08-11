const express = require('express');
const { issueToken, revokeToken } = require('../controllers/token.ctrl');
const { validateTokenRequest, validateRevokeRequest } = require('../dtos/token.dto');
const { requireAuth } = require('../middlewares/auth.mid');

const router = express.Router();

router.post('/token', requireAuth, validateTokenRequest, issueToken);
router.post('/revoke', requireAuth, validateRevokeRequest, revokeToken);

module.exports = router;
