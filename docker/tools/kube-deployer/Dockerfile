FROM registry.com/base/alpine

RUN apk update \
    && apk add --no-cache curl \
        git \
        openssh-client

RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" \
    && install -m 555 kubectl /usr/local/bin/kubectl \
    && rm kubectl

RUN wget https://get.helm.sh/helm-v3.11.1-linux-amd64.tar.gz \
    && tar -zxvf helm-v3.11.1-linux-amd64.tar.gz \
    && install -m 555 linux-amd64/helm /usr/local/bin/helm \
    && rm -r linux-amd64

RUN wget https://github.com/mikefarah/yq/releases/download/v4.28.1/yq_linux_amd64 -O /usr/local/bin/yq \
  && chmod +x /usr/local/bin/yq