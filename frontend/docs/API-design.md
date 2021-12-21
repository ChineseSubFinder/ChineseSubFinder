## 接口认证方式

接口认证通过HTTP头`Authorization: Bearer <token>`传递

## API 列表

`content-type`均为`application/json`

### 应用初始化安装

`POST /setup`

不做权限认证，只在首次安装时有效，用于用户第一次安装程序时的引导页面

请求参数：

```js
{
  username: '',
  password: '',
  movieFolder: '',
  seriesFolder: ''
}
```

### 用户登录

`POST /login`

请求参数：
```js
{
  username: 'user',
  password: 'pass',
  movieFolder: ['/media/电影', '/media/电影2'],
  seriesFolder: ['/media/连续剧'],
}
```

返回 HTTP 码204

返回 HTTP 码200：
```js
{
  accessToken: 'xxxxxx'
}
```

返回列表

### 获取应用信息

`GET /system/info`

请求参数：无

返回 HTTP 码200：
```js
{
  version: '0.0.1', // 版本号
  init: false, // 系统是否已经初始化完成，true或false
}
```

### 获取应用状态

`GET /system/status`

请求参数：无

返回 HTTP 码200：
```js
{
  running: true, // 当前是否正在运行任务，true或false
}
```

### 获取配置信息

`GET /settings`

请求参数： 无

返回 HTTP 码 200：

```js
{
  useProxy: false,
  httpProxyUrl: 'http://127.0.0.1:10809',
  everyTime: '6h',
  threads: 2,
  runAtStartup: true,
  movieFolder: ['/media/电影', '/media/电影2'],
  seriesFolder: ['/media/连续剧'],

  debugMode: false,
  saveOneSeasonSub: false,
  subTypePriority: 0,
  subNameFormatter: 0,
  saveMultiSub: false,
  customVideoExts: ['flv', 'mp4'],
  fixTimeLine: true,

  useEmby: false,
  emby: {
    url: 'http://192.169.x.x:8096',
    apiKey: 'xxxx',
    limitCount: 3000,
    skipWatched: false,
    movieFolderMap: [
      {source: '/mnt/电影', target: '/movies'}
    ],
    seriesFolderMap: []
  }
}
```

### 修改配置信息

`PATCH /settings`

请求参数：

```js
{
  useProxy: false,
  httpProxyUrl: 'http://127.0.0.1:10809',
  everyTime: '6h',
  threads: 2,
  runAtStartup: true,
  movieFolder: ['/media/电影', '/media/电影2'],
  seriesFolder: ['/media/连续剧'],

  debugMode: false,
  saveOneSeasonSub: false,
  subTypePriority: 0,
  subNameFormatter: 0,
  saveMultiSub: false,
  customVideoExts: ['flv', 'mp4'],
  fixTimeLine: true,

  emby: {
    url: 'http://192.169.x.x:8096',
    apiKey: 'xxxx',
    limitCount: 3000,
  }
}
```

可局部更新，只更新传递的字段，例如更新基础信息时提交：

```javascript
{
  useProxy: false,
  httpProxyUrl: 'http://127.0.0.1:10809',
  everyTime: '6h',
  threads: 2,
  runAtStartup: true,
  movieFolder: ['/media/电影', '/media/电影2'],
  seriesFolder: ['/media/连续剧']
}
```

返回 HTTP 码 204

### 检查代理服务器

`POST /check-proxy`

请求参数：

```javascript
{
  httpProxyUrl: 'http://127.0.0.1:10809';
}
```

返回 HTTP 码 200：

```javascript
{
  status: 'OK'; // OK or ERROR
}
```

### 检查目录是否可用

`POST /check-path`

请求参数：

```javascript
{
  path: '/mnt/电影';
}
```

返回 HTTP 码 200：

```javascript
{
  status: 'OK'; // OK or ERROR
}
```

返回 HTTP 码 204

### 停止任务

停止正在运行的任务

`POST /jobs/stop`

请求参数：无

返回 HTTP 码 204

### 开始任务

`POST /jobs/start`

请求参数：无

返回 HTTP 码 204

## 通用错误码

### 401

未登录

### 404

请求内容不存在

### 400

参数验证错误

返回错误信息：

```javascript
{
  message: '代理URL不能为空';
}
```

### 500

其他意外情况导致的错误

```javascript
{
  message: 'xxx';
}
```
