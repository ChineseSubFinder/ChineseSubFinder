<template>
  <q-page class="q-pa-lg">
    <div class="row q-gutter-md">
      <q-select
        v-model="filterForm.hasSubtitle"
        dense
        outlined
        :options="[
          { label: '有字幕', value: true },
          { label: '无字幕', value: false },
        ]"
        label="有无字幕"
        clearable
        style="width: 200px"
      />
      <q-checkbox label="已跳过" :val="true" v-model="filterForm.skipped" />
    </div>
    <q-separator class="q-my-md" />
    <div class="row q-gutter-x-md q-gutter-y-lg">
      <q-card v-for="i in 100" :key="i" flat square>
        <div class="q-mb-sm">
          <q-img src="https://via.placeholder.com/500" class="content-width bg-grey-2" height="230px" no-spinner />
        </div>
        <div class="content-width text-ellipsis-line-2">
          Lorem ipsum dolor sit amet Lorem ipsum dolor sit amet
        </div>
        <div class="row items-center">
          <div class="text-grey">2021-10-31</div>
          <q-space />
          <div>
            <q-btn color="black" round flat dense icon="closed_caption" @click.stop title="已有字幕" />
            <!-- <q-btn color="grey" round flat dense icon="closed_caption" @click.stop title="没有字幕" />-->
            <!-- <q-btn color="black" round flat dense icon="closed_caption_disabled" @click.stop title="已跳过" />-->
          </div>
        </div>
      </q-card>
    </div>
  </q-page>
</template>

<script setup>
import { reactive } from 'vue';

const filterForm = reactive({
  skipped: false,
  hasSubtitle: undefined,
});
</script>

<style scoped>
.content-width {
  width: 160px;
}
.text-ellipsis-line-2 {
  height: 40px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}
</style>
