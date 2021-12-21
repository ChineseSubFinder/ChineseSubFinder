<template>
  <div>
    <q-form @submit="submitAll">
      <q-list dense style="max-width: 600px">
        <q-item tag="label" v-ripple>
          <q-item-section>
            <q-item-label>接口失效通知</q-item-label>
            <q-item-label caption>
              项目维护人员使用，一般人员无需关心
            </q-item-label>
            <div v-if="form.enabled" class="q-mt-sm">
              <q-input v-model="form.bark_server_address" standout dense />
            </div>
          </q-item-section>
          <q-item-section avatar top>
            <q-toggle v-model="form.enabled" />
          </q-item-section>
        </q-item>
      </q-list>

      <q-separator class="q-mt-md" />
      <form-submit-area/>
    </q-form>
  </div>
</template>

<script setup>
import FormSubmitArea from 'pages/settings/FormSubmitArea';
import {formModel, submitAll} from 'pages/settings/useSettings';
import {toRefs} from '@vueuse/core';

const { developer_settings: form } = toRefs(formModel);
</script>
