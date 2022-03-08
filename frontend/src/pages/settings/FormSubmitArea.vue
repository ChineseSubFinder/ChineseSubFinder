<template>
  <div class="q-pa-md q-gutter-md">
    <q-btn :disable="isJobRunning" label="保存" color="primary" type="submit" :loading="submitting"></q-btn>
    <q-btn :disable="isJobRunning" label="重置" @click="resetForm"></q-btn>
    <export-setting-btn-dialog />
  </div>
</template>

<script setup>
import { resetForm, submitting } from 'pages/settings/useSettings';
import ExportSettingBtnDialog from 'pages/settings/exportSettingBtnDialog';
import { isJobRunning } from 'src/store/systemState';
</script>
