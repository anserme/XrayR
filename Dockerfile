# Build go
FROM golang:1.17-alpine AS builder
WORKDIR /app
COPY . .
ENV CGO_ENABLED=0
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download
RUN go build -v -o XrayR -trimpath -ldflags "-s -w -buildid=" ./main

# Release
FROM  alpine:latest
ARG http_proxy=http://172.28.240.1:7890
ARG https_proxy=http://172.28.240.1:7890
# 安装必要的工具包
RUN  apk --update --no-cache add tzdata ca-certificates curl\
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN mkdir /etc/XrayR/
COPY --from=builder /app/XrayR /usr/bin/XrayR/
COPY config.yml /etc/XrayR/config.yml
COPY runXrayR.sh  /usr/bin/XrayR/runXrayR.sh
RUN chmod +x /usr/bin/XrayR/runXrayR.sh
RUN chmod +x /usr/bin/XrayR/XrayR \
    && ln -s /usr/bin/XrayR/XrayR /usr/bin/xrayr \
    && chmod +x /usr/bin/xrayr

ENTRYPOINT [ "sh", "/usr/bin/XrayR/runXrayR.sh"]