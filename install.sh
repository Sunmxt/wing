#! /bin/bash

SOURCING_ENTRYPOINT=/bin/runtime_env

_runtime_env_script_gen_add_import() {
    echo "sar_import runtime.sh" # Entrypoint
}

_runtime_env_script_gen_base() {
    local install_path=$1
    echo "export SAR_BIN_BASE=$install_path"
    echo '
sar_import() {
    local module=$1
    if [ -z "$module" ]; then
        return 1
    fi
    source "'"$install_path"'/$module"
    return $?
}
    '
}

_runtime_env_script_gen() {
    local install_path=$1
    _runtime_env_script_gen_base "$install_path"
    _runtime_env_script_gen_add_import
}

_ensure_directory() {
    local dir=$1

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

    local tmp_script=/tmp/SARTMP$random$random$random
    _runtime_env_script_gen "$bin_dir" > $tmp_script
    source $tmp_script
    rm -f $tmp_script
    source "$bin_dir/settings/component.sh"
    source "$bin_dir/lib.sh"

    loginfo starstudio runtime installing...


    if ! _ensure_directory $install_path; then
        return 1
    fi
    local install_path="`cd "$install_path" ; pwd -P `"

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
        cp -f "$bin" "$target" 
    done

    loginfo [install] generate runtime environment script: $SOURCING_ENTRYPOINT.
    if ! _runtime_env_script_gen "$install_path" > "$SOURCING_ENTRYPOINT"; then
        logerror [install] install runtime environment script failure.
        return 1
    fi

    loginfo runtime installed successfully.
}

if [ -z "$SAR_NOT_CALL_INSTALL" ]; then
    _runtime_install "$0" $*
fi
