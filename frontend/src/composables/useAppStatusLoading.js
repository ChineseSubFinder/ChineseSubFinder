// 程序状态的hook接口
import { Loading } from 'quasar';
import JobApi from 'src/api/JobApi';

export const useAppStatusLoading = () => {
  let timer = null;

  const startLoading = () => {
    Loading.show({
      message: '正在应用程序配置',
      html: true,
    });

    const handler = async () => {
      const [res, err] = await JobApi.getStatus();
      if (res || err?.error?.status <= 401) {
        clearInterval(timer);
        Loading.hide();
      }
    };

    timer = setInterval(async () => {
      handler();
    }, 1000);
    handler();
  };

  return {
    startLoading,
  };
};
