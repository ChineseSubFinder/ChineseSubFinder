<template>
  <span @click.stop="visible = true">
    <slot></slot>
  </span>

  <q-dialog v-model="visible">
    <q-card style="width: 900px; max-width: 900px">
      <q-card-section>
        <div class="text-h6">{{ data.name }} 剧集列表</div>
      </q-card-section>

      <q-tabs v-model="tab" dense active-color="primary" indicator-color="primary" align="justify" narrow-indicator>
        <q-tab
          v-for="item in categoryVideos"
          :key="item.season"
          :name="item.season"
          :label="`第${item.season}季`"
          style="max-width: 150px"
        />
      </q-tabs>

      <q-separator />

      <q-card-section style="max-height: 40vh; overflow: auto">
        <div class="row items-center q-ml-md q-py-none">
          <q-checkbox
            :model-value="selectAllValue"
            indeterminate-value="maybe"
            @click="handleSelectAll"
            title="全选/反选"
          />

          <q-btn
            class="btn-download"
            color="primary"
            label="下载选中"
            flat
            :disable="selection.length === 0"
            @click="downloadSelection"
          ></q-btn>

          <q-btn
            class="btn-download"
            color="primary"
            icon="lock"
            title="锁定选中视频，不进行字幕下载"
            flat
            :disable="selection.length === 0"
            @click="skipAll(true)"
          ></q-btn>

          <q-btn
            class="btn-download"
            color="primary"
            icon="lock_open"
            title="解锁选中视频"
            flat
            :disable="selection.length === 0"
            @click="skipAll(false)"
          ></q-btn>

          <q-space />

          <btn-dialog-search-subtitle
            search-package
            :package-episodes="currentTabEpisodes"
            label="搜索本季字幕包"
            size="md"
          />
          <btn-upload-multiple-for-tv :items="currentTabEpisodes" />
        </div>

        <q-tab-panels v-model="tab" animated>
          <q-tab-panel v-for="{ season, episodes } in categoryVideos" :key="season" :name="season" style="padding: 0">
            <q-list dense>
              <q-item v-for="item in episodes" :key="item.name">
                <q-item-section side top>
                  <q-checkbox v-model="selection" :val="item" />
                </q-item-section>
                <q-item-section>第 {{ pandStart2(item.episode) }} 集</q-item-section>

                <q-item-section v-if="item.sub_f_path_list.length" side>
                  <btn-dialog-preview-video :subtitle-url-list="item.sub_url_list" :path="item.video_f_path" />
                </q-item-section>

                <q-item-section side>
                  <btn-upload-subtitle :path="item.video_f_path" />
                </q-item-section>

                <q-item-section side>
                  <btn-ignore-video :path="item.video_f_path" :video-type="VIDEO_TYPE_TV" />
                </q-item-section>

                <q-item-section side>
                  <q-btn
                    v-if="item.sub_f_path_list.length"
                    color="black"
                    round
                    flat
                    dense
                    icon="closed_caption"
                    @click.stop
                    title="已有字幕"
                  >
                    <q-popup-proxy anchor="top right">
                      <q-list dense>
                        <q-item v-for="(item1, index) in item.sub_url_list" :key="item1">
                          <q-item-section side>{{ index + 1 }}.</q-item-section>

                          <q-item-section class="overflow-hidden ellipsis" :title="item1.split(/\/|\\/).pop()">
                            <a class="text-primary" :href="getUrl(item1)" target="_blank">{{
                              item1.split(/\/|\\/).pop()
                            }}</a>
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
                              @click="doFixSubtitleTimeline(item1)"
                              :disable="!formModel.advanced_settings.fix_time_line"
                            ></q-btn>
                          </q-item-section>
                        </q-item>
                      </q-list>
                    </q-popup-proxy>
                  </q-btn>
                  <q-btn v-else color="grey" round flat dense icon="closed_caption" @click.stop title="没有字幕" />
                </q-item-section>

                <q-item-section side>
                  <btn-dialog-search-subtitle
                    size="md"
                    round
                    :path="item.video_f_path"
                    :season="item.season"
                    :episode="item.episode"
                  />
                </q-item-section>

                <q-item-section side>
                  <q-btn
                    class="btn-download"
                    color="primary"
                    round
                    flat
                    dense
                    icon="download_for_offline"
                    title="添加到下载队列"
                    @click="downloadSubtitle(item)"
                  ></q-btn>
                </q-item-section>
              </q-item>
            </q-list>
          </q-tab-panel>
        </q-tab-panels>
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { computed, ref, watch } from 'vue';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/message';
import { VIDEO_TYPE_TV } from 'src/constants/SettingConstants';
import config from 'src/config';
import { useQuasar } from 'quasar';
import { useSelection } from 'src/composables/use-selection';
import BtnIgnoreVideo from 'pages/library/BtnIgnoreVideo';
import eventBus from 'vue3-eventbus';
import BtnUploadSubtitle from 'pages/library/BtnUploadSubtitle';
import BtnDialogPreviewVideo from 'pages/library/BtnDialogPreviewVideo';
import BtnDialogSearchSubtitle from 'pages/library/BtnDialogSearchSubtitle';
import BtnUploadMultipleForTv from 'pages/library/tvs/BtnUploadMultipleForTv';
import { doFixSubtitleTimeline } from 'pages/library/use-library';
import { formModel } from 'pages/settings/use-settings';

const props = defineProps({
  data: Object,
});

const $q = useQuasar();

const categoryVideos = computed(() => {
  // [{season: episodes: []}]
  const result = [];
  props.data?.one_video_info.forEach((item) => {
    const { season } = item;
    const index = result.findIndex((e) => e.season === season);
    if (index === -1) {
      result.push({
        season,
        episodes: [item],
      });
    } else {
      result[index].episodes.push(item);
    }
  });
  result.sort((a, b) => a.season - b.season);
  result.forEach((item) => {
    item.episodes.sort((a, b) => a.episode - b.episode);
  });
  return result;
});

const tab = ref(null);

watch(categoryVideos, () => {
  if (categoryVideos.value.length && tab.value === null) {
    tab.value = categoryVideos.value[0].season;
  }
});

const currentTabEpisodes = computed(() => categoryVideos.value.find((e) => e.season === tab.value)?.episodes ?? []);

const { selectAllValue, handleSelectAll, selection } = useSelection(currentTabEpisodes);
watch(tab, () => {
  selection.value = [];
});

const pandStart2 = (num) => {
  if (num < 10) {
    return `0${num}`;
  }
  return num;
};

const visible = ref(false);

const getUrl = (path) => config.BACKEND_URL + path.split(/\/|\\/).join('/');

const downloadSubtitle = async (items) => {
  const downloadList = items instanceof Array ? items : [items];
  $q.dialog({
    title: `添加 ${downloadList.length}个 视频任务到下载队列`,
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
    // 下载全部Promises
    const promises = downloadList.map(async (item) => {
      const [, err] = await LibraryApi.downloadSubtitle({
        video_type: VIDEO_TYPE_TV,
        physical_video_file_full_path: item.video_f_path,
        task_priority_level: val, // 一般的队列等级是5，如果想要快，那么可以先默认这里填写3，这样就可以插队
        // 媒体服务器内部视频ID  `video/list` 中 获取到的 media_server_inside_video_id，可以用于自动 Emby 字幕列表刷新用
        media_server_inside_video_id: item.media_server_inside_video_id,
      });
      if (err !== null) {
        return Promise.reject(err);
      }
      return Promise.resolve();
    });

    const result = await Promise.allSettled(promises);

    const successCount = result.filter((item) => item.status === 'fulfilled').length;
    const errorCount = result.filter((item) => item.status === 'rejected').length;

    const msg = `成功添加 ${successCount} 个任务到下载队列${errorCount ? `，失败 ${errorCount} 个` : ''}`;

    SystemMessage.success(msg);
  });
};

const skipAll = async (isSkip) => {
  $q.dialog({
    title: `${isSkip ? '锁定' : '解锁'}选中视频`,
    message: `确定要${isSkip ? '锁定' : '解锁'}选中视频吗？`,
    cancel: true,
    persistent: true,
  }).onOk(async () => {
    const [, err] = await LibraryApi.setSkipInfo({
      video_skip_infos: selection.value.map((item) => ({
        video_type: VIDEO_TYPE_TV,
        physical_video_file_full_path: item.video_f_path,
        is_bluray: false,
        is_skip: isSkip,
      })),
    });
    if (err !== null) {
      SystemMessage.error(err.message);
      return;
    }
    const [res, err2] = await LibraryApi.getSkipInfo({
      video_skip_infos: selection.value.map((item) => ({
        video_type: VIDEO_TYPE_TV,
        physical_video_file_full_path: item.video_f_path,
        is_bluray: false,
        is_skip: true,
      })),
    });
    if (err2 !== null) {
      SystemMessage.error(err2.message);
      return;
    }

    selection.value.forEach((item, index) => {
      eventBus.emit(`refresh-skip-status-${item.video_f_path}`, res.is_skips[index]);
    });

    SystemMessage.success('操作成功');
  });
};

const downloadSelection = () => {
  downloadSubtitle(selection.value);
};
</script>
