<template>
  <q-card flat>
    <header class="title">{{ title }}</header>
    <div class="row">
      <div>
        <q-circular-progress
          show-value
          font-size="12px"
          :value="progress"
          size="100px"
          :thickness="0.22"
          :color="color"
          track-color="grey-3"
          class="q-ma-md"
        >
          {{ progress }}%
        </q-circular-progress>
      </div>
      <div class="col column justify-center">
        <div>{{ current }} / {{ total }}</div>
        <div class="text-grey overflow-hidden ellipsis" style="width: 200px" :title="currentName">
          {{ currentName }}
        </div>
      </div>
    </div>
  </q-card>
</template>

<script setup>
import { computed } from 'vue';
import { isScanMovie, isScanSeries } from 'pages/overview/useJob';

const props = defineProps({
  title: String,
  current: Number,
  total: Number,
  currentName: String,
  color: {
    type: String,
    default: 'secondary',
  },
});

const progress = computed(() => {
  if (props.total === 0) return '100.00';
  // 最后一个处理任务时显示99.99
  if (props.total === props.current && (isScanMovie || isScanSeries)) return '99.99';
  const val = (props.current / props.total) * 100;
  return val.toFixed(2);
});
</script>

<style scoped>
.title {
  font-size: 14px;
  font-weight: bold;
  padding-left: 30px;
}
</style>
