import { createCsfSubtitlesRequest } from 'src/utils/http';
import config from 'src/config';

export class CsfSubtitlesApi {
  // 如果没设置baseUrl，则默认使用当前相对路径
  BaseUrl = config.CSF_SUBTITLES_API_URL;

  prefix = '/share-sub/v1';

  http(url, ...option) {
    return createCsfSubtitlesRequest(`${this.BaseUrl}${url}`, ...option);
  }

  searchMovie = (data) => this.http(`${this.prefix}/search-movie`, data, 'POST');

  searchTvEps = (data) => this.http(`${this.prefix}/search-tv-eps`, data, 'POST');

  searchTvSeasonPackage = (data) => this.http(`${this.prefix}/search-tv-season-package`, data, 'POST');

  searchTvSeasonPackageId = (data) => this.http(`${this.prefix}/search-tv-season-package-id`, data, 'POST');

  getDownloadUrl = (data) => this.http(`${this.prefix}/get-dl-url`, data, 'POST');
}

export default new CsfSubtitlesApi();
