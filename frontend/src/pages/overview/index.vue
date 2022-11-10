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
  </q-page>
</template>

<script setup>
import { getJobsStatus, isJobRunning, systemState } from 'src/store/systemState';
import { useQuasar } from 'quasar';
import { onMounted, ref } from 'vue';
import JobApi from 'src/api/JobApi';
import { SystemMessage } from 'src/utils/message';

const $q = useQuasar();

const submitting = ref(false);

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
</script>
