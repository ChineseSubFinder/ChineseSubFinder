<template>
  <span @click="visible = true">
    <slot></slot>
  </span>

  <q-dialog v-model="visible">
    <q-card style="width: 400px">
      <q-card-section>
        <div class="text-h6">{{ data.name }} 剧集列表</div>
      </q-card-section>

      <q-separator />

      <q-card-section>
        <q-list dense>
          <q-item v-for="item in sortedVideos" :key="item.name">
            <q-item-section>第 {{ item.season }} 季 {{ pandStart2(item.episode) }} 集</q-item-section>
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
                    <q-item v-for="(item1, index) in item.sub_f_path_list" :key="item1">
                      <q-item-section side>{{ index + 1 }}.</q-item-section>

                      <q-item-section class="overflow-hidden ellipsis" :title="item1.split(/\/|\\/).pop()">
                        <a class="text-primary" href="" target="_blank">{{ item1.split(/\/|\\/).pop() }}</a>
                      </q-item-section>
                    </q-item>
                  </q-list>
                </q-popup-proxy>
              </q-btn>
              <q-btn v-else color="grey" round flat dense icon="closed_caption" @click.stop title="没有字幕" />
            </q-item-section>

            <q-item-section side>
              <q-btn
                class="btn-download absolute-bottom-right"
                color="primary"
                round
                flat
                dense
                icon="download_for_offline"
                title="下载字幕"
                @click="downloadSubtitle(item.id)"
              ></q-btn>
            </q-item-section>
          </q-item>
        </q-list>
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { computed, ref } from 'vue';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/Message';

const props = defineProps({
  data: Object,
});

// 按季度、剧集排序
const sortedVideos = computed(() =>
  [...props.data.one_video_info].sort((a, b) => {
    if (a.season > b.season) {
      return 1;
    }
    if (a.season < b.season) {
      return -1;
    }
    if (a.episode > b.episode) {
      return 1;
    }
    if (a.episode < b.episode) {
      return -1;
    }
    return 0;
  })
);

const pandStart2 = (num) => {
  if (num < 10) {
    return `0${num}`;
  }
  return num;
};

const visible = ref(false);

const downloadSubtitle = async (id) => {
  const [, err] = await LibraryApi.downloadSubtitle(id);
  if (err !== null) {
    SystemMessage.error(err.message);
  } else {
    SystemMessage.success('已加入下载队列');
  }
};
</script>
