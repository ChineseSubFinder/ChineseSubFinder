import BaseApi from './BaseApi';

class SettingApi extends BaseApi {
  get = (params) => this.http('/v1/settings', params);

  patchUpdate = (data) => this.http('/v1/settings', data, 'PATCH');

  update = (data) => this.http('/v1/settings', data, 'PUT');
}
export default new SettingApi();
