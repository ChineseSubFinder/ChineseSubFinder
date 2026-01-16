import HttpClient from './http-client';

const httpClient = new HttpClient();
httpClient.registerInterceptorsFromDirectory(require.context('./interceptors', false, /(?<!noscan)\.js$/));
const createRequest = httpClient.createRequest.bind(httpClient);

export { createRequest };
