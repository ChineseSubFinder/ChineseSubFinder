import BaseApi from './BaseApi';

class ProfileApi extends BaseApi {
  changePwd = (params) => this.http('/change-pwd', params, 'POST');
}
export default new ProfileApi();
