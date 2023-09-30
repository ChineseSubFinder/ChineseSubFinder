// 程序状态的hook接口
import { Dialog } from 'quasar';
import SystemApi from 'src/api/SystemApi';
import useInterval from 'src/composables/use-interval';
import { computed, onBeforeUnmount, watch } from 'vue';
import LoadingDialogAppPrepareJobInit from 'components/LoadingDialogAppPrepareError/LoadingDialogAppPrepareJobInit.vue';
import { systemState } from 'src/store/systemState';

let isPreJobLoadingDialogOpened = false;
const openPreJobLoadingDialog = () => {
  if (isPreJobLoadingDialogOpened) {
    return;
  }
  isPreJobLoadingDialogOpened = true;
  Dialog.create({
    component: LoadingDialogAppPrepareJobInit,
  }).onDismiss(() => {
    isPreJobLoadingDialogOpened = false;
  });
};

export const useAppStatusLoading = () => {
  const prepareStatus = computed(() => systemState.preJobStatus);

  const updateDialog = () => {
    if (prepareStatus.value?.is_done !== true) {
      openPreJobLoadingDialog();
    }
  };

  const getPrepareStatus = async () => {
    const [res] = await SystemApi.getPrepareStatus();
    systemState.preJobStatus = res;
  };

  const { resetInterval, stopInterval } = useInterval(
    async () => {
      await getPrepareStatus();
      updateDialog();
    },
    1000,
    false
  );

  const startLoading = async () => {
    await getPrepareStatus();
    updateDialog();
    resetInterval();
  };

  watch(
    () => prepareStatus.value?.is_done,
    (val) => {
      if (val === true) {
        stopInterval();
      }
    }
  );

  onBeforeUnmount(() => {
    stopInterval();
  });

  return {
    startLoading,
  };
};
