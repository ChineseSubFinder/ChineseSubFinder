import { createRequest } from 'src/utils/http';

class BaseApi {
  // 如果没设置baseUrl，则默认使用当前相对路径
  BaseUrl = process.env.BACKEND_URL || new URL(window.location.href).pathname.replace(/\/$/, '');

  http(url, ...option) {
    return createRequest(`${this.BaseUrl}${url}`, ...option);
  }
}

export default BaseApi;
