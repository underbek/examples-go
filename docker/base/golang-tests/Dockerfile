FROM registry-proxy.com/docker

RUN apk add --no-cache git curl make musl-dev ca-certificates gcompat

RUN curl -SL https://go.dev/dl/go1.19.13.linux-amd64.tar.gz -o go1.19.13.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.19.13.linux-amd64.tar.gz

RUN export PATH=$PATH:/usr/local/go/bin && \
      export GOPRIVATE=gitlab.com && \
      export GO111MODULE=on && \
      export GOPATH=$HOME/go && \
      mkdir -p $HOME/go


ENV PATH=$PATH:/usr/local/go/bin
ENV GOPRIVATE=gitlab.com
ENV GOPATH=$HOME/go
ENV GO111MODULE=on
