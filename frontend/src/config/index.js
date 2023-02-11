export default {
  get BACKEND_URL() {
    // 如果.env文件设置了 BACKEND_URL，则使用设置的值，否则使用相对路径
    return process.env.BACKEND_URL || new URL(window.location.href).pathname.replace(/\/$/, '');
  },
  CSF_SUBTITLES_API_URL: process.env.CSF_SUBTITLES_API_URL,
  CSF_SUBTITLES_API_KEY: process.env.CSF_SUBTITLES_API_KEY,
};
