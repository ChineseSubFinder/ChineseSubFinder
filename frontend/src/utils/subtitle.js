/**
 * 从文件名中获取第几集
 * @param filename
 * @returns {null|number}
 */
export const getEpisode = (filename) => {
  let episode = filename.match(/s\d+e(\d+)/i);
  if (episode) {
    return parseInt(episode[1], 10);
  }
  // 第x集
  episode = filename.match(/第(\d+)(集|话|話)/i);
  if (episode) {
    return parseInt(episode[1], 10);
  }

  // [xx] 【xx】 匹配一些动画字幕组
  episode = filename.match(/\[(\d+)\]/i);
  if (episode) {
    return parseInt(episode[1], 10);
  }
  episode = filename.match(/【(\d+)】/i);
  if (episode) {
    return parseInt(episode[1], 10);
  }

  return null;
};
