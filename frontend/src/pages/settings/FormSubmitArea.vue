<template>
  <div class="q-pa-md q-gutter-md">
    <q-btn label="保存" color="primary" type="submit" :loading="submitting"></q-btn>
    <q-btn label="重置" @click="resetForm"></q-btn>
  </div>
</template>

<script setup>
import {resetForm, submitting} from 'pages/settings/useSettings';
</script>
