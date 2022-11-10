// 程序状态的hook接口
import { Loading } from 'quasar';
import SystemApi from 'src/api/SystemApi';

export const useAppStatusLoading = () => {
  let timer = null;

  const startLoading = async () => {
    Loading.show({
      message: '正在应用程序配置',
      html: true,
    });

    const sleep = (ms) =>
      new Promise((resolve) => {
        setTimeout(resolve, ms);
      });
    // 考虑保存配置后HTTP服务没有立即重启的情况，等待几秒再请求状态接口
    await sleep(6000);

    const handler = async () => {
      const [res, err] = await SystemApi.getInfo();
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
