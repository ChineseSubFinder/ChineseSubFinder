<template>
  <div>
    <q-form @submit="submitAll">
      <q-list style="max-width: 600px" dense>
        <q-item tag="label" v-ripple>
          <q-item-section>
            <q-item-label>自动转换字幕文件编码</q-item-label>
            <q-item-label caption>自动转换到目标编码，如果不是特殊情况，不建议开启，仅对新下载字幕生效</q-item-label>
            <q-item v-if="form.auto_change_sub_encode.enable">
              <q-item-section avatar top>
                <q-radio
                  v-for="(v, k) in DESC_ENCODE_TYPE_NAME_MAP"
                  :key="k"
                  :label="v"
                  v-model="form.auto_change_sub_encode.des_encode_type"
                  :val="~~k"
                />
              </q-item-section>
            </q-item>
          </q-item-section>
          <q-item-section avatar top>
            <q-toggle v-model="form.auto_change_sub_encode.enable" />
          </q-item-section>
        </q-item>
      </q-list>
      <q-separator class="q-mt-md" />
      <form-submit-area />
    </q-form>
  </div>
</template>

<script setup>
import { formModel, submitAll } from 'pages/settings/useSettings';
import { toRefs } from '@vueuse/core';
import FormSubmitArea from 'pages/settings/FormSubmitArea';
import {DESC_ENCODE_TYPE_NAME_MAP} from 'src/constants/SettingConstants';


const { experimental_function: form } = toRefs(formModel);
</script>
