import { encode } from 'js-base64';
import BaseApi from './BaseApi';

class LibraryApi extends BaseApi {
  getRefreshStatus = () => this.http('/v1/video/list/refresh-status');

  refreshLibrary = () => this.http('/v1/video/list/refresh_main_list', {}, 'POST');

  getList = () => this.http('/v1/video/list/video_main_list');

  getImdbId = (data) => this.http('/v1/preview/video_f_path_2_imdb_info', data, 'POST');

  downloadSubtitle = (data) => this.http(`/v1/video/list/add`, data, 'POST');

  getMoviePoster = (data) => this.http(`/v1/video/list/movie_poster`, data, 'POST');

  getTvPoster = (data) => this.http(`/v1/video/list/series_poster`, data, 'POST');

  getSkipInfo = (data) => this.http(`/v1/video/list/scan_skip_info`, data, 'POST');

  setSkipInfo = (data) => this.http(`/v1/video/list/scan_skip_info`, data, 'PUT');

  getMovieDetail = (data) => this.http(`/v1/video/list/one_movie_subs`, data, 'POST');

  getTvDetail = (data) => this.http(`/v1/video/list/one_series_subs`, data, 'POST');

  getSubtitleQueueStatus = (data) => this.http(`/v1/subtitles/is_manual_upload_2_local_in_queue`, data, 'POST');

  uploadSubtitle = (data) =>
    this.http(`/v1/subtitles/manual_upload_2_local`, data, 'POST', {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });

  refreshMediaServerSubList = () => this.http(`/v1/subtitles/refresh_media_server_sub_list`, {}, 'POST');

  getSubTitleQueueList = () => this.http(`/v1/subtitles/list_manual_upload_2_local_job`);

  // addPreviewJob = (data) => this.http(`/v1/preview/add`, data, 'POST');
  //
  // getPreviewJobs = () => this.http(`/v1/preview/list`);
  //
  // checkIsPreviewInQueue = (data) => this.http(`/v1/preview/is_in_queue`, data, 'POST');
  //
  // getPreviewJobResult = (data) => this.http(`/v1/preview/job_result`, data, 'POST');
  //
  // getPreviewDistInfo = (data) => this.http(`/v1/preview/export_info`, data, 'POST');

  getSearchSubtitleInfo = (data) => this.http(`/v1/preview/search_other_web`, data, 'POST');

  cleanAllPreviewJobData = () => this.http(`/v1/preview/clean_up`, {}, 'POST');

  getVideoM3u8 = (videoPath) => this.http(`/v1/preview/playlist/${encode(encodeURIComponent(videoPath))}`);

  getUploadInfo = (data) => this.http(`/v1/subtitles/get_generate_upload_url_info`, data, 'POST');
}
export default new LibraryApi();
