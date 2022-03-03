import { reactive, ref, watch } from 'vue';
import SettingApi from 'src/api/SettingApi';
import { SystemMessage } from 'src/utils/Message';
import { deepCopy } from 'src/utils/CommonUtils';

export const settingsState = reactive({
  data: null,
});

export const formModel = reactive({});

// 更新emby目录映射
const updateFolderMap = () => {
  const getFolderMap = (folders, maps) =>
    folders.reduce((r, a) => {
      if (Object.keys(maps).includes(a)) {
        r[a] = maps[a];
      } else {
        r[a] = '';
      }
      return r;
    }, {});

  const commonSettings = settingsState.data?.common_settings;
  const embySettings = settingsState.data.emby_settings;

  embySettings.movie_paths_mapping = getFolderMap(commonSettings.movie_paths, embySettings.movie_paths_mapping || {});
  embySettings.series_paths_mapping = getFolderMap(
    commonSettings.series_paths,
    embySettings.series_paths_mapping || {}
  );
};

watch(
  () => settingsState.data,
  () => {
    updateFolderMap();
    Object.assign(formModel, deepCopy(settingsState.data));
  }
);

export const resetForm = () => {
  Object.assign(formModel, deepCopy(settingsState.data));
};

const getSettings = async () => {
  const [res, err] = await SettingApi.get();
  if (err !== null) {
    SystemMessage.error(err.message);
    return;
  }
  settingsState.data = res;
};

export const useSettings = () => {
  getSettings();
};

export const submitting = ref(false);

export const submitAll = async () => {
  submitting.value = true;
  const [, err] = await SettingApi.update(formModel);
  submitting.value = false;
  if (err !== null) {
    SystemMessage.error(err.message);
    return;
  }
  settingsState.data = { ...settingsState.data, ...deepCopy(formModel) };
  SystemMessage.success('保存成功');
};

/**
 * 获取导出的settings
 * @param includeSensitive
 * @returns {any}
 */
export const getExportSettings = (includeSensitive = false) => {
  const data = deepCopy(settingsState.data);
  if (!includeSensitive) {
    delete data.user_info;
    delete data.common_settings.threads;
    delete data.emby_settings.api_key;
    delete data.emby_settings.address_url;
  }
  return data;
};
