const express = require('express');
const ConsentController = require('../controllers/consent.ctrl');
const { requireAuth } = require('../middlewares/auth.mid');

const router = express.Router();

router.post('/', requireAuth, ConsentController.create);
router.get('/', requireAuth, ConsentController.list);
router.get('/:id', requireAuth, ConsentController.get);
router.post('/:id/revoke', requireAuth, ConsentController.revoke);

module.exports = router;
