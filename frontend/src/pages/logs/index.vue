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
      <log-viewer
        class="full-height"
        :log-lines="currentLogLines"
        :key="logType + currentItem?.log_lines[0]?.date_time"
      />
    </q-card>
  </fix-height-q-page>
</template>

<script setup>
import { useLogList } from 'pages/logs/useLogList';
import FixHeightQPage from 'components/FixHeightQPage';
import { saveText } from 'src/utils/FileDownload';
import { computed, ref } from 'vue';
import { useRealTimeLog } from 'src/composables/useRealTimeLog';
import LogViewer from 'components/LogViewer';
import {getExportSettings, useSettings} from 'pages/settings/useSettings';

const { logList, currentIndex, currentItem } = useLogList();
const logType = ref('rt'); // rt or history

useSettings();
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

const downloadLog = (logLines) => {
  const filename = `${logLines[0]?.date_time || 'output'}.log`;
  const configString = JSON.stringify(getExportSettings(), null, 2);
  const logString = getTextLogContent(logLines);
  const content = `config:\n${configString}\n\nlog:\n${logString}`;
  saveText(filename, content);
};
</script>
