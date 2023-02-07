<template>
  <q-card flat square>
    <div class="area-cover q-mb-sm relative-position">
      <div v-if="!posterInfo?.url" style="width: 160px; height: 200px"></div>
      <q-img
        v-else
        :src="getUrl(posterInfo.url)"
        class="content-width bg-grey-2"
        no-spinner
        style="width: 160px; height: 200px"
        fit="cover"
      />
    </div>
    <div class="content-width text-ellipsis-line-2" :title="data.name">{{ data.name }}</div>
    <div class="row items-center">
      <q-space />
      <div>
        <dialog-t-v-detail :data="detailInfo">
          <q-btn
            v-if="hasSubtitleVideoCount > 0"
            color="black"
            flat
            dense
            icon="closed_caption"
            :label="`${hasSubtitleVideoCount}/${detailInfo.one_video_info.length}`"
            title="已有字幕"
          />
          <q-btn v-else color="grey" round flat dense icon="closed_caption" title="没有字幕" />
        </dialog-t-v-detail>
      </div>
    </div>
  </q-card>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import DialogTVDetail from 'pages/library/tvs/DialogTVDetail';
import LibraryApi from 'src/api/LibraryApi';
import { getUrl, subtitleUploadList } from 'pages/library/use-library';

const props = defineProps({
  data: Object,
});

const posterInfo = ref(null);
const detailInfo = ref(null);

const getPosterInfo = async () => {
  const [res] = await LibraryApi.getTvPoster({
    name: props.data.name,
    main_root_dir_f_path: props.data.main_root_dir_f_path,
    root_dir_path: props.data.root_dir_path,
  });
  posterInfo.value = res;
};

const getDetailInfo = async () => {
  const [res] = await LibraryApi.getTvDetail({
    name: props.data.name,
    main_root_dir_f_path: props.data.main_root_dir_f_path,
    root_dir_path: props.data.root_dir_path,
  });
  detailInfo.value = res;
};

const hasSubtitleVideoCount = computed(
  () => detailInfo.value?.one_video_info.filter((e) => e.sub_f_path_list.length > 0).length
);

watch(subtitleUploadList, (val, oldValue) => {
  // 上传字幕列表当前文件有变化时刷新
  if (
    detailInfo.value?.one_video_info.some((e) => oldValue.map((f) => f.video_f_path).includes(e.video_f_path)) &&
    !detailInfo.value?.one_video_info.some((e) => val.map((f) => f.video_f_path).includes(e.video_f_path))
  ) {
    getDetailInfo();
  }
});

onMounted(() => {
  getPosterInfo();
  getDetailInfo();
});
</script>

<style lang="scss" scoped>
.content-width {
  width: 160px;
}
.text-ellipsis-line-2 {
  height: 40px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.btn-download {
  //display: none;
  opacity: 0;
  transition: all 0.6s ease;
}

.area-cover:hover {
  .btn-download {
    //display: block;
    opacity: 1;
  }
}
</style>
