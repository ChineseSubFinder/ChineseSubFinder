import BaseApi from './BaseApi';

class JobApi extends BaseApi {
  getStatus = () => this.http('/v1/daemon/status');

  start = (data) => this.http('/v1/daemon/start', data, 'POST');

  stop = (data) => this.http('/v1/daemon/stop', data, 'POST');
}
export default new JobApi();
