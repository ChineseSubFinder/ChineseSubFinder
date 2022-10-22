<template>
  <div v-if="isInQueue" class="row items-center q-gutter-xs">
    <q-spinner-hourglass color="primary" size="22px" />
    <div style="font-size: 90%">字幕上传中</div>
  </div>
  <q-btn v-else color="primary" round flat dense icon="upload" title="上传字幕" @click="handleUploadClick" />

  <q-input v-show="false" type="file" ref="qFile" v-model="uploadFile" accept=".srt,.ass,.ssa,.sbv,.webvtt" />
</template>

<script setup>
import { ref, watch, watchEffect } from 'vue';
import { getSubtitleUploadList, subtitleUploadList } from 'pages/library/useLibrary';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/Message';

const props = defineProps({
  data: Object,
});

const uploadFile = ref(null);
const qFile = ref(null);
const isInQueue = ref(false);

watchEffect(() => {
  isInQueue.value = subtitleUploadList.value.some(
    (item) => item.video_f_path === props.data.video_f_path && item.sub_f_path === props.data.sub_f_path
  );
});

const handleUploadClick = () => {
  qFile.value.$el.click();
};

const upload = async () => {
  const formData = new FormData();
  formData.append('video_f_path', props.data.video_f_path);
  formData.append('file', uploadFile.value[0]);
  isInQueue.value = true;
  await LibraryApi.uploadSubtitle(formData);
  SystemMessage.success('字幕上传成功');
  await getSubtitleUploadList();
};

watch(uploadFile, () => {
  upload();
});
</script>
