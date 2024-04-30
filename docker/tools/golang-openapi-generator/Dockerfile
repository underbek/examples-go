FROM registry.com/base/golang as build

RUN go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@57a4b26d477f
RUN cp $GOPATH/bin/oapi-codegen /oapi-codegen

FROM registry.com/base/alpine

COPY --from=build /oapi-codegen /usr/local/bin/oapi-codegen
