<template>
  <q-layout view="lHh Lpr lFf">
    <q-header elevated>
      <q-toolbar class="text-white text-primary">
        <q-btn flat dense round color="white" icon="menu" aria-label="Menu" @click="leftDrawerOpen = !leftDrawerOpen" />
        <div class="text-h5 q-ml-sm">{{$route.meta.title}}</div>
        <q-space />
        <q-btn-dropdown :label="userState.username" icon="account_circle" flat>
          <q-list dense style="min-width: 100px">
            <q-item clickable v-close-popup>
              <q-item-section @click="logout">退出登录</q-item-section>
            </q-item>
          </q-list>
        </q-btn-dropdown>
      </q-toolbar>
    </q-header>

    <q-drawer
      v-model="leftDrawerOpen"
      class="q-pa-md"
      :breakpoint="720"
      :width="280"
      show-if-above
      bordered
      dark
      style="background: #111729;"
      content-class="bg-white"
    >
      <div class="text-h5 q-py-sm q-px-md" style="height: 65px;">
        <div>ChineseSubFinder</div>
        <div class="text-body2">
          {{ systemState.systemInfo?.version }}
          <version-update-badge/>
        </div>
      </div>
      <q-list>
        <menu-item v-for="route in menus" :menu-info="route" :key="`${route.name}${route.path}`" />
      </q-list>
    </q-drawer>

    <q-page-container>
      <router-view />
    </q-page-container>
  </q-layout>
</template>

<script setup>
import routes from 'src/router/routes';
import {ref } from 'vue';
import {useRouter} from 'vue-router';
import MenuItem from 'layouts/MenuItem';
import {systemState} from 'src/store/systemState';
import {userState} from 'src/store/userState';
import {LocalStorage} from 'quasar';
import AccessApi from 'src/api/AccessApi';
import VersionUpdateBadge from 'components/VersionUpdateBadge';

const router = useRouter();

const leftDrawerOpen = ref(false)
const menus = routes.find((e) => e.path === '/').children;

const logout = () => {
  userState.username = '';
  userState.accessToken = undefined;
  LocalStorage.remove('token');
  AccessApi.logout();
  router.push('/access/login')
}
</script>
