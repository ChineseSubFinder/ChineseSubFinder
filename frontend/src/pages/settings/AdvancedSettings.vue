<template>
  <div>
    <q-list dense style="max-width: 600px">
      <q-item tag="label" v-ripple>
        <q-item-section>
          <q-item-label>是否使用代理</q-item-label>
          <q-item-label caption>支持HTTP代理</q-item-label>
        </q-item-section>
        <q-item-section avatar top>
          <q-toggle v-model="form.proxy_settings.use_proxy" />
        </q-item-section>
      </q-item>

      <q-item v-if="form.proxy_settings.use_proxy" class="q-mt-md" dense>
        <q-item-section>
          <div class="row q-gutter-sm no-wrap">
            <q-select
              v-model="form.proxy_settings.use_which_proxy_protocol"
              :options="Object.keys(PROXY_TYPE_NAME_MAP).map((e) => ({ label: PROXY_TYPE_NAME_MAP[e], value: e }))"
              label="协议"
              standout
              dense
              emit-value
              map-options
              style="width: 100px"
            />
            <q-input v-model="form.proxy_settings.input_proxy_address" standout dense label="代理服务器" />
            <q-input v-model="form.proxy_settings.input_proxy_port" standout dense label="代理端口" />
            <q-input v-model="form.proxy_settings.local_http_proxy_server_port" standout dense label="本地端口" />
          </div>

          <div class="q-mt-sm row q-gutter-sm">
            <q-checkbox v-model="form.proxy_settings.need_pwd" left-label label="账号认证" />
            <q-input
              :disable="!form.proxy_settings.need_pwd"
              v-model="form.proxy_settings.input_proxy_username"
              standout
              dense
              label="账号"
            />
            <q-input
              :disable="!form.proxy_settings.need_pwd"
              v-model="form.proxy_settings.input_proxy_password"
              standout
              dense
              label="密码"
            />
          </div>

          <div class="q-mt-sm">
            <proxy-check-btn
              :settings="form.proxy_settings"
              label="测试代理服务"
              size="md"
              icon="bolt"
              color="primary"
            />
          </div>
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

      <q-separator spaced inset></q-separator>

      <q-item tag="label" v-ripple>
        <q-item-section>
          <q-item-label>跳过中文电影</q-item-label>
        </q-item-section>
        <q-item-section avatar>
          <q-toggle v-model="form.scan_logic.skip_chinese_movie" />
        </q-item-section>
      </q-item>

      <q-item tag="label" v-ripple>
        <q-item-section>
          <q-item-label>跳过中文连续剧</q-item-label>
        </q-item-section>
        <q-item-section avatar>
          <q-toggle v-model="form.scan_logic.skip_chinese_series" />
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
          <q-item v-for="item in form.suppliers_settings" :key="item" clickable>
            <q-item-section avatar class="text-bold" style="width: 120px">
              {{ item.name }}
            </q-item-section>
            <q-item-section class="text-grey-8">
              <q-item-label :lines="1">
                {{ item.root_url }}
              </q-item-label>
              <q-item-label style="font-size: 90%">
                每日下载次数限制：{{ item.daily_download_limit }}
              </q-item-label>
            </q-item-section>
            <q-item-section side>
              <edit-sub-source-btn-dialog
                :data="item"
                @update="(data) => handleSubSourceUpdate(item, data)"
              />
            </q-item-section>
          </q-item>
        </q-item-section>
      </q-item>

      <q-separator spaced inset />

      <q-item>
        <q-item-section>
          <q-item-label class="q-mb-sm">队列设置</q-item-label>
          <q-input
            class="col"
            v-model.number="form.task_queue.max_retry_times"
            label="最大重试次数"
            shadow-text="单个任务失败后，最大重试次数，超过后会降一级"
            standout
            dense
            :rules="[(val) => !!val || '不能为空']"
          />
          <q-input
            class="col"
            v-model.number="form.task_queue.one_job_time_out"
            label="任务的超时时间"
            standout
            dense
            suffix="秒"
            :rules="[(val) => !!val || '不能为空']"
          />
          <q-input
            class="col"
            v-model.number="form.task_queue.interval"
            label="下载任务之间的间隔时间"
            shadow-text="防止频率太高触发防爬检测"
            standout
            dense
            suffix="秒"
            :rules="[(val) => !!val || '不能为空']"
          />
          <q-input
            class="col"
            v-model.number="form.task_queue.expiration_time"
            label="下载时效（天）"
            shadow-text="视频创建时间在此时间内，才下载，否则标记为失败"
            standout
            dense
            suffix="天"
            :rules="[(val) => !!val || '不能为空']"
          />
          <q-input
            class="col"
            v-model.number="form.task_queue.download_sub_during_x_days"
            label="有内置字幕的视频下载时效"
            shadow-text="如果创建了 x 天，且有内置的中文字幕，那么也不进行下载了"
            standout
            dense
            suffix="天"
            :rules="[(val) => !!val || '不能为空']"
          />
          <q-input
            class="col"
            v-model.number="form.task_queue.one_sub_download_interval"
            label="单个任务失败后，重新下载的最小间隔（小时）"
            standout
            dense
            suffix="小时"
            :rules="[(val) => !!val || '不能为空']"
          />
          <q-input
            class="col"
            v-model="form.task_queue.check_pulic_ip_target_site"
            label="检查公网IP的目标网站"
            shadow-text="目标网站必须直接返回ip字符串，不需要额外解析。多个站点用 ;（英文分号） 分割"
            standout
            dense
          />
          <div class="text-warning">
            * 默认内置几个检查ip的网站，默认站点失效后才需要手动设置。内置站点列表：
            https://myip.biturl.top/;https://ip4.seeip.org/;https://ipecho.net/plain;https://api-ipv4.ip.sb/ip;
            https://api.ipify.org/;http://myexternalip.com/raw
          </div>
        </q-item-section>
      </q-item>

      <q-separator spaced inset />

      <q-item>
        <q-item-section>
          <q-item-label>下载缓存过期时间设置</q-item-label>
        </q-item-section>
        <q-item-section avatar>
          <div class="row no-wrap q-gutter-xs">
            <q-input class="col" standout dense v-model.number="form.download_file_cache.ttl"> </q-input>
            <q-select
              standout
              dense
              :options="[
                { label: '小时', value: 'hour' },
                { label: '秒', value: 'second' },
              ]"
              emit-value
              map-options
              v-model.number="form.download_file_cache.unit"
            ></q-select>
          </div>
        </q-item-section>
      </q-item>

      <q-separator spaced inset />

      <q-item>
        <q-item-section class="items-start" top>
          <q-item-label>自定义视频扩展名</q-item-label>
          <q-item-label caption>原生支持mp4、mkv、rmvb、iso</q-item-label>
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
  </div>
</template>

<script setup>
import {
  SUB_NAME_FORMAT_EMBY,
  SUB_NAME_FORMAT_NORMAL,
  SUB_NAME_FORMAT_NAME_MAP,
  SUB_TYPE_PRIORITY_NAME_MAP,
  PROXY_TYPE_NAME_MAP,
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
  formModel.advanced_settings.suppliers_settings[item.name].root_url = data.url;
  formModel.advanced_settings.suppliers_settings[item.name].daily_download_limit = data.dailyLimit;
};
</script>
