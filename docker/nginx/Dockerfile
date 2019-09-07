FROM nginx:1.11-alpine

RUN set -x \
    && addgroup -g 82 -S www-data \
    && adduser -u 82 -D -S -G www-data www-data

RUN mkdir -p /etc/nginx/sites-available/ /etc/nginx/sites-enabled/

ADD nginx.conf /etc/nginx/
ADD www.conf /etc/nginx/sites-available/

RUN ln -s /etc/nginx/sites-available/www.conf /etc/nginx/sites-enabled/

RUN echo "upstream php-upstream { server php:9000; }" > /etc/nginx/conf.d/upstream.conf

CMD ["nginx"]

EXPOSE 80
EXPOSE 443