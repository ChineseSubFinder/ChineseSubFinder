<template>
  <q-page class="q-pa-md">
    <q-banner inline-actions class="text-white bg-red" v-if="isJobRunning">
      <template v-slot:avatar>
        <q-icon name="warning" />
      </template>
      任务进程运行中，不能更改配置
      <template v-slot:action>
        <q-btn color="white" label="去总览页面停止" flat @click="$router.push('/overview')" />
      </template>
      <span> </span>
    </q-banner>
    <q-card v-if="isSettingsLoaded" flat>
      <q-tabs
        v-model="tab"
        dense
        active-color="primary"
        indicator-color="primary"
        align="justify"
        narrow-indicator
        style="max-width: 700px"
      >
        <q-tab name="basic" label="基础配置" />
        <q-tab name="advanced" label="进阶配置" />
        <q-tab name="subSource" label="字幕源设置" />
        <q-tab name="emby" label="Emby配置" />
        <q-tab name="development" label="开发人员配置" />
        <q-tab name="experiment" label="实验室" />
      </q-tabs>

      <q-separator />

      <q-form @submit="submitAll">
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

          <q-tab-panel name="subSource">
            <sub-source-settings />
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

        <q-separator />

        <form-submit-area />
      </q-form>
    </q-card>
  </q-page>
</template>

<script setup>
import { computed, ref } from 'vue';
import BasicSettings from 'pages/settings/SettingsPanelBasic';
import AdvancedSettings from 'pages/settings/SettingsPanelAdvanced';
import EmbySettings from 'pages/settings/SettingsPanelEmby';
import DevelopmentSettings from 'pages/settings/SettingsPanelDevelopment';
import { formModel, submitAll, useSettings } from 'pages/settings/use-settings';
import { isJobRunning } from 'src/store/systemState';
import ExperimentSettings from 'pages/settings/SettingsPanelExperiment';
import FormSubmitArea from 'pages/settings/FormSubmitArea';
import SubSourceSettings from 'pages/settings/SettingsPanelSubSource';

const tab = ref('basic');

const isSettingsLoaded = computed(() => Object.keys(formModel).length);

useSettings();
</script>
