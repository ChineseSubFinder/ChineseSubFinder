<template>
  <q-form ref="form">
    <q-list dense>
      <q-item>
        <q-item-section>
          <q-item-label>Emby的内网URL</q-item-label>
        </q-item-section>
        <q-item-section>
          <q-input v-model="setupState.form.emby.url" standout dense :rules="[(val) => !!val || '不能为空']" />
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

      <q-separator spaced inset></q-separator>

      <q-item>
        <q-item-section>
          <q-item-label>电影的目录映射</q-item-label>
          <q-markup-table
            v-if="Object.keys(setupState.form.emby.movieFolderMap).length"
            class="q-mt-sm"
            flat
            separator="none"
          >
            <thead>
              <tr>
                <th>本程序读取到的路径</th>
                <th>Emby读取到的路径</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(target, source) in setupState.form.emby.movieFolderMap" :key="source">
                <td>
                  <q-input model-value="target" disable hint="" standout dense />
                </td>
                <td>
                  <q-input
                    v-model="setupState.form.emby.movieFolderMap[source]"
                    standout
                    dense
                    :rules="[(val) => !!val || '不能为空']"
                  />
                </td>
              </tr>
            </tbody>
          </q-markup-table>
          <div v-else class="text-grey q-pa-md">未填写电影目录</div>
        </q-item-section>
      </q-item>

      <q-item>
        <q-item-section>
          <q-item-label>连续剧的目录映射</q-item-label>
          <q-markup-table
            v-if="Object.keys(setupState.form.emby.seriesFolderMap).length"
            class="q-mt-sm"
            flat
            separator="none"
          >
            <thead>
              <tr>
                <th>本程序读取到的路径</th>
                <th>Emby读取到的路径</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(target, source) in setupState.form.emby.seriesFolderMap" :key="source">
                <td>
                  <q-input model-value="target" disable hint="" standout dense />
                </td>
                <td>
                  <q-input
                    v-model="setupState.form.emby.seriesFolderMap[source]"
                    standout
                    dense
                    :rules="[(val) => !!val || '不能为空']"
                  />
                </td>
              </tr>
            </tbody>
          </q-markup-table>
          <div v-else class="text-grey q-pa-md">未填写连续剧目录</div>
        </q-item-section>
      </q-item>
    </q-list>
  </q-form>
</template>

<script setup>
import { setupState } from 'pages/setup/useSetup';
</script>
