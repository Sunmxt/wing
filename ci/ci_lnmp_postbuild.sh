#! /bin/bash
set -xe
mkdir /apk-cache
apk update --cache-dir /apk-cache 
apk add bash jq procps gettext vim nginx supervisor freetype wget tar curl grep zlib libxml2 readline openssl libjpeg-turbo libpng libmcrypt libwebp icu libpng-dev libwebp-dev libmcrypt-dev openldap-dev libmemcached-dev curl-dev --cache-dir /apk-cache
rm -f /etc/nginx/conf.d/default.conf
cd /
wget https://getcomposer.org/download/1.9.0/composer.phar -P composer-setup.php
php -r "if (hash_file('sha384', 'composer-setup.php') === 'a5c698ffe4b8e849a443b120cd5ba38043260d5c4023dbf93e1558871f1f07f58274fc6f4c93bcfd858c6bd0775cd8d1') { echo 'Installer verified'; } else { echo 'Installer corrupt'; unlink('composer-setup.php'); } echo PHP_EOL;"
php composer-setup.php --install-dir=/bin --filename=composer
php -r "unlink('composer-setup.php');"
mkdir -p /usr/src/php/ext /run/nginx
wget https://pecl.php.net/get/yac-2.0.2.tgz
tar -xvf yac-2.0.2.tgz -C /usr/src/php/ext/
mv /usr/src/php/ext/yac-2.0.2 /usr/src/php/ext/yac
apk add -t build-deps autoconf gcc g++ make automake linux-headers --cache-dir /apk-cache
pecl install mongodb redis memcached mcrypt-1.0.2
docker-php-ext-install pdo mysqli pdo curl fileinfo gd json ldap mbstring pdo_mysql zip yac
docker-php-ext-enable mongodb redis memcached mcrypt
apk del build-deps
rm -rf /apk-cache
