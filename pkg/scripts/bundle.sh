# SAR_LAZY_LOAD_ROOT=
# SAR_LAZY_LOAD_TYPE=
SAR_LAZY_LOAD_TYPE=${SAR_LAZY_LOAD_TYPE:=local}

_sar_resolve_lazy_load_path() {
    SAR_LAZY_LOAD_ROOT="$1"
    if [ -z "$SAR_LAZY_LOAD_ROOT" ]; then
        return 1
    fi

    while [ -L "$SAR_LAZY_LOAD_ROOT" ]; do
        SAR_LAZY_LOAD_ROOT=`readlink "$SAR_LAZY_LOAD_ROOT"`
    done

    SAR_LAZY_LOAD_ROOT="$( cd "$(dirname "$SAR_LAZY_LOAD_ROOT")" && pwd -P || echo '' )"
    if [ -z "$SAR_LAZY_LOAD_ROOT" ]; then
        return 1
    fi
}
if [ "$SAR_LAZY_LOAD_TYPE"="local" ] && [ -z "$SAR_BIN_BASE" ] && ! _sar_resolve_lazy_load_path "${BASH_SOURCE[0]}" && ! _sar_resolve_lazy_load_path `which "$0"`; then
    logerror "[bundle] cannot lazy loading path not detected."
fi


#SAR_BUNDLE_EXCLUDED{{
sar_import log.sh
sar_import utils.sh
sar_import binary.sh
sar_import settings/bundle.sh

__sar_bundle_preprocess() {
    sed -E '
    /^#SAR_BUNDLE_EXCLUDED\{\{/,/^#\}\}SAR_BUNDLE_EXCLUDED/d;
    s/^[[:space:]]+(sar_import.*)$/\1/;
    s/^[[:space:]]+(sar_register_binary.*)$/\1/;
    s/^[[:space:]]+(#.*)$/\1/;

    $s/^(.*)$/\1\n/;

    /^#/d;
    /^sar_import/d;
    /^sar_register_binary/d;
    /^$/d'

}

sar_base64_bundle_binary() {
    cat "$1" | base64 | tr -d '\n'
}

sar_bundle_binaries() {
    local bundle_root=$1

    echo "# binary bundles"
    echo '
SAR_machine=`uname -m | tr -d '\''[:space:]'\'' | tr '\''[:upper:]'\'' '\''[:lower:]'\''`
SAR_arch=`uname -o | tr -d '\''[:space:]'\'' | tr '\''[:upper:]'\'' '\''[:lower:]'\''`
SAR_arch_os=${SAR_arch}_${SAR_machine}
unset SAR_machine
unset SAR_arch
'
    echo "mkdir -p \"$BUNDLE_BINARY_DIR\""
    local has_lazy_loading

    for exec in ${SAR_BINARIES[@]}; do
        eval "local num_of_arch_os=\${#SAR_BINARY_${exec}_ARCH_OS[@]}"
        echo "# binary: $exec"

        if [ $num_of_arch_os -lt 1 ]; then
            continue
        fi

        # filtering
        eval "local arch_oss=(\"\${SAR_BINARY_${exec}_ARCH_OS[@]}\")"
        local -a lazy_arch_oss=()
        local -a online_arch_oss=()
        for arch_os in ${arch_oss[@]}; do
            eval "local no_bundled=\${SAR_BINARY_${exec}_${arch_os}_no_bundled}"

            # skip
            if [ ! -z "$no_bundled" ]; then
                loginfo "[bundle] binary[${arch_os}](skipped): ${exec}"
                continue
            fi

            eval "local lazy_load=\${SAR_BINARY_${exec}_${arch_os}_lazy_load}"
            if [ -z "$lazy_load" ]; then
                online_arch_oss+=("$arch_os")
            else
                lazy_arch_oss+=("$arch_os")
            fi
        done

        # shim to invoke binary.
        _sar_generate_binary_shim "$exec"

        # bundle lazy-load binary.
        echo '
____sar_lazy_load_binray_'${exec}'() {
    local dst="'$BUNDLE_BINARY_DIR/$exec'"
    case ${SAR_arch_os} in
'
        for arch_os in ${lazy_arch_oss[@]}; do
            local has_lazy_loading=1

            # write loading script
            loginfo "[bundle] binary[$arch_os](lazy-loaded): ${exec}"
            local lazy_load_script=`path_join "${bundle_root}" "sar_binary_${exec}_${arch_os}.sh"`
            eval "local full_ref=\${SAR_BINARY_${exec}_${arch_os}_full_ref}"

            echo '#! /usr/bin/env bash
echo '\'"`sar_base64_bundle_binary \"$full_ref\"`"\'' | base64 -d > "'$BUNDLE_BINARY_DIR/$exec'"
' > "$lazy_load_script"


            echo '
        '"${arch_os}"')
            loginfo loading binary: '${exec}' '${arch_os}'
            if ! ____sar_lazy_load_binary_impl "'${exec}'" "'${arch_os}'" | bash; then
                logerror '${exec}' '${arch_os}' not loaded properly.
                return 1
            fi
            ;;
'
        done

        echo '
        *)
            logerror '${exec}' ${SAR_arch_os} unsupported.
            return 1
            ;;
    esac

    if ! [ -e "$dst" ] || ! chmod a+x "$dst"; then
        logerror '${exec}' ${SAR_arch_os} not loaded properly.
        return 1
    else
        SAR_BINREF_'$exec'="'"$BUNDLE_BINARY_DIR/$exec"'"
    fi
}
'

        # bundle normal binary.
        if [ ${#online_arch_oss[@]} -gt 0 ]; then
            echo '
case ${SAR_arch_os} in
'
            for arch_os in ${online_arch_oss[@]}; do
                eval "local full_ref=\${SAR_BINARY_${exec}_${arch_os}_full_ref}"

                loginfo "[bundle] binary[$arch_os]: ${exec}"
                echo '
'"$arch_os"')
    (echo '\'"`sar_base64_bundle_binary \"$full_ref\"`"\'' | base64 -d > "'$BUNDLE_BINARY_DIR/$exec'") && SAR_BINARY_EXTRACTED=1
    ;;
'
            done

            echo '
*)
    logerror '${exec}' ${SAR_arch_os} unsupported.
    ;;
esac
if [ -z "$SAR_BINARY_EXTRACTED" ]; then
    logerror '${exec}' ${SAR_arch_os} not load properly.
else
    chmod a+x "'"$BUNDLE_BINARY_DIR/$exec"'"
    SAR_BINREF_'$exec'="'"$BUNDLE_BINARY_DIR/$exec"'"
fi
unset SAR_BINARY_EXTRACTED
'
        fi

echo '
SAR_'"${exec}"'=____sar_invoke_binary_'${exec}'
'
    done

    # lazy-load utils.
    if [ ! -z "$has_lazy_loading" ]; then
        echo '
____sar_lazy_load_binary_impl() {
    local exec=$1
    local arch_os=$2
    local ref="$SAR_LAZY_LOAD_ROOT/sar_binary_${exec}_${arch_os}.sh"
    case $SAR_LAZY_LOAD_TYPE in
        http)
            curl -L "$ref" -o -
            return $?
            ;;
        local)
            cat "$ref"
            return $?
            ;;
        *)
            logerror "unknown lazy loading type: ${SAR_LAZY_LOAD_TYPE}"
            return 1
            ;;
    esac
}'
    fi

}

sar_bundle_help() {
    echo '
generate script bundle.

usage:
    sar_bundle <bundle_root> <entrypoint_path>
'
}

sar_bundle() {
    local bundle_root=$1
    local entrypoint_path=$2

    if [ -z "$bundle_root" ]; then
        sar_bundle_help
        logerror "[bundle] bundle root cannot be empty."
        return 1
    fi

    if ! mkdir -p "$bundle_root"; then
        logerror "[bundle] cannot encure bundle root."
        return 1
    fi

    local -i idx=0

    entrypoint_path=${entrypoint_path:=main.sh}
    loginfo '[bundle] start build bundle.'
    loginfo '[bundle] --> root: '"$bundle_root"
    loginfo '[bundle] --> entrypoint: '"$entrypoint_path"

    entrypoint_path=`path_join "$bundle_root" "$entrypoint_path"`

    echo '# /usr/bin/env bash' > "$entrypoint_path"
    while [ $idx -lt ${#SAR_IMPORT_TRACE[@]} ]; do
        eval "local src=\${SAR_IMPORT_TRACE[$idx]}"
        eval "local mod=\${SAR_IMPORT_MODULE_TRACE[$idx]}"
        idx+=1

        # get source hash
        local src_hash=`md5sum "$src" | cut -d ' ' -f 1`
 
        # check whether the source is bundled.
        eval "local bundled=\$src_bundled_$src_hash"
        if [ ! -z "$bundled" ]; then
            continue
        fi

        # bundle
        loginfo "[bundle] script: $mod"
        echo "# bundle: $mod" >> "$entrypoint_path"

        # bundle step 1: preprocess 
        cat "$src" | __sar_bundle_preprocess >> "$entrypoint_path"

        eval "local src_bundled_${src_hash}=1"
    done
    echo "# [bundle] binaries." >> "$entrypoint_path"
    sar_bundle_binaries "$bundle_root" | __sar_bundle_preprocess >> "$entrypoint_path" || return 1
}

#}}SAR_BUNDLE_EXCLUDED