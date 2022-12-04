/**
 * 从文件名中获取第几集
 * @param filename
 * @returns {null|number}
 */
export const getEpisode = (filename) => {
  const episode = filename.match(/s\d+e(\d+)/i);
  if (episode) {
    return parseInt(episode[1], 10);
  }
  // 第x集
  const episode2 = filename.match(/第(\d+)集/i);
  if (episode2) {
    return parseInt(episode2[1], 10);
  }
  return null;
};
