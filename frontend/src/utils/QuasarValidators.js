import CommonApi from 'src/api/CommonApi';

export const validateRemotePath = (val) =>
  new Promise((resolve) => {
    CommonApi.checkPath({ path: val }).then(([res, err]) => {
      if (!res?.valid) {
        resolve(err?.message || '目录不可用，请确保目录存在并且拥有权限');
      } else {
        resolve(true);
      }
    });
  });

export const validateEmbyPath = (val, extendData) =>
  new Promise((resolve) => {
    CommonApi.checkPath({
      address_url: extendData.address_url,
      api_key: extendData.api_key,
      path_type: extendData.path_type, // movie / series
      cfs_media_path: extendData.cfs_media_path,
      emby_media_path: val,
    }).then(([res, err]) => {
      if (!res?.media_list?.length) {
        resolve(err?.message || '目录不可用，请输入正确的Emby目录');
      } else {
        resolve(true);
      }
    });
  });

export const validateCronDuration = (val) => /^(-?\d+(ns|us|µs|ms|s|m|h))+$/.test(val) || '格式不正确';
