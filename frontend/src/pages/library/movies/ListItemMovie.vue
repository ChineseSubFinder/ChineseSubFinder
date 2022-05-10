<template>
  <q-card flat square>
    <div class="area-cover q-mb-sm relative-position">
      <q-img src="https://via.placeholder.com/500" class="content-width bg-grey-2" height="230px" no-spinner />
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
                  <a class="text-primary" href="" target="_blank">{{ item.split(/\/|\\/).pop() }}</a>
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

const props = defineProps({
  data: Object,
});

const hasSubtitle = computed(() => props.data.sub_f_path_list.length > 0);

const downloadSubtitle = async () => {
  const [, err] = await LibraryApi.downloadSubtitle(props.data.id);
  if (err !== null) {
    SystemMessage.error(err.message);
  } else {
    SystemMessage.success('已加入下载队列');
  }
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
