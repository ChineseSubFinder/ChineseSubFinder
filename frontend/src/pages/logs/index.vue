<template>
  <fix-height-q-page class="flex column q-pa-md">
    <div class="text-grey-8 q-mb-sm">
      * 如果运行中出现问题，欢迎到Github
      <span class="text-primary cursor-pointer" @click="gotoGithubIssuePage">提交反馈</span>，提交反馈时请附上出错日志。
    </div>
    <q-card class="row col" flat bordered>
      <div class="full-height q-pa-md">
        <q-list style="width: 220px">
          <q-item-label header>
            <div class="row items-center justify-between">
              <div>历史日志</div>
              <div>
                <q-select
                  style="width: 120px"
                  v-model="form.logCount"
                  :options="[
                    { label: '最近3次', value: 3 },
                    { label: '最近5次', value: 5 },
                    { label: '最近10次', value: 10 },
                    { label: '最近20次', value: 20 },
                  ]"
                  dense
                  outlined
                  map-options
                  emit-value
                />
              </div>
            </div>
          </q-item-label>
          <q-separator />
          <q-item
            v-for="item in logList"
            :key="item.index"
            :active="item.index === currentIndex"
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
      <log-viewer class="full-height" :log-lines="currentLogLines" :key="currentItem?.log_lines[0]?.date_time" />
    </q-card>
  </fix-height-q-page>
</template>

<script setup>
import { form, useLogList } from 'pages/logs/useLogList';
import FixHeightQPage from 'components/FixHeightQPage';
import { saveText } from 'src/utils/FileDownload';
import { computed } from 'vue';
import LogViewer from 'components/LogViewer';
import { getExportSettings, useSettings } from 'pages/settings/useSettings';
import { gotoGithubIssuePage } from 'src/utils/CommonUtils';

const { logList, currentIndex, currentItem } = useLogList();

useSettings();
const handleHistoryItemClick = (item) => {
  currentIndex.value = item.index;
};

// eslint-disable-next-line camelcase
const getTexLogLine = ({ level, date_time, content }) => `[${level}]: ${date_time} - ${content}`;

const getTextLogLines = (logLines = []) => logLines.map(getTexLogLine);

const getTextLogContent = (logLines = []) => getTextLogLines(logLines).join('\n');

const currentLogLines = computed(() => {
  const lines = currentItem.value?.log_lines;
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
