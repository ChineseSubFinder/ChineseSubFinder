<template>
  <q-card flat square>
    <div class="area-cover q-mb-sm relative-position">
      <q-img src="https://via.placeholder.com/500" class="content-width bg-grey-2" height="230px" no-spinner />
    </div>
    <div class="content-width text-ellipsis-line-2" :title="data.name">{{ data.name }}</div>
    <div class="row items-center">
      <div class="text-grey">1970-01-01</div>
      <q-space />
      <div>
        <dialog-t-v-detail :data="data">
          <q-btn
            v-if="hasSubtitleVideoCount > 0"
            color="black"
            flat
            dense
            icon="closed_caption"
            :label="`${hasSubtitleVideoCount}/${data.one_video_info.length}`"
            title="已有字幕"
          />
          <q-btn v-else color="grey" round flat dense icon="closed_caption" title="没有字幕" />
        </dialog-t-v-detail>
      </div>
    </div>
  </q-card>
</template>

<script setup>
import { computed } from 'vue';
import DialogTVDetail from 'pages/library/tvs/DialogTVDetail';

const props = defineProps({
  data: Object,
});

const hasSubtitleVideoCount = computed(
  () => props.data.one_video_info.filter((e) => e.sub_f_path_list.length > 0).length
);
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
