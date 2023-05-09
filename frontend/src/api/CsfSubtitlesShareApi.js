import { CsfSubtitlesApi } from 'src/api/CsfSubtitlesApi';

class CsfSubtitlesShareApi extends CsfSubtitlesApi {
  prefix = '/user-share-sub/v1';

  getUploadUrl = (data) => this.http(`${this.prefix}/generate-tmp-upload-url`, data, 'POST');

  setUploadSuccess = (data) => this.http(`${this.prefix}/tmp-upload-done`, data, 'POST');
}

export default new CsfSubtitlesShareApi();
