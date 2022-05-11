import BaseApi from './BaseApi';

class LibraryApi extends BaseApi {
  getRefreshStatus = () => this.http('/v1/video/list/refresh-status');

  refreshLibrary = () => this.http('/v1/video/list/refresh', {}, 'POST');

  getList = () => this.http('/v1/video/list');

  downloadSubtitle = (videoId) => this.http(`/v1/video/subtitle/download`, { id: videoId }, 'POST');
}
export default new LibraryApi();
