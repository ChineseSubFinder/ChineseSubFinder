<template>
  <q-page class="q-pa-lg">
    <div class="row q-gutter-md">
      <q-btn label="更新缓存" color="primary" icon="cached" @click="refreshLibrary" :loading="refreshCacheLoading">
        <template v-slot:loading>
          <q-spinner-hourglass class="on-left" />
          更新缓存中...
        </template>
      </q-btn>

      <q-space />

      <q-select
        v-model="filterForm.hasSubtitle"
        dense
        outlined
        :options="[
          { label: '有字幕', value: true },
          { label: '无字幕', value: false },
        ]"
        label="有无字幕"
        clearable
        emit-value
        map-options
        style="width: 200px"
      />
    </div>

    <q-separator class="q-my-md" />

    <div v-if="movies.length" class="row q-gutter-x-md q-gutter-y-lg">
      <list-item-movie v-for="item in filteredMovies" :data="item" :key="item.name" />
    </div>
    <div v-else class="q-my-md text-grey">当前没有可用视频，点击"更新缓存"按钮可重建缓存</div>
  </q-page>
</template>

<script setup>
import { useLibrary } from 'pages/library/useLibrary';
import { computed, reactive } from 'vue';
import ListItemMovie from './ListItemMovie';

const filterForm = reactive({
  hasSubtitle: null,
});

const { movies, refreshLibrary, refreshCacheLoading } = useLibrary();

const filteredMovies = computed(() => {
  if (filterForm.hasSubtitle === null) {
    return movies.value;
  }
  if (filterForm.hasSubtitle === true) {
    return movies.value.filter((item) => item.sub_f_path_list.length > 0);
  }
  return movies.value.filter((item) => item.sub_f_path_list.length === 0);
});
</script>
