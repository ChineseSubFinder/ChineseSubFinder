import { ref, watch } from 'vue';

// 列表选择hook
export const useSelection = (data) => {
  const selectAllValue = ref(false);
  const selection = ref([]);

  const handleSelectAll = () => {
    if (selectAllValue.value === true) {
      selection.value.splice(0, selection.value.length);
    }
    if (selectAllValue.value === false || selectAllValue.value === 'maybe') {
      selection.value.splice(0, selection.value.length, ...data.value);
    }
  };

  // 全选按钮状态处理
  watch(
    () => selection.value?.length,
    () => {
      if (selection.value === undefined) return;
      if (selection.value.length === data.value.length) {
        selectAllValue.value = true;
      } else if (selection.value.length === 0) {
        selectAllValue.value = false;
      } else {
        selectAllValue.value = 'maybe';
      }
    }
  );

  return {
    selectAllValue,
    selection,
    handleSelectAll,
  };
};
