<template>
  <q-btn label="详情" flat dense @click="show = true" color="primary" />

  <q-dialog v-model="show">
    <q-card style="width: 900px; max-width: 900px">
      <q-card-section>
        <div class="text-body1">任务详情</div>
      </q-card-section>

      <q-separator />

      <q-card-section>
        <table class="job-detail-table">
          <tbody>
            <tr>
              <td>ID</td>
              <td>{{ data.id }}</td>
            </tr>
            <tr>
              <td>状态</td>
              <td>{{ JOB_STATUS_MAP[data.job_status] }}</td>
            </tr>
            <tr>
              <td>类型</td>
              <td>{{ VIDEO_TYPE_NAME_MAP[data.video_type] }}</td>
            </tr>
            <tr>
              <td>路径</td>
              <td>{{ data.video_f_path }}</td>
            </tr>
            <tr>
              <td>名称</td>
              <td>{{ data.video_name }}</td>
            </tr>
            <tr>
              <td>特征码</td>
              <td>{{ data.feature }}</td>
            </tr>
            <template v-if="data.video_type === VIDEO_TYPE_TV">
              <tr>
                <td>连续剧目录</td>
                <td>{{ data.series_root_dir_path }}</td>
              </tr>
              <tr>
                <td>季</td>
                <td>{{ data.season }}</td>
              </tr>
              <tr>
                <td>集</td>
                <td>{{ data.episode }}</td>
              </tr>
            </template>
            <tr>
              <td>媒体服务器ID</td>
              <td>{{ data.media_server_inside_video_id }}</td>
            </tr>
            <tr>
              <td>优先级</td>
              <td>{{ data.task_priority }}</td>
            </tr>
            <tr>
              <td>视频创建时间</td>
              <td>{{ data.created_time }}</td>
            </tr>
            <tr>
              <td>任务创建时间</td>
              <td>{{ data.added_time }}</td>
            </tr>
            <tr>
              <td>更新时间</td>
              <td>{{ data.update_time }}</td>
            </tr>
            <tr>
              <td>错误信息</td>
              <td>{{ data.error_info }}</td>
            </tr>
            <tr>
              <td>下载次数</td>
              <td>{{ data.download_times }}</td>
            </tr>
            <tr>
              <td>重试次数</td>
              <td>{{ data.retry_times }}</td>
            </tr>
          </tbody>
        </table>
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { ref } from 'vue';
import { JOB_STATUS_MAP } from 'src/constants/JobConstants';
import { VIDEO_TYPE_NAME_MAP, VIDEO_TYPE_TV } from 'src/constants/SettingConstants';

defineProps({
  data: {
    type: Object,
  },
});

const show = ref(false);
</script>

<style lang="scss" scoped>
.job-detail-table {
  tr td:first-child {
    padding-right: 20px;
    text-align: right;
    color: $grey;
  }
}
</style>
