import { computed, reactive } from 'vue';
import SystemApi from 'src/api/SystemApi';
import JobApi from 'src/api/JobApi';

export const systemState = reactive({
  systemInfo: null,
  jobStatus: null,
  preJobStatus: null,
});

export const getInfo = async () => {
  const [res] = await SystemApi.getInfo();
  systemState.systemInfo = res;
};

export const isJobRunning = computed(() => systemState.jobStatus?.status === 'running');
export const isRunningInDocker = computed(() => systemState.systemInfo?.is_running_in_docker);

export const getJobsStatus = async () => {
  const [res] = await JobApi.getStatus();
  systemState.jobStatus = res;
};
