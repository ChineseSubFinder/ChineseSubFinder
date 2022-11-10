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
    CommonApi.checkEmbyPath({
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

export const validateCronTime = (val) =>
  // eslint-disable-next-line max-len
  /(@(annually|yearly|monthly|weekly|daily|hourly|reboot))|(@every (\d+(ns|us|µs|ms|s|m|h))+)|((((\d+,)+\d+|(\d+(\/|-)\d+)|\d+|\*) ?){5,7})/.test(
    val
  ) || '格式不正确';
