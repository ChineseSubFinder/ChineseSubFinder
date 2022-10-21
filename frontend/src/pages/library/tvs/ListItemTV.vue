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
      <div class="text-grey">1970-01-01</div>
      <q-space />
      <div>
        <!--        <dialog-t-v-detail :data="data">-->
        <!--          <q-btn-->
        <!--            v-if="hasSubtitleVideoCount > 0"-->
        <!--            color="black"-->
        <!--            flat-->
        <!--            dense-->
        <!--            icon="closed_caption"-->
        <!--            :label="`${hasSubtitleVideoCount}/${data.one_video_info.length}`"-->
        <!--            title="已有字幕"-->
        <!--          />-->
        <!--          <q-btn v-else color="grey" round flat dense icon="closed_caption" title="没有字幕" />-->
        <!--        </dialog-t-v-detail>-->
      </div>
    </div>
  </q-card>
</template>

<script setup>
import { onMounted, ref } from 'vue';
// import DialogTVDetail from 'pages/library/tvs/DialogTVDetail';
import LibraryApi from 'src/api/LibraryApi';
import { getUrl } from 'pages/library/useLibrary';
import { VIDEO_TYPE_TV } from 'src/constants/SettingConstants';

const props = defineProps({
  data: Object,
});

const posterInfo = ref(null);
const isSkipped = ref(null);

const getPosterInfo = async () => {
  const [res] = await LibraryApi.getTvPoster({
    name: props.data.name,
    main_root_dir_f_path: props.data.main_root_dir_f_path,
    root_dir_path: props.data.root_dir_path,
  });
  posterInfo.value = res;
};

const getIsSkipped = async () => {
  const [res] = await LibraryApi.getSkipInfo({
    video_type: VIDEO_TYPE_TV,
    physical_video_file_full_path: props.data.video_f_path,
    is_bluray: false,
    is_skip: true,
  });
  isSkipped.value = res.is_skip;
};

// const hasSubtitleVideoCount = computed(
//   () => props.data.one_video_info.filter((e) => e.sub_f_path_list.length > 0).length
// );

onMounted(() => {
  getPosterInfo();
  getIsSkipped();
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
