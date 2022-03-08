<template>
  <div v-if="subJobsDetail">
    <div class="q-my-xs">
      任务状态：
      <q-badge outline v-if="isPreparing" color="secondary">已启动</q-badge>
      <q-badge outline v-if="isRunning" color="positive">运行中</q-badge>
      <q-badge outline v-else-if="isWaiting" color="warning">等待下一次定时任务启动</q-badge>
    </div>

    <div class="text-grey row items-center">
      <div>
        <template v-if="isRunning">
          已运行：<time-counter :time="startTimestamp" v-slot="props">
            {{ getTimeCountString(props) }}
          </time-counter>
        </template>
        <template v-else-if="isWaiting">
          <time-counter :time="startTimestamp" v-slot="props">
            {{ getTimeCountString(props) }}
          </time-counter>
          后开始
        </template>
      </div>
    </div>

    <section v-if="isRunning" class="row q-mt-md">
      <job-progress-card
        title="任务进度"
        :current="subJobsDetail.working_unit_index"
        :total="subJobsDetail.unit_count"
        :current-name="subJobsDetail.working_unit_name"
      />

      <job-progress-card
        v-if="subJobsDetail.video_count"
        title="子任务（视频）进度"
        :current="subJobsDetail.working_video_index"
        :total="subJobsDetail.video_count"
        :current-name="subJobsDetail.working_video_name"
      />
    </section>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import JobProgressCard from 'pages/jobs/JobProgressCard';
import dayjs from 'dayjs';
import TimeCounter from 'components/TimeCounter';
import { subJobsDetail, isWaiting, isPreparing, isRunning } from 'pages/jobs/useJob';

const startTimestamp = computed(() => dayjs(subJobsDetail.value.started_time).unix());

const getTimeCountString = (props) => {
  const { days, hours, minutes, seconds } = props;
  return `${days ? `${days} 天 ` : ''}${hours} 小时 ${minutes} 分 ${seconds} 秒`;
};
</script>
