<template>
  <div class="markdown-body" v-html="html"></div>
</template>

<script setup>
import {markdown} from 'markdown';
import {computed, defineProps} from 'vue';

const props = defineProps({
  source: {
    type: String,
    required: true
  },
});

const html = computed(() => markdown.toHTML(props.source));
</script>

<style scoped>
@import "~github-markdown-css/github-markdown.css";
</style>
