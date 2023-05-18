import RouterPlaceholder from 'components/RouterPlaceholder';

const routes = [
  {
    path: '/',
    component: () => import('layouts/MainLayout.vue'),
    redirect: { name: 'overview' },
    children: [
      {
        name: 'overview',
        path: 'overview',
        component: () => import('pages/overview/index.vue'),
        meta: { title: '总览', icon: 'home' },
      },
      {
        name: 'library',
        path: 'library',
        component: RouterPlaceholder,
        meta: { title: '库', icon: 'video_library' },
        children: [
          {
            name: 'library.movie.list',
            path: 'library/movies',
            component: () => import('pages/library/movies/index.vue'),
            meta: { title: '电影', icon: 'movie' },
          },
          {
            name: 'library.tv.list',
            path: 'library/tvs',
            component: () => import('pages/library/tvs/index.vue'),
            meta: { title: '连续剧', icon: 'live_tv' },
          },
        ],
      },
      {
        name: 'jobs',
        path: 'jobs',
        component: () => import('pages/jobs/index.vue'),
        meta: { title: '下载队列', icon: 'assignment' },
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

  {
    path: '/test',
    component: () => import('components/ShareSubtitle/ShareSubtitlePanel.test.vue'),
  },

  // Always leave this as last one,
  // but you can also remove it
  {
    path: '/:catchAll(.*)*',
    component: () => import('pages/Error404.vue'),
  },
];

export default routes;
