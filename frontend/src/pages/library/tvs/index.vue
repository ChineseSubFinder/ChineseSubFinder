<template>
  <q-page class="q-pa-lg">
    <div class="row q-gutter-md">
      <btn-dialog-library-refresh/>

      <q-space />

      <q-select
        v-model="filterForm.hasSubtitle"
        dense
        outlined
        :options="[
          { label: '无字幕', value: 1 },
          { label: '部分有字幕', value: 2 },
          { label: '全部有字幕', value: 3 },
        ]"
        label="有无字幕"
        clearable
        map-options
        emit-value
        style="width: 160px"
      />

      <q-input v-model="filterForm.search" outlined dense label="输入关键字搜索">
        <template #append>
          <q-icon name="search" />
        </template>
      </q-input>
    </div>

    <q-separator class="q-my-md" />

    <div v-if="tvs.length" class="row q-gutter-x-md q-gutter-y-lg">
      <q-intersection v-for="item in filteredTvs" :key="item.name" style="width: 160px; height: 280px">
        <list-item-t-v :data="item" />
      </q-intersection>
    </div>
    <div v-else class="q-my-md text-grey">当前没有可用视频，点击"更新缓存"按钮可重建缓存</div>
  </q-page>
</template>

<script setup>
import { useLibrary } from 'pages/library/useLibrary';
import { computed, reactive } from 'vue';
import ListItemTV from 'pages/library/tvs/ListItemTV';
import BtnDialogLibraryRefresh from 'pages/library/BtnDialogLibraryRefresh';

const filterForm = reactive({
  hasSubtitle: null,
  search: '',
});

const { tvs } = useLibrary();

const filteredTvs = computed(() => {
  let res = tvs.value;

  const getSubtitleCount = (item) => item.one_video_info.filter((e) => e.sub_f_path_list.length > 0).length;

  if (filterForm.hasSubtitle === 1) {
    res = res.filter((item) => getSubtitleCount(item) === 0);
  }
  if (filterForm.hasSubtitle === 2) {
    res = res.filter((item) => getSubtitleCount(item) > 0 && getSubtitleCount(item) < item.one_video_info.length);
  }
  if (filterForm.hasSubtitle === 3) {
    res = res.filter((item) => getSubtitleCount(item) === item.one_video_info.length);
  }

  if (filterForm.search !== '') {
    res = res.filter((item) => item.name.toLowerCase().includes(filterForm.search.toLowerCase()));
  }

  return res;
});
</script>
