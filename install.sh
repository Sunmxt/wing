#! /usr/bin/env bash

_ensure_directory() {
    local dir=$1
    if [ -z "$dir" ] || [ "$dir" = "." ] || [ "$dir" = ".." ]; then
        return
    fi 

    if ! [ -d "$dir" ]; then
        loginfo [pre_install] create dictionary: $dir.
        if ! mkdir -p "$dir"; then
            logerror [pre_install] cannot make dictionary: $dir.
            return 1
        fi
    fi
}


_runtime_install() {
    local install_path=$2
    local bin_dir="`cd "$(dirname "$1")" ; pwd -P `"

    SAR_BIN_BASE="$bin_dir" source "$bin_dir/bin/sar_activate"
    source "$bin_dir/settings/component.sh"
    source "$bin_dir/settings/install.sh"
    source "$bin_dir/lib.sh"

    loginfo starstudio runtime installing...

    if ! _ensure_directory "$install_path"; then
        return 1
    fi

    if ! _ensure_directory "$INSTALL_SYSTEM_BINARY_DIR"; then
        return 1
    fi
    local install_path="`cd "$install_path" ; pwd -P `"

    local -i idx=${#COMPONENTS[@]}
    for bin in `cd "$bin_dir/bin"; find . -type f | sed 's/^\.\//bin\//g'`; do
        eval "COMPONENTS[$idx]=\"$bin\""
        local -i idx=idx+1
    done

    for relative in "${COMPONENTS[@]}"; do 
        local bin=`path_join "$bin_dir" "$relative"`
        local target=`path_join "$install_path" "$relative"`
        loginfo [check] $bin
        if [ ! -e "$bin" ]; then
            loginfo "\"$bin\" doesn't exist"
            return 1
        fi

        if ! _ensure_directory `dirname "$target"`; then
            return 1
        fi
    done

    for relative in "${COMPONENTS[@]}"; do
        local bin=`path_join "$bin_dir" "$relative"`
        local target=`path_join "$install_path" "$relative"`
        loginfo [install] "$bin --> $target"
        if [ ! -e "$bin" ]; then
            loginfo [install] "\"$bin\" doesn't exist"
            return 1
        fi
        cp -rf "$bin" "$target" 
    done

    local installed_cmd_path="$install_path/bin"
    for cmd in "${EXPORT_COMMANDS[@]}"; do
        local cmd_bin_path="$installed_cmd_path/$cmd"
        if [ -e "$cmd_bin_path" ]; then
            logwarn "[install] static command binary overrides export command \"$cmd\"."
            continue
        fi
        loginfo [install] export command: $cmd
        echo '#! /bin/bash
sar_execute '$cmd' $* 
' > "$cmd_bin_path"
    done

    for cmd in `cd "$installed_cmd_path"; find . -type f | sed 's/^\.\///g'`; do
        local cmd_source=`path_join "$installed_cmd_path" "$cmd"`
        local cmd_link_to=`path_join "$INSTALL_SYSTEM_BINARY_DIR" "$cmd"`
        chmod a+x "$cmd_source"
        loginfo "[install] create symbol link: $cmd_source -> $cmd_link_to" 
        if ! ln -s "$cmd_source" "$cmd_link_to"; then
            logerror "[install] create symbol link failure."
            return 1
        fi
    done

    local cmd_source=`path_join "$installed_cmd_path" "sar_activate"`
    local cmd_link_to=`path_join "$INSTALL_SYSTEM_BINARY_DIR" "runtime_env"`
    loginfo "[install] create symbol link: $cmd_source -> $cmd_link_to" 
    if ! ln -s "$cmd_source" "$cmd_link_to"; then
        logerror "[install] create symbol link failure."
        return 1
    fi

    loginfo runtime installed successfully.
}

if [ -z "$SAR_NOT_CALL_INSTALL" ]; then
    _runtime_install "$0" $*
fi
