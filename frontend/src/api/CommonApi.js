import BaseApi from './BaseApi';

class CommonAPi extends BaseApi {
  setup = (params) => this.http('/setup', params, 'POST');

  getDefaultSettings = (params) => this.http('/def-settings', params);

  checkProxy = (params) => this.http('/check-proxy', params, 'POST', { timeout: 2 * 60 * 1000 });

  checkPath = (params) => this.http('/check-path', params, 'POST');

  checkEmbyPath = (data) => this.http('/check-emby-path', data, 'POST');

  checkEmbyServer = (data) => this.http('/check-emby-settings', data, 'POST');

  checkTmdbApiKey = (data) => this.http('/check-tmdb-api-settings', data, 'POST');
}
export default new CommonAPi();
