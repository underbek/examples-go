FROM openapitools/openapi-generator-cli:v7.0.1

RUN ln -s /usr/local/bin/docker-entrypoint.sh /usr/local/bin/openapi-generator

ENV NODE_MAJOR=20 \
    REDOCLY_TELEMETRY=off

RUN apt-get update \
    && apt-get install -y ca-certificates curl gnupg \
    && mkdir -p /etc/apt/keyrings \
    && curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key | gpg --dearmor -o /etc/apt/keyrings/nodesource.gpg \
    && echo "deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_$NODE_MAJOR.x nodistro main" | tee /etc/apt/sources.list.d/nodesource.list \
    && apt update  \
    && apt install -y nodejs \
    && apt-get clean \
    && npm i -g @redocly/cli@latest \
    && npm cache clean --force
