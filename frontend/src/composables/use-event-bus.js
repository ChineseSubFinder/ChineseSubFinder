/**
 * Created by Lijie on 2021/07/12 17:45.
 * eventbus
 */
import eventBus from 'vue3-eventbus';
import { onBeforeUnmount } from 'vue';

const useEventBus = (eventName, fn) => {
  const eventBusFn = (...args) => {
    fn(...args);
  };
  eventBus.on(eventName, eventBusFn);
  onBeforeUnmount(() => {
    eventBus.off(eventName, eventBusFn);
  });
};

export default useEventBus;
