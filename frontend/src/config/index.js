export default {
  get BACKEND_URL() {
    return process.env.BACKEND_URL || new URL(window.location.href).pathname.replace(/\/$/, '');
  },

  // 19037为静态服务默认端口
  BACKEND_STATIC_URL:
    process.env.BACKEND_STATIC_URL || `${window.location.protocol}//${window.location.hostname}:19037`,
};
