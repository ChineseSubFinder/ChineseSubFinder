import BaseApi from './BaseApi';

class AccessApi extends BaseApi {
  login = (params) => this.http('/login', params, 'POST');

  logout = (params) => this.http('/logout', params, 'POST');
}
export default new AccessApi();
