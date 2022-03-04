import LogApi from 'src/api/LogApi';
import { SystemMessage } from 'src/utils/Message';
import { computed, onMounted, ref } from 'vue';

export const useLogList = () => {
  const logList = ref([]);
  const currentIndex = ref(null);
  const currentItem = computed(() => logList.value.find((item) => item.index === currentIndex.value));

  const getData = async () => {
    const [res, err] = await LogApi.getList();
    if (err !== null) {
      SystemMessage.error(err.message);
      return;
    }
    logList.value = res.recent_logs;
  };

  onMounted(() => {
    getData();
  });

  return {
    logList,
    currentIndex,
    currentItem,
  };
};
