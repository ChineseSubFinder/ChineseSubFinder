<template>
  <div style="min-height: 300px">
    <q-banner
      v-if="apiLimitInfo"
      dense
      class="bg-grey-3"
      :class="{
        // 使用超过4/5时，显示黄色警告
        'bg-negative': apiLimitInfo.dailyCount >= apiLimitInfo.dayliLimit,
        'bg-warning': apiLimitInfo.dailyCount / apiLimitInfo.dayliLimit > 4 / 5,
      }"
    >
      <div class="text-bold">
        每日限制：{{ apiLimitInfo.dailyCount }} / {{ apiLimitInfo.dailyLimit }}，ApiKey 过期时间：{{
          dayjs.unix(apiLimitInfo.expireTime).format('YYYY-MM-DD HH:mm:ss')
        }}
      </div>
    </q-banner>
    <q-list v-if="csfSearchResult?.length" separator>
      <q-item v-for="(item, index) in csfSearchResult" :key="item.sub_sha256">
        <q-item-section> {{ index + 1 }}. {{ item.title }} </q-item-section>
        <q-item-section side>
          <div class="row">
            <btn-dialog-preview-video
              :path="path"
              :sub-list="[selectedSubUrl]"
              :on-btn-click="(callback) => handlePreviewClick(item, callback)"
              :subtitle-type="selectedItem?.ext.replace('.', '')"
            />
            <q-btn color="primary" icon="download" flat dense @click="handleDownloadCsfSub(item)" title="下载" />
          </div>
        </q-item-section>
      </q-item>
    </q-list>
    <div v-else-if="!loading" class="text-grey">
      <div>未搜索到数据，<q-btn flat label="重试" color="primary" dense @click="searchCsf" /></div>
      <div>如果报错信息提示没有 ApiKey，请到<b>配置中心-字幕源设置</b>，填写SubtitleBest的ApiKey</div>
    </div>
    <q-inner-loading :showing="loading">
      <q-spinner size="50px" color="primary" />
      <div>{{ loadingMsg }}</div>
      <div v-if="countdownLoading">预计 {{ nextRequestCountdownSecond }} 秒后取得数据</div>
    </q-inner-loading>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/message';
import CsfSubtitlesApi from 'src/api/CsfSubtitlesApi';
import BtnDialogPreviewVideo from 'pages/library/BtnDialogPreviewVideo.vue';
import { getSubtitleUploadList } from 'pages/library/use-library';
import eventBus from 'vue3-eventbus';
import { LocalStorage } from 'quasar';
import useInterval from 'src/composables/use-interval';
import useEventBus from 'src/composables/use-event-bus';
import dayjs from 'dayjs';

const props = defineProps({
  path: String,
  isMovie: {
    type: Boolean,
    default: false,
  },
  searchPackage: {
    type: Boolean,
    default: false,
  },
  season: {
    type: Number,
  },
  episode: {
    type: Number,
  },
});

// 上次请求时间
let lastRequestApiTime = LocalStorage.getItem('lastRequestApiTime') || 0;
// 最小请求间隔
const minRequestApiInterval = 15 * 1000;
// 下次请求倒数时间
const nextRequestCountdownSecond = ref(0);
// api限制信息
const apiLimitInfo = ref(null);
useEventBus('subtitle-best-api-limit-info', (info) => {
  apiLimitInfo.value = info;
});

useInterval(() => {
  const v = Math.ceil((lastRequestApiTime + minRequestApiInterval - Date.now()) / 1000);
  nextRequestCountdownSecond.value = v > 0 ? v : 0;
}, 100);

const loading = ref(false);
const countdownLoading = ref(false);
const loadingMsg = ref('');
const csfSearchResult = ref(null);
const selectedSubBlob = ref(null);
const selectedItem = ref(null);
const imdbId = ref(null);

// blob缓存
const cacheBlob = new Map();
const selectedSubUrl = computed(() => {
  if (selectedSubBlob.value) {
    return URL.createObjectURL(selectedSubBlob.value);
  }
  return null;
});

const checkOk = () => {
  const now = Date.now();
  if (now - lastRequestApiTime < minRequestApiInterval) {
    return false;
  }
  lastRequestApiTime = now;
  LocalStorage.set('lastRequestApiTime', now);
  return true;
};

const waitRequestReady = async () => {
  countdownLoading.value = true;
  // 每100ms检查一次，直到请求间隔大于最小请求间隔
  while (!checkOk()) {
    // eslint-disable-next-line no-await-in-loop
    await new Promise((resolve) => {
      setTimeout(resolve, 100);
    });
  }
  countdownLoading.value = false;
};

const searchCsf = async () => {
  loading.value = true;
  loadingMsg.value = '正在获取字幕列表...';
  const [d, e] = await LibraryApi.getImdbId({
    is_movie: props.isMovie,
    video_f_path: props.path,
  });
  if (e) {
    SystemMessage.error(e.message);
    loading.value = false;
    return;
  }
  imdbId.value = d?.ImdbId;
  await waitRequestReady();
  if (props.isMovie) {
    const [data, err] = await CsfSubtitlesApi.searchMovie({
      imdb_id: imdbId.value,
    });
    if (err !== null) {
      SystemMessage.error(err.message);
    } else {
      csfSearchResult.value = data.subtitles;
    }
  } else if (!props.searchPackage) {
    const [data, err] = await CsfSubtitlesApi.searchTvEps({
      imdb_id: imdbId.value,
      season: props.season,
      episode: props.episode,
    });
    if (err !== null) {
      SystemMessage.error(err.message);
    } else {
      csfSearchResult.value = data.subtitles;
    }
  } else {
    // TODO: search package
  }
  loadingMsg.value = '';
  loading.value = false;
};

const fetchSubtitleBlob = async (item) => {
  selectedItem.value = item;
  if (cacheBlob.has(item.sub_sha256)) {
    selectedSubBlob.value = cacheBlob.get(item.sub_sha256);
    return;
  }
  selectedSubBlob.value = null;
  loading.value = true;
  loadingMsg.value = '正在获取下载地址...';
  await waitRequestReady();

  loadingMsg.value = '正在下载字幕...';
  const [data, err] = await CsfSubtitlesApi.getDownloadUrl({
    ...item,
    imdb_id: imdbId.value,
  });
  if (err !== null) {
    SystemMessage.error(err.message);
  } else {
    // fetch资源，获取blob url
    const res = await fetch(data.download_link);
    const blob = await res.blob();
    cacheBlob.set(item.sub_sha256, blob);
    selectedSubBlob.value = blob;
  }
  loadingMsg.value = '';
  loading.value = false;
};

const handleDownloadCsfSub = async (item) => {
  await fetchSubtitleBlob(item);

  if (!selectedSubBlob.value) {
    return;
  }

  // 上传
  const formData = new FormData();
  formData.append('video_f_path', props.path);
  formData.append('file', new File([selectedSubBlob.value], item.title, { type: 'text/plain' }));
  await LibraryApi.uploadSubtitle(formData);
  await getSubtitleUploadList();
  eventBus.emit('subtitle-uploaded');

  SystemMessage.success('已下载到库中');
};

const handlePreviewClick = async (item, callback) => {
  await fetchSubtitleBlob(item);
  if (selectedSubUrl.value) {
    callback(true);
  } else {
    callback(false);
  }
};

onMounted(() => {
  searchCsf();
});
</script>
