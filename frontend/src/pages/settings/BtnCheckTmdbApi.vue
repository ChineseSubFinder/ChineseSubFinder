<template>
  <q-btn label="检测TMDB API" color="secondary" :loading="loading" @click="check" />
</template>

<script setup>
import { ref } from 'vue';
import CommonApi from 'src/api/CommonApi';
import { SystemMessage } from 'src/utils/Message';
import { formModel } from 'pages/settings/useSettings';

const loading = ref(false);

const check = async () => {
  loading.value = true;
  console.log(formModel);
  const [res, err] = await CommonApi.checkTmdbApiKey({
    proxy_settings: formModel.advanced_settings.proxy_settings,
    api_key: formModel.advanced_settings.tmdb_api_settings.api_key,
  });
  if (err !== null) {
    SystemMessage.error(err.message);
  } else if (res.message !== 'true') {
    SystemMessage.error('TMDB API连接异常，请确认ApiKey是否正确，或者启用代理后重试');
  } else {
    SystemMessage.success('TMDB服务连接正常');
  }
  loading.value = false;
};
</script>
