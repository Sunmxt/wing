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