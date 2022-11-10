import { ref } from 'vue';
import { useWebSocketApi } from 'src/composables/use-web-socket-api';

export const useRealTimeLog = () => {
  const logLines = ref([]);

  useWebSocketApi('running_log', (data) => {
    logLines.value.push(...(data.log_lines ?? []));
  });

  return {
    logLines,
  };
};
