<template>
  <q-btn
    color="primary"
    flat
    icon="upload"
    v-bind="$attrs"
    :label="dense ? '' : '上传整季字幕'"
    @click="handleUploadClick"
    title="选择多个字幕上传"
  />

  <q-input
    v-show="false"
    type="file"
    ref="qFile"
    v-model="uploadFile"
    @update:model-value="handleFileChange"
    multiple
    accept=".srt,.ass,.ssa,.sbv,.webvtt"
  />

  <q-dialog v-model="show">
    <q-card style="min-width: 900px">
      <q-card-section>
        <div class="text-body1">批量上传</div>
      </q-card-section>

      <q-separator />

      <q-card-section style="min-height: 500px">
        <q-list separator>
          <q-item v-for="item in uploadFile" :key="item.name">
            <q-item-section>{{ item.name }}</q-item-section>
            <q-item-section side>
              <q-select
                v-model="mapForm[item.name]"
                :options="items"
                dense
                filled
                style="width: 200px"
                label="选择对应的集数"
                clearable
              >
                <template v-slot:selected>
                  <span v-if="mapForm[item.name]"> 第{{ mapForm[item.name].episode }}集 </span>
                </template>
                <template v-slot:option="scope">
                  <q-item v-bind="scope.itemProps" clickable v-ripple>
                    <q-item-section>第{{ scope.opt.episode }}集</q-item-section>
                  </q-item>
                </template>
              </q-select>
            </q-item-section>
          </q-item>
        </q-list>
      </q-card-section>

      <q-separator />

      <q-card-actions align="right">
        <q-btn class="q-px-md" v-close-popup flat color="primary" label="关闭" />
        <q-btn class="q-px-md" type="submit" color="primary" :loading="submitting" @click="upload" label="上传">
          <template v-slot:loading>
            <q-spinner-facebook />
          </template>
        </q-btn>
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { reactive, ref } from 'vue';
import { getSubtitleUploadList } from 'pages/library/use-library';
import LibraryApi from 'src/api/LibraryApi';
import { SystemMessage } from 'src/utils/message';
import eventBus from 'vue3-eventbus';
import { getEpisode } from 'src/utils/subtitle';

const props = defineProps({
  items: Array,
  dense: {
    type: Boolean,
    default: false,
  },
});

const uploadFile = ref(null);
const qFile = ref(null);
const show = ref(false);
const submitting = ref(false);
const mapForm = reactive({});

const handleUploadClick = () => {
  qFile.value.$el.click();
};

const upload = async () => {
  submitting.value = true;
  // eslint-disable-next-line no-restricted-syntax
  for (const name of Object.keys(mapForm)) {
    const item = mapForm[name];
    if (item) {
      const formData = new FormData();
      formData.append('video_f_path', item.video_f_path);
      formData.append(
        'file',
        [].find.call(uploadFile.value, (file) => file.name === name)
      );
      // eslint-disable-next-line no-await-in-loop
      await LibraryApi.uploadSubtitle(formData);
      // eslint-disable-next-line no-await-in-loop
      await getSubtitleUploadList();
      eventBus.emit('subtitle-uploaded');
    }
  }
  SystemMessage.success('字幕上传成功。如果设置开启了“自动校正时间轴”，处理需要一些时间，请耐心等待', {
    timeout: 3000,
  });
  submitting.value = false;
  show.value = false;
};

const handleFileChange = (val) => {
  if (val.length > 0) {
    show.value = true;
    [].forEach.call(val, (e) => {
      const episode = getEpisode(e.name);
      const item = props.items.find((i) => i.episode === episode);
      if (item) {
        mapForm[e.name] = item;
      } else {
        mapForm[e.name] = null;
      }
    });
  }
};
</script>
