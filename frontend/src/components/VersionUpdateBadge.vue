<template>
  <q-badge v-if="latestVersion
   && systemState.systemInfo
   &&latestVersion.tag_name !== systemState.systemInfo?.version"
           class="cursor-pointer"
           label="new"
           title="有新的版本更新"
           @click="visible = true"
  />
  <q-dialog v-if="latestVersion" v-model="visible"
             max-width="400px"
             max-height="400px">
    <q-card>
      <q-card-section>
        <div class="text-h5">{{latestVersion.tag_name}}更新日志</div>
      </q-card-section>

      <q-separator/>

      <q-card-section>
        <markdown :source="latestVersion.body" />
      </q-card-section>

      <q-separator/>

      <q-card-section align="right">
        <q-btn
          color="primary"
          @click="navigateToReleasePage"
        >
          前往更新
        </q-btn>
      </q-card-section>
    </q-card>
  </q-dialog>
</template>

<script setup>
import {onMounted, ref} from 'vue';
import Markdown from 'components/Markdown';
import {systemState} from 'src/store/systemState';

const latestVersion = ref(null);
const visible = ref(false);


const getLatestVersion = async () => {
  const data = await fetch('https://api.github.com/repos/allanpk716/chinesesubfinder/releases/latest').then(
    (res) => res.json()
  );
  latestVersion.value = data;
};

const navigateToReleasePage = () => {
  window.open(latestVersion.value.html_url);
  visible.value = false;
}

onMounted(getLatestVersion);
</script>
