<template>
  <q-form ref="form" class="q-gutter-md">
    <div class="row items-center q-gutter-sm">
      <div class="text-h6">电影目录</div>
      <q-btn
        icon="add"
        color="primary"
        dense
        rounded
        size="xs"
        title="新增"
        @click="setupState.form.movieFolder.push('')"
      ></q-btn>
    </div>
    <template v-for="(item, i) in setupState.form.movieFolder" :key="i">
      <div class="row items-center q-gutter-xs">
        <q-input
          class="col-11"
          filled
          v-model="setupState.form.movieFolder[i]"
          :label="`电影目录${i > 0 ? i + 1 : ''}`"
          hint="输入需要扫描的电影文件夹目录"
          lazy-rules
          :rules="[(val) => (val && val.length > 0) || '不能为空', validateRemotePath]"
        />
        <q-btn
          v-if="i !== 0"
          icon="remove"
          color="negative"
          dense
          rounded
          size="xs"
          title="删除"
          @click="setupState.form.movieFolder.splice(i, 1)"
        ></q-btn>
      </div>
    </template>

    <div class="row items-center q-gutter-sm">
      <div class="text-h6">连续剧目录</div>
      <q-btn
        icon="add"
        color="primary"
        dense
        rounded
        size="xs"
        title="新增"
        @click="setupState.form.seriesFolder.push('')"
      ></q-btn>
    </div>
    <template v-for="(item, i) in setupState.form.seriesFolder" :key="i">
      <div class="row items-center q-gutter-xs">
        <q-input
          class="col-11"
          filled
          v-model="setupState.form.seriesFolder[i]"
          :label="`连续剧目录${i > 0 ? i + 1 : ''}`"
          hint="输入需要扫描的连续剧文件夹目录"
          lazy-rules
          :rules="[(val) => (val && val.length > 0) || '不能为空', validateRemotePath]"
        />
        <q-btn
          v-if="i !== 0"
          icon="remove"
          color="negative"
          dense
          rounded
          size="xs"
          title="删除"
          @click="setupState.form.seriesFolder.splice(i, 1)"
        ></q-btn>
      </div>
    </template>
  </q-form>
</template>

<script setup>
import { validateRemotePath } from 'src/utils/quasar-validators';
import { setupState } from 'pages/setup/use-setup';
</script>
