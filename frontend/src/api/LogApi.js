import BaseApi from './BaseApi';

class LogApi extends BaseApi {
  getList = () => this.http('/running-log', { the_last_few_times: 3 });
}
export default new LogApi();
