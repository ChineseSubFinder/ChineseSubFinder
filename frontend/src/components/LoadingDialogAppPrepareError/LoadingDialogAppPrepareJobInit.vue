<template>
  <q-dialog ref="dialogRef" class="relative-position" persistent maximized transition-duration="0">
    <q-card class="column full-height overflow-hidden">
      <q-card-section class="bg-primary row items-center text-white">
        <div class="col text-h6">初始化</div>
      </q-card-section>

      <q-card-section>
        <div class="row items-center q-gutter-x-sm">
          <q-spinner v-if="!isDone" color="primary" />
          <q-icon v-else name="check_circle" color="primary" size="20px" />
          <div class="text-h6">{{ prejobStatus.stage_name }}</div>
        </div>
        <div class="text-grey">{{ prejobStatus.now_process_info }}</div>
      </q-card-section>

      <q-card-section v-if="hasError" class="col column justify-center no-wrap">
        <div class="text-bold">错误信息</div>
        <pre class="col bg-grey-3 q-pa-sm q-ma-none overflow-auto">{{ prejobStatus.g_error_info }}</pre>
        <template v-if="prejobStatus?.rename_err_results?.length">
          <div class="text-bold q-mt-md">修改名称失败的字幕文件列表</div>
          <pre class="col bg-grey-3 q-pa-sm q-ma-none overflow-auto">{{
            prejobStatus.rename_err_results.join('\n')
          }}</pre>
        </template>
      </q-card-section>

      <q-card-section class="row justify-center" v-if="showCloseButton">
        <q-btn color="negative" label="关闭" v-close-popup />
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { useDialogPluginComponent } from 'quasar';
import { computed, watch } from 'vue';
import { systemState } from 'src/store/systemState';

const { dialogRef } = useDialogPluginComponent();

const prejobStatus = computed(() => systemState.preJobStatus);
const isDone = computed(() => systemState.preJobStatus.is_done);
const hasError = computed(
  () => !!systemState.preJobStatus.g_error_info || systemState.preJobStatus.rename_err_results?.length > 0
);
const showCloseButton = computed(() => hasError.value);

watch(isDone, (val) => {
  if (val && !showCloseButton.value) {
    dialogRef.value?.hide();
  }
});
</script>
