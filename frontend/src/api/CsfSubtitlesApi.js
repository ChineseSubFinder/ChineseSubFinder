import { createCsfSubtitlesRequest } from 'src/utils/http';
import config from 'src/config';

class CsfSubtitlesApi {
  // 如果没设置baseUrl，则默认使用当前相对路径
  BaseUrl = config.CSF_SUBTITLES_API_URL;

  http(url, ...option) {
    return createCsfSubtitlesRequest(`${this.BaseUrl}${url}`, ...option);
  }

  searchMovie = (data) => this.http('/v1/search-movie', data, 'POST');

  searchTvEps = (data) => this.http('/v1/search-tv-eps', data, 'POST');

  searchTvSeasonPackage = (data) => this.http('/v1/search-tv-season-package', data, 'POST');

  searchTvSeasonPackageId = (data) => this.http('/v1/search-tv-season-package-id', data, 'POST');

  getDownloadUrl = (data) => this.http('/v1/get-dl-url', data, 'POST');
}

export default new CsfSubtitlesApi();
