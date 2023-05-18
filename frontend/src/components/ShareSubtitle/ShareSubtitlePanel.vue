<template>
  <div class="q-pa-md">
    <section class="relative-position">
      <div class="text-bold">1. 确认视频信息</div>
      <q-markup-table flat dense>
        <q-tr>
          <q-td style="width: 150px">IMDB ID</q-td>
          <q-td>{{ tmdbInfo?.ImdbId }}</q-td>
        </q-tr>
        <q-tr>
          <q-td>TMDB ID</q-td>
          <q-td>{{ tmdbInfo?.TmdbId }}</q-td>
        </q-tr>
        <q-tr>
          <q-td>原始语言</q-td>
          <q-td>{{ tmdbInfo?.OriginalLanguage }}</q-td>
        </q-tr>
        <q-tr>
          <q-td>原始标题</q-td>
          <q-td>{{ tmdbInfo?.OriginalTitle }}</q-td>
        </q-tr>
        <q-tr>
          <q-td>中文标题</q-td>
          <q-td>{{ tmdbInfo?.TitleCn }}</q-td>
        </q-tr>
        <q-tr>
          <q-td>季</q-td>
          <q-td>第{{ mediaData?.[0]?.season }}季</q-td>
        </q-tr>
        <q-tr>
          <q-td>英文标题</q-td>
          <q-td>{{ tmdbInfo?.TitleEn }}</q-td>
        </q-tr>
        <q-tr>
          <q-td>年份</q-td>
          <q-td>{{ tmdbInfo?.Year }}</q-td>
        </q-tr>
      </q-markup-table>
      <q-inner-loading :showing="loading" />
    </section>

    <section class="q-mt-md">
      <div class="text-bold">2. 确认 SubtitleBest 中该视频已存在的字幕</div>
      <div class="relative-position q-card--bordered" style="height: 320px; overflow: auto">
        <search-panel-csf-api
          v-if="isMovie"
          is-movie
          :path="mediaData.video_f_path"
          hide-download
          hide-limit
          @get-result="handleGetResult"
        />
        <search-panel-csf-api-tv-package
          v-else
          :episodes="mediaData"
          hide-download
          hide-limit
          @get-result="handleGetResult"
        />
      </div>
    </section>

    <section class="q-mt-md">
      <div class="text-bold">3. 选择要上传的字幕</div>
      <q-list style="max-height: 600px">
        <q-item v-for="(item, index) in uploadListForm" :key="index">
          <q-item-section v-if="!isMovie" side style="width: 80px">
            <div class="row items-center q-gutter-xs">
              <q-spinner v-if="submitting && index === uploadFinishedCount" />
              <q-icon v-if="uploadFinishedCount > index" name="check_circle" color="positive" title="上传完成" />
              <div>第{{ item.episode }}集</div>
            </div>
          </q-item-section>
          <q-item-section>
            <q-select v-model="item.selected" :options="item.list" option-label="name" option-value="id" filled dense />
          </q-item-section>
          <q-item-section side>
            <btn-dialog-preview-video
              size="sm"
              :subtitle-url-list="[item.selected.subUrl]"
              :path="item.selected.path"
            />
          </q-item-section>
        </q-item>
      </q-list>
    </section>

    <section class="q-mt-md">
      <div class="text-bold">4. 填写字幕上传信息</div>

      <div class="row items-center">
        <div style="width: 120px">这是机翻吗？</div>
        <div class="q-gutter-sm">
          <q-radio v-model="form.markType" :val="2" label="是" />
          <q-radio v-model="form.markType" :val="1" label="否" />
        </div>
      </div>

      <div class="row items-center">
        <div style="width: 120px">字幕质量如何？</div>
        <div class="q-gutter-sm">
          <q-radio v-model="form.ratting" :val="1" label="不匹配" />
          <q-radio v-model="form.ratting" :val="2" label="差" />
          <q-radio v-model="form.ratting" :val="3" label="一般" />
          <q-radio v-model="form.ratting" :val="4" label="好" />
        </div>
      </div>

      <div class="q-mt-sm">
        <q-checkbox v-model="form.confirmMatch" dense label="我确认上传的字幕和视频匹配" />
      </div>

      <div class="q-mt-sm">
        <q-checkbox
          v-if="isSubtitleBestHasResult"
          v-model="form.confirmBetter"
          dense
          label="我确认上传的字幕比 SubtitleBest 上已有的字幕更好"
        />
      </div>
    </section>

    <q-separator class="q-my-md" />

    <section>
      <div class="q-mt-md row items-center q-gutter-md">
        <q-btn color="primary" label="提交" :disable="isSubmitDisabled" :loading="submitting" @click="handleUpload" />
        <div v-if="isSubmitDisabled" class="text-negative">
          <template v-if="form.ratting === 1 || form.ratting === 2">
            不能上传字幕质量为"不匹配"或者"差"的字幕
          </template>
          <template v-else>请完善表单</template>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
/* eslint-disable no-await-in-loop */
import { computed, onMounted, reactive, ref } from 'vue';
import BtnDialogPreviewVideo from 'pages/library/BtnDialogPreviewVideo.vue';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/message';
import SearchPanelCsfApi from 'pages/library/SearchPanelCsfApi.vue';
import { getUrl } from 'pages/library/use-library';
import CsfSubtitlesShareApi from 'src/api/CsfSubtitlesShareApi';
import SearchPanelCsfApiTvPackage from 'pages/library/SearchPanelCsfApiTvPackage.vue';

const props = defineProps({
  isMovie: {
    type: Boolean,
    default: false,
  },
  mediaData: {
    type: Object,
  },
});

const tmdbInfo = ref(null);
const loading = ref(false);
const submitting = ref(false);
const uploadFinishedCount = ref(0);
const isSubtitleBestHasResult = ref(false);

const form = reactive({
  markType: null,
  ratting: null,
  confirmMatch: false,
  confirmBetter: false,
});

const isSubmitDisabled = computed(
  () =>
    !form.markType ||
    !form.ratting ||
    [1, 2].includes(form.ratting) ||
    !form.confirmMatch ||
    (isSubtitleBestHasResult.value && !form.confirmBetter)
);

const getUploadListForm = () => {
  if (props.isMovie) {
    // 构建这样一个数组，子项内容为：{selected: listItem, list: [{name: xxx, path: xxx, url: xxx}]}
    const list = props.mediaData?.sub_url_list.map((e, index) => ({
      name: e.split(/[/\\]/).pop(),
      path: props.mediaData.video_f_path,
      subUrl: e,
      subPath: props.mediaData.sub_f_path_list[index],
    }));
    return [
      {
        selected: list[0],
        list,
      },
    ];
  }

  // 连续剧
  const result = props.mediaData
    .map((e) => {
      const list = e.sub_url_list.map((subUrl, index) => ({
        name: subUrl.split(/[/\\]/).pop(),
        path: e.video_f_path,
        subUrl,
        subPath: e.sub_f_path_list[index],
      }));
      return {
        episode: e.episode,
        video_f_path: e.video_f_path,
        selected: list[0],
        list,
      };
    })
    .sort((a, b) => a.episode - b.episode);
  return result;
};

const uploadListForm = reactive(getUploadListForm());

const getImdbId = async () => {
  loading.value = true;
  const videoFilePath = props.isMovie ? props.mediaData.video_f_path : props.mediaData[0].video_f_path;
  const [res, err] = await LibraryApi.getImdbId({
    video_f_path: videoFilePath,
    is_movie: props.isMovie,
  });
  loading.value = false;
  if (err) {
    SystemMessage.error(err.message);
    return;
  }
  tmdbInfo.value = res;
};

const handleGetResult = (result) => {
  isSubtitleBestHasResult.value = result?.length > 0;
};

const handleUpload = async () => {
  submitting.value = true;
  uploadFinishedCount.value = 0;
  const tokens = [];

  for (let i = 0; i < uploadListForm.length; i += 1) {
    const uploadListFormI = uploadListForm[i];

    // 获取上传信息
    const [res, err] = await LibraryApi.getUploadInfo({
      is_movie: props.isMovie,
      video_f_path: props.isMovie ? props.mediaData.video_f_path : uploadListFormI.video_f_path,
      sub_f_path: uploadListFormI.selected.subPath,
      season: props.mediaData?.[0]?.season,
      episode: uploadListFormI.episode,
    });
    if (err !== null) {
      SystemMessage.error(err.message);
      submitting.value = false;
      return;
    }

    // 获取上传地址
    const [res2, err2] = await CsfSubtitlesShareApi.getUploadUrl({
      ...res,
      score: form.ratting,
      mark_type: form.markType,
    });
    if (err2 !== null) {
      SystemMessage.error(err2.message);
      submitting.value = false;
      return;
    }

    const { subUrl } = uploadListFormI.selected;
    // 下载字幕，并将其存储为File对象
    const res3 = await fetch(getUrl(subUrl));
    const blob = await res3.blob();
    const file = new File([blob], subUrl.split(/[/\\]/).pop(), { type: 'text/plain' });
    // 上传字幕
    await fetch(res2.upload_url, {
      method: 'PUT',
      body: file,
    });

    // 通知单个视频上传完成
    const [res4, err4] = await CsfSubtitlesShareApi.setUploadSuccess({
      token: res2.token,
    });
    if (err4 !== null) {
      SystemMessage.error(err4.message);
      submitting.value = false;
      return;
    }
    tokens.push(res4.token);
    uploadFinishedCount.value += 1;
  }

  // 如果是电视剧，通知所有视频上传完成
  if (!props.isMovie) {
    const [, err] = await CsfSubtitlesShareApi.setUploadSuccessForTv({
      tokens,
    });
    if (err !== null) {
      SystemMessage.error(err.message);
      submitting.value = false;
      return;
    }
  }
  submitting.value = false;
  SystemMessage.success('上传成功');
};

onMounted(() => {
  getImdbId();
});
</script>
