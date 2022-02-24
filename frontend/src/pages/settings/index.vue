<template>
  <q-page class="q-pa-md">
    <q-banner inline-actions class="text-white bg-red" v-if="isJobRunning">
      <template v-slot:avatar>
        <q-icon name="warning" />
      </template>
      当前有任务正在运行中，不能更改配置
    </q-banner>
    <q-card v-if="isSettingsLoaded" flat>
      <q-tabs
        v-model="tab"
        dense
        active-color="primary"
        indicator-color="primary"
        align="justify"
        narrow-indicator
        style="max-width: 500px"
      >
        <q-tab name="basic" label="基础配置" />
        <q-tab name="advanced" label="进阶配置" />
        <q-tab name="emby" label="Emby配置" />
        <q-tab name="development" label="开发人员配置" />
        <q-tab name="experiment" label="实验室" />
      </q-tabs>

      <q-separator />

      <q-tab-panels
        v-model="tab"
        animated
        :class="{ disabled: isJobRunning }"
        :style="{ pointerEvents: isJobRunning ? 'none' : 'default' }"
      >
        <q-tab-panel name="basic">
          <basic-settings />
        </q-tab-panel>

        <q-tab-panel name="advanced">
          <advanced-settings />
        </q-tab-panel>

        <q-tab-panel name="emby">
          <emby-settings />
        </q-tab-panel>

        <q-tab-panel name="development">
          <development-settings />
        </q-tab-panel>

        <q-tab-panel name="experiment">
          <experiment-settings />
        </q-tab-panel>
      </q-tab-panels>
    </q-card>
  </q-page>
</template>

<script setup>
import { computed, ref } from 'vue';
import BasicSettings from 'pages/settings/BasicSettings';
import AdvancedSettings from 'pages/settings/AdvancedSettings';
import EmbySettings from 'pages/settings/EmbySettings';
import DevelopmentSettings from 'pages/settings/DevelopmentSettings';
import { settingsState, useSettings } from 'pages/settings/useSettings';
import { isJobRunning } from 'src/store/systemState';
import ExperimentSettings from 'pages/settings/ExperimentSettings';

const tab = ref('basic');

const isSettingsLoaded = computed(() => Object.keys(settingsState.data ?? {}).length);

useSettings();
</script>
