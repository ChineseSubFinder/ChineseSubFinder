<template>
  <span v-if="hasNewVersion" @click="visible = true">
    <slot v-if="$slots.default"></slot>
    <q-badge v-else class="cursor-pointer" label="new" title="有新的版本更新" />
  </span>
  <q-dialog v-if="latestVersion" v-model="visible">
    <q-card class="column" style="width: 600px; min-height: 400px">
      <q-card-section>
        <div class="text-h5">{{ latestVersion.tag_name }}</div>
      </q-card-section>

      <q-tabs
        v-model="tab"
        dense
        class="text-grey"
        active-color="primary"
        indicator-color="primary"
        align="justify"
        narrow-indicator
      >
        <q-tab name="log" label="更新日志" />
        <q-tab name="update" label="升级方式" />
      </q-tabs>

      <q-separator />

      <q-tab-panels class="col" v-model="tab" animated>
        <q-tab-panel name="log">
          <markdown :source="latestVersion.body" />
        </q-tab-panel>
        <q-tab-panel name="update">
          <section>
            <div class="text-h6">Windows</div>
            <div>
              下载最新版本替换，
              <a :href="latestVersion.html_url" target="_blank"> 下载地址 </a>
            </div>
          </section>

          <section>
            <div class="text-h6">Docker</div>
            <div>
              参考教程
              <!-- eslint-disable-next-line max-len -->
              <a
                href="https://github.com/ChineseSubFinder/ChineseSubFinder/blob/master/docker/readme.md"
                target="_blank"
              >
                Docker部署教程
              </a>
            </div>
            <div class="text-grey">
              * 新版本发布到Docker发布完成可能需要一小时左右，如果发现Docker拉取的版本没有变化，请耐心等待一段时间
            </div>
          </section>
        </q-tab-panel>
      </q-tab-panels>

      <q-separator />

      <q-card-actions align="right">
        <q-btn color="primary" @click="navigateToReleasePage"> 前往更新 </q-btn>
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue';
import Markdown from 'components/Markdown';
import { systemState } from 'src/store/systemState';
import { LocalStorage } from 'quasar';

const latestVersion = ref(LocalStorage.getItem('latestVersion') ?? null);
const visible = ref(false);
const tab = ref('log');

const hasNewVersion = computed(() => {
  const v = systemState.systemInfo?.version.replace(/\s+(L|l)ite$/, '');
  return latestVersion.value?.tag_name && v && latestVersion.value.tag_name !== v;
});

const getLatestVersion = async () => {
  try {
    const data = await fetch('https://api.github.com/repos/ChineseSubFinder/ChineseSubFinder/releases/latest').then(
      (res) => {
        if (res.ok) {
          return res.json();
        }
        return Promise.reject(res);
      }
    );
    latestVersion.value = data;
    // 接口请求速率过高有可能403，本地存一份
    LocalStorage.set('latestVersion', data);
  } catch (e) {
    // do nothing
  }
};

const navigateToReleasePage = () => {
  window.open(latestVersion.value.html_url);
  visible.value = false;
};

onMounted(getLatestVersion);
</script>

<style lang="scss" scoped>
a {
  color: $primary;
}
</style>
