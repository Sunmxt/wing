sar_import builder/common.sh
sar_import builder/ci.sh
sar_import builder/validate.sh
sar_import settings/image.sh
sar_import utils.sh

_runtime_image_stash_prefix() {
    local prefix=$1
    local env=$2
    local tag=$3
    `hash_for_key "$prefix" "$env" "$tag"`
}

_runtime_image_stash_prefix_by_context() {
    eval "local stash_prefix=\$_SAR_RT_BUILD_${context}_STASH_PREFIX"
    echo $stash_prefix
}

_generate_runtime_image_dockerfile_add_os_deps_alpine() {
    loginfo "[runtime_image_build] add os dependencies with apk."

    echo '
RUN set -xe;\
    mkdir -p /tmp/apk-cache;\
    [ ! -z "'`strip $SAR_RUNTIME_ALPINE_APK_MIRROR`'" ] && sed -Ei "s/dl-cdn\.alpinelinux\.org/'$SAR_RUNTIME_ALPINE_APK_MIRROR'/g" /etc/apk/repositories;\
    apk update --cache-dir /tmp/apk-cache;\
    apk add '${SAR_RUNTIME_ALPINE_DEPENDENCIES[@]}' --cache-dir /tmp/apk-cache;\
    rm -rf /tmp/apk-cache;\
    pip install '${SAR_RUNTIME_SYS_PYTHON_DEPENDENCIES[@]}'
'
}

_generate_runtime_image_dockerfile_add_os_deps_centos() {
    logerror "[runtime_image_builder] Centos will be supported soon."
}

_generate_runtime_image_dockerfile_add_os_deps_debian() {
    logerror "[runtime_image_builder] Debian will be supported soon."
}

_generate_runtime_image_dockerfile_add_supervisor_services() {
    loginfo "[runtime_image_build] add supervisor services."

    local supervisor_root_config=/tmp/supervisor-$RANDOM$RANDOM$RANDOM.ini

    echo '
[unix_http_server]
file=/run/supervisord.sock

[include]
files = /etc/supervisor.d/services/*.ini

[supervisord]
logfile=/var/log/supervisord.log
logfile_maxbytes=0           
loglevel=info                
pidfile=/run/runtime/supervisord.pid 
nodaemon=true                

[rpcinterface:supervisor]
supervisor.rpcinterface_factory = supervisor.rpcinterface:make_main_rpcinterface

[supervisorctl]
serverurl=unix:///run/supervisord.sock
' > "$supervisor_root_config"

    # generate services.
}

_generate_runtime_image_dockerfile() {
    local context=$1
    local package_ref=$2
    local package_env=$3
    local pakcage_tag=$4

    local build_id=$RANDOM$RANDOM$RANDOM$RANDOM

    eval "local -i deps_count=\${#_SAR_RT_BUILD_${context}_DEPS[@]}"
    local failure=0
    local -i idx=1
    while [ $idx -le $deps_count ]; do
        eval "local key=\${_SAR_RT_BUILD_${context}_DEPS[$idx]}"
        eval "local pkg_env_name=\${_SAR_RT_BUILD_${context}_DEP_${key}_ENV}"
        eval "local pkg_prefix=\${_SAR_RT_BUILD_${context}_DEP_${key}_PREFIX}"
        eval "local pkg_tag=\${_SAR_RT_BUILD_${context}_DEP_${key}_TAG}"
        local pkg_image_ref=`_ci_get_package_ref "$pkg_prefix" "$pkg_env_name" "$pkg_tag"`

        loginfo "[runtime_image_builder][pre_check] check package: $pkg_image_ref"
        if ! _validate_dependency_package "$pkg_prefix" "$pkg_prefix" "$pkg_tag"; then
            local failure=1
        fi

        local -i idx=idx+1
    done
    if [ $failure -ne 0 ]; then
        logerror "[runtime_image_builder]" dependency package validation failure.
        return 1
    fi

    # Multi-stage image layers.
    local -i idx=1
    while [ $idx -le $deps_count ]; do
        eval "local key=\${_SAR_RT_BUILD_${context}_DEPS[$idx]}"
        eval "local pkg_env_name=\${_SAR_RT_BUILD_${context}_DEP_${key}_ENV}"
        eval "local pkg_prefix=\${_SAR_RT_BUILD_${context}_DEP_${key}_PREFIX}"
        eval "local pkg_tag=\${_SAR_RT_BUILD_${context}_DEP_${key}_TAG}"
        local pkg_image_ref=`_ci_get_package_ref "$pkg_prefix" "$pkg_env_name" "$pkg_tag"`

        loginfo "[runtime_image_builder] package $pkg_image_ref used."
        echo "FROM $pkg_image_ref AS sar_stage_`hash_for_key $build_id $pkg_image_ref`"
        local -i idx=idx+1
    done

    eval "local base_image=\$_SAR_RT_BUILD_${context}_BASE_IMAGE"
    _validate_base_image "$base_image" || return 1
    echo "FROM $base_image" # 暂时假设 base image 里有构建需要的各种工具

    # Run pre-build scripts.

    # Place packages.
    local -i idx=1
    while [ $idx -le $deps_count ]; do
        eval "local key=\${_SAR_RT_BUILD_${context}_DEPS[$idx]}"
        eval "local pkg_env_name=\${_SAR_RT_BUILD_${context}_DEP_${key}_ENV}"
        eval "local pkg_prefix=\${_SAR_RT_BUILD_${context}_DEP_${key}_PREFIX}"
        eval "local pkg_tag=\${_SAR_RT_BUILD_${context}_DEP_${key}_TAG}"
        eval "local placed_path=\${_SAR_RT_BUILD_${context}_DEP_${key}_PLACE_PATH}"
        local pkg_image_ref=`_ci_get_package_ref "$pkg_prefix" "$pkg_env_name" "$pkg_tag"`

        loginfo "[runtime_image_builder] place package $pkg_image_ref --> $placed_path"
        echo "COPY --from=sar_stage_`hash_for_key $build_id $pkg_image_ref` /package/data \"$placed_path\""
        local -i idx=idx+1
    done

    # add system dependencies.
    local pkg_mgr=`determine_os_package_manager`
    case $pkg_mgr in
        apk)
            _generate_runtime_image_dockerfile_add_os_deps_alpine || return 1
            ;;
        yum)
            _generate_runtime_image_dockerfile_add_os_deps_centos || return 1
            ;;
        apt)
            _generate_runtime_image_dockerfile_add_os_deps_debian || return 1
            ;;
        *)
            logerror "[runtime_image_builder] unsupported package manager type: $pkg_mgr"
            return 1
            ;;
    esac

    _generate_runtime_image_dockerfile_add_supervisor_services

    # save runtime image metadata and install runtime.
    echo '
RUN set -xe;\
    [ -d '\''/_sar_package/runtime_install'\'' ] && (echo install runtime; bash /_sar_package/runtime_install/install.sh /opt/runtime );\
    mkdir -p /_sar_package;\
    echo PKG_REF='\\\'$package_ref\\\'' > /_sar_package/meta;\
    echo PKG_ENV='\\\'$package_env\\\'' >> /_sar_package/meta;\
    echo PKG_TAG='\\\'$pakcage_tag\\\'' >> /_sar_package/meta;\
    echo PKG_TYPE=runtime_image >> /_sar_package/meta;\
    echo PKG_APP_NAME='\\\'$application_name\\\'' >> /_sar_package/meta;

CMD ["supervisord"]

'
}

build_runtime_image_help() {
    echo '
Build runtime image.

usage:
    build_runtime_image <build_mode> [options] -t <tag> -e <environment_varaible_name> -r prefix -- [docker build options]

mode:
  docker
  gitlab-docker

options:
    -c <context_name>       specified build context. default: system

example:
    build_runtime_image -t latest -e ENV -r registry.stuhome.com/mine/myproject
'
}

build_runtime_image() {
    OPTIND=0
    while getopts 't:e:r:c:' opt; do
        case $opt in
            t)
                local ci_image_tag=$OPTARG
                ;;
            e)
                local ci_image_env_name=`_ci_get_env_value "$OPTARG"`
                ;;
            r)
                local ci_image_prefix=$OPTARG
                ;;
            c)
                local context=$OPTARG
                ;;
            *)
                runtime_image_help
                logerror "[runtime_image_builder]" unexcepted options -$opt.
                ;;
        esac
    done

    eval "local __=\$$OPTIND"
    local -i optind=$OPTIND
    if [ "$__" != "--" ]; then
        if [ ! -z "$__" ]; then
            logerror "[runtime_image_builder] build_runtime_image: got unexcepted non-option argument: \"$__\"."
            return 1
        fi
        local -i optind=optind-1
    fi

    if [ -z "$context" ]; then
        local context=system
    fi

    if [ -z "$ci_image_tag" ]; then
        logerror "[runtime_image_builder]" empty runtime image tag.
        return 1
    fi
    if [ -z "$ci_image_prefix" ]; then
        logerror "[runtime_image_builder]" empty runtime image prefix.
        return 1
    fi

    # add runtime
    runtime_image_add_dependency -c "$context" -r "$SAR_RUNTIME_PKG_PREFIX" -e "$SAR_RUNTIME_PKG_ENV" -t "$SAR_RUNTIME_PKG_TAG" /_sar_package/runtime_install

    local dockerfile=/tmp/Dockerfile-RuntimeImage-$RANDOM$RANDOM$RANDOM
    if ! _generate_runtime_image_dockerfile "$context" "$ci_image_prefix" "$ci_image_env_name" "$ci_image_tag" > "$dockerfile" ; then
        logerror "[runtime_image_builder]" generate runtime image failure.
        return 1
    fi


    local opts=
    local -i idx=1
    while [ $idx -le $optind ]; do
        eval "local opts=\"\$opts \$$idx\""
        local -i idx=idx+1
    done
    shift $optind
    
    eval "log_exec _ci_docker_build $opts -- -f \"$dockerfile\" $* ."
}

runtime_image_base_image_help() {
    echo '
Set base image of runtime image.

usage:
    build_runtime_image_base_image [options] <image reference>

options:
    -c <context_name>       specified build context. default: system

example:
    build_runtime_image_base_image alpine:3.7
    build_runtime_image_base_image -c context2 alpine:3.7
'
}

runtime_image_base_image() {
    OPTIND=0
    while getopts 'c:' opt; do
        case $opt in
            c)
                local context=$OPTARG
                ;;
            *)
                runtime_image_base_image_help
                logerror "[runtime_image_builder]" unexcepted options -$opt.
                ;;
        esac
    done
    eval "local base_image=\$$OPTIND"
    if [ -z "$base_image" ]; then
        logerror "[runtime_image_builder] base image not specifed."
        return 1
    fi

    if [ -z "$context" ]; then
        local context=system
    fi
    eval "_SAR_RT_BUILD_${context}_BASE_IMAGE=$base_image"
}

runtime_image_add_dependency_help() {
    echo '
Add package dependency. Packages will be placed to during building runtime image.

usage:
    runtime_image_add_dependency -t <tag> -e <environment_varaible_name> -r prefix [options] <path>

options:
    -c <context_name>       specified build context. default: system

example:
    runtime_image_add_dependency -t c3adea1d -e ENV -r registry.stuhome.com/be/recruitment-fe /app/statics
'
}

runtime_image_add_dependency() {
    OPTIND=0
    while getopts 't:e:r:c:' opt; do
        case $opt in
            t)
                local ci_package_tag=$OPTARG
                ;;
            e)
                local ci_package_env_name=$OPTARG
                ;;
            r)
                local ci_package_prefix=$OPTARG
                ;;
            c)
                local context=$OPTARG
                ;;
            *)
                runtime_image_add_dependency_help
                logerror "[runtime_image_builder]" unexcepted options -$opt.
                ;;
        esac
    done
    if [ -z "$context" ]; then
        local context=system
    fi
    eval "local place_path=\$$OPTIND"
    if [ -z "$place_path" ]; then
        runtime_image_add_dependency_help
        logerror "[runtime_image_builder] runtime_image_add_dependency: Target path cannot be empty."
        return 1
    fi
    local dependency_key=`hash_for_key $ci_package_prefix "$ci_package_env_name" "$ci_package_tag"`
    eval "_SAR_RT_BUILD_${context}_DEP_${dependency_key}_ENV=$ci_package_env_name"
    eval "_SAR_RT_BUILD_${context}_DEP_${dependency_key}_PREFIX=$ci_package_prefix"
    eval "_SAR_RT_BUILD_${context}_DEP_${dependency_key}_TAG=$ci_package_tag"
    eval "_SAR_RT_BUILD_${context}_DEP_${dependency_key}_PLACE_PATH=$place_path"
    eval "local -i dep_count=\${#_SAR_RT_BUILD_${context}_DEPS[@]}+1"
    eval "_SAR_RT_BUILD_${context}_DEPS[$dep_count]=$dependency_key"
}

runtime_image_bootstrap_run() {
    return 0
}

runtime_image_pre_build_run() {
    return 0
}

runtime_image_post_build_run() {
    return 0
}

runtime_image_pre_build_script() {
    return 0
}

runtime_image_post_build_script() {
    return 0
}

runtime_image_health_check_script() {
    return 0
}