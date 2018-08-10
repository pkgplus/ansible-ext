FROM golang:1.9.4-alpine3.6
MAINTAINER Xue Bing <xuebing1110@gmail.com>

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

# move to GOPATH
RUN mkdir -p /go/src/github.com/xuebing1110/ansible-ext
COPY . $GOPATH/src/github.com/xuebing1110/ansible-ext/
WORKDIR $GOPATH/src/github.com/xuebing1110/ansible-ext

# copy config
RUN mkdir -p /app
COPY proto/ansible.swagger.json /app/
COPY etc/init.d/* /etc/init.d/
COPY etc/playbook/* /app/playbook/
RUN mkdir -p /app/playbook/playbook.d
COPY etc/sysconfig/* /etc/sysconfig/
COPY etc/systemd/* /etc/systemd/system/

# build
RUN go build -o /app/ansible-ext cmd/main.go

EXPOSE 50051
WORKDIR /app
CMD ["/app/ansible-ext"]
