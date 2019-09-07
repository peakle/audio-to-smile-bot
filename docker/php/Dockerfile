FROM registry.wake-app.net/wakeapp/php:7.2-fpm

RUN apt-get update && apt install -y \
		libfreetype6-dev \
		libjpeg62-turbo-dev \
		libpng-dev \
		nghttp2 \
	&& docker-php-ext-configure gd --with-freetype-dir=/usr/include/ --with-jpeg-dir=/usr/include/ \
	&& docker-php-ext-install -j$(nproc) gd

RUN apt-get install sox
RUN apt-get install -yq mysql-client

WORKDIR /app