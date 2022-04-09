<template>
  <div>
    <q-list dense style="max-width: 600px">
      <q-item tag="label" v-ripple>
        <q-item-section>
          <q-item-label>是否使用代理</q-item-label>
          <q-item-label caption>支持HTTP代理</q-item-label>
        </q-item-section>
        <q-item-section avatar top>
          <q-toggle v-model="form.proxy_settings.use_http_proxy" />
        </q-item-section>
      </q-item>

      <q-item v-if="form.proxy_settings.use_http_proxy" class="q-mt-md" dense>
        <q-item-section>
          <q-input
            v-model="form.proxy_settings.http_proxy_address"
            standout
            dense
            label="代理服务器地址"
            style="width: 400px"
          >
            <template v-slot:after>
              <div style="width: 80px">
                <proxy-check-btn :url="form.proxy_settings.http_proxy_address" />
              </div>
            </template>
          </q-input>
        </q-item-section>
      </q-item>

      <q-separator spaced inset></q-separator>

      <q-item tag="label" v-ripple>
        <q-item-section>
          <q-item-label>调试模式</q-item-label>
        </q-item-section>
        <q-item-section avatar>
          <q-toggle v-model="form.debug_mode" />
        </q-item-section>
      </q-item>

      <q-separator spaced inset></q-separator>

      <q-item tag="label" v-ripple>
        <q-item-section>
          <q-item-label>保存整季的缓存字幕</q-item-label>
        </q-item-section>
        <q-item-section avatar>
          <q-toggle v-model="form.save_full_season_tmp_subtitles" />
        </q-item-section>
      </q-item>

      <q-separator spaced inset></q-separator>

      <q-item>
        <q-item-section>
          <q-item-label>字幕格式下载优先级</q-item-label>
        </q-item-section>
        <q-item-section avatar>
          <div class="row">
            <q-radio
              v-for="(v, k) in SUB_TYPE_PRIORITY_NAME_MAP"
              :key="k"
              v-model="form.sub_type_priority"
              :val="~~k"
              :label="v"
            />
          </div>
        </q-item-section>
      </q-item>

      <q-separator spaced inset></q-separator>

      <q-item>
        <q-item-section>
          <q-item-label>字幕保存的命名格式</q-item-label>
          <q-item v-for="(v, k) in SUB_NAME_FORMAT_NAME_MAP" :key="k" tag="label" v-ripple>
            <q-item-section avatar top>
              <q-radio v-model="form.sub_name_formatter" :val="~~k" />
            </q-item-section>
            <q-item-section>
              <q-item-label>{{ v }}</q-item-label>
              <q-item-label caption>
                {{ subNameFormatDescMap[k] }}
              </q-item-label>
            </q-item-section>
          </q-item>
        </q-item-section>
      </q-item>

      <q-item v-if="SUB_NAME_FORMAT_EMBY === form.sub_name_formatter" tag="label" v-ripple>
        <q-item-section>
          <q-item-label>保存多字幕</q-item-label>
          <q-item-label caption>每个视频下面保存每个网站找到的最佳字幕，需要选择Emby格式</q-item-label>
        </q-item-section>
        <q-item-section avatar>
          <q-toggle v-model="form.save_multi_sub" />
        </q-item-section>
      </q-item>

      <q-separator spaced inset></q-separator>

      <q-item>
        <q-item-section>
          <q-item-label class="q-mb-sm">字幕源设置</q-item-label>
          <q-item v-for="item in ['xunlei', 'shooter', 'subhd', 'zimuku']" :key="item" clickable>
            <q-item-section avatar class="text-bold" style="width: 120px">
              {{ form.suppliers_settings[item].name }}
            </q-item-section>
            <q-item-section class="text-grey-8">
              <q-item-label :lines="1">
                {{ form.suppliers_settings[item].root_url }}
              </q-item-label>
              <q-item-label style="font-size: 90%">
                每日下载次数限制：{{ form.suppliers_settings[item].daily_download_limit }}
              </q-item-label>
            </q-item-section>
            <q-item-section side>
              <edit-sub-source-btn-dialog
                :data="form.suppliers_settings[item]"
                @update="(data) => handleSubSourceUpdate(item, data)"
              />
            </q-item-section>
          </q-item>
        </q-item-section>
      </q-item>

      <q-separator spaced inset />

      <q-item>
        <q-item-section class="items-start" top>
          <q-item-label>自定义视频扩展名</q-item-label>
          <q-item-label caption>原生支持mp4、mkv、rmvb、iso、m2ts</q-item-label>
          <template v-for="(item, i) in form.custom_video_exts" :key="i">
            <div class="row items-center q-gutter-x-md" :class="{ 'q-mt-md': i === 0 }">
              <q-input
                v-model="form.custom_video_exts[i]"
                placeholder=""
                standout
                dense
                :rules="[(val) => !!val || '不能为空']"
              />
              <q-btn
                icon="remove"
                color="negative"
                dense
                rounded
                size="xs"
                title="删除"
                @click="form.custom_video_exts.splice(i, 1)"
              ></q-btn>
            </div>
          </template>
        </q-item-section>
        <q-item-section side top>
          <q-btn
            icon="add"
            color="primary"
            dense
            rounded
            size="xs"
            title="新增"
            @click="form.custom_video_exts.push('')"
          ></q-btn>
        </q-item-section>
      </q-item>

      <q-separator spaced inset></q-separator>

      <q-item tag="label" v-ripple>
        <q-item-section>
          <q-item-label>自动校正字幕时间轴</q-item-label>
        </q-item-section>
        <q-item-section avatar>
          <q-toggle v-model="form.fix_time_line" />
        </q-item-section>
      </q-item>
    </q-list>
    <q-separator class="q-mt-md" />
  </div>
</template>

<script setup>
import {
  SUB_NAME_FORMAT_EMBY,
  SUB_NAME_FORMAT_NORMAL,
  SUB_NAME_FORMAT_NAME_MAP,
  SUB_TYPE_PRIORITY_NAME_MAP,
} from 'src/constants/SettingConstants';
import { formModel } from 'pages/settings/useSettings';
import { toRefs } from '@vueuse/core';
import ProxyCheckBtn from 'components/ProxyCheckBtn';
import EditSubSourceBtnDialog from 'pages/settings/EditSubSourceBtnDialog';

const subNameFormatDescMap = {
  [SUB_NAME_FORMAT_NORMAL]: '兼容性更好，AAA.zh.ass or AAA.zh.default.ass。',
  [SUB_NAME_FORMAT_EMBY]: 'AAA.chinese(简英,subhd).ass or AAA.chinese(简英,xunlei).default.ass。',
};

const { advanced_settings: form } = toRefs(formModel);

const handleSubSourceUpdate = (item, data) => {
  formModel.advanced_settings.suppliers_settings[item].root_url = data.url;
  formModel.advanced_settings.suppliers_settings[item].daily_download_limit = data.dailyLimit;
};
</script>
