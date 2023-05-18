<template>
  <q-btn
    color="primary"
    flat
    icon="upload"
    v-bind="$attrs"
    :label="dense ? '' : '本地上传整季字幕'"
    @click="handleUploadClick"
    title="选择多个字幕上传"
  />

  <q-input
    v-show="false"
    type="file"
    ref="qFile"
    v-model="inputFiles"
    @update:model-value="handleFileChange"
    multiple
    accept=".srt,.ass,.ssa,.sbv,.webvtt"
  />

  <q-dialog v-model="show" @before-show="handleBeforeShow">
    <q-card
      class="relative-position"
      style="min-width: 900px"
      @dragover="handleDragenter"
      @drop="handleDrop"
      @dragleave="handleDragleave"
    >
      <div class="drag-mask row items-center justify-center text-white text-h5" v-if="isDragover">
        放置字幕文件到此处
      </div>
      <q-card-section>
        <div class="text-body1">批量上传</div>
      </q-card-section>

      <q-separator />

      <q-card-section style="min-height: 500px">
        <div
          v-if="uploadFiles.length === 0"
          class="upload-area column justify-center items-center q-gutter-sm"
          style="height: 500px"
        >
          <div>拖拽文件上传</div>
          <div>或者</div>
          <q-btn color="primary" label="选择文件" dense flat @click="() => qFile.$el.click()" />
        </div>
        <template v-else>
          <q-btn color="primary" label="添加字幕文件" @click="() => qFile.$el.click()" class="q-mb-sm" />
          <q-list separator>
            <q-item v-for="item in uploadFiles" :key="item.name">
              <q-item-section>{{ item.name }}</q-item-section>
              <q-item-section side>
                <q-select
                  :disable="submitting"
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

              <q-item-section side>
                <q-btn color="negative" flat icon="close" dense title="删除" rounded @click="handleRemoveFile(item)" />
              </q-item-section>
            </q-item>
          </q-list>
        </template>
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

const inputFiles = ref(null);
const qFile = ref(null);
const show = ref(false);
const submitting = ref(false);
const mapForm = reactive({});
const isDragover = ref(false);
const uploadFiles = ref([]);

const handleUploadClick = () => {
  show.value = true;
  // qFile.value.$el.click();
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
        uploadFiles.value.find((file) => file.name === name)
      );
      // eslint-disable-next-line no-await-in-loop
      await LibraryApi.uploadSubtitle(formData);
      // eslint-disable-next-line no-await-in-loop
      await getSubtitleUploadList();
      // sleep 1s
      // eslint-disable-next-line no-await-in-loop,no-promise-executor-return
      await new Promise((resolve) => setTimeout(resolve, 1000));
      eventBus.emit('subtitle-uploaded');
    }
  }
  SystemMessage.success('字幕上传成功。如果设置开启了“自动校正时间轴”，处理需要一些时间，请耐心等待', {
    timeout: 3000,
  });
  submitting.value = false;
  show.value = false;
};

const addMapForm = (file) => {
  const episode = getEpisode(file.name);
  const item = props.items.find((i) => i.episode === episode);
  if (item) {
    mapForm[file.name] = item;
  } else {
    mapForm[file.name] = null;
  }
};

const handleNewFileAdded = (file) => {
  if (/\.srt|\.ass|\.ssa|\.sbv|\.webvtt/.test(file.name)) {
    if (!uploadFiles.value.some((f) => f.name === file.name)) {
      uploadFiles.value.push(file);
      addMapForm(file);
    }
  }
};

const handleFileChange = (val) => {
  [].forEach.call(val, (file) => {
    handleNewFileAdded(file);
  });
};

const handleDragenter = (e) => {
  e.preventDefault();
  isDragover.value = true;
};

const handleDragleave = (e) => {
  e.preventDefault();
};

const handleDrop = (e) => {
  e.preventDefault();
  const { files } = e.dataTransfer;
  [].forEach.call(files, (file) => {
    handleNewFileAdded(file);
  });
  handleFileChange(files);
  isDragover.value = false;
};

const handleRemoveFile = (file) => {
  uploadFiles.value = uploadFiles.value.filter((f) => f.name !== file.name);
  delete mapForm[file.name];
};

const handleBeforeShow = () => {
  uploadFiles.value = [];
  Object.keys(mapForm).forEach((key) => {
    delete mapForm[key];
  });
};
</script>

<style lang="scss">
.drag-mask {
  position: absolute;
  width: 100%;
  height: 100%;
  z-index: 999;
  background-color: rgba(0, 0, 0, 0.5);
}

.upload-area {
  border: 2px dashed #ccc;
  border-radius: 4px;
}
</style>
