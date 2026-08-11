const express = require('express');
const { accessData } = require('../controllers/access.ctrl');
const { smartAuth } = require('../middlewares/auth.mid');

const router = express.Router();

router.post('/data', smartAuth, accessData);

module.exports = router;
