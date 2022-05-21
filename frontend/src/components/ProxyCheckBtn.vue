<template>
  <q-btn standout dense flat size="sm" title="检查连接" :loading="checking" @click="checkProxy" v-bind="$attrs" />
  <q-dialog v-model="show">
    <q-card style="width: 400px">
      <q-card-section>
        <div class="text-body1">代理检测结果</div>
      </q-card-section>

      <q-separator />

      <q-card-section>
        <q-list dense>
          <q-item v-for="item in checkList" :key="item.name">
            <q-item-section>{{ item.name }}</q-item-section>
            <q-item-section side>
              <div class="row items-center q-gutter-sm" v-if="item.valid">
                <span class="text-positive">{{ item.speed }}ms</span>
                <q-icon name="done" size="18px" color="positive"></q-icon>
              </div>
              <q-icon v-else name="close" size="18px" color="negative"></q-icon>
            </q-item-section>
          </q-item>
        </q-list>
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script setup>
import { ref, defineProps } from 'vue';
import CommonApi from 'src/api/CommonApi';

const props = defineProps({
  settings: Object,
});

const show = ref(false);

const checking = ref(false);
const checkList = ref([]);

const checkProxy = async () => {
  checking.value = true;
  const [res] = await CommonApi.checkProxy({ proxy_settings: props.settings });
  checking.value = false;
  checkList.value = res?.sub_site_status || [];
  show.value = true;
};
</script>
