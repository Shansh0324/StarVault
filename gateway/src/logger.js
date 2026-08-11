function formatLog(level, event, reqId, extra = {}) {
    return JSON.stringify({
        timestamp: new Date().toISOString(),
        level,
        service: 'gateway',
        requestId: reqId || 'system',
        event,
        ...extra
    });
}

const logger = {
    info: (event, reqId, extra) => console.log(formatLog('info', event, reqId, extra)),
    warn: (event, reqId, extra) => console.warn(formatLog('warn', event, reqId, extra)),
    error: (event, reqId, extra) => console.error(formatLog('error', event, reqId, extra))
};

module.exports = logger;
