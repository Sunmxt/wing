#! /usr/bin/env sh

set -xe
docker build -f docker/Dockerfile-minimum -t wing:`cat ./VERSION` .