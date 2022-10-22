import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/Message';
import { until } from '@vueuse/core';
import config from 'src/config';
import { LocalStorage } from 'quasar';

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
export const refreshCacheLoading = ref(false);
export const libraryRefreshStatus = ref(null);
export const subtitleUploadList = ref([]);

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

export const refreshLibrary = async () => {
  refreshCacheLoading.value = true;
  const [, err] = await LibraryApi.refreshLibrary();
  if (err !== null) {
    SystemMessage.error(err.message);
  } else {
    libraryRefreshStatus.value = null;
    getRefreshStatusTimer = setInterval(() => {
      getLibraryRefreshStatus();
    }, 1000);
    await until(libraryRefreshStatus).toBe('stopped');
    clearInterval(getRefreshStatusTimer);
    getRefreshStatusTimer = null;
    await getLibraryList();
    SystemMessage.success('更新成功');
  }
  refreshCacheLoading.value = false;
};

export const getSubtitleUploadList = async () => {
  const [res] = await LibraryApi.getSubTitleQueueList();
  subtitleUploadList.value = res.jobs;
};

export const useLibrary = () => {
  const getSubtitleUploadListTimer = setInterval(() => {
    getSubtitleUploadList();
  }, 5000);

  onMounted(() => {
    getLibraryList();
    getLibraryRefreshStatus();
    getSubtitleUploadList();
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
