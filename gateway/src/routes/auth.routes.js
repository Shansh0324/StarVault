const express = require('express');
const AuthController = require('../controllers/auth.ctrl');
const { requireAuth } = require('../middlewares/auth.mid');

const router = express.Router();

router.post('/register', AuthController.register);
router.post('/login', AuthController.login);
router.get('/me', requireAuth, AuthController.me);

router.post('/webauthn/register/begin', AuthController.webauthnRegisterBegin);
router.post('/webauthn/register/finish', AuthController.webauthnRegisterFinish);
router.post('/webauthn/login/begin', AuthController.webauthnLoginBegin);
router.post('/webauthn/login/finish', AuthController.webauthnLoginFinish);

module.exports = router;
