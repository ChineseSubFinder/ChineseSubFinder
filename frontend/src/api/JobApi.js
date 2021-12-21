import BaseApi from './BaseApi';

class JobApi extends BaseApi {
  getStatus = () => this.http('/v1/jobs/status');

  start = (data) => this.http('/v1/jobs/start', data, 'POST');

  stop = (data) => this.http('/v1/jobs/stop', data, 'POST');
}
export default new JobApi();
