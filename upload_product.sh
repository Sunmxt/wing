#! /usr/bin/env bash

mkdir ./output -p
cp -v ./bin/wing ./output/
source controller/sae/runtime/bin/sar_activate
ci_build gitlab-package -e CI_COMMIT_REF_NAME ./output