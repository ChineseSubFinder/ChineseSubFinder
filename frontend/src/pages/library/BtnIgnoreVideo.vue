<template>
  <q-btn
    v-if="isSkipped"
    color="warning"
    round
    flat
    dense
    icon="lock"
    @click.stop
    title="当前视频已锁定，不进行字幕下载"
    @click="skip"
    v-bind="$attrs"
  />
  <q-btn
    v-else
    color="grey"
    round
    flat
    dense
    icon="lock"
    @click.stop
    title="点击锁定视频，不进行字幕下载"
    @click="skip"
    v-bind="$attrs"
  />
</template>

<script setup>
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/message';
import { useQuasar } from 'quasar';
import { onMounted, ref } from 'vue';
import useEventBus from 'src/composables/use-event-bus';
import { checkIsVideoLocked } from 'pages/library/use-library';

const props = defineProps({
  path: String,
  videoType: Number,
});

const $q = useQuasar();

const isSkipped = ref(null);

const getIsSkipped = async () => {
  isSkipped.value = await checkIsVideoLocked({
    video_type: props.videoType,
    physical_video_file_full_path: props.path,
    is_bluray: false,
    is_skip: true,
  });
};

const skip = async () => {
  $q.dialog({
    title: '提示',
    message: isSkipped.value ? '确定要解锁该视频吗？' : `确定要锁定该视频，不进行字幕下载吗？`,
    cancel: true,
    persistent: true,
  }).onOk(async () => {
    const [res] = await LibraryApi.setSkipInfo({
      video_skip_infos: [
        {
          video_type: props.videoType,
          physical_video_file_full_path: props.path,
          is_bluray: false,
          is_skip: !isSkipped.value,
        },
      ],
    });
    if (res) {
      SystemMessage.success('操作成功');
      getIsSkipped();
    }
  });
};

useEventBus(`refresh-skip-status-${props.path}`, (flag) => {
  if (flag !== undefined) {
    isSkipped.value = flag;
  } else {
    getIsSkipped();
  }
});

onMounted(() => {
  getIsSkipped();
});
</script>
