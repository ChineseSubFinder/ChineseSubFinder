import { router } from 'src/router';

const handleError = (error) => {
  // eslint-disable-next-line
  console.error('interceptor catch the error!\n', error);
  let errorMessageText = error.data?.message || error.message || '网络错误';
  // 权限不足时的处理
  if (error.status === 401) {
    errorMessageText = error.data.message || '权限不足，请登录重试';
    router.push('/access/login');
  }

  const rtData = {
    error,
    message: errorMessageText,
  };

  return Promise.reject(rtData);
};

export default {
  onRequestRejected: (error) => handleError(error),
  onResponseFullFilled: (response) => {
    const { data } = response;
    // 正常返回但是code是错误码的情况也需要异常处理
    if (data?.code && data?.code > 300) {
      return handleError(response);
    }
    return response;
  },
  onResponseRejected: (error) => handleError(error?.response || error),
};
