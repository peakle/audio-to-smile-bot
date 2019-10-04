FROM php:7.2-fpm

RUN apt-get update && apt-get install -y \
		libfreetype6-dev \
		libjpeg62-turbo-dev \
		libpng-dev \
		nghttp2 \
	&& docker-php-ext-configure gd --with-freetype-dir=/usr/include/ --with-jpeg-dir=/usr/include/ \
	&& docker-php-ext-install pdo_mysql \
	&& docker-php-ext-install -j$(nproc) gd

RUN apt-get install -y sox && apt-get install unzip
RUN apt-get install supervisor && apt-get install nano
RUN apt-get install -y default-mysql-client
RUN curl -sS https://getcomposer.org/installer -o composer-setup.php &&  php composer-setup.php --install-dir=/usr/local/bin --filename=composer

WORKDIR /app
