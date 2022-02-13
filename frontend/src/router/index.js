import { route } from 'quasar/wrappers';
import { createRouter, createMemoryHistory, createWebHistory, createWebHashHistory } from 'vue-router';
import { getInfo, systemState } from 'src/store/systemState';
import routes from './routes';

/*
 * If not building with SSR mode, you can
 * directly export the Router instantiation;
 *
 * The function below can be async too; either use
 * async/await or return a Promise which resolves
 * with the Router instance.
 */
// eslint-disable-next-line no-nested-ternary
const createHistory = process.env.SERVER
  ? createMemoryHistory
  : process.env.VUE_ROUTER_MODE === 'history'
  ? createWebHistory
  : createWebHashHistory;

const Router = createRouter({
  scrollBehavior: () => ({ left: 0, top: 0 }),
  routes,

  // Leave this as is and make changes in quasar.conf.js instead!
  // quasar.conf.js -> build -> vueRouterMode
  // quasar.conf.js -> build -> publicPath
  history: createHistory(process.env.MODE === 'ssr' ? void 0 : process.env.VUE_ROUTER_BASE),
});

Router.beforeEach(async (to, from, next) => {
  // 获取是否初始化的信息
  if (!systemState.systemInfo) {
    await getInfo();
  }
  // 未初始化时跳转到初始化页面
  if (systemState.systemInfo?.is_setup === false && to.path !== '/setup') {
    next('/setup');
  }
  // 初始化后禁用初始化页面
  if (systemState.systemInfo?.is_setup === true && to.path === '/setup') {
    next('/');
  }
  next();
});

export default route((/* { store, ssrContext } */) => Router);

export { Router as router };
