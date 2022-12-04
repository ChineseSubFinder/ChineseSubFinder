<template>
  <div>
    <q-list dense style="max-width: 600px">
      <q-item tag="label" v-ripple>
        <q-item-section>
          <q-item-label>是否开启</q-item-label>
        </q-item-section>
        <q-item-section avatar>
          <q-toggle v-model="form.enable" />
        </q-item-section>
      </q-item>

      <template v-if="form.enable">
        <q-item>
          <q-item-section>
            <q-item-label>Emby的内网URL</q-item-label>
          </q-item-section>
          <q-item-section avatar>
            <q-input
              v-model="form.address_url"
              standout
              dense
              :rules="[
                (val) => (form.enable && !!val) || '不能为空',
                (val) => val.match(/^https?:\/\/\w+(\.\w+)*(:[0-9]+)?\/?(\/[.\w]*)*$/) || '请输入正确的URL',
              ]"
            />
          </q-item-section>
        </q-item>
        <q-item>
          <q-item-section>
            <q-item-label>APIKey</q-item-label>
          </q-item-section>
          <q-item-section avatar>
            <q-input v-model="form.api_key" standout dense :rules="[(val) => (form.enable && !!val) || '不能为空']" />
          </q-item-section>
        </q-item>

        <q-item> <btn-check-emby-server /> </q-item>
        <q-item>
          <q-item-section>
            <q-item-label>获取最多的剧集数量</q-item-label>
          </q-item-section>
          <q-item-section avatar>
            <q-input
              v-model.number="form.max_request_video_number"
              standout
              dense
              :rules="[(val) => (form.enable && !!val) || '不能为空', (val) => /^\d+$/.test(val) || '必须是整数']"
            />
          </q-item-section>
        </q-item>
        <q-item tag="label" v-ripple>
          <q-item-section>
            <q-item-label>是否跳过已观看的</q-item-label>
          </q-item-section>
          <q-item-section avatar>
            <q-toggle v-model="form.skip_watched" />
          </q-item-section>
        </q-item>

        <q-separator spaced inset></q-separator>

        <q-item :class="{ disabled: form.auto_or_manual }" tag="label" v-ripple>
          <q-item-section>
            <q-item-label>自动匹配IMDB ID</q-item-label>
          </q-item-section>
          <q-item-section avatar>
            <q-toggle v-model="form.auto_or_manual" :disable="form.auto_or_manual" />
          </q-item-section>
        </q-item>
      </template>
    </q-list>
  </div>
</template>

<script setup>
import { formModel } from 'pages/settings/use-settings';
import { toRefs } from '@vueuse/core';
import BtnCheckEmbyServer from 'pages/settings/BtnCheckEmbyServer';

const { emby_settings: form } = toRefs(formModel);
</script>
