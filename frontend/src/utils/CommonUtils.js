export const deepCopy = (obj) => JSON.parse(JSON.stringify(obj));

export const gotoGithubIssuePage = () => {
  const searchParams = new URLSearchParams();
  searchParams.append('template', '----bug----.md');
  window.open(`https://github.com/allanpk716/ChineseSubFinder/issues/new?${searchParams.toString()}`, '_blank');
};
