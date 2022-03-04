<template>
  <div class="text-bold">运行日志：</div>
  <log-viewer class="q-mt-sm" :log-lines="rtLogLines" style="height: 400px;" />
</template>

<script setup>
import {useRealTimeLog} from 'src/composables/useRealTimeLog';
import LogViewer from 'components/LogViewer';

const { logLines: rtLogLines } = useRealTimeLog();
</script>
