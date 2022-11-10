<template>
  <div v-if="isInQueue" class="row items-center q-gutter-xs">
    <q-spinner-hourglass color="primary" size="22px" />
    <div v-if="!dense" style="font-size: 90%">字幕上传中</div>
  </div>
  <q-btn
    v-else
    color="primary"
    flat
    dense
    icon="upload"
    v-bind="$attrs"
    :label="dense ? '' : '上传本地字幕'"
    @click="handleUploadClick"
    title="上传本地字幕"
  />

  <q-input v-show="false" type="file" ref="qFile" v-model="uploadFile" accept=".srt,.ass,.ssa,.sbv,.webvtt" />
</template>

<script setup>
import { ref, watch, watchEffect } from 'vue';
import { getSubtitleUploadList, subtitleUploadList } from 'pages/library/use-library';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/message';
import eventBus from 'vue3-eventbus';

const props = defineProps({
  path: String,
  dense: {
    type: Boolean,
    default: false,
  },
});

const uploadFile = ref(null);
const qFile = ref(null);
const isInQueue = ref(false);

watchEffect(() => {
  isInQueue.value = subtitleUploadList.value.some((item) => item.video_f_path === props.path);
});

const handleUploadClick = () => {
  qFile.value.$el.click();
};

const upload = async () => {
  const formData = new FormData();
  formData.append('video_f_path', props.path);
  formData.append('file', uploadFile.value[0]);
  isInQueue.value = true;
  await LibraryApi.uploadSubtitle(formData);
  SystemMessage.success('字幕上传成功。如果设置开启了“自动校正时间轴”，处理需要一些时间，请耐心等待', {
    timeout: 3000,
  });
  await getSubtitleUploadList();
  eventBus.emit('subtitle-uploaded');
};

watch(uploadFile, () => {
  upload();
});
</script>
