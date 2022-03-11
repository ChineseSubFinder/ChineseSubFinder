import LogApi from 'src/api/LogApi';
import { SystemMessage } from 'src/utils/Message';
import { computed, onMounted, reactive, ref, watch } from 'vue';

export const form = reactive({
  logCount: 3,
});

export const useLogList = () => {
  const logList = ref([]);
  const currentIndex = ref(null);
  const currentItem = computed(() => logList.value.find((item) => item.index === currentIndex.value));

  const getData = async () => {
    const [res, err] = await LogApi.getList({
      the_last_few_times: form.logCount,
    });
    if (err !== null) {
      SystemMessage.error(err.message);
      return;
    }
    logList.value = res.recent_logs;
  };

  watch(() => form.logCount, getData);

  onMounted(() => {
    getData();
  });

  return {
    logList,
    currentIndex,
    currentItem,
  };
};
