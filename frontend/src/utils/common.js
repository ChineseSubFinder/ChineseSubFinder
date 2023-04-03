export const deepCopy = (obj) => JSON.parse(JSON.stringify(obj));

export const gotoGithubIssuePage = () => {
  const searchParams = new URLSearchParams();
  searchParams.append('template', '----bug----.md');
  window.open(`https://github.com/ChineseSubFinder/ChineseSubFinder/issues/new?${searchParams.toString()}`, '_blank');
};

export const isImdbId = (str) => {
  if (!str) return false;
  if (str === 'tt00000') return false;
  return /^tt\d{7,8}$/.test(str);
};
