<template>
  <q-btn color="primary" icon="search" size="sm" flat dense v-bind="$attrs" @click="visible = true" title="字幕搜索" />

  <q-dialog v-model="visible" transition-show="slide-up" transition-hide="slide-down" @before-show="handleBeforeShow">
    <q-card style="min-width: 70vw">
      <q-card-section>
        <div class="text-h6 text-grey-8">字幕搜索</div>
        <div class="text-grey">点击关键字跳转到网站搜索</div>
      </q-card-section>

      <q-separator />

      <q-card-section v-if="searchInfo">
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
      </q-card-section>
      <q-card-section v-else>
        <div class="row items-center justify-center" style="height: 200px">
          <q-spinner size="30px" />
        </div>
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { ref } from 'vue';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/Message';

const props = defineProps({
  path: String,
  isMovie: {
    type: Boolean,
    default: false,
  },
});

const visible = ref(false);
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

const handleBeforeShow = () => {
  getSearchInfo();
};
</script>
