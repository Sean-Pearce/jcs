# 开发指南

## 系统架构（WIP）

![系统架构](https://i.loli.net/2020/05/13/WIL2owSJi7DTXhl.png)

`web-ui` 是存储系统对外提供的网页端访问方式，对应项目中的 `portal` 模块。`portal` 模块基于 [vue-admin-template](https://github.com/PanJiaChen/vue-admin-template) 开发，详细说明文档在 [这里](https://panjiachen.gitee.io/vue-element-admin-site/zh/guide/)。

`http-server` 是 `portal` 模块的后端，对外提供 `upload`, `download`, `list` 等 http 接口。它会把来自 `portal` 的读写请求封装为 s3 请求然后转发给 `storage` 节点。

`scheduler` 可以根据用户自定义的放置策略和存储服务信息实时计算数据放置方案，对外提供 grpc 接口。

`storage` 根据 `scheduler` 返回的放置方案，将来自 `http-server` 的 s3 请求转发给其他存储服务或本地对象存储后端 minio。
