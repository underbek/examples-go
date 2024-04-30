FROM php:8.0.27-fpm-alpine3.16

RUN rm -rf /var/www \
    && mkdir /var/www \
    && chown www-data: /var/www

USER www-data

WORKDIR /var/www
