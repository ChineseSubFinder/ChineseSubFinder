import HttpClient from './http-client';

const httpClient = new HttpClient();
httpClient.registerInterceptorsFromDirectory(require.context('./interceptors', false, /(?<!noscan)\.js$/));
const createRequest = httpClient.createRequest.bind(httpClient);

const csfSubtitlesHttpClient = new HttpClient();
csfSubtitlesHttpClient.registerInterceptorsFromDirectory(
  require.context('./csf-subtitles-interceptors', false, /(?<!noscan)\.js$/)
);
const createCsfSubtitlesRequest = csfSubtitlesHttpClient.createRequest.bind(csfSubtitlesHttpClient);

export { createRequest, createCsfSubtitlesRequest };
