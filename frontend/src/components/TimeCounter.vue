<template>
  <span>
    <slot :days="days" :hours="hours" :minutes="minutes" :seconds="seconds"></slot>
  </span>
</template>

<script setup>
import {computed, onBeforeUnmount, onMounted, ref} from 'vue';
import dayjs from 'dayjs';

const props = defineProps({
  // 单位秒
  time: {
    type: Number,
    required: true,
  },
  interval: {
    type: Number,
    default: 1,
  },
});

const nowTime = ref(0);

const dur = computed(() => Math.abs(props.time - nowTime.value));

const days = computed(() => Math.floor(dur.value / (24 * 60 * 60)));
const hours = computed(() => Math.floor((dur.value % (24 * 60 * 60)) / (60 * 60)));
const minutes = computed(() => Math.floor((dur.value % (60 * 60)) / 60));
const seconds = computed(() => Math.floor(dur.value % 60));

const updateNowTime = () => {
  nowTime.value = dayjs().unix();
};

updateNowTime();

const timer = ref(null);

onMounted(() => {
  timer.value = setInterval(updateNowTime, props.interval * 1000);
})

onBeforeUnmount(() => {
  clearInterval(timer.value);
})

</script>
