<template>
  <q-page class="q-pa-md">
    <q-table
      :columns="columns"
      :rows="filteredData"
      flat
      bordered
      selection="multiple"
      v-model:selected="selected"
      :pagination="{ rowsPerPage: 20 }"
    >
      <template v-slot:top>
        <div class="col">
          <div class="row">
            <div class="col-2 q-table__title">下载队列</div>

            <q-space />

            <div class="q-gutter-sm row">
              <q-select
                label="状态"
                v-model.number="form.status"
                :options="statusOptions"
                outlined
                dense
                map-options
                emit-value
                style="width: 120px"
              ></q-select>
              <q-select
                v-model.number="form.videoType"
                :options="videoTypeOptions"
                label="类型"
                emit-value
                map-options
                outlined
                dense
                style="width: 100px"
              ></q-select>
              <q-select
                v-model="form.priority"
                :options="priorityOptions"
                label="优先级"
                outlined
                dense
                map-options
                emit-value
                style="width: 130px"
              ></q-select>
              <q-input v-model="form.search" outlined label="输入关键字搜索" dense></q-input>
            </div>
          </div>
          <div class="q-mt-sm q-gutter-xs">
            <q-btn
              :disable="selected.length === 0"
              size="sm"
              icon="expand_less"
              label="提升优先级"
              color="primary"
              @click="batchUpdatePriority('high')"
            />
            <q-btn
              :disable="selected.length === 0"
              size="sm"
              icon="expand_more"
              label="降低优先级"
              color="primary"
              @click="batchUpdatePriority('low')"
            />
            <q-btn :disable="selected.length === 0" size="sm" label="删除" color="negative" @click="batchDeleteJobs" />
          </div>
        </div>
      </template>

      <template v-slot:body-cell-jobStatus="{ row }">
        <q-td>
          <span
            :style="{
              background: JOB_STATUS_COLOR_MAP[row.job_status],
              color: 'white',
              borderRadius: '3px',
              padding: '1px 3px',
              fontSize: '12px',
            }"
            >{{ JOB_STATUS_MAP[row.job_status] }}</span
          >
        </q-td>
      </template>

      <template v-slot:body-cell-actions="{ row }">
        <q-td>
          <job-detail-btn-dialog :data="row" />
          <job-log-btn-dialog :data="row" />
        </q-td>
      </template>
    </q-table>
  </q-page>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue';
import JobApi from 'src/api/JobApi';
import { SystemMessage } from 'src/utils/Message';
import { VIDEO_TYPE_NAME_MAP } from 'src/constants/SettingConstants';
import { JOB_STATUS_COLOR_MAP, JOB_STATUS_MAP, JOB_STATUS_OPTIONS } from 'src/constants/JobConstants';
import { useQuasar } from 'quasar';
import JobLogBtnDialog from 'pages/jobs/JobLogBtnDialog';
import JobDetailBtnDialog from 'pages/jobs/JobDetailBtnDialog';

const $q = useQuasar();

const columns = [
  // { label: 'ID', field: 'id' },
  { label: '状态', field: 'job_status', name: 'jobStatus' },
  { label: '类型', field: 'video_type', format: (val) => VIDEO_TYPE_NAME_MAP[val] },
  // { label: '路径', field: 'video_f_path' },
  { label: '名称', field: 'video_name', width: '100px' },
  // { label: '特征码', field: 'feature' },
  // { label: '连续剧目录', field: 'series_root_dir_path' },
  // { label: '季', field: 'season' },
  // { label: '集', field: 'episode' },
  { label: '优先级', field: 'task_priority' },
  // { label: '视频创建时间', field: 'created_time' },
  { label: '创建时间', field: 'added_time' },
  { label: '更新时间', field: 'update_time' },
  // { label: '媒体服务器ID', field: 'media_server_inside_video_id' },
  { label: '错误信息', field: 'error_info' },
  { label: '下载次数', field: 'download_times' },
  { label: '重试次数', field: 'retry_times' },
  { label: '操作', name: 'actions' },
];

const data = ref([]);
const selected = ref([]);
const form = reactive({
  search: '',
  status: null,
  videoType: null,
  priority: null,
});

const filteredData = computed(() => {
  const { search, status, videoType, priority } = form;
  return data.value.filter((item) => {
    if (search !== '') {
      if (
        !(
          item.video_name.includes(search) ||
          item.video_f_path.includes(search) ||
          item.series_root_dir_path.includes(search) ||
          String(item.media_server_inside_video_id) === search
        )
      ) {
        return false;
      }
    }
    if (status !== null && item.job_status !== status) {
      return false;
    }
    if (videoType !== null && item.video_type !== videoType) {
      return false;
    }
    if (priority !== null && item.task_priority !== priority) {
      if (priority === 'high' && item.task_priority > 3) {
        return false;
      }
      if (priority === 'low' && item.task_priority < 7) {
        return false;
      }
      if (priority === 'middle' && (item.task_priority >= 7 || item.task_priority <= 3)) {
        return false;
      }
    }
    return true;
  });
});

const statusOptions = [{ label: '全部', value: null }, ...JOB_STATUS_OPTIONS];
const videoTypeOptions = [
  { label: '全部', value: null },
  ...Object.keys(VIDEO_TYPE_NAME_MAP).map((key) => ({ label: VIDEO_TYPE_NAME_MAP[key], value: parseInt(key, 10) })),
];
const priorityOptions = [
  { label: '全部', value: null },
  { label: '低（7-10）', value: 'low' },
  { label: '中（4-6）', value: 'middle' },
  { label: '高（1-3）', value: 'high' },
];

const batchUpdatePriority = async (priority) => {
  $q.dialog({
    title: '操作确认',
    message: `确认修改优先级？`,
    cancel: true,
    persistent: true,
    focus: 'none',
  }).onOk(async () => {
    const selectedIds = selected.value.map((item) => item.id);
    const results = await Promise.allSettled(
      selectedIds.map((id) =>
        JobApi.update(id, {
          task_priority: priority,
        })
      )
    );
    const errorCount = results.filter(({ value: [, err] }) => err !== null).length;
    if (errorCount > 0) {
      SystemMessage.error(`${errorCount}个任务修改优先级失败！`);
    } else {
      SystemMessage.success('成功修改优先级');
    }
  });
};

const batchDeleteJobs = async () => {
  $q.dialog({
    title: '操作确认',
    message: `确认删除选中任务？`,
    cancel: true,
    persistent: true,
    focus: 'none',
  }).onOk(async () => {
    const selectedIds = selected.value.map((item) => item.id);
    const results = await Promise.allSettled(selectedIds.map((id) => JobApi.delete(id)));
    const errorCount = results.filter(({ value: [, err] }) => err !== null).length;
    if (errorCount > 0) {
      SystemMessage.error(`${errorCount}个任务删除失败！`);
    } else {
      SystemMessage.success('删除成功');
    }
  });
};

const getData = async () => {
  const [res, err] = await JobApi.getList();
  if (err !== null) {
    SystemMessage.error(err.message);
  } else {
    data.value = res.all_jobs;
  }
};

onMounted(() => {
  getData();
});
</script>
