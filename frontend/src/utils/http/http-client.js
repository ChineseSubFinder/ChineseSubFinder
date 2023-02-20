/**
 * 数据请求公共方法
 */

import axios from 'axios';

class HttpClient {
  axiosInstance = null;

  constructor() {
    this.axiosInstance = axios.create({
      timeout: 60000, // 超时时间
      withCredentials: false, // 跨域传递cookie值
    });
  }

  /**
   * 注册拦截器
   */
  registerInterceptor(interceptor) {
    const { onRequestFullFilled, onRequestRejected, onResponseFullFilled, onResponseRejected } = interceptor;
    if (onRequestFullFilled || onRequestRejected) {
      this.axiosInstance.interceptors.request.use(onRequestFullFilled, onRequestRejected);
    }
    if (onResponseFullFilled || onResponseRejected) {
      this.axiosInstance.interceptors.response.use(onResponseFullFilled, onResponseRejected);
    }
  }

  /**
   * 根据目录自动注册拦截器
   * @param contextModules
   */
  registerInterceptorsFromDirectory(contextModules) {
    const handlers = contextModules.keys().reduce((cur, key) => {
      cur.push(contextModules(key).default);
      return cur;
    }, []);
    handlers.forEach((hanlder) => {
      this.registerInterceptor(hanlder);
    });
  }

  // 通用请求方法
  createRequest(url = '', data = {}, type = 'GET', config = {}) {
    config.headers = config.headers || {};
    const axiosConfig = Object.assign(config, {
      method: type.toUpperCase(),
      url,
      headers: { ...config.headers },
    });
    if (['DELETE', 'GET'].includes(type.toUpperCase())) {
      config.params = data;
    } else {
      config.data = data;
    }
    return (
      this.axiosInstance
        .request(axiosConfig)
        // 分别处理直接返回的数据源和{result: 1, message: '', data: {}|[]}形式的数据源
        .then((response) => [response.data, null, response])
        .catch((error) => [null, error])
    );
  }
}

export default HttpClient;
