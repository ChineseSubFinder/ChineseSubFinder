import { createRequest } from 'src/utils/http';

class BaseApi {
  BaseUrl = process.env.BACKEND_URL;

  http(url, ...option) {
    return createRequest(`${this.BaseUrl}${url}`, ...option);
  }
}

export default BaseApi;
