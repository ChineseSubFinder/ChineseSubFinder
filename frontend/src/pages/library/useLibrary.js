import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/Message';
import { until } from '@vueuse/core';
import config from 'src/config';
import { LocalStorage } from 'quasar';

export const getUrl = (basePath) => config.BACKEND_URL + basePath.split(/\/|\\/).join('/');

// 封面规则
export const coverRule = ref(LocalStorage.getItem('coverRule') ?? 'poster.jpg');

export const useLibrary = () => {
  const originMovies = ref([]);
  const originTvs = ref([]);
  const refreshCacheLoading = ref(false);
  const libraryRefreshStatus = ref(null);

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

  let getRefreshStatusTimer = null;

  const getLibraryRefreshStatus = async () => {
    const [res] = await LibraryApi.getRefreshStatus();
    libraryRefreshStatus.value = res.status;
  };

  const getLibraryList = async () => {
    const [res, err] = await LibraryApi.getList();
    if (err !== null) {
      SystemMessage.error(err.message);
    } else {
      originMovies.value = res.movie_infos_v2;
      originTvs.value = res.season_infos_v2;
    }
  };

  const refreshLibrary = async () => {
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

  onMounted(() => {
    getLibraryList();
    getLibraryRefreshStatus();
  });

  onBeforeUnmount(() => {
    clearInterval(getRefreshStatusTimer);
  });

  return {
    movies,
    tvs,
    refreshLibrary,
    refreshCacheLoading,
  };
};
