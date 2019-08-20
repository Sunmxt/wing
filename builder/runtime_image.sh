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

_generate_supervisor_system_service() {
    local name=$1
    local workdir=$2
    shift 2
    local exec="$*"
    echo '
[program:'$name']
command='$exec'
startsecs=20
autorestart=true
stdout_logfile=/var/log/application/'$name.log'
stderr_logfile=/var/log/application/'$name.err.log'
directory='$workdir'
'
}

_generate_supervisor_cron_service() {
    local name=$1
    local cron=$2
    local workdir=$3
    shift 3
    local exec="$*"
    echo '
'
}

_generate_supervisor_normal_service() {
    _generate_supervisor_system_service $*
    return $?
}

_generate_runtime_image_dockerfile_add_supervisor_services() {
    local context=$1
    loginfo "[runtime_image_build] add supervisor services."

    local supervisor_root_config=supervisor-$RANDOM$RANDOM$RANDOM.ini

    echo '
[unix_http_server]
file=/run/supervisord.sock

[include]
files = /etc/supervisor.d/services/*.conf

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

    echo "COPY $supervisor_root_config /etc/sar_supervisor.conf"

    local making_dirs="/var/log/application /run/runtime"
    # generate services.
    eval "local -i count=\${#_SAR_RT_BUILD_${context}_SVCS[@]}"
    local -i idx=1
    while [ $idx -le $count ]; do
        eval "local key=\${_SAR_RT_BUILD_${context}_SVCS[$idx]}"
        eval "local type=\${_SAR_RT_BUILD_${context}_SVC_${key}_TYPE}"
        eval "local name=\${_SAR_RT_BUILD_${context}_SVC_${key}_NAME}"
        eval "local exec=\${_SAR_RT_BUILD_${context}_SVC_${key}_EXEC}"
        eval "local working_dir=\${_SAR_RT_BUILD_${context}_SVC_${key}_WORKING_DIR}"
        if [ -z "$working_dir" ]; then
            local working_dir=$SAR_RUNTIME_APP_DEFAULT_WORKING_DIR
        fi
        local making_dirs="$making_dirs '$working_dir'"

        local file="supervisor-svc-$key.conf"
        case $type in
            cron)
                eval "locak cron=\${_SAR_RT_BUILD_${context}_SVC_${key}_CRON}"
                if ! _generate_supervisor_cron_service "$name" "$cron" "$working_dir" $exec > "$file"; then
                    logerror "[runtime_image_builder] Generate cronjob $name configuration failure." 
                    return 1
                fi
                ;;
            system)
                if ! _generate_supervisor_system_service "$name" "$working_dir" $exec > "$file"; then
                    logerror "[runtime_image_builder] Generate system service $name configuration failure." 
                    return 1
                fi
                ;;
            normal)
                if ! _generate_supervisor_normal_service "$name" "$working_dir" $exec > "$file"; then
                    logerror "[runtime_image_builder] Generate normal service $name configuration failure." 
                    return 1
                fi
                ;;
            *)
                logerror "[runtime_image_builder] Unsupported service type."
                return 1
                ;;
        esac
        local svc_files="$svc_files $file"
        local -i idx=idx+1
    done
    if [ ! -z "$svc_files" ]; then
        echo "COPY $svc_files /etc/supervisor.d/services/"
    fi
    if [ ! -z "$making_dirs" ]; then
        echo '
RUN set -xe;\
    mkdir -p '`echo $making_dirs | xargs -n 1 echo | sort | uniq | sed -E 's/(.*)/'\\\\\''\1'\\\\\''/g' | xargs echo`'
'
    fi
}

_generate_runtime_image_dockerfile_prebuild_scripts() {
    local context=$1
    eval "local pre_build_script_keys=\$_SAR_RT_BUILD_${context}_PRE_BUILD_SCRIPTS"
    if [ ! -z "$pre_build_script_keys" ]; then
        local pre_build_script_keys=`echo $pre_build_script_keys | xargs -n 1 echo | sort | uniq | sed -E 's/(.*)/'\\\\\''\1'\\\\\''/g' | xargs echo`
        local pre_build_work_dirs=
        for key in `echo $pre_build_script_keys`; do
            eval "local pre_build_script_path=\$_SAR_RT_BUILD_${context}_PRE_BUILD_SFRIPT_${key}_PATH"
            eval "local pre_build_script_workdir=\$_SAR_RT_BUILD_${context}_PRE_BUILD_SCRIPT_${key}_WORKDIR"
            if [ -z "$pre_build_script_workdir" ]; then
                local pre_build_script_workdir=/
            fi
            local pre_build_work_dirs="$pre_build_work_dirs '$pre_build_script_workdir'"
            local pre_build_scripts="$pre_build_scripts '$pre_build_script_path'"
            # TODO: checksum here.
        done
        local pre_build_scripts=`echo $pre_build_scripts | xargs -n 1 echo | sort | uniq | sed -E 's/(.*)/'\\\\\''\1'\\\\\''/g' | xargs echo`
        local pre_build_work_dirs=`echo $pre_build_work_dirs | xargs -n 1 echo | sort | uniq | sed -E 's/(.*)/'\\\\\''\1'\\\\\''/g' | xargs echo`
        # run pre-build scripts.
        local failure=0
        echo $pre_build_scripts | xargs -n 1 -I {} test -f {} || (local failure=1; logerror some pre-build script missing.)
        if [ ! $failure -eq 0 ]; then
            return 1
        fi
        echo "COPY $pre_build_scripts /_sar_package/pre_build_scripts/"
        echo -n '
RUN     set -xe;\
        cd /_sar_package/pre_build_scripts;\
        chmod a+x *; mkdir -p '$pre_build_work_dirs';'

        for key in `echo $pre_build_script_keys`; do
            eval "local pre_build_script_path=\$_SAR_RT_BUILD_${context}_PRE_BUILD_SFRIPT_${key}_PATH"
            eval "local pre_build_script_workdir=\$_SAR_RT_BUILD_${context}_PRE_BUILD_SCRIPT_${key}_WORKDIR"
            if [ -z "$pre_build_script_workdir" ]; then
                local pre_build_script_workdir=/
            fi
            local script_name=`eval "basename $pre_build_script_path"`
            local script_name=`strip $script_name`
            local script_name=`path_join /_sar_package/pre_build_scripts $script_name`
            echo -n "cd $pre_build_script_workdir; $script_name;"
        done
        echo
    fi
}

_generate_runtime_image_dockerfile_postbuild_scripts() {
    local context=$1
    eval "local post_build_script_keys=\$_SAR_RT_BUILD_${context}_POST_BUILD_SCRIPTS"
    if [ ! -z "$post_build_script_keys" ]; then
        local post_build_script_keys=`echo $post_build_script_keys | xargs -n 1 echo | sort | uniq | sed -E 's/(.*)/'\\\\\''\1'\\\\\''/g' | xargs echo`
        for key in `echo $post_build_script_keys`; do
            eval "local post_build_script_path=\$_SAR_RT_BUILD_${context}_POST_BUILD_SFRIPT_${key}_PATH"
            eval "local post_build_script_workdir=\$_SAR_RT_BUILD_${context}_POST_BUILD_SCRIPT_${key}_WORKDIR"
            local post_build_scripts="$post_build_scripts '$post_build_script_path'"
            if [ -z "$post_build_script_workdir" ]; then
                local post_build_script_workdir=/
            fi
            local post_build_work_dirs="$post_build_work_dirs '$post_build_script_workdir'"
            # TODO: checksum here.
        done
        local post_build_scripts=`echo $post_build_scripts | xargs -n 1 echo | sort | uniq | sed -E 's/(.*)/'\\\\\''\1'\\\\\''/g' | xargs echo`
        local post_build_work_dirs=`echo $post_build_work_dirs | xargs -n 1 echo | sort | uniq | sed -E 's/(.*)/'\\\\\''\1'\\\\\''/g' | xargs echo`
        # run post-build scripts.
        local failure=0
        echo $post_build_scripts | xargs -n 1 -I {} test -f {} || (local failure=1; logerror some post-build script missing.)
        if [ ! $failure -eq 0 ]; then
            return 1
        fi
        echo "COPY $post_build_scripts /_sar_package/post_build_scripts/"
        echo -n '
RUN     set -xe;\
        cd /_sar_package/post_build_scripts;\
        chmod a+x *; mkdir -p '$post_build_work_dirs';'

        for key in `echo $post_build_script_keys`; do
            eval "local post_build_script_path=\$_SAR_RT_BUILD_${context}_POST_BUILD_SFRIPT_${key}_PATH"
            eval "local post_build_script_workdir=\$_SAR_RT_BUILD_${context}_POST_BUILD_SCRIPT_${key}_WORKDIR"
            if [ -z "$post_build_script_workdir" ]; then
                local post_build_script_workdir=/
            fi
            local script_name=`eval "basename $post_build_script_path"`
            local script_name=`strip $script_name`
            local script_name=`path_join /_sar_package/post_build_scripts $script_name`
            echo -n "cd $post_build_script_workdir; $script_name;"
        done
        echo
    fi
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

    # pack pre-build and post-build scripts.
    if ! _generate_runtime_image_dockerfile_prebuild_scripts $context; then
        return 1
    fi
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

    if ! _generate_runtime_image_dockerfile_add_supervisor_services $context; then
        logerror "[runtime_image_builder] failed to add supervisor services."
        return 1
    fi

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
'

    # run post-build scripts.
    if ! _generate_runtime_image_dockerfile_postbuild_scripts $context; then
        return 1
    fi
echo 'CMD ["supervisord", "-c", "/etc/sar_supervisor.conf"]'

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
                build_runtime_image_help
                logerror "[runtime_image_builder]" unexcepted options -$opt.
                ;;
        esac
    done

    eval "local __=\$$OPTIND"
    local -i optind=$OPTIND
    if [ "$__" != "--" ]; then
        if [ ! -z "$__" ]; then
            build_runtime_image_help
            logerror "[runtime_image_builder] build_runtime_image: got unexcepted non-option argument: \"$__\"."
            return 1
        fi
        local -i optind=optind-1
    fi

    if [ -z "$context" ]; then
        local context=system
    fi

    if [ -z "$ci_image_tag" ]; then
        build_runtime_image_help
        logerror "[runtime_image_builder]" empty runtime image tag.
        return 1
    fi
    if [ -z "$ci_image_prefix" ]; then
        build_runtime_image_help
        logerror "[runtime_image_builder]" empty runtime image prefix.
        return 1
    fi

    # add runtime
    runtime_image_add_dependency -c "$context" -r "$SAR_RUNTIME_PKG_PREFIX" -e "$SAR_RUNTIME_PKG_ENV" -t "$SAR_RUNTIME_PKG_TAG" /_sar_package/runtime_install

    local dockerfile=/tmp/Dockerfile-RuntimeImage-$RANDOM$RANDOM$RANDOM
    if ! _generate_runtime_image_dockerfile "$context" "$ci_image_prefix" "$ci_image_env_name" "$ci_image_tag" > "$dockerfile" ; then
        build_runtime_image_help
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
    runtime_image_base_image [options] <image reference>

options:
    -c <context_name>       specified build context. default: system

example:
    runtime_image_base_image alpine:3.7
    runtime_image_base_image -c context2 alpine:3.7
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
        runtime_image_base_image_help
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

runtime_image_add_service_help() {
    echo '

Add service to image. Services will be started automatically after conainter started.

usage:
    runtime_image_add_service [options] <type> ...
    runtime_image_add_service [options] system <service_name> <command>
    runtime_image_add_service [options] normal <service_name> <command>
    runtime_image_add_service [options] cron <service_name> <command>

options:
    -d <path>               working directory.
    -c <context_name>       specified build context. default: system


runtime_image_add_service system conf_updator /opt/runtime/bin/runtime_conf_update.sh
runtime_image_add_service cron runtime_conf_update "5 0 * * *"
runtime_image_add_service normal nginx /sbin/nginx

'
}

_runtime_image_add_service() {
    local context=$1
    local working_dir=$2
    local type=$3
    local name=$4
    shift 4
    local -i idx=1
    local exec=
    while [ $idx -le $# ]; do
        eval "local param=\$$idx"
        if echo "$param" | grep ' ' -q ; then
            local exec="$exec '$param'"
        else
            local exec="$exec $param"
        fi
        local -i idx=idx+1
    done

    local key=`hash_for_key $name`
    eval "_SAR_RT_BUILD_${context}_SVC_${key}_TYPE=$type"
    eval "_SAR_RT_BUILD_${context}_SVC_${key}_NAME=$name"
    eval "_SAR_RT_BUILD_${context}_SVC_${key}_EXEC=\"$exec\""
    eval "_SAR_RT_BUILD_${context}_SVC_${key}_WORKING_DIR=$working_dir"
    eval "local -i count=\${#_SAR_RT_BUILD_${context}_SVCS[@]}+1"
    eval "_SAR_RT_BUILD_${context}_SVCS[$count]=$key"
}

_runtime_image_add_service_cron() {
    local context=$1
    local working_dir=$2
    local name=$3
    local cron=$4
    shift 3

    local -i idx=1
    local exec=
    while [ $idx -le $# ]; do
        eval "local param=\$$idx"
        if echo "$param" | grep ' ' -q ; then
            local exec="$exec '$param'"
        else
            local exec="$exec $param"
        fi
        local -i idx=idx+1
    done

    local key=`hash_for_key $name`
    eval "_SAR_RT_BUILD_${context}_SVC_${key}_TYPE=cron"
    eval "_SAR_RT_BUILD_${context}_SVC_${key}_NAME=$name"
    eval "_SAR_RT_BUILD_${context}_SVC_${key}_CRON=$cron"
    eval "_SAR_RT_BUILD_${context}_SVC_${key}_EXEC=\"$exec\""
    eval "_SAR_RT_BUILD_${context}_SVC_${key}_WORKING_DIR=$working_dir"
    eval "local -i count=\${#_SAR_RT_BUILD_${context}_SVCS[@]}+1"
    eval "_SAR_RT_BUILD_${context}_SVCS[$count]=$key"
}

runtime_image_add_service() {
    OPTIND=0
    while getopts 'c:d:' opt; do
        case $opt in
            c)
                local context=$OPTARG
                ;;
            d)
                local working_dir=$OPTARG
                ;;
            *)
                runtime_image_add_service_help
                logerror "[runtime_image_builder]" unexcepted options -$opt.
                ;;
        esac
    done

    local -i optind=$OPTIND-1
    shift $optind

    if [ -z "$context" ]; then
        local context=system
    fi

    local type=$1
    case $type in
        system)
            shift 1
            _runtime_image_add_service $context "$working_dir" system $* || return 1
            ;;
        cron)
            shift 1
            _runtime_image_add_service_cron $context "$working_dir" $* || return 1
            ;;
        normal)
            shift 1
            _runtime_image_add_service $context "$working_dir" normal $* || return 1
            ;;
        *)
            logerror "[runtime_image_builder] unknown service type: $type"
            ;;
    esac
}

#runtime_image_pre_build_run() {
#    return 1
#}
#
#runtime_image_post_build_run() {
#    return 1
#}

runtime_image_pre_build_script_help() {
    echo '
Run pre-build script within building of runtime image.

usage:
    runtime_image_pre_build_script [options] <script_path>

options:
    -d <path>               working directory.
    -c <context_name>       specified build context. default: system

example:
    runtime_image_pre_build_script install_lnmp.sh
    runtime_image_pre_build_script -c my_context install_nginx.sh
'
}

runtime_image_pre_build_script() {
    OPTIND=0
    while getopts 'c:d:' opt; do
        case $opt in
            c)
                local context=$OPTARG
                ;;
            d)
                local working_dir=$OPTARG
                ;;
            *)
                runtime_image_pre_build_script_help
                logerror "[runtime_image_builder]" unexcepted options -$opt.
                ;;
        esac
    done
    if [ -z "$context" ]; then
        local context=system
    fi
    local -i optind=$OPTIND-1
    shift $optind

    local -i idx=1
    local -i failure=0
    local script_appended=
    while [ $idx -le $# ]; do
        eval "local script=\$$idx"
        if ! [ -f "$script" ]; then
            logerror "[runtime_image_builder] not a script file: $script"
            local -i failure=1
        fi
        local script_appended="$script_appended '$script'"
        local -i idx=idx+1
    done
    if [ -z `strip $script_appended` ]; then
        runtime_image_pre_build_script_help
        logerror "script missing."
        return 1
    fi
    if [ $failure -gt 0 ]; then
        return 1
    fi

    local key=`eval "hash_file_for_key $script_appended"`
    eval "_SAR_RT_BUILD_${context}_PRE_BUILD_SCRIPTS=\"\$_SAR_RT_BUILD_${context}_PRE_BUILD_SCRIPTS \$key\""
    eval "_SAR_RT_BUILD_${context}_PRE_BUILD_SFRIPT_${key}_PATH=\"$script_appended\""
    eval "_SAR_RT_BUILD_${context}_PRE_BUILD_SCRIPT_${key}_WORKDIR=$working_dir"
}

runtime_image_post_build_script_help() {
    echo '
Run post-build script within building of runtime image.

usage:
    runtime_image_post_build_script [options] <script_path>

options:
    -d <path>               working directory.
    -c <context_name>       specified build context. default: system

example:
    runtime_image_post_build_script cleaning.sh
    runtime_image_post_build_script -c my_context send_notification.sh
'
}

runtime_image_post_build_script() {
    OPTIND=0
    while getopts 'c:d:' opt; do
        case $opt in
            c)
                local context=$OPTARG
                ;;
            d)
                local working_dir=$OPTARG
                ;;
            *)
                runtime_image_post_build_script_help
                logerror "[runtime_image_builder]" unexcepted options -$opt.
                ;;
        esac
    done
    if [ -z "$context" ]; then
        local context=system
    fi
    local -i optind=$OPTIND-1
    shift $optind

    local -i idx=1
    local -i failure=0
    local script_appended=
    while [ $idx -le $# ]; do
        eval "local script=\$$idx"
        if ! [ -f "$script" ]; then
            logerror "[runtime_image_builder] not a script file: $script"
            local -i failure=1
        fi
        local script_appended="$script_appended '$script'"
        local -i idx=idx+1
    done
    if [ $failure -gt 0 ]; then
        return 1
    fi

    if [ -z `strip $script_appended` ]; then
        runtime_image_post_build_script_help
        logerror "script missing."
        return 1
    fi

    local key=`eval "hash_file_for_key $script_appended"`
    eval "_SAR_RT_BUILD_${context}_POST_BUILD_SCRIPTS=\"\$_SAR_RT_BUILD_${context}_POST_BUILD_SCRIPTS \$key\""
    eval "_SAR_RT_BUILD_${context}_POST_BUILD_SFRIPT_${key}_PATH=\"$script_appended\""
    eval "_SAR_RT_BUILD_${context}_POST_BUILD_SCRIPT_${key}_WORKDIR=$working_dir"
}

#runtime_image_health_check_script() {
#    return 1
#}