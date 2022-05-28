<template>
  <div>
    <q-list style="max-width: 600px" dense>
      <q-item tag="label">
        <q-item-section>
          <q-item-label>Assrt（https://assrt.net/api/doc）</q-item-label>
          <q-item-label caption>
            <div>注册：https://assrt.net/user/register.xml，用户面板：https://assrt.net/usercp.php</div>
            <ul class="q-pl-md">
              <li>一般用户是 5c次/min 的 API 请求限制</li>
              <li>建议设置完 Token 后，重启程序或者容器！</li>
              <li>搜索字幕效果未知，如果不用就关闭即可</li>
              <li>建议配合“保存多字幕”的选项服用（如果你使用 Emby 的话）</li>
            </ul>
          </q-item-label>
        </q-item-section>
        <q-item-section avatar top>
          <q-toggle v-model="form.assrt_settings.enabled" />
        </q-item-section>
      </q-item>

      <q-item class="q-mt-sm">
        <q-item-section>
          <q-input
            :disable="!form.assrt_settings.enabled"
            v-model="form.assrt_settings.token"
            placeholder="填写你的API Token"
            label="API Token"
            standout
            dense
            :rules="[(val) => !!val || '不能为空']"
          />
        </q-item-section>
      </q-item>
    </q-list>
  </div>
</template>

<script setup>
import { formModel } from 'pages/settings/useSettings';
import { toRefs } from '@vueuse/core';

const { subtitle_sources: form } = toRefs(formModel);
</script>
