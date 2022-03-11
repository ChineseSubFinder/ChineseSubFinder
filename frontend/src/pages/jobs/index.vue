<template>
  <q-page class="q-pa-md">
    <q-card v-if="systemState.jobStatus" flat>
      <header class="row items-center q-gutter-lg">
        <div>
          <div>
            守护进程：
            <q-badge v-if="isJobRunning" color="positive">运行中</q-badge>
            <q-badge v-else color="grey">未运行</q-badge>
          </div>
          <div class="text-grey">用于处理定时任务、启动扫描程序</div>
        </div>
        <div>
          <q-btn v-if="isJobRunning" label="强制停止" color="negative" @click="stopJobs" :loading="submitting"></q-btn>
          <q-btn v-else label="立即运行" color="primary" @click="startJobs" :loading="submitting"></q-btn>
        </div>
      </header>
    </q-card>

    <template v-if="subJobsDetail && isJobRunning">
      <q-separator class="q-my-md" />

      <job-detail-panel />

      <q-separator class="q-my-md" />

      <job-r-t-log-panel />
    </template>

    <div v-else-if="isJobRunning" class="q-mt-lg row items-center">
      <q-spinner-facebook color="primary" size="2em" />
      <div class="text-primary q-ml-sm">正在获取任务执行情况...</div>
    </div>
  </q-page>
</template>

<script setup>
import { getJobsStatus, isJobRunning, systemState } from 'src/store/systemState';
import { useQuasar } from 'quasar';
import { onBeforeUnmount, onMounted, ref } from 'vue';
import JobApi from 'src/api/JobApi';
import { SystemMessage } from 'src/utils/Message';
import JobRTLogPanel from 'pages/jobs/JobRTLogPanel';
import { subJobsDetail, useJob } from 'pages/jobs/useJob';
import JobDetailPanel from 'pages/jobs/JobDetailPanel';
import { wsManager } from 'src/composables/useWebSocketApi';

const $q = useQuasar();

const submitting = ref(false);

useJob();

const startJobs = () => {
  $q.dialog({
    title: '是否立即运行？',
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
    title: '是否强制停止？',
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
});

onBeforeUnmount(() => {
  wsManager.close();
  wsManager.ws = null;
});
</script>
