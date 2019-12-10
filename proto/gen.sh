#!/bin/bash

########## Reference documentation ##############
# https://grpc.io/blog/coreos
# https://github.com/grpc-ecosystem/grpc-gateway
# require: https://github.com/protocolbuffers/protobuf/releases
#################################################

echo "generate hostmanager stub..."
protoc -I/usr/local/include \
    -I$GOPATH/src \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    -Ihostmanager/ \
    --go_out=plugins=grpc:hostmanager \
    hostmanager/hostmanager.proto

## generate gateway
echo "generate hostmanager gateway..."
protoc -I/usr/local/include \
    -I$GOPATH/src \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    -Ihostmanager/ \
    --grpc-gateway_out=logtostderr=true:hostmanager \
    hostmanager/hostmanager.proto

## generate swagger
echo "generate hostmanager swagger..."
protoc -I/usr/local/include \
    -I$GOPATH/src \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    -Ihostmanager/ \
    --swagger_out=logtostderr=true:. \
    hostmanager/hostmanager.proto

echo "generate ansible stub..."
protoc -I/usr/local/include \
    -I$GOPATH/src \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    -Iansible/ \
    --go_out=plugins=grpc:ansible \
    ansible/ansible.proto

## generate gateway
echo "generate ansible gateway..."
protoc -I/usr/local/include \
    -I$GOPATH/src \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    -Iansible/ \
    --grpc-gateway_out=logtostderr=true:ansible \
    ansible/ansible.proto

## generate swagger
echo "generate ansible swagger..."
protoc -I/usr/local/include \
    -I$GOPATH/src \
    -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    -Iansible/ \
    --swagger_out=logtostderr=true:. \
    ansible/ansible.proto

echo "over!"
