const express = require('express');
const UserController = require('../controllers/user.ctrl');
const { requireAuth } = require('../middlewares/auth.mid');

const router = express.Router();

router.use(requireAuth); // Protect all user routes

router.put('/key', UserController.updateKey);
router.delete('/key', UserController.revokeKey);

module.exports = router;
