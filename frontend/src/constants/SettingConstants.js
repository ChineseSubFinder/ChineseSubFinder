export const SUB_TYPE_PRIORITY_AUTO = 0;
export const SUB_TYPE_PRIORITY_SRT = 1;
export const SUB_TYPE_PRIORITY_ASS = 2;

export const SUB_TYPE_PRIORITY_NAME_MAP = {
  [SUB_TYPE_PRIORITY_AUTO]: '自动',
  [SUB_TYPE_PRIORITY_SRT]: 'srt',
  [SUB_TYPE_PRIORITY_ASS]: 'ass',
};

export const SUB_NAME_FORMAT_EMBY = 0;
export const SUB_NAME_FORMAT_NORMAL = 1;
export const SUB_NAME_VIDEO = 2;

export const SUB_NAME_FORMAT_NAME_MAP = {
  [SUB_NAME_FORMAT_EMBY]: 'Emby格式',
  [SUB_NAME_FORMAT_NORMAL]: '常规格式',
  [SUB_NAME_VIDEO]: '与视频名称一致',
};

export const DESC_ENCODE_TYPE_UTF8 = 0;
export const DESC_ENCODE_TYPE_GBK = 1;

export const DESC_ENCODE_TYPE_NAME_MAP = {
  [DESC_ENCODE_TYPE_UTF8]: 'UTF-8',
  [DESC_ENCODE_TYPE_GBK]: 'GBK',
};

export const AUTO_CONVERT_LANG_CHS = 0;
export const AUTO_CONVERT_LANG_CHT = 1;

export const AUTO_CONVERT_LANG_NAME_MAP = {
  [AUTO_CONVERT_LANG_CHS]: '转简体',
  [AUTO_CONVERT_LANG_CHT]: '轉繁體',
};

export const DEFAULT_SUB_SOURCE_URL_MAP = {
  xunlei: 'http://sub.xmp.sandai.net:8000/subxl/%s.json',
  shooter: 'https://www.shooter.cn/api/subapi.php',
  subhd: 'https://subhd.tv',
  zimuku: 'https://zimuku.org',
  assrt: 'https://api.assrt.net/v1',
  a4k: 'https://www.a4k.net',
};

export const PROXY_TYPE_HTTP = 'http';
export const PROXY_TYPE_SOCKS5 = 'socks5';

export const PROXY_TYPE_NAME_MAP = {
  [PROXY_TYPE_HTTP]: 'HTTP',
  [PROXY_TYPE_SOCKS5]: 'SOCKS5',
};

export const VIDEO_TYPE_MOVIE = 0;
export const VIDEO_TYPE_TV = 1;

export const VIDEO_TYPE_NAME_MAP = {
  [VIDEO_TYPE_MOVIE]: '电影',
  [VIDEO_TYPE_TV]: '电视剧',
};
