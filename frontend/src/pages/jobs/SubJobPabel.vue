<template>
  <div v-if="subJobsDetail">
    <div class="text-grey row items-center">
      <div>
        开始于 {{ subJobsDetail.started_time }}
      </div>

      <div class="q-my-xs q-ml-md">
        <q-badge outline v-if="subJobsDetail.status === 'running'" color="positive">运行中</q-badge>
        <q-badge outline v-else-if="subJobsDetail.status === 'waiting'" color="warning">等待中</q-badge>
      </div>
    </div>

    <section class="row q-mt-md">
      <job-progress-card
        title="子任务进度"
        :current="subJobsDetail.working_unit_index"
        :total="subJobsDetail.unit_count"
        :current-name="subJobsDetail.working_unit_name"
      />

      <job-progress-card
        title="视频进度"
        :current="subJobsDetail.working_video_index"
        :total="subJobsDetail.video_count"
        :current-name="subJobsDetail.working_video_name"
      />

    </section>
  </div>
</template>

<script setup>
import { useWebSocketApi } from 'src/composables/useWebSocketApi';
import { ref } from 'vue';
import JobProgressCard from 'pages/jobs/JobProgressCard';

const subJobsDetail = ref(null);

useWebSocketApi('sub_download_jobs_status', (data) => {
  subJobsDetail.value = data;
});
</script>
