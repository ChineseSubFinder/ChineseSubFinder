<template>
  <q-page class="q-pa-md">
    <q-form class="q-gutter-md" @submit="handleSubmit" style="width: 500px">
      <q-input
        filled
        v-model="form.oldPwd"
        type="password"
        label="原始密码"
        hint=""
        :rules="[(val) => (val && val.length > 0) || '请输入密码']"
      />

      <q-input
        filled
        v-model="form.newPwd"
        type="password"
        label="新密码"
        hint="密码必须在6-36位之间"
        :rules="[
          (val) => (val && val.length > 0) || '请输入密码',
          (val) =>
            /^([a-z]|[A-Z]|[\d~@#$%!\*-\+=:,\\?\[\]\{}]){6,36}$/.test(val) ||
            '密码必须在6-36位之间，且只能由字母、数字、特殊符号组成',
        ]"
      />

      <q-input
        filled
        v-model="form.confirmPwd"
        type="password"
        label="确认密码"
        hint=""
        :rules="[(val) => val === form.newPwd || '两次输入的密码不一致']"
      />

      <div>
        <q-btn
          unelevated
          size="lg"
          color="primary"
          type="submit"
          :loading="submitting"
          class="full-width text-white"
          label="修改密码"
        />
      </div>
    </q-form>
  </q-page>
</template>

<script setup>
import { reactive, ref } from 'vue';
import ProfileApi from 'src/api/ProfileApi';
import { SystemMessage } from 'src/utils/message';

const submitting = ref(false);
const form = reactive({
  oldPwd: '',
  newPwd: '',
  confirmPwd: '',
});

const handleSubmit = async () => {
  submitting.value = true;
  const [, err] = await ProfileApi.changePwd({
    org_pwd: form.oldPwd,
    new_pwd: form.newPwd,
  });
  if (err !== null) {
    SystemMessage.error(err.message);
  } else {
    SystemMessage.success('修改密码成功');
  }
  submitting.value = false;
};
</script>

<style scoped lang="scss"></style>
