<template>
  <q-btn label="导出" @click="visible = true" />
  <q-dialog v-model="visible">
    <q-card style="width: 800px; max-width: 800px">
      <q-card-section>
        <div class="text-h6 text-grey-8">导出</div>
      </q-card-section>

      <q-separator />

      <q-card-section>
        <q-toggle v-model="hideSensitive" color="primary" label="去除敏感配置（反馈问题用）" />
        <div class="code-area relative-position overflow-auto bg-grey-2 q-pa-sm">
          <copy-to-clipboard-btn class="copy-btn absolute-top-right q-ma-sm hidden" :text="settingsString" />
          <pre style="height: 300px">{{ settingsString }}</pre>
        </div>
      </q-card-section>

      <q-separator />

      <q-card-actions align="right" class="q-pa-md">
        <q-btn class="q-px-md" v-close-popup flat color="primary" label="关闭" />
        <q-btn class="q-px-md" type="submit" color="primary" label="导出文件" @click="exportSettings" />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { computed, ref } from 'vue';
import { getExportSettings } from 'pages/settings/use-settings';
import { saveText } from 'src/utils/file-download';
import CopyToClipboardBtn from 'components/CopyToClipboardBtn';

const visible = ref(false);
const hideSensitive = ref(false);

const settingsString = computed(() => {
  const settings = getExportSettings(!hideSensitive.value);
  return JSON.stringify(settings, null, 2);
});

const exportSettings = () => {
  saveText('ChineseSubFinderSettings.json', settingsString.value);
};
</script>

<style lang="scss" scoped>
.code-area:hover {
  .copy-btn {
    display: block !important;
  }
}
</style>
