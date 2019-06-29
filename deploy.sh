#! /bin/bash

source star_ci_lib

export ENV=$1
export IMAGE_BASE=`docker_base_tag_by_env $1`

if [ -z "$IMAGE_BASE" ] || [ -z "$ENV" ]; then
    exit 1
fi

envsubst < ./k8s-deploy.template.yml > k8s.yml
kubectl cluster-info
kubectl apply -f k8s.yml
