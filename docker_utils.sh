DOCKERCLI_CONFIG=~/.docker/config.json
if ! [ -e ~/.docker ]; then
    mkdir ~/.docker
fi

enable_dockercli_experimentals() {
    # temp file is for the stupid jq beheavior.
    if ! [ -e "$DOCKERCLI_CONFIG" ] ; then
        echo "{}" > "$DOCKERCLI_CONFIG"
    fi
    TMP=/tmp/enable_dockercli_experimentals$RANDOM$RANDOM
    cat $DOCKERCLI_CONFIG | jq 'setpath(["experimental"];"enabled")' >> "$TMP"
    cat "$TMP" > "$DOCKERCLI_CONFIG"
}

disable_dockercli_experimentals() {
    if ! [ -e "$DOCKERCLI_CONFIG" ] ; then
        echo "{}" > "$DOCKERCLI_CONFIG"
    fi
    TMP=/tmp/enable_dockercli_experimentals$RANDOM$RANDOM
    cat $DOCKERCLI_CONFIG | jq 'setpath(["experimental"];"disabled")' >> "$TMP"
    cat "$TMP" > "$DOCKERCLI_CONFIG"
}

is_image_exists() {
    docker manifest inspect $1 2>&1 >/dev/null
    return $?
}