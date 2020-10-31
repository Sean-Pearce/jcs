# jcs - Joint Cloud Storage

[![pipeline status](http://gitlab.act.buaa.edu.cn/jointcloudstorage/jcs/badges/master/pipeline.svg)](http://gitlab.act.buaa.edu.cn/jointcloudstorage/jcs/-/commits/master)

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

## 开发流程

1. 添加新 `Issue`，在 `Issue` 页面描述想要修改/添加的功能
2. 确定实现方案后，在 `Issue` 页面创建相应的 `Merge Request`
3. 向新分支提交代码，并在 `Merge Request` 页面追踪进度
4. 功能实现完成，且自行测试通过后，请其他成员进行 `Code Review`（比较小的修改可以跳过这步）
5. `Code Review` 通过后，在将修改 `merge` 进目标分支前，请务必把本分支的修改 `rebase` 到最新的目标分支上，并在 `Merge Request` 页面勾选 `Squash Commits` 选项

注意：请不要直接在 `master` 分支提交代码。对 `master` 分支的任何修改都应该遵循上述流程。

## Pipeline 使用指南

本项目集成了 `GitLab CI/CD`，每次 `commit` 都会触发一个 `Pipeline` 任务，每个 `Pipeline` 任务分为五个 `stage`，分别是：
- `build`：构建镜像
- `test`：进行集成测试
- `deliver`：将镜像推送到镜像仓库
- `staging`：将镜像部署到预发环境（实验室内部虚拟机）
- `deploy`：将镜像部署到生产环境（ jcc 项目服务器）

其中，非 `master` 分支上的 `commit` 只会触发前两个 `stage`，如果需要，也可以在 `Pipeline` 页面手动触发 `staging`，部署完成后可以在 [这里](http://182.92.117.116:15004) 看到最新修改效果。而 `master` 分支上的 `commit` 会自动触发前三个 `stage`，且可以手动触发 `staging` 和 `deploy`，后者将直接把镜像部署到 jcc 项目服务器，访问地址在 [这里](http://117.50.133.184/jcs)。
