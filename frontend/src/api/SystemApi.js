import BaseApi from './BaseApi';

class SystemApi extends BaseApi {
  getInfo = (params) => this.http('/system-status', params);

  getPrepareStatus = () => this.http('/pre-job', {}, 'POST');
}
export default new SystemApi();
