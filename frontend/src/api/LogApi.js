import BaseApi from './BaseApi';

class LogApi extends BaseApi {
  getList = (params) => this.http('/running-log', params);
}
export default new LogApi();
