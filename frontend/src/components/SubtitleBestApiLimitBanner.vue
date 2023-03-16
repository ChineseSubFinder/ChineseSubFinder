<template>
  <q-banner
    v-if="apiLimitInfo"
    dense
    class="bg-grey-3"
    :class="{
      // 使用超过4/5时，显示黄色警告
      'bg-negative': apiLimitInfo.dailyCount >= apiLimitInfo.dayliLimit,
      'bg-warning': apiLimitInfo.dailyCount / apiLimitInfo.dayliLimit > 4 / 5,
    }"
  >
    <div class="text-bold text-center">
      每日限制：{{ apiLimitInfo.dailyCount }} / {{ apiLimitInfo.dailyLimit }}，ApiKey 过期时间：{{
        dayjs.unix(apiLimitInfo.expireTime).format('YYYY-MM-DD HH:mm:ss')
      }}
    </div>
  </q-banner>
</template>

<script setup>
import { ref } from 'vue';
import useEventBus from 'src/composables/use-event-bus';
import dayjs from 'dayjs';

const apiLimitInfo = ref(null);

useEventBus('subtitle-best-api-limit-info', (info) => {
  apiLimitInfo.value = info;
});
</script>
