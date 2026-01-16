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
        <search-panel-manual :is-movie="isMovie" :path="path" />
      </template>
      <template v-else>
        <div class="q-pa-md text-grey">字幕包搜索已下线，请改用手动搜索。</div>
      </template>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { ref } from 'vue';
import SearchPanelManual from 'pages/library/SearchPanelManual.vue';

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
</script>
