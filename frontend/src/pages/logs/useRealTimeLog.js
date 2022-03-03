import { ref } from 'vue';
import { useWebSocketApi, wsManager } from 'src/composables/useWebSocketApi';

export const useRealTimeLog = () => {
  const logLines = ref([]);

  useWebSocketApi('running_log', (data) => {
    logLines.value.push(...(data.log_lines ?? []));
    wsManager.send({
      type: 'common_reply',
      message: 'running log recv ok',
    });
  });

  return {
    logLines,
  };
};
