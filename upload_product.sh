#! /usr/bin/env bash

mkdir ./output -p
cp -v ./bin/wing ./output/
sae/runtime/install.sh /opt/runtime
source runtime_env
ci_build gitlab-package -e CI_COMMIT_REF_NAME ./output