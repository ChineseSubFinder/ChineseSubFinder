import { store as qStore } from 'quasar/wrappers';
import { createStore } from 'vuex';

// import example from './module-example'

/*
 * If not building with SSR mode, you can
 * directly export the Store instantiation;
 *
 * The function below can be async too; either use
 * async/await or return a Promise which resolves
 * with the Store instance.
 */

const store = createStore({
  modules: {},

  // enable strict mode (adds overhead!)
  // for dev mode and --debug builds only
  strict: process.env.DEBUGGING,
});

export default qStore((/* { ssrContext } */) => store);
// 暴露store供非组件js代码中使用
export { store };
