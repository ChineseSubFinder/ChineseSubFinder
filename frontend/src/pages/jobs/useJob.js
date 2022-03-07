import { onMounted, ref, watch, computed } from 'vue';
import { useWebSocketApi } from 'src/composables/useWebSocketApi';
import { systemState } from 'src/store/systemState';

export const subJobsDetail = ref(null);

export const isRunning = computed(() => subJobsDetail.value.status === 'running');

export const isPreparing = computed(() => subJobsDetail.value.status === 'preparing');

export const isWaiting = computed(() => subJobsDetail.value.status === 'waiting');

export const useJob = () => {
  watch(
    () => systemState.jobStatus?.status,
    () => {
      subJobsDetail.value = null;
    }
  );

  useWebSocketApi('sub_download_jobs_status', (data) => {
    subJobsDetail.value = data;
  });

  onMounted(() => {
    subJobsDetail.value = null;
  });
};
