# jcs - Joint Cloud Storage

一个用 Golang 实现的云际存储原型系统，提供 web ui 和 s3 api (WIP) 两种访问方式。

## 部署

### 构建镜像

```shell
# 构建 jcs-httpserver 镜像，使用 all_proxy 可以加快依赖拉取速度
docker build -t jcs-httpserver --build-arg all_proxy=$all_proxy --target=httpserver .

# 构建 jcs-storage 镜像
docker build -t jcs-storage --build-arg all_proxy=$all_proxy --target=storage .

# 构建 jcs-scheduler 镜像
docker build -t jcs-scheduler --build-arg all_proxy=$all_proxy --target=scheduler .

# 构建 jcs-portal 镜像
cd portal
docker build . -t jcs-portal
```

### 启动服务

```shell
docker-compose -d up
```

### 访问 Web UI

http://localhost:8080

### 停止服务

```shell
docker-compose down
```

## 开发指南

注意：`master` 分支开启了分支保护，请不要直接在 `master` 分支提交代码。

1. 创建新分支
2. 提交 `pull request`
3. 向新分支提交代码，并在 `pull request` 页面追踪进度
4. 功能实现完成，且测试通过后，请其他成员进行 code review
5. code review 通过后，merge 进目标分支