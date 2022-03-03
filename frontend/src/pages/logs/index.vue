<template>
  <fix-height-q-page class="flex row q-pa-md">
    <q-card class="row col" flat bordered>
      <div class="full-height q-pa-md">
        <q-list style="width: 220px">
          <q-item clickable @click="logType = 'rt'" :active="logType === 'rt'" active-class="text-primary text-bold">
            <q-item-section>实时日志</q-item-section>

            <q-item-section side>
              <q-btn
                flat
                round
                icon="file_download"
                size="sm"
                @click.stop="downloadLog(rtLogLines)"
                title="下载日志"
              ></q-btn>
            </q-item-section>
          </q-item>

          <q-separator class="q-my-xs" />

          <q-item-label header>历史日志</q-item-label>
          <q-item
            v-for="item in logList"
            :key="item.index"
            :active="logType === 'history' && item.index === currentIndex"
            active-class="text-primary text-bold"
            @click="handleHistoryItemClick(item)"
            clickable
          >
            <q-item-section>{{ item.log_lines[0]?.date_time }}</q-item-section>
            <q-item-section side>
              <q-btn
                flat
                round
                icon="file_download"
                size="sm"
                @click.stop="downloadLog(item.log_lines)"
                title="下载日志"
              ></q-btn>
            </q-item-section>
          </q-item>
        </q-list>
      </div>
      <q-separator vertical />
      <div class="full-height col bg-grey-2 overflow-auto" :key="logType + currentItem?.log_lines[0]?.date_time">
        <q-virtual-scroll
          v-model.number="virtualListIndex"
          ref="logArea"
          class="full-height q-pa-sm"
          :items="currentLogLines"
          :items-size="1000"
        >
          <template v-slot="{ item, index }">
            <div :key="index" style="white-space: nowrap; line-height: 2">
              {{ getTexLogLine(item) }}
            </div>
          </template>
        </q-virtual-scroll>
      </div>
    </q-card>
  </fix-height-q-page>
</template>

<script setup>
import { useLogList } from 'pages/logs/useLogList';
import FixHeightQPage from 'components/FixHeightQPage';
import { saveText } from 'src/utils/FileDownload';
import { useRealTimeLog } from 'pages/logs/useRealTimeLog';
import { computed, nextTick, ref, watch } from 'vue';
import { templateRef } from '@vueuse/core';

const { logList, currentIndex, currentItem } = useLogList();
const logType = ref('rt'); // rt or history

const { logLines: rtLogLines } = useRealTimeLog();

const handleHistoryItemClick = (item) => {
  currentIndex.value = item.index;
  logType.value = 'history';
};

// eslint-disable-next-line camelcase
const getTexLogLine = ({ level, date_time, content }) => `[${level}]: ${date_time} - ${content}`;

const getTextLogLines = (logLines = []) => logLines.map(getTexLogLine);

const getTextLogContent = (logLines = []) => getTextLogLines(logLines).join('\n');

const currentLogLines = computed(() => {
  const lines = logType.value === 'rt' ? rtLogLines.value : currentItem.value?.log_lines;
  return lines || [];
});

const logArea = templateRef('logArea');

// 自动滚动到底部
watch(
  () => rtLogLines.value.length,
  () => {
    if (logType.value !== 'rt') return;
    const element = logArea.value.$el;
    // console.log(element.scrollTop, element.clientHeight, element.scrollHeight);
    // 如果当前正处于底部，则自动滚动
    if (element.scrollTop + element.clientHeight >= element.scrollHeight - 10) {
      nextTick(() => {
        logArea.value.scrollTo(rtLogLines.value.length - 1);
      });
    }
  }
);

const downloadLog = (logLines) => {
  const filename = `${logLines[0]?.date_time || 'output'}.log`;
  saveText(filename, getTextLogContent(logLines));
};
</script>
