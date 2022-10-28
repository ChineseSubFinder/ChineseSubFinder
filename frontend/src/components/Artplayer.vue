<template><div ref="artRef"></div></template>
<script setup>
import Artplayer from 'artplayer';
import { nextTick, onBeforeUnmount, onMounted, ref, watchEffect } from 'vue';

const props = defineProps({
  option: { type: Object, required: true },
});

const emit = defineEmits(['get-instance']);

const instance = ref(null);
const artRef = ref(null);

const setup = () => {
  instance.value = new Artplayer({
    ...props.option,
    container: artRef.value,
  });
  nextTick(() => {
    emit('get-instance', instance.value);
  });
};

watchEffect(
  () => props.option,
  () => {
    instance.value?.destroy();
    setup();
  }
);

onMounted(() => {
  setup();
});

onBeforeUnmount(() => {
  instance.value?.destroy();
});
</script>
