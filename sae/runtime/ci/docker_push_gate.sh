sar_import lib.sh

SAR_CI_DOCKER_GATE_ORIGIN_BIN=`which docker`

docker() {
    if [ "push" = "$1" ]; then
        logerror "[CI_DOCKER_PUSH_GATE] should not push image. something went wrong. check your code."
        return 1
    fi
    "$SAR_CI_DOCKER_GATE_ORIGIN_BIN" $*
}