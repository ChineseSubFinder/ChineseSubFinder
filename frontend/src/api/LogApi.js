import BaseApi from './BaseApi';

class LogApi extends BaseApi {
  getList = () => this.http('/running-log', {}, 'POST');
}
export default new LogApi();
