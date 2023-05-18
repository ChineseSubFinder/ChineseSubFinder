<template>
  <q-btn color="primary" flat dense icon="cloud_upload" v-bind="$attrs" @click="handleBtnClick" title="共享字幕" />

  <q-dialog v-model="visible" persistent>
    <q-card class="overflow-hidden column" style="width: 80vw; max-width: 80vw; height: 80vh">
      <q-card-section>
        <div class="row justify-between items-center">
          <div class="text-h6">共享字幕</div>
          <q-btn icon="close" flat round dense @click="visible = false" />
        </div>
      </q-card-section>

      <q-separator />

      <q-card-section class="q-pa-none col overflow-auto">
        <share-subtitle-panel :media-data="mediaData" :is-movie="isMovie" />
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { ref } from 'vue';
import ShareSubtitlePanel from 'components/ShareSubtitle/ShareSubtitlePanel.vue';
import { checkIsVideoLocked } from 'pages/library/use-library';
import { SystemMessage } from 'src/utils/message';
import { VIDEO_TYPE_MOVIE } from 'src/constants/SettingConstants';

const props = defineProps({
  path: String,
  isMovie: {
    type: Boolean,
    default: false,
  },
  dense: {
    type: Boolean,
    default: false,
  },
  mediaData: Object,
});

const visible = ref(false);

const handleBtnClick = async () => {
  if (props.isMovie) {
    if (!props.mediaData.sub_url_list?.length) {
      SystemMessage.warn('当前视频没有字幕，无法共享');
      return;
    }

    const isLock = await checkIsVideoLocked({
      video_type: VIDEO_TYPE_MOVIE,
      physical_video_file_full_path: props.mediaData.video_f_path,
    });

    if (!isLock) {
      SystemMessage.warn('请先锁定视频，再进行字幕共享');
      return;
    }
  }
  visible.value = true;
};
</script>
