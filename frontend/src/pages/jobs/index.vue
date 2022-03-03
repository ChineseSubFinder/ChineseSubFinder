<template>
  <q-page class="q-pa-md">
    <q-card v-if="systemState.jobStatus" flat>
      <header class="row items-center justify-between q-gutter-md">
        <div>
          当前任务状态：
          <q-badge v-if="isRunning" color="positive">运行中</q-badge>
          <q-badge v-else color="grey">未运行</q-badge>
        </div>
        <div>
          <q-btn
            v-if="isRunning"
            label="强制停止"
            color="negative"
            @click="stopJobs"
            :loading="submitting"
          ></q-btn>
          <q-btn v-else label="立即运行" color="primary" @click="startJobs" :loading="submitting"></q-btn>
        </div>
      </header>
    </q-card>

    <q-separator class="q-my-md"/>

    <sub-job-pabel/>

    <q-separator class="q-my-md"/>

    <job-r-t-log-panel/>
  </q-page>
</template>

<script setup>
import {getJobsStatus, systemState} from 'src/store/systemState';
import { useQuasar } from 'quasar';
import {computed, onMounted, ref} from 'vue';
import JobApi from 'src/api/JobApi';
import { SystemMessage } from 'src/utils/Message';
import SubJobPabel from 'pages/jobs/SubJobPabel';
import JobRTLogPanel from 'pages/jobs/JobRTLogPanel';

const $q = useQuasar();

const submitting = ref(false);

const isRunning = computed(() => systemState.jobStatus?.status === 'running')

const startJobs = () => {
  $q.dialog({
    title: '立即运行任务？',
    cancel: true,
  }).onOk(async () => {
    submitting.value = true;
    const [, err] = await JobApi.start();
    submitting.value = false;
    if (err !== null) {
      SystemMessage.error(err.message);
      return;
    }
    getJobsStatus();
    SystemMessage.success('启动成功');
  });
};
const stopJobs = () => {
  $q.dialog({
    title: '强制停止当前任务？',
    cancel: true,
  }).onOk(async () => {
    submitting.value = true;
    const [, err] = await JobApi.stop();
    submitting.value = false;
    if (err !== null) {
      SystemMessage.error(err.message);
      return;
    }
    getJobsStatus();
    SystemMessage.success('停止成功');
  });
};

onMounted(() => {
  getJobsStatus();
})
</script>
