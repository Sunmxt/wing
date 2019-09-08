sar_import log.sh

path_join() {
    local -i idx=1
    local result=
    while [ $idx -le $# ]; do
        eval "local current=\$$idx"
        local -i idx=idx+1
        if [ -z "$result" ] || [ "${current:0:1}" = "/" ] ; then
            local result=$current
            local current=
        fi

        local dir=`dirname "$result/$current"`
        local base=`basename "$result/$current"`
        if [ "$dir" = '/' ]; then
            local dir=
        fi
        local result=$dir/$base
    done
    echo $result
}

hash_for_key() {
    echo "$*" | md5sum - | cut -d ' ' -f 1
}

hash_file_for_key() {
    cat $* | md5sum - | cut -d ' ' -f 1
}

strip() {
    echo "$*" | xargs
}

determine_os_package_manager() {
    echo apk
}

full_path_of() {
    echo `cd "$1"; pwd -P `
}

hash_content_for_key() {
    local target=$1
    if [ -d "$target" ]; then
        find "$target" -type f | xargs cat | md5sum - | cut -d ' ' -f 1
    elif [ -f "$target" ]; then
        cat "$target" | md5sum - | cut -d ' ' -f 1
    elif [ -e "$target" ]; then
        logerror \"$target\" not exists.
        return 1
    else
        logerror unknown file type of \"$target\"
        return 1
    fi
}

case "$SAR_CURRENT_PLATFORM" in
    darwin)
        SAR_YQ_BIN=`path_join "$SAR_BIN_BASE" libexec/yq_darwin_amd64`
        ;;
    linux)
        SAR_YQ_BIN=`path_join "$SAR_BIN_BASE" libexec/yq_linux_amd64`
        ;;
    *)
        logwarn "yq doesn't support current platform: $SAR_CURRENT_PLATFORM."
        ;;
esac

yq() {
    "$SAR_YQ_BIN" $*
}