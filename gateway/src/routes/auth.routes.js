const express = require('express');
const AuthController = require('../controllers/auth.ctrl');
const { requireAuth } = require('../middlewares/auth.mid');

const router = express.Router();

router.post('/register', AuthController.register);
router.post('/login', AuthController.login);
router.get('/me', requireAuth, AuthController.me);

module.exports = router;
