<template>
  <div>
    <q-list style="max-width: 600px" dense>
      <q-item>
        <q-item-section>
          <q-item-label>字幕扫描时机</q-item-label>
          <q-item>
            <q-item-section avatar top>
              <q-radio v-model="form.interval_or_assign_or_custom" :val="0" />
            </q-item-section>
            <q-item-section>
              <q-item-label>扫描的间隔</q-item-label>
              <q-item-label caption> 间隔小时数 </q-item-label>
            </q-item-section>
            <q-item-section side>
              <q-select
                v-model="scanCronString0"
                :options="scanIntervalOptions"
                standout
                dense
                style="width: 200px"
                :rules="[(val) => !!val || '不能为空']"
                emit-value
                map-options
                :disable="form.interval_or_assign_or_custom !== 0"
                @update:model-value="handleScanIntervalChange"
              />
            </q-item-section>
          </q-item>

          <q-item>
            <q-item-section avatar top>
              <q-radio v-model="form.interval_or_assign_or_custom" :val="1" />
            </q-item-section>
            <q-item-section>
              <q-item-label>指定扫描时间</q-item-label>
              <q-item-label caption> 选择每天固定时间点 </q-item-label>
            </q-item-section>
            <q-item-section side>
              <q-select
                v-model="scanCronString1"
                :options="scanSpecTimeOptions"
                standout
                dense
                emit-value
                map-options
                style="width: 200px"
                :rules="[
                  (val) => !!val || !!val?.length || '不能为空',
                  (val) => val.length <= 4 || '最多选择4个时间点',
                ]"
                :disable="form.interval_or_assign_or_custom !== 1"
                @update:model-value="handleScanSpecTimeChange"
                multiple
              />
            </q-item-section>
          </q-item>

          <q-item>
            <q-item-section avatar top>
              <q-radio v-model="form.interval_or_assign_or_custom" :val="2" />
            </q-item-section>
            <q-item-section>
              <q-item-label>自定义规则</q-item-label>
              <q-item-label caption>
                详细规则参考
                <a href="https://pkg.go.dev/github.com/robfig/cron/v3" target="_blank" class="text-primary"
                  >robfig/cron 文档</a
                >
              </q-item-label>
            </q-item-section>
            <q-item-section side>
              <q-input
                v-model="scanCronString2"
                standout
                dense
                style="width: 200px"
                :rules="[(val) => !!val || '不能为空', validateCronTime]"
                @update:model-value="handleScanCustomChange"
                :disable="form.interval_or_assign_or_custom !== 2"
              />
            </q-item-section>
          </q-item>
        </q-item-section>
      </q-item>

      <q-separator spaced inset></q-separator>

      <q-item>
        <q-item-section>
          <q-item-label>设备性能选择</q-item-label>
        </q-item-section>
        <q-item-section avatar>
          <div class="row">
            <q-radio v-model="form.threads" :val="1" label="弱鸡（1线程）" />
            <q-radio v-model="form.threads" :val="3" label="一般（3线程）" />
            <q-radio v-model="form.threads" :val="6" label="超猛（6线程）" />
          </div>
        </q-item-section>
      </q-item>

      <q-separator spaced inset></q-separator>

      <!--        <q-item>-->
      <!--          <q-item-section>-->
      <!--            <q-item-label>并发数</q-item-label>-->
      <!--          </q-item-section>-->
      <!--          <q-item-section avatar>-->
      <!--            <q-input-->
      <!--              v-model.number="form.threads"-->
      <!--              standout-->
      <!--              dense-->
      <!--              :rules="[(val) => !!val || '不能为空', (val) => /^\d+$/.test(val) || '必须是整数']"-->
      <!--            />-->
      <!--          </q-item-section>-->
      <!--        </q-item>-->

      <!--        <q-separator spaced inset></q-separator>-->

      <q-item tag="label" v-ripple>
        <q-item-section>
          <q-item-label>程序启动立即开启扫描</q-item-label>
        </q-item-section>
        <q-item-section avatar>
          <q-toggle v-model="form.run_scan_at_start_up" />
        </q-item-section>
      </q-item>

      <q-separator spaced inset></q-separator>

      <q-item>
        <q-item-section class="items-start" top>
          <q-item-label>电影的目录</q-item-label>
        </q-item-section>
        <q-item-section avatar>
          <template v-for="(item, i) in form.movie_paths" :key="i">
            <div class="row items-center q-gutter-x-md">
              <q-input
                v-model="form.movie_paths[i]"
                placeholder="/media/电影"
                standout
                dense
                lazy-rules
                :rules="[(val) => !!val || '不能为空', validateRemotePath]"
                style="width: 200px"
              />
              <q-btn
                v-if="i === 0"
                icon="add"
                color="primary"
                dense
                rounded
                size="xs"
                title="新增"
                @click="form.movie_paths.push('')"
              ></q-btn>
              <q-btn
                v-else
                icon="remove"
                color="negative"
                dense
                rounded
                size="xs"
                title="删除"
                @click="form.movie_paths.splice(i, 1)"
              ></q-btn>
            </div>
          </template>
        </q-item-section>
      </q-item>

      <q-separator spaced inset></q-separator>

      <q-item>
        <q-item-section class="items-start" top>
          <q-item-label>连续剧的目录</q-item-label>
        </q-item-section>
        <q-item-section avatar>
          <template v-for="(item, i) in form.series_paths" :key="i">
            <div class="row items-center q-gutter-md">
              <q-input
                v-model="form.series_paths[i]"
                placeholder="/media/连续剧"
                standout
                dense
                :rules="[(val) => !!val || '不能为空', validateRemotePath]"
                style="width: 200px"
              />
              <q-btn
                v-if="i === 0"
                icon="add"
                color="primary"
                dense
                rounded
                size="xs"
                title="新增"
                @click="form.series_paths.push('')"
              ></q-btn>
              <q-btn
                v-else
                icon="remove"
                color="negative"
                dense
                rounded
                size="xs"
                title="删除"
                @click="form.series_paths.splice(i, 1)"
              ></q-btn>
            </div>
          </template>
        </q-item-section>
      </q-item>
    </q-list>
  </div>
</template>

<script setup>
import { formModel } from 'pages/settings/useSettings';
import { validateCronTime, validateRemotePath } from 'src/utils/QuasarValidators';
import { toRefs } from '@vueuse/core';
import { ref, watch } from 'vue';

const { common_settings: form } = toRefs(formModel);

const scanCronString0 = ref('');
const scanCronString1 = ref([]);
const scanCronString2 = ref('');

if (form.value.interval_or_assign_or_custom === 0) {
  scanCronString0.value = form.value.scan_interval.split(' ').pop();
} else if (form.value.interval_or_assign_or_custom === 1) {
  scanCronString1.value = form.value.scan_interval.split(' ')[1].split(',');
} else if (form.value.interval_or_assign_or_custom === 2) {
  scanCronString2.value = form.value.scan_interval;
}

const scanIntervalOptions = [
  { label: '每4小时', value: '4h' },
  { label: '每5小时', value: '5h' },
  { label: '每6小时', value: '6h' },
  { label: '每7小时', value: '7h' },
  { label: '每8小时', value: '8h' },
  { label: '每9小时', value: '9h' },
  { label: '每10小时', value: '10h' },
];
const scanSpecTimeOptions = [
  { label: '00:00', value: '0' },
  { label: '01:00', value: '1' },
  { label: '02:00', value: '2' },
  { label: '03:00', value: '3' },
  { label: '04:00', value: '4' },
  { label: '05:00', value: '5' },
  { label: '06:00', value: '6' },
  { label: '07:00', value: '7' },
  { label: '08:00', value: '8' },
  { label: '09:00', value: '9' },
  { label: '10:00', value: '10' },
  { label: '11:00', value: '11' },
  { label: '12:00', value: '12' },
  { label: '13:00', value: '13' },
  { label: '14:00', value: '14' },
  { label: '15:00', value: '15' },
  { label: '16:00', value: '16' },
  { label: '17:00', value: '17' },
  { label: '18:00', value: '18' },
  { label: '19:00', value: '19' },
  { label: '20:00', value: '20' },
  { label: '21:00', value: '21' },
  { label: '22:00', value: '22' },
  { label: '23:00', value: '23' },
];

const handleScanIntervalChange = () => {
  formModel.common_settings.scan_interval = `@every ${scanCronString0.value}`;
};

const handleScanSpecTimeChange = () => {
  formModel.common_settings.scan_interval = `0 ${scanCronString1.value.join(',')} * * *`;
};

const handleScanCustomChange = () => {
  formModel.common_settings.scan_interval = `${scanCronString2.value}`;
};

// 同步更新emby的线程设置
watch(
  () => formModel.common_settings.threads,
  (val) => {
    formModel.emby_settings.threads = val;
  }
);
</script>
