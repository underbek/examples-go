FROM registry.com/base/alpine

RUN apk update \
  && apk add --no-cache curl \
    wget \
    zip

RUN wget https://github.com/mikefarah/yq/releases/download/v4.28.1/yq_linux_amd64 -O /usr/local/bin/yq \
  && chmod +x /usr/local/bin/yq