import config from 'src/config';

export default {
  onRequestFullFilled: (req) => {
    req.headers.Authorization = `Bearer ${config.CSF_SUBTITLES_API_KEY}`;
    return req;
  },
};
