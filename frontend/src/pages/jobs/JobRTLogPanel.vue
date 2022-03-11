<template>
  <div class="text-bold">运行日志：</div>
  <log-viewer class="q-mt-sm" :log-lines="rtLogLines" style="height: 400px" />
  <div class="text-grey-10 q-mt-sm">
    * 如果运行中出现问题，欢迎到Github
    <span class="text-primary cursor-pointer" @click="gotoGithubIssuePage">提交反馈</span
    >，提交反馈时请附上出错日志，任务运行结束后可以在日志页面下载日志。
  </div>
</template>

<script setup>
import { useRealTimeLog } from 'src/composables/useRealTimeLog';
import LogViewer from 'components/LogViewer';
import { gotoGithubIssuePage } from 'src/utils/CommonUtils';

const { logLines: rtLogLines } = useRealTimeLog();
</script>
