<template>
  <router-view />
</template>
<script setup>
import {getJobsStatus, systemState} from 'src/store/systemState';
import useInterval from 'src/composables/useInterval';

useInterval(() => {
  if (systemState.systemInfo?.is_setup) {
    getJobsStatus();
  }
}, 10000)
</script>
