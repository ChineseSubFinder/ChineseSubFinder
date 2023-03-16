import { LocalStorage } from 'quasar';
import { ref } from 'vue';
import useInterval from 'src/composables/use-interval';

export const useApiLimit = (storageKey, minRequestInterval = 2 * 1000) => {
  // 上次请求时间
  let lastRequestApiTime = LocalStorage.getItem(storageKey) || 0;
  // 最小请求间隔
  const minRequestApiInterval = minRequestInterval;
  // 下次请求倒数时间，秒
  const nextRequestCountdownSecond = ref(0);
  // 是否正在读秒
  const countdownLoading = ref(false);

  useInterval(() => {
    const v = Math.ceil((lastRequestApiTime + minRequestApiInterval - Date.now()) / 1000);
    nextRequestCountdownSecond.value = v > 0 ? v : 0;
  }, 100);

  /**
   * 检查请求是否可用
   * @returns {boolean}
   */
  const checkRequestReady = () => {
    const now = Date.now();
    if (now - lastRequestApiTime < minRequestApiInterval) {
      return false;
    }
    lastRequestApiTime = now;
    LocalStorage.set(storageKey, now);
    return true;
  };

  /**
   * 等待请求可用
   * @returns {Promise<void>}
   */
  const waitRequestReady = async () => {
    countdownLoading.value = true;
    // 每100ms检查一次，直到请求间隔大于最小请求间隔
    while (!checkRequestReady()) {
      // eslint-disable-next-line no-await-in-loop
      await new Promise((resolve) => {
        setTimeout(resolve, 100);
      });
    }
    countdownLoading.value = false;
  };

  return {
    checkRequestReady,
    waitRequestReady,
    countdownLoading,
    nextRequestCountdownSecond,
  };
};
