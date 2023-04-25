<template>
  <div class="q-pa-md">
    <section class="relative-position">
      <div>确认视频信息</div>
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

    <q-separator class="q-my-md" />

    <section>
      <div>确认要上传的字幕</div>
      <q-list dense style="max-height: 600px">
        <q-item v-for="item in previewList" :key="item.url" clickable>
          <q-item-section>
            <q-item-label>{{ item.name }}</q-item-label>
          </q-item-section>
          <q-item-section side>
            <btn-dialog-preview-video size="sm" :subtitle-url-list="[item.url]" :path="item.path" />
          </q-item-section>
        </q-item>
      </q-list>
    </section>

    <q-separator class="q-my-md" />

    <section>
      <div>这是机翻吗？</div>
      <div class="q-gutter-sm">
        <q-radio v-model="form.isMachineTranslate" :val="1" label="是" />
        <q-radio v-model="form.isMachineTranslate" :val="0" label="否" />
      </div>

      <div>给字幕一个评分吧</div>
      <q-rating v-model="form.ratting" size="2em" :max="4" color="primary">
        <template v-slot:tip-1>
          <q-tooltip>不匹配</q-tooltip>
        </template>
        <template v-slot:tip-2>
          <q-tooltip>差</q-tooltip>
        </template>
        <template v-slot:tip-3>
          <q-tooltip>一般</q-tooltip>
        </template>
        <template v-slot:tip-4>
          <q-tooltip>好</q-tooltip>
        </template>
      </q-rating>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue';
import BtnDialogPreviewVideo from 'pages/library/BtnDialogPreviewVideo.vue';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/message';

// interface TmdbInfo {
//   ImdbId: string;
//   OriginalLanguage: string;
//   OriginalTitle: string;
//   TmdbId: string;
//   TitleCn: string;
//   TitleEn: string;
//   Year: string;
// }

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

const form = reactive({
  isMachineTranslate: null,
  ratting: null,
});

const previewList = computed(() => {
  if (props.isMovie) {
    return props.mediaData?.sub_url_list.map((e) => ({
      name: e.split(/[/\\]/).pop(),
      path: props.mediaData.video_f_path,
      url: e,
    }));
  }
  return props.mediaData?.reduce((acc, cur) => {
    const { sub_url_list: subUrlList } = cur;
    if (subUrlList) {
      acc.push(
        ...subUrlList.map((e) => ({
          name: e.split(/[/\\]/).pop(),
          path: cur.video_f_path,
          url: e,
        }))
      );
    }
    return acc;
  }, []);
});

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

onMounted(() => {
  getImdbId();
});
</script>
