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

set_var() {
    eval "$1="\'"$2"\'
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

next_long_opt() {
    local "store_to=$1"
    if [ -z "$store_to" ]; then
        logerror next_long_opt: empty store variable name.
        return 1
    fi
    shift 1

    LONGOPTCANELIMATE=1
    local -i idxarg=$LONGOPTINDARG
    LONGOPTINDARG=0
    local -i idx=$LONGOPTIND
    idx=idx+idxarg
    while [ $idx -le $# ]; do
        idx=idx+1
        eval "local ____=\${$idx}"
        if [ "${____}" = "--" ]; then
            return 1
        fi
        if [ "${____:0:2}" = "--" ]; then
            eval "$store_to=\"${____:2}\""
            break
        fi
    done
    LONGOPTIND=$idx
    [ $idx -le $# ]
}

get_long_opt_arg() {
    local -i idxarg=$LONGOPTINDARG
    local -i idx=$LONGOPTIND
    local -i incre=$idxarg+1
    idx=idx+incre
    eval echo "\${$idx}"
    LONGOPTINDARG=$incre
    [ $idx -le $# ]
}

eliminate_long_opt() {
    if [ -z "$LONGOPTCANELIMATE" ]; then
        return
    fi
    local -i idxarg=$LONGOPTINDARG
    local -i idx=$LONGOPTIND
    local -i ref=idxarg+idx+1
    echo '
local -i __sar_idx=1;
local -a __sar_args=();
while test $__sar_idx -lt '$idx' ; do
    eval "__sar_args+=(\"\${$__sar_idx}\")";
    __sar_idx=__sar_idx+1;
done;
local -i __sar_ref='$ref';
while test $__sar_ref -le $# ; do
    eval "__sar_args+=(\"\${$__sar_ref}\")";
    __sar_ref=__sar_ref+1; __sar_idx=__sar_idx+1;
done;
set -- ${__sar_args[@]};'
    idx=idx-1
    echo "LONGOPTIND=$idx;"
    echo "unset LONGOPTCANELIMATE;"
}

binary_deps() {
    while [ $# -gt 0 ]; do
        if which "$1" 2>&1 > /dev/null || (set | grep -E '^'"$1"'\s+()' 2>&1 >/dev/null ); then
            shift
            continue
        fi
        MISSING_BINARY_DEP="$1"
        return 1
    done
    return 0
}