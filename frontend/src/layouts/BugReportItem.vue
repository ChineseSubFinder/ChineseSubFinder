<template>
  <q-item clickable @click="gotoGithubIssuePage" title="跳转到Github issue页面">
    <q-item-section>
      <div class="row items-center">
        <span><q-icon name="bug_report" size="18px" /></span>
        <span class="q-ml-xs">问题反馈</span>
      </div>
    </q-item-section>
  </q-item>
</template>

<script setup>
import { gotoGithubIssuePage } from 'src/utils/common';
</script>
