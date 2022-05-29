<template>
  <q-dialog v-model="visible" persistent>
    <q-card style="width: 800px; max-width: 800px">
      <q-card-section>
        <div class="text-h6 text-grey-8">新增功能介绍</div>
      </q-card-section>

      <q-separator />

      <q-card-section>
        <ul>
          <li>新增共享字幕功能，可以在 <span class="text-bold">配置中心-实验室</span> 开启</li>
        </ul>
      </q-card-section>

      <q-separator />

      <q-card-actions align="right" class="q-pa-md">
        <q-btn class="q-px-md" flat color="primary" label="我知道了" @click="handleAgree" />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<script setup>
import {computed, onMounted, ref} from 'vue';
import {LocalStorage} from 'quasar';
import {until} from '@vueuse/core';
import {systemState} from 'src/store/systemState';

const visible = ref(false);

const currentVersion = computed(() => systemState.systemInfo?.version);

const noticeFlagItemKey = computed(() => `noticeFlag-${currentVersion.value}`);

const handleAgree = () => {
  visible.value = false;
  LocalStorage.set(noticeFlagItemKey.value, true);
}

onMounted(async () => {
  await until(() => currentVersion.value !== undefined).toBe(true);
  const noticeFlag = LocalStorage.getItem(noticeFlagItemKey.value);
  if (!noticeFlag) {
    visible.value = true;
  }
})
</script>
