FROM registry.com/base/python

RUN apt -yq update \
    && apt -yq upgrade \
    && apt -yq install curl wget \
    && wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb \
    && apt -yq install ./google-chrome-stable_current_amd64.deb \
    && pip install poetry
