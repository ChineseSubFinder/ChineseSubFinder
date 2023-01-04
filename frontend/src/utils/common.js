export const deepCopy = (obj) => JSON.parse(JSON.stringify(obj));

export const gotoGithubIssuePage = () => {
  const searchParams = new URLSearchParams();
  searchParams.append('template', '----bug----.md');
  window.open(`https://github.com/ChineseSubFinder/ChineseSubFinder/issues/new?${searchParams.toString()}`, '_blank');
};
