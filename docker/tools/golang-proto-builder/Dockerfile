FROM registry.com/base/golang-alpine

RUN apk update \
  && apk add protobuf-dev make \
  && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
  && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
