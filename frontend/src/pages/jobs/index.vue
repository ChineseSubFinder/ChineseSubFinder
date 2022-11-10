<template>
  <q-page class="q-pa-md">
    <div class="row items-center">
      <div class="q-gutter-xs">
        <q-btn
          :disable="selected.length === 0"
          size="md"
          icon="expand_less"
          label="升级"
          color="primary"
          @click="batchUpdatePriority('high')"
        />

        <q-btn
          :disable="selected.length === 0"
          size="md"
          icon="expand_more"
          label="降级"
          color="primary"
          @click="batchUpdatePriority('low')"
        />

        <q-btn :disable="selected.length === 0" size="md" label="修改状态" color="primary" @click="batchUpdateStatus" />
      </div>

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

    <q-separator class="q-mt-md" />

    <q-table
      :columns="columns"
      :rows="filteredData"
      flat
      selection="multiple"
      v-model:selected="selected"
      class="sticky-column-table"
      :pagination="{ rowsPerPage: 20 }"
    >
      <template v-slot:body-cell-jobStatus="{ row }">
        <q-td>
          <span
            :style="{
              background: JOB_STATUS_COLOR_MAP[row.job_status],
              color: 'white',
              borderRadius: '5px',
              padding: '2px 6px',
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
import { SystemMessage } from 'src/utils/message';
import { VIDEO_TYPE_NAME_MAP } from 'src/constants/SettingConstants';
import {
  JOB_STATUS_COLOR_MAP,
  JOB_STATUS_IGNORE,
  JOB_STATUS_MAP,
  JOB_STATUS_OPTIONS,
  JOB_STATUS_PENDING,
} from 'src/constants/JobConstants';
import { useQuasar } from 'quasar';
import JobLogBtnDialog from 'pages/jobs/JobLogBtnDialog';
import JobDetailBtnDialog from 'pages/jobs/JobDetailBtnDialog';

const $q = useQuasar();

const columns = [
  // { label: 'ID', field: 'id' },
  { label: '状态', field: 'job_status', name: 'jobStatus', align: 'left' },
  { label: '类型', field: 'video_type', format: (val) => VIDEO_TYPE_NAME_MAP[val], align: 'left' },
  // { label: '路径', field: 'video_f_path' },
  { label: '名称', field: 'video_name', width: '100px', align: 'left' },
  // { label: '特征码', field: 'feature' },
  // { label: '连续剧目录', field: 'series_root_dir_path' },
  // { label: '季', field: 'season' },
  // { label: '集', field: 'episode' },
  { label: '优先级', field: 'task_priority', align: 'left' },
  // { label: '视频创建时间', field: 'created_time' },
  { label: '创建时间', field: 'added_time', align: 'left' },
  { label: '更新时间', field: 'update_time', align: 'left' },
  // { label: '媒体服务器ID', field: 'media_server_inside_video_id' },
  { label: '错误信息', field: 'error_info', align: 'left' },
  { label: '下载次数', field: 'download_times', align: 'left' },
  { label: '重试次数', field: 'retry_times', align: 'left' },
  { label: '操作', name: 'actions', align: 'left', headerClasses: 'sticky-column-header' },
];

const data = ref([]);
const selected = ref([]);
const form = reactive({
  search: '',
  status: null,
  videoType: null,
  priority: null,
});

const JOB_PRIORITY_NUM2STR_MAP = {
  0: '高',
  1: '高',
  2: '高',
  3: '高',
  4: '中',
  5: '中',
  6: '中',
  7: '低',
  8: '低',
  9: '低',
  10: '低',
};

const getData = async () => {
  const [res, err] = await JobApi.getList();
  if (err !== null) {
    SystemMessage.error(err.message);
  } else {
    data.value = res.all_jobs;
  }
};

const refresh = () => {
  selected.value = [];
  getData();
};

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

    const betweenOfNumber = (num, min, max) => num >= min && num <= max;
    if (priority !== null && item.task_priority !== priority) {
      // 0-3为高
      if (priority === 'high' && !betweenOfNumber(item.task_priority, 0, 3)) {
        return false;
      }
      // 7-10为低
      if (priority === 'low' && !betweenOfNumber(item.task_priority, 7, 10)) {
        return false;
      }
      // 4-6为中
      if (priority === 'middle' && !betweenOfNumber(item.task_priority, 4, 6)) {
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
  { label: '高（0-3）', value: 'high' },
];

const batchUpdatePriority = async (priority) => {
  $q.dialog({
    title: '操作确认',
    message: `确认修改优先级？`,
    cancel: true,
    persistent: true,
    focus: 'none',
  }).onOk(async () => {
    const results = await Promise.allSettled(
      selected.value.map((item) =>
        JobApi.update(item.id, {
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

    refresh();
  });
};

const batchUpdateStatus = async () => {
  $q.dialog({
    title: '修改状态',
    message: '需要变更成哪个状态？',
    options: {
      type: 'radio',
      items: [
        { label: JOB_STATUS_MAP[JOB_STATUS_PENDING], value: JOB_STATUS_PENDING },
        { label: JOB_STATUS_MAP[JOB_STATUS_IGNORE], value: JOB_STATUS_IGNORE },
      ],
    },
    cancel: true,
    persistent: true,
  }).onOk(async (val) => {
    const results = await Promise.allSettled(
      selected.value.map((item) =>
        JobApi.update(item.id, {
          job_status: val,
          task_priority: JOB_PRIORITY_NUM2STR_MAP[item.task_priority],
        })
      )
    );
    const errorCount = results.filter(({ value: [, err] }) => err !== null).length;
    if (errorCount > 0) {
      SystemMessage.error(`${errorCount}个任务修改状态失败！`);
    } else {
      SystemMessage.success('成功修改任务状态');
    }

    refresh();
  });
};

onMounted(() => {
  getData();
});
</script>

<style lang="scss">
.sticky-column-table {
  thead tr:last-child th:last-child {
    background-color: #fff;
  }

  td:last-child {
    background-color: #fff;
  }

  th:last-child,
  td:last-child {
    position: sticky;
    right: 0;
    z-index: 1;
    box-shadow: -5px 0px 5px -1px #ddd;
  }
  td:last-child {
    //border-left: 1px solid $grey-3;
  }
}
</style>
