<template>
  <q-btn dense flat color="primary" label="修改" @click="visible = true"></q-btn>
  <common-form-dialog
    :title="title"
    v-model:visible="visible"
    :submit-fn="handleSubmit"
    @before-show="handleBeforeShow"
  >
    <q-input v-model="form.url" label="RootUrl" outlined :rules="[(val) => !!val || '不能为空']">
      <template v-slot:after>
        <q-btn dense flat label="重置" color="primary" title="重置成默认url" @click="resetUrl" />
      </template>
    </q-input>
    <q-input
      v-if="data.name !== 'csf'"
      v-model.number="form.dailyLimit"
      type="number"
      label="每日下载次数下载"
      placeholder="0为禁用字幕源，-1为不限制次数。最高建议100"
      :rules="[(val) => val !== '' || '不能为空']"
      outlined
    />
  </common-form-dialog>
</template>

<script setup>
import CommonFormDialog from 'components/CommonFormDialog';
import { computed, reactive, ref } from 'vue';
import { DEFAULT_SUB_SOURCE_URL_MAP } from 'src/constants/SettingConstants';

const props = defineProps({
  data: Object,
});

const emit = defineEmits(['update']);

const form = reactive({
  url: '',
  dailyLimit: -1,
});

const visible = ref(false);

const title = computed(() => `修改字幕源：${props.data?.name}`);

const handleSubmit = () => {
  emit('update', form);
  visible.value = false;
};

const handleBeforeShow = () => {
  form.url = props.data?.root_url;
  form.dailyLimit = props.data?.daily_download_limit;
};

const resetUrl = () => {
  form.url = DEFAULT_SUB_SOURCE_URL_MAP[props.data?.name];
};
</script>
