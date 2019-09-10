sar_import log.sh 

DOCKERCLI_CONFIG=~/.docker/config.json
if ! [ -e ~/.docker ]; then
    mkdir ~/.docker
fi

is_dockercli_experimentals_enabled() {
    local state=`jq '.experimental' "$DOCKERCLI_CONFIG"`
    if [ "$state" = "enabled" ]; then
        return 0
    fi
    return 1
}

enable_dockercli_experimentals() {
    # temp file is for the stupid jq beheavior.
    if is_dockercli_experimentals_enabled; then
        return
    fi
    if ! [ -e "$DOCKERCLI_CONFIG" ] ; then
        echo "{}" > "$DOCKERCLI_CONFIG"
    fi
    TMP=/tmp/enable_dockercli_experimentals$RANDOM$RANDOM
    
    if ! cat $DOCKERCLI_CONFIG | jq 'setpath(["experimental"];"enabled")' >> "$TMP"; then
        logerror cannot enable docker experimental functions.
        return 1
    fi
    cat "$TMP" > "$DOCKERCLI_CONFIG"
}

disable_dockercli_experimentals() {
    if ! is_dockercli_experimentals_enabled; then
        return
    fi
    if ! [ -e "$DOCKERCLI_CONFIG" ] ; then
        echo "{}" > "$DOCKERCLI_CONFIG"
    fi
    TMP=/tmp/enable_dockercli_experimentals$RANDOM$RANDOM
    if cat $DOCKERCLI_CONFIG | jq 'setpath(["experimental"];"disabled")' >> "$TMP"; then
        logerror cannot disable docker experimental functions.
        return 1
    fi
    cat "$TMP" > "$DOCKERCLI_CONFIG"
}

is_image_exists() {
    enable_dockercli_experimentals
    docker manifest inspect $1 2>&1 >/dev/null
    return $?
}


docker_installed() {
    docker version -f '{{ .Client.Version }}' 2>&1 >/dev/null
    return $?
}