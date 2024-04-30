FROM python:3.10.9-slim
RUN pip install poetry==1.4.2 \
    && mkdir -p /root/.config/pypoetry \
    && poetry config virtualenvs.in-project true \
    && poetry self add poetry-exec-plugin \
    && apt update && apt install -y gcc git