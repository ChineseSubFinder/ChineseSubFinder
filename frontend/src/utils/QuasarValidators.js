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

export const validateCronDuration = (val) => /^(-?\d+(ns|us|µs|ms|s|m|h))+$/.test(val) || '格式不正确';
