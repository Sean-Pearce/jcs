# jcs
Joint Cloud Storage

## 部署

### 构建镜像

```shell
cd root/of/this/project
# 构建 jcs-httpserver 镜像，使用 all_proxy 可以加快依赖拉取速度
docker build -t jcs-httpserver --build-arg all_proxy=$all_proxy --target=httpserver .
# 构建 jcs-storage 镜像
docker build -t jcs-storage --build-arg all_proxy=$all_proxy --target=storage .
# 构建 jcs-scheduler 镜像
docker build -t jcs-scheduler --build-arg all_proxy=$all_proxy --target=scheduler .
cd portal
# 构建 jcs-portal 镜像
docker build . -t jcs-portal
```

### 启动服务

```shell
cd root/of/this/project
docker-compose -d up
```

### 停止服务

```shell
cd root/of/this/project
docker-compose down
```