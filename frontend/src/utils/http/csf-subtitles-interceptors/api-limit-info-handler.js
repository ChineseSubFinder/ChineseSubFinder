import eventBus from 'vue3-eventbus';

export default {
  onResponseFullFilled: (response) => {
    try {
      if (response.headers['x-daily-limit']) {
        const limitInfo = {
          dailyLimit: response.headers['x-daily-limit'],
          dailyCount: response.headers['x-daily-count'],
          expireTime: response.headers['x-api-expiration-time'],
        };
        eventBus.emit('subtitle-best-api-limit-info', limitInfo);
      }
    } catch (e) {
      // ignore
    }
    return response;
  },
};
