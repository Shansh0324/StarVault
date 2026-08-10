const express = require('express');
const { accessData } = require('../controllers/access.ctrl');
const { requireAuth } = require('../middlewares/auth.mid');

const router = express.Router();

router.post('/data', requireAuth, accessData);

module.exports = router;
