<template>
  <q-form ref="form">
    <q-list dense>
      <q-item>
        <q-item-section>
          <q-item-label>Emby的内网URL</q-item-label>
        </q-item-section>
        <q-item-section>
          <q-input
            v-model="setupState.form.emby.url"
            standout
            dense
            :rules="[
              (val) => !!val || '不能为空',
              (val) => val.match(/^https?:\/\/\w+(\.\w+)*(:[0-9]+)?\/?(\/[.\w]*)*$/) || '请输入正确的URL',
            ]"
          />
        </q-item-section>
      </q-item>
      <q-item>
        <q-item-section>
          <q-item-label>APIKey</q-item-label>
        </q-item-section>
        <q-item-section>
          <q-input v-model="setupState.form.emby.apiKey" standout dense :rules="[(val) => !!val || '不能为空']" />
        </q-item-section>
      </q-item>
      <q-item>
        <q-item-section>
          <q-item-label>获取最多的剧集数量</q-item-label>
        </q-item-section>
        <q-item-section>
          <q-input
            v-model.number="setupState.form.emby.limitCount"
            standout
            dense
            :rules="[(val) => !!val || '不能为空', (val) => /^\d+$/.test(val) || '必须是整数']"
          />
        </q-item-section>
      </q-item>
      <q-item tag="label" v-ripple>
        <q-item-section>
          <q-item-label>是否跳过已观看的</q-item-label>
        </q-item-section>
        <q-item-section>
          <q-toggle v-model="setupState.form.emby.skipWatched" />
        </q-item-section>
      </q-item>

      <q-item :class="{ disabled: setupState.form.emby.autoOrManual }" tag="label" v-ripple>
        <q-item-section>
          <q-item-label>自动匹配IMDB ID</q-item-label>
        </q-item-section>
        <q-item-section>
          <q-toggle v-model="setupState.form.emby.autoOrManual" :disable="setupState.form.emby.autoOrManual" />
        </q-item-section>
      </q-item>
    </q-list>
  </q-form>
</template>

<script setup>
import { setupState } from 'pages/setup/use-setup';
</script>
