<template>
  <router-view />
</template>
<script setup>
import {getJobsStatus, systemState} from 'src/store/systemState';

if (systemState.systemInfo?.is_setup) {
  getJobsStatus();
}
</script>
