const express = require('express');
const AppController = require('../controllers/app.ctrl');

const router = express.Router();

// Registration is unauthenticated (public) for the MVP
router.post('/register', AppController.register);

module.exports = router;
