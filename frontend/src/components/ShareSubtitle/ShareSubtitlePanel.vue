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
      <div class="relative-position q-card--bordered" style="height: 120px; overflow: auto">
        <search-panel-csf-api
          :is-movie="isMovie"
          :path="mediaData.video_f_path"
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
        <q-btn color="primary" label="提交" :disable="isSubmitDisabled" @click="handleUpload" />
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
import { computed, onMounted, reactive, ref } from 'vue';
import BtnDialogPreviewVideo from 'pages/library/BtnDialogPreviewVideo.vue';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/message';
import SearchPanelCsfApi from 'pages/library/SearchPanelCsfApi.vue';
import { getUrl } from 'pages/library/use-library';
import CsfSubtitlesShareApi from 'src/api/CsfSubtitlesShareApi';

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
  return null;
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
  const [res, err] = await LibraryApi.getUploadInfo({
    is_movie: props.isMovie,
    video_f_path: props.mediaData.video_f_path,
    sub_f_path: uploadListForm[0].selected.subPath,
    season: null,
    episode: null,
  });
  if (err !== null) {
    SystemMessage.error(err.message);
    return;
  }
  const [res2, err2] = await CsfSubtitlesShareApi.getUploadUrl({
    ...res,
    score: form.ratting,
    mark_type: form.markType,
  });
  if (err2 !== null) {
    SystemMessage.error(err2.message);
    return;
  }
  const { subUrl } = uploadListForm[0].selected;
  // 下载字幕，并将其存储为File对象
  const res3 = await fetch(getUrl(subUrl));
  const blob = await res3.blob();
  const file = new File([blob], subUrl.split(/[/\\]/).pop(), { type: 'text/plain' });
  // 上传字幕
  await fetch(res2.upload_url, {
    method: 'PUT',
    body: file,
  });
  await CsfSubtitlesShareApi.setUploadSuccess({
    token: res2.token,
  });
};

onMounted(() => {
  getImdbId();
});
</script>
