<template>
  <q-btn label="导出" @click="visible = true" />
  <q-dialog v-model="visible">
    <q-card style="width: 800px; max-width: 800px">
      <q-card-section>
        <div class="text-h6 text-grey-8">导出</div>
      </q-card-section>

      <q-separator />

      <q-card-section>
        <q-toggle
          v-model="hideSensitive"
          color="primary"
          label="去除敏感配置（反馈问题用）"
        />
        <div class="code-area relative-position overflow-auto bg-grey-2 q-pa-sm">
          <q-btn
            class="copy-btn absolute-top-right q-ma-sm hidden"
            flat
            dense
            icon="content_copy"
            color="grey-8"
            @click="copy(settingsString)"
          />
          <pre style="height: 300px">{{ settingsString }}</pre>
        </div>
      </q-card-section>

      <q-separator />

      <q-card-actions align="right" class="q-pa-md">
        <q-btn class="q-px-md" v-close-popup flat color="primary" label="关闭" />
        <q-btn class="q-px-md" type="submit" color="primary" label="导出文件" @click="exportSettings"/>
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { computed, ref } from 'vue';
import { settingsState } from 'pages/settings/useSettings';
import { copyToClipboard } from 'quasar';
import { SystemMessage } from 'src/utils/Message';
import {deepCopy} from 'src/utils/CommonUtils';

const visible = ref(false);
const hideSensitive = ref(false);

const settingsString = computed(() => {
  const settings = deepCopy(settingsState.data);
  delete settings.user_info;
  if (hideSensitive.value) {
    delete settings.common_settings.threads;
    delete settings.emby_settings.api_key;
    delete settings.emby_settings.address_url;
  }
  return JSON.stringify(settings, null, 2);
});

const exportSettings = () => {
  const element = document.createElement('a');
  element.setAttribute('href', `data:text/plain;charset=utf-8,${encodeURIComponent(settingsString.value)}`);
  element.setAttribute('download', 'ChineseSubFinderSettings.json');

  element.style.display = 'none';
  document.body.appendChild(element);

  element.click();

  document.body.removeChild(element);
};

const copy = (str) =>
  copyToClipboard(str).then(() => {
    SystemMessage.success('已复制到剪贴板');
  });
</script>

<style lang="scss" scoped>
.code-area:hover {
  .copy-btn {
    display: block !important;
  }
}
</style>
