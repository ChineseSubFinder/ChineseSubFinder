import { reactive } from 'vue';
import SystemApi from 'src/api/SystemApi';
import JobApi from 'src/api/JobApi';

export const systemState = reactive({
  systemInfo: null,
  jobStatus: null,
});

export const getInfo = async () => {
  const [res] = await SystemApi.getInfo();
  systemState.systemInfo = res;
};

export const getJobsStatus = async () => {
  const [res] = await JobApi.getStatus();
  systemState.jobStatus = res;
};
