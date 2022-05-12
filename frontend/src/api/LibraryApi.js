import BaseApi from './BaseApi';

class LibraryApi extends BaseApi {
  getRefreshStatus = () => this.http('/v1/video/list/refresh-status');

  refreshLibrary = () => this.http('/v1/video/list/refresh', {}, 'POST');

  getList = () => this.http('/v1/video/list');

  downloadSubtitle = (data) => this.http(`/v1/video/list/add`, data, 'POST');
}
export default new LibraryApi();
