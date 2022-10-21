import BaseApi from './BaseApi';

class LibraryApi extends BaseApi {
  getRefreshStatus = () => this.http('/v1/video/list/refresh-status');

  refreshLibrary = () => this.http('/v1/video/list/refresh_main_list', {}, 'POST');

  getList = () => this.http('/v1/video/list/video_main_list');

  downloadSubtitle = (data) => this.http(`/v1/video/list/add`, data, 'POST');

  getMoviePoster = (data) => this.http(`/v1/video/list/movie_poster`, data, 'POST');

  getTvPoster = (data) => this.http(`/v1/video/list/series_poster`, data, 'POST');

  getSkipInfo = (data) => this.http(`/v1/video/list/scan_skip_info`, data, 'POST');

  setSkipInfo = (data) => this.http(`/v1/video/list/scan_skip_info`, data, 'PUT');
}
export default new LibraryApi();
