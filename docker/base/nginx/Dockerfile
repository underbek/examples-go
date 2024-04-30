FROM nginx:1.23.2-alpine

ADD configs/nginx/probe.conf /etc/nginx/
ADD configs/nginx/fpm-status.conf /etc/nginx/
ADD configs/nginx/nginx.conf /etc/nginx/

RUN rm -rf /var/www \
    && mkdir /var/www \
    && mkdir /run/nginx \
    && rm -rf /etc/nginx/conf.d/* \
    && chown -R nginx: /var/www \
    && chown -R nginx: /var/cache/nginx \
    && chown -R nginx: /var/log/nginx \
    && chown -R nginx: /run/nginx

WORKDIR /var/www
USER nginx

EXPOSE 8090
EXPOSE 8091