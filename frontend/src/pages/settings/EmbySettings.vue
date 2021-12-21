<template>
  <div>
    <q-form @submit="submitAll">
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
                :rules="[(val) => (form.enable && !!val) || '不能为空']"
              />
            </q-item-section>
          </q-item>
          <q-item>
            <q-item-section>
              <q-item-label>APIKey</q-item-label>
            </q-item-section>
            <q-item-section avatar>
              <q-input
                v-model="form.api_key"
                standout
                dense
                :rules="[(val) => (form.enable && !!val) || '不能为空']"
              />
            </q-item-section>
          </q-item>
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

          <q-item>
            <q-item-section>
              <q-item-label>电影的目录映射</q-item-label>
              <q-markup-table class="q-mt-sm" flat separator="none">
                <thead>
                  <tr>
                    <th>本程序读取到的路径</th>
                    <th>Emby读取到的路径</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(target, source) in form.movie_paths_mapping" :key="source">
                    <td>
                      <q-input :model-value="source" disable hint="" standout dense />
                    </td>
                    <td>
                      <q-input
                        v-model="form.movie_paths_mapping[source]"
                        standout
                        dense
                        :rules="[(val) => (form.enable && !!val) || '不能为空']"
                      />
                    </td>
                  </tr>
                </tbody>
              </q-markup-table>
            </q-item-section>
          </q-item>

          <q-separator spaced inset></q-separator>

          <q-item>
            <q-item-section>
              <q-item-label>连续剧的目录映射</q-item-label>
              <q-markup-table class="q-mt-sm" flat separator="none">
                <thead>
                  <tr>
                    <th>本程序读取到的路径</th>
                    <th>Emby读取到的路径</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(target, source) in form.series_paths_mapping" :key="source">
                    <td>
                      <q-input :model-value="source" disable hint="" standout dense />
                    </td>
                    <td>
                      <q-input
                        v-model="form.series_paths_mapping[source]"
                        standout
                        dense
                        :rules="[(val) => (form.enable && !!val) || '不能为空']"
                      />
                    </td>
                  </tr>
                </tbody>
              </q-markup-table>
            </q-item-section>
          </q-item>
        </template>
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

const { emby_settings: form } = toRefs(formModel);
</script>
