## generate stub
```sh
protoc -I/usr/local/include \
    -I/Users/xuebing/WorkSpace/go/src \
    -I/Users/xuebing/WorkSpace/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    -IHostManager/ \
    --go_out=plugins=grpc:hostManager \
    HostManager/hostManager.proto
```

## generate gateway
```sh
protoc -I/usr/local/include \
    -I/Users/xuebing/WorkSpace/go/src \
    -I/Users/xuebing/WorkSpace/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    -IHostManager/ \
    --grpc-gateway_out=logtostderr=true:hostManager \
    HostManager/hostManager.proto

```

## generate swagger
```sh
protoc -I/usr/local/include \
    -I/Users/xuebing/WorkSpace/go/src \
    -I/Users/xuebing/WorkSpace/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
    -IHostManager/ \
    --swagger_out=logtostderr=true:hostManager \
    HostManager/hostManager.proto
```