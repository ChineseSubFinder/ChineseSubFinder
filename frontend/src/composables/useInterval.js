import { onBeforeUnmount, ref } from 'vue';

const useInterval = (fn, ms, autoStart = true) => {
  const timer = ref(null);
  if (autoStart) {
    timer.value = setInterval(() => {
      fn();
    }, ms);
    fn();
  }
  const resetInterval = () => {
    clearInterval(timer.value);
    timer.value = setInterval(() => {
      fn();
    }, ms);
    fn();
  };
  const stopInterval = () => {
    clearInterval(timer.value);
  };
  onBeforeUnmount(() => {
    clearInterval(timer.value);
  });
  return {
    timer,
    resetInterval,
    stopInterval,
  };
};

export default useInterval;
