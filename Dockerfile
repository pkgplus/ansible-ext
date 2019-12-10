## build stage
FROM golang:1.13.5-alpine3.10 as build-env

# repo
RUN cp /etc/apk/repositories /etc/apk/repositories.bak
RUN echo "http://mirrors.aliyun.com/alpine/v3.6/main/" > /etc/apk/repositories
RUN echo "http://mirrors.aliyun.com/alpine/v3.6/community/" >> /etc/apk/repositories

# git
RUN apk update
RUN apk add --no-cache git

# move to GOPATH
RUN mkdir -p /app
WORKDIR /app

# go mod
ENV GOPROXY=https://goproxy.cn
COPY go.mod .
COPY go.sum .
RUN go mod download

# build
COPY . .
COPY etc /app/
RUN go build -o /app/ansible-ext cmd/main.go

## docker image stage
FROM alpine:3.10

# repo
RUN cp /etc/apk/repositories /etc/apk/repositories.bak
RUN echo "http://mirrors.aliyun.com/alpine/v3.6/main/" > /etc/apk/repositories
RUN echo "http://mirrors.aliyun.com/alpine/v3.6/community/" >> /etc/apk/repositories

# timezone
RUN apk update
RUN apk add --no-cache py-pip ansible openssh && pip install paramiko
RUN apk add --no-cache tzdata \
    && echo "Asia/Shanghai" > /etc/timezone \
    && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

# Add Tini
RUN apk add --no-cache tini
ENTRYPOINT ["/sbin/tini", "--"]

COPY --from=build-env /app /app

# copy config
COPY proto/ansible.swagger.json /app/
COPY etc/init.d/* /etc/init.d/
COPY etc/playbook/* /app/playbook/
RUN mkdir -p /app/playbook/playbook.d
COPY etc/sysconfig/* /etc/sysconfig/
COPY etc/systemd/* /etc/systemd/system/

ENV PORT=50051
EXPOSE 50051
WORKDIR /app
CMD ["/app/ansible-ext"]


