FROM registry.com/base/alpine

RUN apk update \
  && apk add --no-cache openssh-client git

RUN echo 'StrictHostKeyChecking no' >> /etc/ssh/ssh_config
RUN git config --global user.email "ci@test.com" \
    && git config --global user.name "ci"