<template>
  <q-dialog
    :model-value="visible"
    @update:model-value="(val) => $emit('update:visible', val)"
    v-bind="$attrs"
    persistent
  >
    <q-card :style="{ width: width }">
      <q-card-section>
        <div class="text-h6">{{ title }}</div>
      </q-card-section>
      <q-separator />
      <q-form ref="editForm" @submit="submit" class="q-gutter-md">
        <q-card-section>
          <slot></slot>
        </q-card-section>
        <q-separator />
        <q-card-actions align="right">
          <q-btn class="q-px-md" v-close-popup flat color="primary" label="关闭" />
          <q-btn class="q-px-md" type="submit" color="primary" :loading="submitting" label="提交">
            <template v-slot:loading>
              <q-spinner-facebook />
            </template>
          </q-btn>
        </q-card-actions>
      </q-form>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { ref, defineProps } from 'vue';

const props = defineProps({
  title: String,
  visible: {
    type: Boolean,
    default: false,
  },
  width: {
    type: String,
    default: '600px',
  },
  submitFn: {
    type: Function,
    default: (callback) => {
      callback();
    },
  },
});

const submitting = ref(false);
const submit = async () => {
  submitting.value = true;
  try {
    await props.submitFn();
  } catch (e) {
    // handle error
  }
  submitting.value = false;
};
</script>

<style scoped>
::v-deep(.q-card__actions .q-btn) {
  padding: 8px 18px;
}
</style>
