<template>
  <q-page class="q-pa-md">
    <q-card class="q-pa-md" flat>
      <header class="row items-center">
        <div>
          当前任务状态：
          <q-badge v-if="systemState.running" color="positive">运行中</q-badge>
          <q-badge v-else color="grey">未运行</q-badge>
        </div>
        <q-space/>
        <q-btn
          v-if="systemState.running"
          label="强制停止"
          color="negative"
          @click="stopJobs"
          :loading="submitting"
        ></q-btn>
        <q-btn v-else label="立即运行" color="primary" @click="startJobs" :loading="submitting"></q-btn>
      </header>
      <q-separator class="q-my-md"/>
      <job-list-table class="q-mt-md"/>
    </q-card>
  </q-page>
</template>

<script setup>
import { systemState } from 'src/store/systemState';
import { useQuasar } from 'quasar';
import { ref } from 'vue';
import JobApi from 'src/api/JobApi';
import { SystemMessage } from 'src/utils/Message';
import JobListTable from 'pages/jobs/JobListTable';

const $q = useQuasar();

const submitting = ref(false);

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
    SystemMessage.success('停止成功');
  });
};
</script>
