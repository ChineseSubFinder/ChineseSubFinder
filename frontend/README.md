# ChineseSubFinder Panel

管理面板

## 安装

```bash
npm install
```

> 如果是本地调试，可以使用 `npm ci` 命令装依赖，不会修改文件

### 运行

```bash
# 启动后端
go run cmd/chinesesubfinder/main.go

# 运行前端开发环境
npm run dev
```

### 构建

```bash
npm run build
```

## 配置文件说明

| 文件名           | 环境     |
| ---------------- | -------- |
| .env.development | 开发环境 |
| .env.production  | 正式环境 |

`.env.development`中配置的`/api`，是开发环境下对后端服务的代理，代理规则可在`quasar.conf.js > devServer`中查看。

如果本地运行需要修改后端地址请新增`.env.development.local`文件

配置项说明：

| key         | value             |
| ----------- | ----------------- |
| BACKEND_URL | 后台 API 服务地址 |

## WebUI 进入后更新说明

修改这个文件的内容即可：`/front/NOTIFY.md`
