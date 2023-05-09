<template>
  <share-subtitle-panel :is-movie="isMovie" :media-data="mediaData" />
</template>

<script setup>
import ShareSubtitlePanel from 'components/ShareSubtitle/ShareSubtitlePanel.vue';
import { useSettings } from 'pages/settings/use-settings';

const isMovie = true;
const mediaData = {
  video_f_path: '/Users/myon/temp/ccxx/test3/Chainsaw Man - S01E02 - ARRIVAL IN TOKYO HDTV-1080p.mp4',
  sub_url_list: [
    '/movie_dir_0/test3/Chainsaw Man - S01E02 - ARRIVAL IN TOKYO HDTV-1080p.ass',
    '/movie_dir_0/test3/Chainsaw Man - S01E02 - ARRIVAL IN TOKYO HDTV-1080p.srt',
  ],
  sub_f_path_list: [
    '/Users/myon/temp/ccxx/test3/Chainsaw Man - S01E02 - ARRIVAL IN TOKYO HDTV-1080p.ass',
    '/Users/myon/temp/ccxx/test3/Chainsaw Man - S01E02 - ARRIVAL IN TOKYO HDTV-1080p.srt',
  ],
};

useSettings();
</script>
