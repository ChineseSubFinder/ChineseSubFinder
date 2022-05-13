import { createRequest } from 'src/utils/http';
import config from 'src/config';

class BaseApi {
  // 如果没设置baseUrl，则默认使用当前相对路径
  BaseUrl = config.BACKEND_URL;

  http(url, ...option) {
    return createRequest(`${this.BaseUrl}${url}`, ...option);
  }
}

export default BaseApi;
