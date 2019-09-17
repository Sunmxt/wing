#! /bin/bash

source star_ci_lib

export ENV=$1
export IMAGE_BASE=`docker_base_tag_by_env $1`
starconf_set_entry remote https://config.stuhome.com/production/wing/wing.json -t $STAGING_TOKEN

if [ -z "$IMAGE_BASE" ] || [ -z "$ENV" ]; then
    exit 1
fi

export IMAGE_BASE=$IMAGE_BASE:${CI_COMMIT_SHA:0:10}

star_boot envsubst < ./k8s-deploy.template.yml > k8s.yml
kubectl cluster-info
kubectl apply -f k8s.yml
