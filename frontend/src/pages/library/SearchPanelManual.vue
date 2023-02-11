<template>
  <div style="min-height: 300px">
    <div class="text-grey">点击关键字跳转到网站搜索</div>
    <q-list separator>
      <q-item v-for="url in searchInfo?.search_url" :key="url">
        <q-item-section top side style="width: 200px" class="text-bold text-black">
          {{ getDomain(url) }}
        </q-item-section>
        <q-item-section>
          <div class="row q-gutter-sm">
            <a
              v-for="item in searchInfo?.key_words"
              :key="item"
              :href="getSearchUrl(url, item)"
              target="_blank"
              style="text-decoration: none"
            >
              <q-badge class="cursor-pointer" color="secondary" title="点击跳转到网站搜索">{{ item }}</q-badge>
            </a>
          </div>
        </q-item-section>
      </q-item>
    </q-list>
    <q-inner-loading :showing="!searchInfo">
      <q-spinner size="50px" color="primary" />
    </q-inner-loading>
  </div>
</template>

<script setup>
import { onMounted, ref } from 'vue';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/message';

const props = defineProps({
  path: String,
  isMovie: {
    type: Boolean,
    default: false,
  },
});

const searchInfo = ref(null);

const getSearchInfo = async () => {
  const [data, err] = await LibraryApi.getSearchSubtitleInfo({
    video_f_path: props.path,
    is_movie: props.isMovie,
  });
  if (err !== null) {
    SystemMessage.error(err.message);
  }
  searchInfo.value = data;
};

const getDomain = (url) => {
  const reg = /https?:\/\/([^/]+)/;
  const result = reg.exec(url);
  return result[1];
};

const getSearchUrl = (url, keyword) => {
  if (url.includes('%s')) {
    return url.replace('%s', encodeURIComponent(keyword));
  }
  return url + encodeURIComponent(keyword);
};

onMounted(() => {
  getSearchInfo();
});
</script>
