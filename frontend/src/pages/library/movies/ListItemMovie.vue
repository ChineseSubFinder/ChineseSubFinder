<template>
  <q-card flat square>
    <div class="area-cover q-mb-sm relative-position">
      <div v-if="!posterInfo?.url" :style="{ width, height: coverHeight }"></div>
      <q-img
        v-else
        :src="getUrl(posterInfo.url)"
        class="content-width bg-grey-2"
        no-spinner
        :style="{ width, height: coverHeight }"
        fit="cover"
      />
    </div>
    <div class="content-width text-ellipsis-line-2" :title="data.name">{{ data.name }}</div>
    <div class="row items-center">
      <btn-dialog-preview-video
        v-if="hasSubtitle"
        size="sm"
        :subtitle-url-list="detialInfo?.sub_url_list"
        :path="data.video_f_path"
      />

      <div>
        <q-btn
          v-if="hasSubtitle"
          size="sm"
          color="black"
          round
          flat
          dense
          icon="closed_caption"
          @click.stop
          title="已有字幕"
        >
          <q-popup-proxy>
            <q-list dense>
              <q-item v-for="(item, index) in detialInfo.sub_url_list" :key="item">
                <q-item-section side>{{ index + 1 }}.</q-item-section>

                <q-item-section class="overflow-hidden ellipsis" :title="item.split`(/\/|\\/)`.pop()">
                  <a class="text-primary" :href="getUrl(item)" target="_blank">{{ item.split(/\/|\\/).pop() }}</a>
                </q-item-section>
                <q-item-section side>
                  <q-btn
                    color="primary"
                    round
                    flat
                    dense
                    icon="construction"
                    :title="`字幕时间轴校准${
                      !formModel.advanced_settings.fix_time_line
                        ? '（此功能需要在进阶设置里开启自动校正字幕时间轴，检测到你当前尚未开启此选项）'
                        : ''
                    }`"
                    @click="doFixSubtitleTimeline(item)"
                    :disable="!formModel.advanced_settings.fix_time_line"
                  ></q-btn>
                </q-item-section>
              </q-item>
            </q-list>
          </q-popup-proxy>
        </q-btn>
        <q-btn v-else color="grey" size="sm" round flat dense icon="closed_caption" @click.stop title="没有字幕" />
      </div>

      <btn-dialog-search-subtitle :path="props.data.video_f_path" is-movie />
      <q-space />

      <btn-upload-subtitle :path="data.video_f_path" dense size="sm" />

      <q-btn
        class="btn-download"
        color="primary"
        round
        flat
        dense
        icon="download_for_offline"
        title="添加到下载队列"
        @click="downloadSubtitle"
        size="sm"
      ></q-btn>

      <div>
        <btn-ignore-video :path="props.data.video_f_path" :video-type="VIDEO_TYPE_MOVIE" size="sm" />
      </div>
    </div>
  </q-card>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/message';
import { VIDEO_TYPE_MOVIE } from 'src/constants/SettingConstants';
import { useQuasar } from 'quasar';
import { doFixSubtitleTimeline, getUrl, subtitleUploadList } from 'pages/library/use-library';
import BtnIgnoreVideo from 'pages/library/BtnIgnoreVideo';
import BtnUploadSubtitle from 'pages/library/BtnUploadSubtitle';
import BtnDialogPreviewVideo from 'pages/library/BtnDialogPreviewVideo';
import BtnDialogSearchSubtitle from 'pages/library/BtnDialogSearchSubtitle';
import { formModel } from 'pages/settings/use-settings';

const props = defineProps({
  data: Object,
  width: {
    type: String,
    default: '160px',
  },
  coverHeight: {
    type: String,
    default: '200px',
  },
});

const $q = useQuasar();

const posterInfo = ref(null);
const detialInfo = ref(null);

const getPosterInfo = async () => {
  const [res] = await LibraryApi.getMoviePoster({
    name: props.data.name,
    main_root_dir_f_path: props.data.main_root_dir_f_path,
    video_f_path: props.data.video_f_path,
  });
  posterInfo.value = res;
};

const getDetailInfo = async () => {
  const [res] = await LibraryApi.getMovieDetail({
    name: props.data.name,
    main_root_dir_f_path: props.data.main_root_dir_f_path,
    video_f_path: props.data.video_f_path,
  });
  detialInfo.value = res;
};

const hasSubtitle = computed(() => detialInfo.value?.sub_url_list.length > 0);

const downloadSubtitle = async () => {
  $q.dialog({
    title: '添加到下载队列',
    message: '选择下载任务的类型：',
    options: {
      model: 3,
      type: 'radio',
      items: [
        { label: '插队任务', value: 3 },
        { label: '一次性任务（执行成功后忽略该任务）', value: 0 },
      ],
    },
    cancel: true,
    persistent: true,
  }).onOk(async (val) => {
    const [, err] = await LibraryApi.downloadSubtitle({
      video_type: VIDEO_TYPE_MOVIE,
      physical_video_file_full_path: props.data.video_f_path,
      task_priority_level: val, // 一般的队列等级是5，如果想要快，那么可以先默认这里填写3，这样就可以插队
      // 媒体服务器内部视频ID  `video/list` 中 获取到的 media_server_inside_video_id，可以用于自动 Emby 字幕列表刷新用
      media_server_inside_video_id: props.data.media_server_inside_video_id,
    });
    if (err !== null) {
      SystemMessage.error(err.message);
    } else {
      SystemMessage.success('已加入下载队列');
    }
  });
};

watch(subtitleUploadList, (val, oldVal) => {
  // 上传字幕列表当前文件有变化时刷新
  if (
    (val.find((e) => e.video_f_path === props.data.video_f_path) &&
      !oldVal.find((e) => e.video_f_path === props.data.video_f_path)) ||
    (!val.find((e) => e.video_f_path === props.data.video_f_path) &&
      oldVal.find((e) => e.video_f_path === props.data.video_f_path))
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
  width: v-bind(width);
}
.text-ellipsis-line-2 {
  height: 40px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.area-cover:hover {
  .btn-download {
    //display: block;
    opacity: 1;
  }
}
</style>
