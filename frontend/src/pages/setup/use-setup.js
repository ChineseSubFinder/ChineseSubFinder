import { onMounted, reactive, watch } from 'vue';
import CommonApi from 'src/api/CommonApi';
import { SystemMessage } from 'src/utils/message';

export const setupState = reactive({
  defaultSettings: null,
  form: {
    username: '',
    password: '',
    confirmPassword: '',
    movieFolder: [''],
    seriesFolder: [''],
    mediaServer: '',
    emby: {
      url: '',
      apiKey: '',
      limitCount: 3000,
      skipWatched: true,
      autoOrManual: true,
      movieFolderMap: {},
      seriesFolderMap: {},
    },
  },
});

const getDefaultSettings = async () => {
  const [res, err] = await CommonApi.getDefaultSettings();
  if (err !== null) {
    SystemMessage.error(err.message);
    return;
  }
  setupState.defaultSettings = res;
};

const getFolderMap = (folders, maps) =>
  folders.reduce((r, a) => {
    if (Object.keys(maps).includes(a)) {
      r[a] = maps[a];
    } else {
      r[a] = '';
    }
    return r;
  }, {});

watch(
  () => setupState.form.movieFolder,
  () => {
    setupState.form.emby.movieFolderMap = getFolderMap(
      setupState.form.movieFolder,
      setupState.form.emby.movieFolderMap
    );
  },
  { deep: true }
);

watch(
  () => setupState.form.seriesFolder,
  () => {
    setupState.form.emby.seriesFolderMap = getFolderMap(
      setupState.form.seriesFolder,
      setupState.form.emby.seriesFolderMap
    );
  },
  { deep: true }
);

export const useSetup = () => {
  onMounted(() => {
    getDefaultSettings();
  });
};
