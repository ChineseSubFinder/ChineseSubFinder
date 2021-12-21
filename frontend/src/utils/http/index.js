import { createRequest, registerInterceptorsFromDirectory } from './http-client';

registerInterceptorsFromDirectory(require.context('./interceptors', false, /(?<!noscan)\.js$/));

export { createRequest };
