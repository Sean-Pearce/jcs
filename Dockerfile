FROM golang:1.14.2-buster as build-stage
ARG all_proxy
WORKDIR /jcs
# install tools
RUN echo \
    deb https://mirrors.tuna.tsinghua.edu.cn/debian/ buster main contrib non-free\
    deb https://mirrors.tuna.tsinghua.edu.cn/debian/ buster-updates main contrib non-free\
    deb https://mirrors.tuna.tsinghua.edu.cn/debian/ buster-backports main contrib non-free\
    deb https://mirrors.tuna.tsinghua.edu.cn/debian-security buster/updates main contrib non-free\
    > /etc/apt/sources.list
RUN cat /etc/apt/sources.list
RUN apt update && apt install -y unzip
# install protoc toolchain
RUN PB_REL="https://github.com/protocolbuffers/protobuf/releases" \
    && curl -LO $PB_REL/download/v3.11.4/protoc-3.11.4-linux-x86_64.zip \
    && unzip protoc-3.11.4-linux-x86_64.zip -d $HOME/.local
ENV PATH="$PATH:/root/.local/bin"
ENV PATH="$PATH:/go/bin"
# copy source code
COPY service ./service
COPY go.* ./
# install dependencies
RUN go env -w GOPROXY=https://goproxy.cn,direct \
    && go mod download \
    && go get github.com/golang/protobuf/protoc-gen-go
# gen proto
RUN protoc --go_out=plugins=grpc:. ./service/scheduler/proto/*.proto
# build httpserver
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
      -ldflags='-w -s -extldflags "-static"' \
      ./service/httpserver
# build storage
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
      -ldflags='-w -s -extldflags "-static"' \
      ./service/storage
# build scheduler
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
      -ldflags='-w -s -extldflags "-static"' \
      ./service/scheduler

FROM scratch as httpserver
WORKDIR /httpserver
COPY --from=build-stage /jcs/httpserver /httpserver/
ENTRYPOINT [ "./httpserver" ]

FROM scratch as storage
WORKDIR /storage
COPY --from=build-stage /jcs/storage /storage/
ENTRYPOINT [ "./storage" ]

FROM scratch as scheduler
WORKDIR /scheduler
COPY --from=build-stage /jcs/scheduler /scheduler/
ENTRYPOINT [ "./scheduler" ]