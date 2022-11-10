<template>
  <q-form ref="form" class="q-gutter-md">
    <q-input
      filled
      v-model="setupState.form.username"
      label="输入您的用户名"
      hint="用户名必须在3-16个字符以内，只能包含字母、数字、下划线"
      lazy-rules
      :rules="[
        (val) => (val && val.length > 0) || '请输入用户名',
        (val) => /^[_a-zA-Z0-9]{3,16}$/.test(val) || '用户名必须在3-16个字符以内，只能包含字母、数字、下划线',
      ]"
    />

    <q-input
      filled
      v-model="setupState.form.password"
      type="password"
      label="输入您的密码"
      hint="密码必须在6-30位之间"
      lazy-rules
      :rules="[
        (val) => (val && val.length > 0) || '请输入密码',
        (val) =>
          /^([a-z]|[A-Z]|[\d~@#$%\*-\+=:,\\?\[\]\{}]){6,30}$/.test(val) ||
          '密码必须在6-30位之间，且只能由字母、数字、特殊符号组成',
      ]"
    />

    <q-input
      filled
      v-model="setupState.form.confirmPassword"
      type="password"
      label="确认密码"
      hint=""
      lazy-rules
      :rules="[(val) => val === setupState.form.password || '两次输入的密码不一致']"
    />
  </q-form>
</template>

<script setup>
import { setupState } from 'pages/setup/use-setup';
</script>
