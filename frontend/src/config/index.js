export default {
  get BACKEND_URL() {
    return process.env.BACKEND_URL || new URL(window.location.href).pathname.replace(/\/$/, '');
  },
};
