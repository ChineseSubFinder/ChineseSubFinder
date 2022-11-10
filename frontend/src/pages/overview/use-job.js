import { onMounted, ref, watch, computed } from 'vue';
import { useWebSocketApi } from 'src/composables/use-web-socket-api';
import { systemState } from 'src/store/systemState';

export const subJobsDetail = ref(null);

export const isRunning = computed(() => subJobsDetail.value.status === 'running');
export const isPreparing = computed(() => subJobsDetail.value.status === 'preparing');
export const isWaiting = computed(() => subJobsDetail.value.status === 'waiting');
export const isScanMovie = computed(() => subJobsDetail.value.status === 'scan-movie');
export const isScanSeries = computed(() => subJobsDetail.value.status === 'scan-series');

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
