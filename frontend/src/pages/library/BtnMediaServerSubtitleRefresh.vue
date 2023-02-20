<template>
  <q-btn label="更新媒体服务器字幕" v-if="show" color="primary" icon="cached" @click="confirm">
    <template v-slot:loading>
      <q-spinner-hourglass class="on-left" />
      更新缓存中...
    </template>
  </q-btn>
</template>

<script setup>
import { useQuasar } from 'quasar';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/message';
import { computed } from 'vue';
import { settingsState } from 'src/store/settingsState';

const $q = useQuasar();

const show = computed(() => settingsState.settings?.emby_settings.enable);

const confirm = () => {
  $q.dialog({
    title: '更新媒体服务器字幕',
    message:
      '此操作会刷新最近的10000个视频的字幕，可能需要一段时间才能生效。' +
      '提交后会触发媒体服务器的刷新，异步操作，无需频繁点击触发。服务器刷新生效也需要看各自的媒体文件数量。',
    persistent: true,
    ok: '确定',
    cancel: '取消',
  }).onOk(async () => {
    const [, err] = await LibraryApi.refreshMediaServerSubList();
    if (err !== null) {
      SystemMessage.error(err.message);
    } else {
      SystemMessage.success('刷新成功');
    }
  });
};
</script>
