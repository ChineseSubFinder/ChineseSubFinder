<template>
  <q-btn color="primary" icon="search" size="sm" flat dense v-bind="$attrs" @click="visible = true" title="字幕搜索" />

  <q-dialog v-model="visible" transition-show="slide-up" transition-hide="slide-down" persistent>
    <q-card style="min-width: 70vw">
      <q-card-section>
        <div class="row justify-between items-center">
          <div class="text-h6 text-grey-8">字幕搜索</div>
          <q-btn icon="close" flat round dense @click="visible = false" />
        </div>
        <div class="text-warning">* 下载字幕包是在浏览器端进行处理的，下载过程中请不要关闭页面</div>
      </q-card-section>
      <q-separator />

      <template v-if="!searchPackage">
        <q-tabs
          v-model="tab"
          dense
          active-color="primary"
          indicator-color="primary"
          align="justify"
          narrow-indicator
          style="display: inline-block"
        >
          <q-tab name="csf" label="Subtitle.Best API" />
          <q-tab name="manual" label="手动搜索" />
        </q-tabs>

        <q-tab-panels v-model="tab" animated keep-alive>
          <q-tab-panel name="csf">
            <search-panel-csf-api :path="path" :is-movie="isMovie" :season="season" :episode="episode" />
          </q-tab-panel>

          <q-tab-panel name="manual">
            <search-panel-manual :is-movie="isMovie" :path="path" />
          </q-tab-panel>
        </q-tab-panels>
      </template>
      <template v-else>
        <q-tabs
          v-model="tab"
          dense
          active-color="primary"
          indicator-color="primary"
          align="justify"
          narrow-indicator
          style="display: inline-block"
        >
          <q-tab name="csf" label="Subtitle.Best API" />
        </q-tabs>

        <q-tab-panels v-model="tab" animated keep-alive>
          <q-tab-panel name="csf">
            <search-panel-csf-api-tv-package :episodes="packageEpisodes" />
          </q-tab-panel>

          <q-tab-panel name="csf-share">
            <search-panel-csf-api-tv-package :episodes="packageEpisodes" use-user-share-api />
          </q-tab-panel>
        </q-tab-panels>
      </template>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { ref } from 'vue';
import SearchPanelManual from 'pages/library/SearchPanelManual.vue';
import SearchPanelCsfApi from 'pages/library/SearchPanelCsfApi.vue';
import SearchPanelCsfApiTvPackage from 'pages/library/SearchPanelCsfApiTvPackage.vue';

defineProps({
  path: String,
  isMovie: {
    type: Boolean,
    default: false,
  },
  searchPackage: {
    type: Boolean,
    default: false,
  },
  season: {
    type: Number,
  },
  episode: {
    type: Number,
  },
  packageEpisodes: {
    type: Array,
  },
});

const visible = ref(false);
const tab = ref('csf');
</script>
