#! /usr/bin/env sh
set -xe

# Build Dashboard
make dashboard/dist

# Build main executable
make bin/wing $MAKE_ENV_ARGV SKIP_FE_BUILD=1
chmod a+x bin/wing