<template>
  <share-subtitle-panel :is-movie="isMovie" :media-data="mediaData" />
</template>

<script setup>
import ShareSubtitlePanel from 'components/ShareSubtitle/ShareSubtitlePanel.vue';
import { useSettings } from 'pages/settings/use-settings';

const isMovie = false;
// const mediaData = {
//   video_f_path: '/Users/myon/temp/ccxx/test3/Chainsaw Man - S01E02 - ARRIVAL IN TOKYO HDTV-1080p.mp4',
//   sub_url_list: [
//     '/movie_dir_0/test3/Chainsaw Man - S01E02 - ARRIVAL IN TOKYO HDTV-1080p.ass',
//     '/movie_dir_0/test3/Chainsaw Man - S01E02 - ARRIVAL IN TOKYO HDTV-1080p.srt',
//   ],
//   sub_f_path_list: [
//     '/Users/myon/temp/ccxx/test3/Chainsaw Man - S01E02 - ARRIVAL IN TOKYO HDTV-1080p.ass',
//     '/Users/myon/temp/ccxx/test3/Chainsaw Man - S01E02 - ARRIVAL IN TOKYO HDTV-1080p.srt',
//   ],
// };

const mediaData = [
  {
    name: 'Game of Thrones - S01E01 - Winter Is Coming Bluray-1080p.mkv',
    video_f_path:
      '/Users/myon/temp/cctv/Game of Thrones/Season 1/Game of Thrones - S01E01 - Winter Is Coming Bluray-1080p.mkv',
    video_url: '/series_dir_0/Game of Thrones/Season 1/Game of Thrones - S01E01 - Winter Is Coming Bluray-1080p.mkv',
    season: 1,
    episode: 1,
    sub_f_path_list: [
      '/Users/myon/temp/cctv/Game of Thrones/Season 1/Game of Thrones - S01E01 - Winter Is Coming Bluray-1080p.ass',
      '/Users/myon/temp/cctv/Game of Thrones/Season 1/Game of Thrones - S01E01 - Winter Is Coming Bluray-1080p.srt',
    ],
    sub_url_list: [
      '/series_dir_0/Game of Thrones/Season 1/Game of Thrones - S01E01 - Winter Is Coming Bluray-1080p.ass',
      '/series_dir_0/Game of Thrones/Season 1/Game of Thrones - S01E01 - Winter Is Coming Bluray-1080p.srt',
    ],
    media_server_inside_video_id: '',
  },
  {
    name: 'Game of Thrones - S01E02 - Winter Is Coming Bluray-1080p.mkv',
    video_f_path:
      '/Users/myon/temp/cctv/Game of Thrones/Season 1/Game of Thrones - S01E02 - Winter Is Coming Bluray-1080p.mkv',
    video_url: '/series_dir_0/Game of Thrones/Season 1/Game of Thrones - S01E02 - Winter Is Coming Bluray-1080p.mkv',
    season: 1,
    episode: 2,
    sub_f_path_list: [
      '/Users/myon/temp/cctv/Game of Thrones/Season 1/Game of Thrones - S01E02 - Winter Is Coming Bluray-1080p.ass',
      '/Users/myon/temp/cctv/Game of Thrones/Season 1/Game of Thrones - S01E02 - Winter Is Coming Bluray-1080p.srt',
    ],
    sub_url_list: [
      '/series_dir_0/Game of Thrones/Season 1/Game of Thrones - S01E02 - Winter Is Coming Bluray-1080p.ass',
      '/series_dir_0/Game of Thrones/Season 1/Game of Thrones - S01E02 - Winter Is Coming Bluray-1080p.srt',
    ],
    media_server_inside_video_id: '',
  },
  {
    name: 'Game of Thrones - S01E03 - Lord Snow Bluray-1080p.mkv',
    video_f_path:
      '/Users/myon/temp/cctv/Game of Thrones/Season 1/Game of Thrones - S01E03 - Lord Snow Bluray-1080p.mkv',
    video_url: '/series_dir_0/Game of Thrones/Season 1/Game of Thrones - S01E03 - Lord Snow Bluray-1080p.mkv',
    season: 1,
    episode: 3,
    sub_f_path_list: [
      '/Users/myon/temp/cctv/Game of Thrones/Season 1/Game of Thrones - S01E03 - Lord Snow Bluray-1080p.ass',
      '/Users/myon/temp/cctv/Game of Thrones/Season 1/Game of Thrones - S01E03 - Lord Snow Bluray-1080p.srt',
    ],
    sub_url_list: [
      '/series_dir_0/Game of Thrones/Season 1/Game of Thrones - S01E03 - Lord Snow Bluray-1080p.ass',
      '/series_dir_0/Game of Thrones/Season 1/Game of Thrones - S01E03 - Lord Snow Bluray-1080p.srt',
    ],
    media_server_inside_video_id: '',
  },
];

useSettings();
</script>
