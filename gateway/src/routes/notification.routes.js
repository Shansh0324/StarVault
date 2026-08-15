const express = require('express');
const router = express.Router();
const NotificationController = require('../controllers/notification.ctrl');
const { requireAuth } = require('../middlewares/auth.mid');

router.get('/access', requireAuth, NotificationController.streamAccessEvents);

module.exports = router;
