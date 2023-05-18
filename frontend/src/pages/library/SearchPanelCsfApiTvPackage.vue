<template>
  <div :key="currentVideoFilePath" style="min-height: 300px">
    <subtitle-best-api-limit-banner v-if="!hideLimit" />

    <q-splitter v-if="packages.length" class="q-mt-md overflow-hidden" v-model="splitterModel" unit="px">
      <template v-slot:before>
        <div class="q-px-md text-bold">字幕包列表</div>
        <q-list>
          <q-item
            v-for="(item, index) in packages"
            :key="item"
            clickable
            @click="selectedPackage = item"
            :active="selectedPackage === item"
          >
            <q-item-section class="overflow-hidden ellipsis no-wrap"> {{ index + 1 }}. {{ item }} </q-item-section>
          </q-item>
        </q-list>
      </template>

      <template v-slot:after>
        <template v-if="csfSearchResult?.length">
          <div class="q-px-md row justify-between">
            <div class="text-bold">字幕列表</div>
            <q-btn v-if="!hideDownload" label="下载全部" color="primary" icon="download" flat @click="downloadAll" />
          </div>
          <q-list separator>
            <q-item v-for="(item, index) in csfSearchResult" :key="item.sub_sha256">
              <q-item-section>
                <div class="row items-center q-gutter-sm">
                  <div>{{ index + 1 }}. {{ item.title }}</div>
                  <q-badge color="primary">{{ LANGUAGES[item.language] }}</q-badge>
                  <q-badge
                    v-if="cacheBlob[item.sub_sha256]"
                    color="grey"
                    title="已缓存到浏览器，再次预览和下载不消耗次数，关闭窗口后失效"
                    >已缓存</q-badge
                  >
                  <q-badge
                    v-if="downloadedEpisodeSubtitleMap[item.episode] === item.sub_sha256"
                    color="grey"
                    title="视频已下载该字幕"
                    >已下载</q-badge
                  >
                </div>
              </q-item-section>
              <q-item-section side>
                <div class="row" v-if="getVideoFilePathOfSubtitle(item)">
                  <btn-dialog-preview-video
                    :disable="!getVideoFilePathOfSubtitle(item)"
                    :path="getVideoFilePathOfSubtitle(item)"
                    :subtitle-url-list="[selectedSubUrl]"
                    :on-btn-click="(callback) => handlePreviewClick(item, callback)"
                    :subtitle-type="selectedItem?.ext.replace('.', '')"
                  />
                  <q-btn
                    v-if="!hideDownload"
                    color="primary"
                    icon="download"
                    flat
                    dense
                    @click="handleDownloadCsfSub(item, true, true)"
                    title="下载"
                  />
                </div>
                <div v-else class="text-grey">找不到该字幕对应的视频</div>
              </q-item-section>
            </q-item>
          </q-list>
        </template>
        <div v-else class="text-grey q-pa-md">请先选择字幕包</div>
      </template>
    </q-splitter>
    <div v-else-if="!loading" class="text-grey">
      <div>
        <span>ImdbId: {{ imdbId || '-' }}</span>
        <span v-if="imdbId && !isImdbId(imdbId)" class="q-ml-md text-negative">这是个无效的ImdbId</span>
      </div>
      <template v-if="tmdbErrorMsg">
        <div class="text-negative">{{ tmdbErrorMsg }}</div>
        <div><q-btn flat label="重试" color="primary" dense @click="searchPackages" /></div>
      </template>
      <template v-else-if="subtitleBestApiErrorMsg">
        <div class="text-negative">获取字幕列表失败，错误信息：{{ subtitleBestApiErrorMsg }}</div>
        <div><q-btn flat label="重试" color="primary" dense @click="searchPackages" /></div>
      </template>
      <template v-else>
        <div>未搜索到数据，<q-btn flat label="重试" color="primary" dense @click="searchPackages" /></div>
        <div>如果报错信息提示没有 ApiKey，请到<b>配置中心-字幕源设置</b>，填写SubtitleBest的ApiKey</div>
      </template>
    </div>
    <q-inner-loading :showing="loading || isDownloadingAll">
      <q-spinner size="50px" color="primary" />
      <div v-if="isDownloadingAll" class="text-bold">({{ downloadedCountOfList }} / {{ episodes.length }})</div>
      <div>{{ loadingMsg }}</div>
      <div v-if="countdownLoading">预计 {{ nextRequestCountdownSecond }} 秒后取得数据</div>
    </q-inner-loading>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/message';
import CsfSubtitlesApi from 'src/api/CsfSubtitlesApi';
import BtnDialogPreviewVideo from 'pages/library/BtnDialogPreviewVideo.vue';
import { getSubtitleUploadList } from 'pages/library/use-library';
import eventBus from 'vue3-eventbus';
import { useQuasar } from 'quasar';
import { LANGUAGES } from 'src/constants/LibraryConstants';
import { VIDEO_TYPE_TV } from 'src/constants/SettingConstants';
import { settingsState } from 'src/store/settingsState';
import SubtitleBestApiLimitBanner from 'components/SubtitleBestApiLimitBanner.vue';
import { useApiLimit } from 'src/composables/use-api-limit';
import { isImdbId } from 'src/utils/common';
import CsfSubtitlesShareApi from 'src/api/CsfSubtitlesShareApi';

const props = defineProps({
  episodes: Array,
  // 隐藏下载按钮
  hideDownload: {
    type: Boolean,
    default: false,
  },
  // 隐藏额度
  hideLimit: {
    type: Boolean,
    default: false,
  },
  // 使用用户共享的字幕API替换默认API
  useUserShareApi: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['getResult']);

const $q = useQuasar();
const { nextRequestCountdownSecond, countdownLoading, waitRequestReady } = useApiLimit(
  'lastSubtitleBestRequestTime',
  5 * 1000
);

const splitterModel = ref(300);
const tmdbErrorMsg = ref('');
const subtitleBestApiErrorMsg = ref('');
const packages = ref([]);
const selectedPackage = ref(null);

const loading = ref(false);
const isDownloadingAll = ref(false);
const loadingMsg = ref('');
const csfSearchResult = ref(null);
const selectedSubBlob = ref(null);
const selectedItem = ref(null);
const imdbId = ref(null);

const subtitleApi = computed(() => (props.useUserShareApi ? CsfSubtitlesShareApi : CsfSubtitlesApi));

const currentVideoFilePath = computed(() => props.episodes[0]?.video_f_path);
const currentSeason = computed(() => props.episodes[0]?.season);

// blob缓存
const cacheBlob = ref({});
const downloadedEpisodeSubtitleMap = ref({});
const downloadedCountOfList = computed(
  () => csfSearchResult?.value.filter((e) => e.sub_sha256 === downloadedEpisodeSubtitleMap.value[e.episode]).length || 0
);

const selectedSubUrl = computed(() => {
  if (selectedSubBlob.value) {
    return URL.createObjectURL(selectedSubBlob.value);
  }
  return null;
});

const setLock = async (paths) => {
  const [, err] = await LibraryApi.setSkipInfo({
    video_skip_infos: paths.map((path) => ({
      video_type: VIDEO_TYPE_TV,
      physical_video_file_full_path: path,
      is_bluray: false,
      is_skip: true,
    })),
  });
  if (err !== null) {
    SystemMessage.error(err.message);
  } else {
    // 通知列表页锁定成功
    paths.forEach((path) => eventBus.emit(`refresh-skip-status-${path}`, true));
    SystemMessage.success('操作成功');
  }
};

const searchPackages = async () => {
  loading.value = true;
  subtitleBestApiErrorMsg.value = '';
  loadingMsg.value = '正在从TMDB获取视频详细信息...';
  const [d, e] = await LibraryApi.getImdbId({
    is_movie: false,
    video_f_path: currentVideoFilePath.value,
  });
  if (e) {
    if (settingsState.settings.advanced_settings.tmdb_api_settings.enable) {
      tmdbErrorMsg.value =
        '从 TMDB 获取数据失败，检测到你当前正在使用自己的 TMDB ApiKey ，请检查进阶设置-TMDB API中设置的ApiKey是否有效。';
    } else {
      tmdbErrorMsg.value =
        '从 TMDB 获取数据失败，检测到你正在使用公共查询接口，可能是使用人数过多导致查询失败，可以尝试在进阶设置里启用 TMDB API，填写自己的 ApiKey。';
    }
    loading.value = false;
    loadingMsg.value = '';
    SystemMessage.error(e.message);
    return;
  }
  tmdbErrorMsg.value = '';
  loadingMsg.value = '正在从 SubtitleBest 获取字幕包列表...';
  imdbId.value = d?.ImdbId;

  if (!isImdbId(imdbId.value)) {
    loadingMsg.value = '';
    loading.value = false;
    return;
  }

  const [data, err] = await subtitleApi.value.searchTvSeasonPackage({
    imdb_id: imdbId.value,
    season: currentSeason.value,
  });
  if (err !== null) {
    SystemMessage.error(err.message);
    subtitleBestApiErrorMsg.value = err.message;
  } else {
    packages.value = data.season_package_ids;
  }
  loadingMsg.value = '';
  loading.value = false;
  emit('getResult', packages.value);
};

const searchPackageSubtitles = async (packageId) => {
  loading.value = true;
  loadingMsg.value = '正在获取字幕列表...';
  const [data, err] = await subtitleApi.value.searchTvSeasonPackageId({
    imdb_id: imdbId.value,
    season_package_id: packageId,
  });
  if (err !== null) {
    SystemMessage.error(err.message);
  } else {
    csfSearchResult.value = data.subtitles;
  }
  loadingMsg.value = '';
  loading.value = false;
};

watch(
  () => selectedPackage.value,
  (v) => {
    if (v) {
      searchPackageSubtitles(v);
    }
  }
);

const fetchSubtitleBlob = async (item) => {
  selectedItem.value = item;
  if (cacheBlob.value[item.sub_sha256]) {
    selectedSubBlob.value = cacheBlob.value[item.sub_sha256];
    return;
  }
  selectedSubBlob.value = null;
  loading.value = true;
  loadingMsg.value = '正在获取下载地址...';
  await waitRequestReady();

  loadingMsg.value = '正在下载字幕...';
  const [data, err] = await subtitleApi.value.getDownloadUrl({
    ...item,
    imdb_id: imdbId.value,
  });
  if (err !== null) {
    SystemMessage.error(err.message);
  } else {
    // fetch资源，获取blob url
    const res = await fetch(data.download_link);
    const blob = await res.blob();
    cacheBlob.value = {
      ...cacheBlob.value,
      [item.sub_sha256]: blob,
    };
    selectedSubBlob.value = blob;
  }
  loadingMsg.value = '';
  loading.value = false;
};

const getVideoFilePathOfSubtitle = (item) => props.episodes.find((e) => e.episode === item.episode)?.video_f_path;

const handleDownloadCsfSub = async (item, isConfirmLock = false, showMessage = false) => {
  await fetchSubtitleBlob(item);

  if (!selectedSubBlob.value) {
    return;
  }

  const videoFilePath = getVideoFilePathOfSubtitle(item);

  // 上传
  const formData = new FormData();
  formData.append('video_f_path', videoFilePath);
  formData.append('file', new File([selectedSubBlob.value], item.title, { type: 'text/plain' }));
  await LibraryApi.uploadSubtitle(formData);
  await getSubtitleUploadList();
  eventBus.emit('subtitle-uploaded');

  if (isConfirmLock) {
    $q.dialog({
      title: '操作确认',
      message: `已下载到库中，是否锁定该视频，无需再次自动下载字幕？`,
      cancel: true,
      persistent: true,
      focus: 'none',
    }).onOk(async () => {
      setLock([videoFilePath]);
    });
  }

  downloadedEpisodeSubtitleMap.value[item.episode] = item.sub_sha256;

  if (showMessage) {
    SystemMessage.success('已下载到库中');
  }
};

const handlePreviewClick = async (item, callback) => {
  await fetchSubtitleBlob(item);
  if (selectedSubUrl.value) {
    callback(true);
  } else {
    callback(false);
  }
};

const downloadAll = async () => {
  if (csfSearchResult.value.length === 0) {
    SystemMessage.warning('没有可下载的字幕');
    return;
  }
  isDownloadingAll.value = true;
  try {
    // eslint-disable-next-line no-restricted-syntax
    for (const item of csfSearchResult.value) {
      if (props.episodes.some((e) => e.episode === item.episode)) {
        // eslint-disable-next-line no-await-in-loop
        await handleDownloadCsfSub(item, false, false);
      }
    }
    $q.dialog({
      title: '操作确认',
      message: `已下载到库中，是否锁定本季视频，无需再次自动下载字幕？`,
      cancel: true,
      persistent: true,
      focus: 'none',
    }).onOk(async () => {
      setLock(props.episodes.map((e) => e.video_f_path));
    });
  } catch (e) {
    SystemMessage.error(e);
  }
  isDownloadingAll.value = false;
};
onMounted(() => {
  searchPackages();
});
</script>
