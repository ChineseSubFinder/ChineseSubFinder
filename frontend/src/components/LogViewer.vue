<template>
  <div class="col bg-grey-2 overflow-auto q-pa-sm" :key="logType + currentItem?.log_lines[0]?.date_time">
    <q-virtual-scroll
      ref="logArea"
      class="full-height"
      :items="logLines"
    >
      <template v-slot="{ item, index }">
        <div :key="index" style="white-space: nowrap; line-height: 2">
          {{ getTexLogLine(item) }}
        </div>
      </template>
    </q-virtual-scroll>
  </div>
</template>

<script setup>
// 自动滚动到底部
import {nextTick, watch} from 'vue';
import {templateRef} from '@vueuse/core';

const props = defineProps({
  logLines: {
    type: Array,
    default: () => [],
  }
});

const logArea = templateRef('logArea');

// eslint-disable-next-line camelcase
const getTexLogLine = ({ level, date_time, content }) => `[${level}]: ${date_time} - ${content}`;

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
