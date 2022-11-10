<template>
  <q-btn flat dense icon="content_copy" color="grey-8" @click="copy" v-bind="$attrs" />
</template>

<script setup>
import { copyToClipboard } from 'quasar';
import { SystemMessage } from 'src/utils/message';

const props = defineProps({
  text: String,
  getText: Function,
});

const copy = async () => {
  let copyPromise = null;
  if (typeof props.getText === 'function') {
    copyPromise = copyToClipboard(props.getText());
  } else {
    copyPromise = copyToClipboard(props.text);
  }

  await copyPromise;
  SystemMessage.success('已复制到剪贴板');
};
</script>
