<template>
  <q-btn label="更新缓存" color="primary" icon="refresh" @click="confirm" :loading="refreshCacheLoading">
    <template v-slot:loading>
      <q-spinner-hourglass class="on-left" />
      更新缓存中...
    </template>
  </q-btn>
</template>

<script setup>
import { refreshCacheLoading, refreshLibrary } from 'pages/library/use-library';
import { useQuasar } from 'quasar';

const $q = useQuasar();

const confirm = () => {
  $q.dialog({
    title: '更新缓存',
    message:
      '刷新缓存并不会自动提交下载字幕的任务，仅仅是方便手动选择一个视频去下载字幕。这个是一个长耗时任务，请在有手动的需求下操作。暂时不支持动态更新缓存，需要手动执行完整的缓存刷新操作。',
    persistent: true,
    ok: '确定',
    cancel: '取消',
  })
    .onOk(() => {
      refreshLibrary();
    })
    .onCancel(() => {
      // console.log('>>>> Cancel')
    });
};
</script>
