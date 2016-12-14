# protoc-gen-ttndoc

Generate docs for TTN API

## Installation

```
go install
```

## Usage

```
protoc -I/usr/local/include -I$GOPATH/src -I$GOPATH/src/github.com/TheThingsNetwork -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --ttndoc_out=logtostderr=true,.handler.ApplicationManager=all:. $GOPATH/src/github.com/TheThingsNetwork/ttn/api/handler/handler.proto
```
