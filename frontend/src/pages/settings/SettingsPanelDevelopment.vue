<template>
  <div>
    <q-list dense style="max-width: 600px">
      <q-item tag="label" v-ripple>
        <q-item-section>
          <q-item-label>接口失效通知</q-item-label>
          <q-item-label caption> 项目维护人员使用，一般人员无需关心 </q-item-label>
          <div v-if="form.enable" class="q-mt-sm">
            <q-input v-model="form.bark_server_address" standout dense />
          </div>
        </q-item-section>
        <q-item-section avatar top>
          <q-toggle v-model="form.enable" />
        </q-item-section>
      </q-item>
    </q-list>

    <q-separator class="q-mt-md" />
  </div>
</template>

<script setup>
import { formModel } from 'pages/settings/use-settings';
import { toRefs } from '@vueuse/core';

const { developer_settings: form } = toRefs(formModel);
</script>
