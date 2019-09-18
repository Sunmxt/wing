#! /usr/bin/env sh
set -xe

# Build Dashboard
make bin/dashboard

# Build main executable
make bin/wing $MAKE_ENV_ARGV SKIP_FE_BUILD=1
chmod a+x bin/wing