<template>
  <q-card flat square>
    <div class="area-cover q-mb-sm relative-position">
      <q-img
        :src="data.cover"
        class="content-width bg-grey-2"
        no-spinner
        style="width: 160px; height: 200px"
        fit="cover"
      />
      <q-btn
        class="btn-download absolute-bottom-right"
        color="primary"
        round
        flat
        dense
        icon="download_for_offline"
        title="下载字幕"
        @click="downloadSubtitle"
      ></q-btn>
    </div>
    <div class="content-width text-ellipsis-line-2" :title="data.name">{{ data.name }}</div>
    <div class="row items-center">
      <div class="text-grey">1970-01-01</div>
      <q-space />
      <div>
        <q-btn v-if="hasSubtitle" color="black" round flat dense icon="closed_caption" @click.stop title="已有字幕">
          <q-popup-proxy>
            <q-list dense>
              <q-item v-for="(item, index) in data.sub_f_path_list" :key="item">
                <q-item-section side>{{ index + 1 }}.</q-item-section>

                <q-item-section class="overflow-hidden ellipsis" :title="item.split(/\/|\\/).pop()">
                  <a class="text-primary" :href="getUrl(item)" target="_blank">{{ item.split(/\/|\\/).pop() }}</a>
                </q-item-section>
              </q-item>
            </q-list>
          </q-popup-proxy>
        </q-btn>
        <q-btn v-else color="grey" round flat dense icon="closed_caption" @click.stop title="没有字幕" />
      </div>
    </div>
  </q-card>
</template>

<script setup>
import { computed } from 'vue';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/Message';
import { VIDEO_TYPE_MOVIE } from 'src/constants/SettingConstants';
import { useQuasar } from 'quasar';
import { getUrl } from 'pages/library/useLibrary';

const props = defineProps({
  data: Object,
});

const $q = useQuasar();

const hasSubtitle = computed(() => props.data.sub_f_path_list.length > 0);

const downloadSubtitle = async () => {
  $q.dialog({
    title: '添加到下载队列',
    message: '选择下载任务的类型：',
    options: {
      model: 3,
      type: 'radio',
      items: [
        { label: '正常任务', value: 3 },
        { label: '一次性任务（下载后设置这个任务的状态为"忽略"）', value: 0 },
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
