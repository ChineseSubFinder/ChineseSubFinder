<template>
  <div>
    <q-form @submit="submitAll">
      <q-list style="max-width: 600px" dense>
        <q-item>
          <q-item-section>
            <q-item-label>扫描间隔</q-item-label>
            <q-item-label caption>
              格式参考
              <a href="https://pkg.go.dev/github.com/robfig/cron#hdr-Intervals" target="_blank" class="text-primary"
                >robfig/cron 文档</a
              >
            </q-item-label>
          </q-item-section>
          <q-item-section avatar>
            <q-input
              v-model="form.scan_interval"
              standout
              dense
              :rules="[(val) => !!val || '不能为空', validateCronDuration]"
            />
          </q-item-section>
        </q-item>

        <q-separator spaced inset></q-separator>

        <q-item>
          <q-item-section>
            <q-item-label>并发数</q-item-label>
          </q-item-section>
          <q-item-section avatar>
            <q-input
              v-model.number="form.threads"
              standout
              dense
              :rules="[(val) => !!val || '不能为空', (val) => /^\d+$/.test(val) || '必须是整数']"
            />
          </q-item-section>
        </q-item>

        <q-separator spaced inset></q-separator>

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
      <q-separator class="q-mt-md" />
      <form-submit-area />
    </q-form>
  </div>
</template>

<script setup>
import { formModel, submitAll } from 'pages/settings/useSettings';
import { validateCronDuration, validateRemotePath } from 'src/utils/QuasarValidators';
import { toRefs } from '@vueuse/core';
import FormSubmitArea from 'pages/settings/FormSubmitArea';

const { common_settings: form } = toRefs(formModel);
</script>
