#! /bin/bash

source star_ci_lib

export k8s_target_image_tag=`docker_base_tag_by_env $env`
envsubst < ./k8s-deploy.template.yml > k8s.yml
kubectl apply -f k8s.yml --kubeconfig=/var/run/kube/config