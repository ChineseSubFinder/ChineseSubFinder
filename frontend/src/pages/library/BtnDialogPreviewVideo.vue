<template>
  <q-btn color="primary" icon="smart_display" flat dense v-bind="$attrs" @click="visible = true" title="预览" />

  <q-dialog
    v-model="visible"
    persistent
    transition-show="slide-up"
    transition-hide="slide-down"
    maximized
    @before-show="handleBeforeShow"
    @before-hide="handleBeforeHide"
  >
    <q-card class="column">
      <q-bar>
        <div class="text-bold">字幕预览</div>
        <q-space />
        <q-btn dense flat icon="close" v-close-popup title="关闭" />
      </q-bar>

      <q-card-section class="col row items-center justify-center">
        <div class="row q-pa-md justify-center" v-if="isJobSuccess">
          <artplayer :option="artOption" style="height: 80vh; width: calc(1920 / 1080 * 80vh)"></artplayer>
        </div>
        <div v-else-if="isJobFailed || isJobNotExists" class="column items-center">
          <div class="text-grey-4" style="font-size: 120px; letter-spacing: 10px">(;-;)</div>
          <div class="text-negative q-mt-md text-bold">
            {{ jobResult.message || '未知错误' }}
          </div>
          <div class="q-mt-md">
            <q-btn label="点我重试" @click="preview" outline color="primary" />
          </div>
        </div>
        <div v-else class="column items-center">
          <q-spinner-cube size="xl" color="primary" />
          <div class="q-mt-sm text-grey-8">视频转码中...</div>
        </div>
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { computed, ref } from 'vue';
import Hls from 'hls.js';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/Message';
import useInterval from 'src/composables/useInterval';
import { until } from '@vueuse/core';
import Artplayer from 'components/Artplayer';
import config from 'src/config';
import { useQuasar } from 'quasar';

const $q = useQuasar();

const START_TIME = '0';
const END_TIME = '300';

const getPreviewUrl = (url) => `${config.BACKEND_URL}/static/preview/${url}`;

const props = defineProps({
  path: String,
  subList: {
    type: Array,
    default: () => [],
  },
});

const visible = ref(false);
const loading = ref(false);

const selectedSub = ref(null);
const previewInfo = ref(null);
const jobResult = ref(null);

const isJobSuccess = computed(() => jobResult.value?.message === 'ok');
const isJobNotExists = computed(() => jobResult.value?.message === '');
const isJobFailed = computed(() => jobResult.value?.message && !isJobSuccess.value && !isJobNotExists.value);

const submitForm = computed(() => ({
  video_f_path: props.path,
  sub_f_path: selectedSub.value,
  start_time: START_TIME,
  end_time: END_TIME,
}));

const getIsInQueue = async () => {
  const [res] = await LibraryApi.checkIsPreviewInQueue(submitForm.value);

  return res?.message === 'true';
};

const { resetInterval: startCheckQueue, stopInterval: stopCheckQueue } = useInterval(
  async () => {
    loading.value = await getIsInQueue();
  },
  5000,
  false
);

const checkQueue = async () => {
  loading.value = true;
  startCheckQueue();
  await until(loading).toBe(false);
  stopCheckQueue();
};

const getPreviewInfo = async () => {
  const [res, err1] = await LibraryApi.getPreviewDistInfo(submitForm.value);
  if (err1 !== null) {
    SystemMessage.error(err1.message);
  } else {
    previewInfo.value = res;
  }
};

const preview = async () => {
  previewInfo.value = null;
  const isInQueue = await getIsInQueue();
  if (!isInQueue) {
    jobResult.value = null;
    const [, err] = await LibraryApi.addPreviewJob(submitForm.value);
    if (err !== null) {
      SystemMessage.error(err.message);
    } else {
      // 等待预览任务完成
      await checkQueue();
      const [res1] = await LibraryApi.getPreviewJobResult(submitForm.value);
      jobResult.value = res1;
    }
  }
  loading.value = false;
};

const artOption = computed(() => ({
  autoplay: true,
  autoSize: true,
  url: getPreviewUrl(previewInfo.value.video_f_path),
  subtitle: {
    url: getPreviewUrl(previewInfo.value.sub_f_path),
  },
  customType: {
    m3u8(video, url) {
      if (Hls.isSupported()) {
        const hls = new Hls();
        hls.loadSource(url);
        hls.attachMedia(video);
      } else if (video.canPlayType('application/vnd.apple.mpegurl')) {
        video.src = url;
      } else {
        // art.notice.show = '不支持播放格式：m3u8';
      }
    },
  },
  controls:
    props.subList.length === 0
      ? []
      : [
          {
            disable: false,
            name: 'button',
            index: 10,
            position: 'right',
            html: '选择字幕',
            tooltip: '选择字幕',
            style: {
              color: 'red',
            },
            click() {
              $q.dialog({
                title: '选择字幕',
                style: 'width: 800px',
                options: {
                  type: 'radio',
                  model: selectedSub.value,
                  items: props.subList.map((e) => ({ label: e, value: e })),
                },
                cancel: true,
                persistent: true,
              }).onOk((data) => {
                selectedSub.value = data;
                getPreviewInfo();
                preview();
              });
            },
            mounted() {
              // console.log('自定义按钮挂载完成1');
            },
          },
        ],
}));

const handleBeforeShow = () => {
  loading.value = true;
  selectedSub.value = props.subList?.[0];
  getPreviewInfo();
  preview();
};

const handleBeforeHide = () => {
  stopCheckQueue();
  previewInfo.value = null;
  jobResult.value = null;
  LibraryApi.cleanAllPreviewJobData();
};
</script>
