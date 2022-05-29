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

      <q-item tag="label" v-ripple>
        <q-item-section>
          <q-item-label>共享字幕</q-item-label>
          <q-item-label caption
          >上传字幕能够让本程序变得更加好用，发扬共享精神，一起来共享字幕吧！</q-item-label
          >
        </q-item-section>
        <q-item-section avatar top>
          <q-toggle v-model="form.share_sub_settings.share_sub_enabled" />
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
              placeholder="ws://192.168.xx.xx:9222"
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

      <q-separator spaced inset />

      <q-item>
        <q-item-section>
          <q-item-label>本地Chrome</q-item-label>
          <q-item-label caption>
            如果本程序能够自动下载 Chrome 就不建议自己制定 Chrome 版本，因为如果本程序更新了， Chrome
            也会自动下载最新的，但是你指定的，我是没法更新的，有问题也只有你自己去手动更新 Chrome。建
            议还是优先还是解决网络问题去下载 Chrome。下载 Chrome 是又 go-rod 进行的，有问题也只能去提
            issues。注意以下几点：
            <div>
              <ol>
                <li>
                  如果是 Docker 用户，推荐映射你解压后的 /volume1/docker/chinesesubfinder/Chrome 文件 夹 到
                  /app/cache/Plugin/Chrome 文件夹中，那么你需要填写的 Chrome 容器内的完整路径应 该是（举例，按自己下载的
                  Chrome 来改）: /app/cache/Plugin/Chrome/chrome
                </li>
                <li>如果是 Windows 用户，那么就是你 Chrome.exe 的完整路径</li>
                <li>Chrome 版本不要太低</li>
                <li>请确认指定的chrome和对应平台、CPU架构一致</li>
              </ol>
            </div>
          </q-item-label>
        </q-item-section>
        <q-item-section avatar top>
          <q-toggle v-model="form.local_chrome_settings.enabled" />
        </q-item-section>
      </q-item>

      <template v-if="form.local_chrome_settings.enabled">
        <q-item>
          <q-item-section>
            <q-input
              v-model="form.local_chrome_settings.local_chrome_exe_f_path"
              label="Chrome(.exe) 的完整路径"
              placeholder="/your/chrome/path/chrome.exe"
              standout
              dense
              :rules="[(val) => !!val || '不能为空']"
            />
          </q-item-section>
        </q-item>
      </template>

      <q-separator spaced inset />

      <q-item>
        <q-item-section>
          <q-item-label>API key</q-item-label>
          <q-item-label caption>
            本程序提供一些接口给开发者使用，通过API key鉴权，具体参见
            <!-- eslint-disable -->
            <a
              href="https://github.com/allanpk716/ChineseSubFinder/blob/docs/DesignFile/ApiKey%E8%AE%BE%E8%AE%A1/ApiKey%E8%AE%BE%E8%AE%A1.md"
              class="text-primary"
              target="_blank"
              >开发文档</a
            >
          </q-item-label>
        </q-item-section>
        <q-item-section avatar top>
          <q-toggle v-model="form.api_key_settings.enabled" />
        </q-item-section>
      </q-item>

      <template v-if="form.api_key_settings.enabled">
        <q-item>
          <q-btn label="重新生成密钥" color="primary" size="sm" @click="generateApiKey" />
        </q-item>
        <q-item class="q-mt-sm">
          <q-item-section>
            <q-input
              v-model="form.api_key_settings.key"
              standout
              dense
              :rules="[(val) => !!val || '不能为空']"
              readonly
            >
              <template #append>
                <copy-to-clipboard-btn v-if="form.api_key_settings.key" :text="form.api_key_settings.key" size="sm" />
              </template>
            </q-input>
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
import {computed, watch} from 'vue';
import CopyToClipboardBtn from 'components/CopyToClipboardBtn';
import {useQuasar} from 'quasar';

const $q = useQuasar();

const { experimental_function: form } = toRefs(formModel);

const isChsChtChangerEnable = computed(
  () =>
    formModel.experimental_function.auto_change_sub_encode?.enable &&
    formModel.experimental_function.auto_change_sub_encode?.des_encode_type === DESC_ENCODE_TYPE_UTF8
);

const generateUuid = () =>
  'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    // eslint-disable-next-line no-bitwise
    const r = (Math.random() * 16) | 0;
    // eslint-disable-next-line no-bitwise
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });

const generateApiKey = () => {
  const uuid = generateUuid();
  formModel.experimental_function.api_key_settings.key = uuid;
};

watch(() => formModel.experimental_function.share_sub_settings.share_sub_enabled, (val) => {
  if (val) {
    $q.dialog({
      title: '共享字幕说明',
      style: 'width: 600px',
      message: `<b>开启“共享字幕”功能后：</b>
<ul>
  <li>本程序会收集、上传 Emby 中已经观看的视频对应的字幕（如果你没有开启 Emby 那么就不会收集这个部分的字幕）
  <li>如果你使用了本程序提供的 Http API 提交了已经观看的视频和字幕信息，本程序也会收集这个部分
  <li>如果上述两点你都没有符合条件，那么“共享字幕”功能暂时是不会收集你本地的其他字幕的（因为没有视频对应关系，收集的意义不大）
  <li>如果有任何疑问欢迎去看本程序的上传字幕部分的代码
</ul>

<b>字幕的去向、用途：</b>
<ul>
  <li>字幕上传后，评估通过的字幕会存储在共享服务器中
  <li>后续本程序自身提供的字幕搜索会有两种形式：
    <ul style="padding-left: 20px;">
      <li>a. 类似 xunlei、shooter 的接口，通过计算视频文件唯一 ID （这个ID的算法在本程序内，暂时没有整理出来）去查询
      <li>b. 支持 IMDB、TMDB ID （注意，这个ID是电影或者是连续剧的 ID，不是一集的 ID），加上 SxxExx 这样的信息去查询
    </ul>
   PS： 暂时不会支持关键词查询，除非后续有特殊情况出现
  </li>
</ul>
<div style="color: red;">* 开启共享后，需要重启本程序或者docker容器才能生效</div>
`,
      persistent: true,
      html: true,
      ok: '共享',
      cancel: '不共享',
    }).onOk(() => {
      // 共享字幕
    }).onCancel(() => {
      formModel.experimental_function.share_sub_settings.share_sub_enabled = false;
    })
  }
})
</script>
