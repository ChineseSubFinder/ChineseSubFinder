<template>
  <q-btn color="primary" icon="search" size="sm" flat dense v-bind="$attrs" @click="visible = true" title="字幕搜索" />

  <q-dialog v-model="visible" transition-show="slide-up" transition-hide="slide-down">
    <q-card style="min-width: 70vw">
      <q-card-section>
        <div class="text-h6 text-grey-8">字幕搜索</div>
      </q-card-section>

      <q-separator />

      <q-tabs
        v-model="tab"
        dense
        active-color="primary"
        indicator-color="primary"
        align="justify"
        narrow-indicator
        style="max-width: 300px"
      >
        <q-tab name="csf" label="Subtitle.Best API" />
        <q-tab name="manual" label="手动搜索" />
      </q-tabs>

      <q-tab-panels v-model="tab" animated keep-alive>
        <q-tab-panel name="csf">
          <search-panel-csf-api
            :path="path"
            :imdb-id="imdbId"
            :is-movie="isMovie"
            :season="season"
            :episode="episode"
            :search-package="searchPackage"
          />
        </q-tab-panel>

        <q-tab-panel name="manual">
          <search-panel-manual :is-movie="isMovie" :path="path" />
        </q-tab-panel>
      </q-tab-panels>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { ref } from 'vue';
import SearchPanelManual from 'pages/library/SearchPanelManual.vue';
import SearchPanelCsfApi from 'pages/library/SearchPanelCsfApi.vue';

defineProps({
  path: String,
  imdbId: String,
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
});

const visible = ref(false);
const tab = ref('csf');
</script>
