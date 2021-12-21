<template>
  <q-table title="任务列表" :columns="columns" :rows="[{}, {}, {}]">
    <template v-slot:top-right>
      <q-select
        :options="[{ label: '成功', value: 'success' }, { label: '失败', value: 'failed' }]"
        outlined
        model-value=""
        option-label="label"
        option-value="value"
        label="运行结果"
        dense
        style="width: 200px"
      ></q-select>
    </template>
  </q-table>
</template>

<script setup>
const columns = [
  { label: 'ID', field: 'id' },
  { label: '状态', field: 'status' },
  { label: '处理总数', field: 'total' },
  { label: '跳过数量', field: 'skipped' },
  { label: '新增数量', field: 'added' },
  { label: '失败数量', field: 'added' },
  { label: '运行结果', field: 'added' },
].map((e) => ({ align: 'left', ...e }));
</script>
