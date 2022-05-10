<template>
  <div>
    <q-list style="max-width: 600px" dense>
      <q-item>
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

      <q-separator spaced inset></q-separator>

      <q-item tag="label" :disable="!isChsChtChangerEnable" v-ripple>
        <q-item-section>
          <q-item-label>简、繁字幕互转功能</q-item-label>
          <q-item-label caption
            >需要开启"自动转换字幕文件编码"功能，并设置为转码"UTF-8"，否则无法启用和生效</q-item-label
          >
          <q-item v-if="form.chs_cht_changer.enable">
            <q-item-section avatar top>
              <q-radio
                :disable="!isChsChtChangerEnable"
                v-for="(v, k) in AUTO_CONVERT_LANG_NAME_MAP"
                :key="k"
                :label="v"
                v-model="form.chs_cht_changer.des_chinese_language_type"
                :val="~~k"
              />
            </q-item-section>
          </q-item>
        </q-item-section>
        <q-item-section avatar top>
          <q-toggle :disable="!isChsChtChangerEnable" v-model="form.chs_cht_changer.enable" />
        </q-item-section>
      </q-item>

      <q-separator spaced inset></q-separator>

      <q-item>
        <q-item-section>
          <q-item-label>远程Chrome</q-item-label>
          <q-item-label caption>
            本功能能够将本程序使用的 Chrome 操作移到一个有算力和资源的硬件上，这样部署本程序的资源要求进一步降低。<br />
            需要自行参看<a
              class="text-primary"
              href="https://go-rod.github.io/i18n/zh-CN/#/custom-launch?id=远程管理启动器"
              target="_blank"
              >https://go-rod.github.io/i18n/zh-CN/#/custom-launch?id=远程管理启动器</a
            >文档部署实验性功能，可用性和稳定性存疑，未必会继续支持更新。除非 go-rod 更新。
          </q-item-label>
        </q-item-section>
        <q-item-section avatar>
          <q-toggle v-model="form.remote_chrome_settings.enable" />
        </q-item-section>
      </q-item>

      <template v-if="form.remote_chrome_settings.enable">
        <q-item>
          <q-item-section>
            <q-item-label>远程 Docker 地址</q-item-label>
          </q-item-section>
          <q-item-section avatar>
            <q-input
              v-model="form.remote_chrome_settings.remote_docker_url"
              placeholder="Ws://192.168.xx.xx:9222"
              standout
              dense
              :rules="[(val) => (form.remote_chrome_settings.enable && !!val) || '不能为空']"
            />
          </q-item-section>
        </q-item>

        <q-item>
          <q-item-section>
            <q-item-label>远程 Docker 中的 ADBlocker 目录</q-item-label>
          </q-item-section>
          <q-item-section avatar>
            <q-input
              v-model="form.remote_chrome_settings.remote_adblock_path"
              placeholder="/mnt/share/adblock1"
              standout
              dense
              :rules="[(val) => (form.remote_chrome_settings.enable && !!val) || '不能为空']"
            />
          </q-item-section>
        </q-item>

        <q-item>
          <q-item-section>
            <q-item-label>远程 Docker 中的缓存文件夹目录</q-item-label>
          </q-item-section>
          <q-item-section avatar>
            <q-input
              v-model="form.remote_chrome_settings.remote_user_data_dir"
              placeholder="/mnt/share/tmp"
              standout
              dense
              :rules="[(val) => (form.remote_chrome_settings.enable && !!val) || '不能为空']"
            />
          </q-item-section>
        </q-item>
      </template>
    </q-list>
  </div>
</template>

<script setup>
import { formModel } from 'pages/settings/useSettings';
import { toRefs } from '@vueuse/core';
import {
  AUTO_CONVERT_LANG_NAME_MAP,
  DESC_ENCODE_TYPE_NAME_MAP,
  DESC_ENCODE_TYPE_UTF8,
} from 'src/constants/SettingConstants';
import { computed } from 'vue';

const { experimental_function: form } = toRefs(formModel);

const isChsChtChangerEnable = computed(
  () =>
    formModel.experimental_function.auto_change_sub_encode?.enable &&
    formModel.experimental_function.auto_change_sub_encode?.des_encode_type === DESC_ENCODE_TYPE_UTF8
);
</script>
