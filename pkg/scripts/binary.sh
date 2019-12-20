#! /use/bin/env bash

#SAR_BUNDLE_EXCLUDED{{
sar_import log.sh
sar_import utils.sh

SAR_BINARY_REGISTRY=()
SAR_BINARIES=()

sar_register_binary_help() {
    echo '
register executable binary.

usage:
    sar_register_binary [options] <exec_name> <arch> <os> <relative_path>

options:
    --no-bundled        do not bundle this binary.
    --lazy-load         lazy load binary. (used by bundled only)
    -h, --help          show help
'
}

sar_register_binary() {
    LONGOPTIND=0
    while next_long_opt opt $*; do
        case $opt in
            no-bundled)
                local no_bundled=1
                ;;
            lazy-load)
                local lazy_load=1
                ;;
            help)
                sar_register_binary_help
                return 0
                ;;
        esac
        eval `eliminate_long_opt`
    done
    
    OPTIND=0
    while getopts 'h' opt; do
        case $opt in
            h)
                sar_register_binary_help
                return 0
                ;;
            *)
                sar_register_binary_help
                logerror "unknown option \"""$opt""\"."
                return 1
        esac
    done
    local -i num_of_opt=$OPTIND-1
    if [ $num_of_opt -gt 0 ]; then
        shift $num_of_opt
    fi

    local exec=$1
    local arch=$2
    local machine=$3
    local ref="$4"

    test -z "$ref" && return 0


    eval "local cur_full_ref=\$SAR_BINARY_${exec}_${arch}_${machine}_full_ref"
    eval "SAR_BINARY_${exec}_${arch}_${machine}_ref=\"$ref\""
    if [ ! -z "$no_bundled" ]; then
        eval "SAR_BINARY_${exec}_${arch}_${machine}_no_bundled=1"
    fi
    if [ ! -z "$lazy_load" ]; then
        eval "SAR_BINARY_${exec}_${arch}_${machine}_lazy_load=1"
    fi
    eval "SAR_BINARY_${exec}_${arch}_${machine}_full_ref=\"$SAR_BIN_BASE/$ref\""

    eval "local cur_exec=\$SAR_BINARY_$exec"
    if [ -z "$cur_exec" ]; then
        eval "SAR_BINARY_$exec=1"
        eval "SAR_BINARY_${exec}_ARCH_OS=()"
        SAR_BINARIES+=("$exec")
    fi

    if [ -z "$cur_full_ref" ]; then
        SAR_BINARY_REGISTRY+=("${exec}_${arch}_${machine}")
        eval "SAR_BINARY_${exec}_ARCH_OS+=(\"${arch}_${machine}\")"
    fi
}

_sar_generate_binary_shim() {
    local exec=$1
    echo '
____sar_invoke_binary_'"${exec}"'() {
    if [ -z "${SAR_BINREF_'$exec'}" ]; then
        ____sar_lazy_load_binray_'${exec}' || return 1
    fi
    ${SAR_BINREF_'$exec'} $*
}

'$exec'() {
    ${SAR_'$exec'} $*
}
'
}

sar_binary_init() {
    local machine=`uname -m | tr -d '[:space:]' | tr '[:upper:]' '[:lower:]'`
    local arch=`uname | tr -d '[:space:]' | tr '[:upper:]' '[:lower:]'`
    for exec in $SAR_BINARIES; do
        eval "full_ref=\$SAR_BINARY_${exec}_${arch}_${machine}_full_ref"
        eval "declare -g SAR_$exec"
        if [ ! -z "$full_ref" ]; then
            eval "SAR_$exec=\"$full_ref\""
        else
            sar_binary_arch_unsupported ${exec} ${arch} ${machine}
            return 1
        fi
        eval "`_sar_generate_binary_shim $exec`"
    done
}

sar_binary_arch_unsupported() {
    local exec=$1
    local arch=$2
    local machine=$3

    logerror "$exec unsupported $arch $machine."
    return 1
}

#}}SAR_BUNDLE_EXCLUDED