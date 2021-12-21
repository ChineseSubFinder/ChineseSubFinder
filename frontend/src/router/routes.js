import RouterPlaceholder from 'components/RouterPlaceholder';

const routes = [
  {
    path: '/',
    component: () => import('layouts/MainLayout.vue'),
    redirect: { name: 'settings' },
    children: [
      {
        name: 'home',
        path: '/home',
        component: () => import('pages/index.vue'),
        meta: { title: '首页', icon: 'home' },
      },
      {
        name: 'library',
        path: 'library',
        component: () => import('pages/library/index.vue'),
        meta: { title: '库', icon: 'video_library' },
      },
      {
        name: 'jobs',
        path: 'jobs',
        component: () => import('pages/jobs/index.vue'),
        meta: { title: '任务列表', icon: 'cloud_queue' },
      },
      {
        name: 'settings',
        path: 'settings',
        component: () => import('pages/settings/index.vue'),
        meta: { title: '配置中心', icon: 'settings' },
      },
    ],
  },

  {
    path: '/access',
    component: RouterPlaceholder,
    children: [
      {
        path: 'login',
        component: () => import('pages/access/login/index.vue'),
      },
    ],
  },

  {
    path: '/setup',
    component: () => import('pages/setup/index.vue'),
  },

  // Always leave this as last one,
  // but you can also remove it
  {
    path: '/:catchAll(.*)*',
    component: () => import('pages/Error404.vue'),
  },
];

export default routes;
