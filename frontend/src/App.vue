<template>
  <router-view />
</template>
<script setup>
import { getJobsStatus, systemState } from 'src/store/systemState';
import useInterval from 'src/composables/use-interval';
import { watch } from 'vue';
import { userState } from 'src/store/userState';
import { useAppStatusLoading } from 'src/composables/use-app-status-loading';

const { startLoading } = useAppStatusLoading();

const getSystemJobStatus = () => {
  if (userState.accessToken && systemState.systemInfo?.is_setup) {
    getJobsStatus();
  }
};

useInterval(() => {
  getSystemJobStatus();
}, 8000);

watch(
  () => systemState.systemInfo?.is_setup,
  (val) => {
    getSystemJobStatus();
    if (val) {
      startLoading();
    }
  }
);
</script>
