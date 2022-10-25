/**
 * Created by Lijie on 2021/07/12 17:45.
 * eventbus
 */
import eventBus from 'vue3-eventbus';
import { onBeforeUnmount } from 'vue';

const useEventBus = (eventName, fn) => {
  eventBus.on(eventName, fn);
  onBeforeUnmount(() => {
    eventBus.off(eventName, fn);
  });
};

export default useEventBus;
