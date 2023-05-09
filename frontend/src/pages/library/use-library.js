import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/message';
import { until } from '@vueuse/core';
import config from 'src/config';
import { LocalStorage } from 'quasar';
import { useSettings } from 'pages/settings/use-settings';

export const getUrl = (basePath) => config.BACKEND_URL + basePath.split(/\/|\\/).join('/');

// 封面规则
export const coverRule = ref(LocalStorage.getItem('coverRule') ?? 'poster.jpg');

export const originMovies = ref([]);
export const originTvs = ref([]);
const movies = computed(() =>
  originMovies.value.map((movie) => ({
    ...movie,
  }))
);

const tvs = computed(() =>
  originTvs.value.map((tv) => ({
    ...tv,
  }))
);
export const libraryRefreshStatus = ref(null);
export const subtitleUploadList = ref([]);

export const refreshCacheLoading = computed(() => libraryRefreshStatus.value === 'running');

let getRefreshStatusTimer = null;

export const getLibraryRefreshStatus = async () => {
  const [res] = await LibraryApi.getRefreshStatus();
  libraryRefreshStatus.value = res.status;
};

export const getLibraryList = async () => {
  const [res, err] = await LibraryApi.getList();
  if (err !== null) {
    SystemMessage.error(err.message);
  } else {
    originMovies.value = res.movie_infos_v2;
    originTvs.value = res.season_infos_v2;
  }
};

export const checkLibraryRefreshStatus = async () => {
  libraryRefreshStatus.value = null;
  await getLibraryRefreshStatus();
  getRefreshStatusTimer = setInterval(() => {
    getLibraryRefreshStatus();
  }, 1000);
  await until(libraryRefreshStatus).toBe('stopped');
  clearInterval(getRefreshStatusTimer);
  getRefreshStatusTimer = null;
  await getLibraryList();
};

export const refreshLibrary = async () => {
  const [, err] = await LibraryApi.refreshLibrary();
  if (err !== null) {
    SystemMessage.error(err.message);
  } else {
    await checkLibraryRefreshStatus();
    SystemMessage.success('更新缓存成功');
  }
};

export const getSubtitleUploadList = async () => {
  const [res] = await LibraryApi.getSubTitleQueueList();
  subtitleUploadList.value = res.jobs;
};

export const useLibrary = () => {
  useSettings();

  const getSubtitleUploadListTimer = setInterval(() => {
    getSubtitleUploadList();
  }, 5000);

  onMounted(() => {
    getLibraryList();
    getLibraryRefreshStatus();
    getSubtitleUploadList();
    checkLibraryRefreshStatus();
  });

  onBeforeUnmount(() => {
    clearInterval(getRefreshStatusTimer);
    clearInterval(getSubtitleUploadListTimer);
  });

  return {
    movies,
    tvs,
    refreshLibrary,
    refreshCacheLoading,
  };
};

export const doFixSubtitleTimeline = async (path) => {
  const formData = new FormData();
  formData.append('video_f_path', path);
  const subtitleUrl = getUrl(path);
  // 先下载字幕到内存，生成file文件
  const res = await fetch(subtitleUrl);
  if (!res.ok) {
    SystemMessage.error('获取字幕文件失败');
    return;
  }
  const blob = await res.blob();
  const file = new File([blob], path.split(/\/|\\/).pop());
  formData.append('file', file);
  await LibraryApi.uploadSubtitle(formData);
  SystemMessage.success('已提交时间轴校准', {
    timeout: 3000,
  });
  await getSubtitleUploadList();
};

/**
 * 检查一个视频是否锁定
 * @param videoInfo {video_type, physical_video_file_full_path, is_bluray, is_skip}
 * @returns {Promise<boolean>}
 */
export const checkIsVideoLocked = async (videoInfo) => {
  const [res] = await LibraryApi.getSkipInfo({
    video_skip_infos: [videoInfo],
  });
  return !!res.is_skips?.[0];
};
