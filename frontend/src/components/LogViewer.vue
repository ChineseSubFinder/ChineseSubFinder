<template>
  <div class="log-viewer col bg-grey-2 overflow-auto relative-position" :key="logLines[0]?.date_time">
    <q-virtual-scroll ref="logArea" class="full-height q-pa-sm" :items="logLines">
      <template v-slot="{ item, index }">
        <div :key="index" style="white-space: nowrap; line-height: 2">
          {{ getTextLogLine(item) }}
        </div>
      </template>
    </q-virtual-scroll>
    <copy-to-clipboard-btn
      v-if="logLines.length"
      class="copy-btn hidden absolute-top-right q-ma-md"
      :get-text="getTextLogs"
    />
  </div>
</template>

<script setup>
// 自动滚动到底部
import { nextTick, watch } from 'vue';
import { templateRef } from '@vueuse/core';
import CopyToClipboardBtn from 'components/CopyToClipboardBtn';

const props = defineProps({
  logLines: {
    type: Array,
    default: () => [],
  },
});

const logArea = templateRef('logArea');

// eslint-disable-next-line camelcase
const getTextLogLine = ({ level, date_time, content }) => `[${level}]: ${date_time} - ${content}`;

const getTextLogs = () => props.logLines.map(getTextLogLine).join('\n');

watch(
  () => props.logLines.length,
  () => {
    const element = logArea.value.$el;
    // console.log(element.scrollTop, element.clientHeight, element.scrollHeight);
    // 如果当前正处于底部，则自动滚动
    if (element.scrollTop + element.clientHeight >= element.scrollHeight - 10) {
      nextTick(() => {
        logArea.value.scrollTo(props.logLines.length - 1);
      });
    }
  }
);
</script>

<style lang="scss" scoped>
.log-viewer:hover {
  .copy-btn {
    display: block !important;
  }
}
</style>
