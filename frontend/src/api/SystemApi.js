import BaseApi from './BaseApi';

class SystemApi extends BaseApi {
  getInfo = (params) => this.http('/system-status', params);

  getStatus = (params) => this.http('/system/status', params);
}
export default new SystemApi();
