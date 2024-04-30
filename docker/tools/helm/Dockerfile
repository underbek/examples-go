FROM registry.com/base/alpine

RUN apk update \
  && apk add --no-cache curl wget

RUN wget https://get.helm.sh/helm-v3.10.0-linux-amd64.tar.gz -O - \
  | tar xz && mv linux-amd64/helm /usr/local/bin/

RUN wget https://github.com/mikefarah/yq/releases/download/v4.28.1/yq_linux_amd64 -O /usr/local/bin/yq \
  && chmod +x /usr/local/bin/yq