import config from 'src/config';
import { settingsState } from 'src/store/settingsState';

export default {
  onRequestFullFilled: (req) => {
    // 设置api key
    const apiKey = settingsState.settings?.subtitle_sources?.subtitle_best_settings?.api_key;
    if (['get', 'delete'].includes(req.method)) {
      req.params = req.params || {};
      req.params.api_key = apiKey;
    } else {
      req.data = req.data || {};
      req.data.api_key = apiKey;
    }
    // 设置bearer头
    req.headers.Authorization = `Bearer ${config.CSF_SUBTITLES_API_KEY}`;
    return req;
  },
};
