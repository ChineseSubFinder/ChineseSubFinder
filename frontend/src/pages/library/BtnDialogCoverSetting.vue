<template>
  <q-btn dense flat color="primary" label="封面规则设置" @click="visible = true"></q-btn>
  <common-form-dialog
    title="封面规则设置"
    v-model:visible="visible"
    :submit-fn="handleSubmit"
    @before-show="handleBeforeShow"
  >
    <q-input v-model="form.coverRule" label="封面图片文件名" outlined :rules="[(val) => !!val || '不能为空']">
      <template v-slot:after>
        <q-btn dense flat label="重置" color="primary" title="重置成默认url" @click="resetCover" />
      </template>
    </q-input>
    <div class="text-grey">同级目录下的封面图片文件名，暂不支持动态设置</div>
  </common-form-dialog>
</template>

<script setup>
import CommonFormDialog from 'components/CommonFormDialog';
import { reactive, ref } from 'vue';
import { LocalStorage } from 'quasar';
import { coverRule } from 'pages/library/use-library';

const emit = defineEmits(['update']);

const form = reactive({
  coverRule: '',
});

const visible = ref(false);

const handleSubmit = () => {
  coverRule.value = form.coverRule;
  LocalStorage.set('coverRule', coverRule.value);
  emit('update', coverRule.value);
  visible.value = false;
};

const handleBeforeShow = () => {
  form.coverRule = coverRule.value;
};

const resetCover = () => {
  form.coverRule = 'poster.jpg';
};
</script>
