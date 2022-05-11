import { onBeforeUnmount, onMounted, ref } from 'vue';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/Message';
import { until } from '@vueuse/core';

export const useLibrary = () => {
  const movies = ref([]);
  const tvs = ref([]);
  const refreshCacheLoading = ref(false);
  const libraryRefreshStatus = ref(null);
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
      movies.value = res.movie_infos;
      tvs.value = res.season_infos;
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
      getLibraryList();
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
