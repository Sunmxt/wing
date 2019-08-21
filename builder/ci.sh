sar_import lib.sh
sar_import builder/common.sh
sar_import settings/image.sh

_ci_build_generate_registry_tag() {
    local tag=$1
    if [ -z "$tag" ]; then
        local tag=latest
    fi
    echo $tag
}

_ci_build_generate_registry_path() {
    local prefix=$1
    local env=$2

    if ! [ -z "$prefix" ]; then
        local prefix=`echo "$prefix" | sed -E 's/(\/)*$//g'`
    fi
    local registry_path="$prefix"

    if [ ! -z "$env" ]; then
        eval "local path_appended=\$$env"
        if [ ! -z "$path_appended" ]; then
            local registry_path=$registry_path/$path_appended
        fi
    fi

    echo $registry_path
    if [ -z "$registry_path" ]; then
        logwarn Empty image reference.
        return 1
    fi
}

# _ci_docker_build [options] -- [docker build options]
#   options:
#       -t <tag>           
#       -e <environment_variable_name>      Identify docker path by environment variable.
#       -r <ref_prefix>
_ci_docker_build() {
    # Parse normal options
    OPTIND=0
    while getopts 't:e:r:' opt; do
        case $opt in
            t)
                local ci_build_docker_tag=$OPTARG
                ;;
            e)
                local ci_build_docker_env_name=$OPTARG
                ;;
            r)
                local ci_registry_image=$OPTARG
                ;;
        esac
    done
    eval "local __=\$$OPTIND"
    local -i optind=$OPTIND
    if [ "$__" = "--" ]; then
        local -i optind=optind+1
    fi
    local -i shift_opt_cnt=optind-1
    shift $shift_opt_cnt

    local ci_build_docker_ref=`_ci_build_generate_registry_path "$ci_registry_image" "$ci_build_docker_env_name"`

    if [ -z "$ci_build_docker_tag" ]; then
        local ci_build_docker_tag=latest
    fi
    if [ -z "$ci_build_docker_ref" ]; then
        logerror Empty image reference.
        return 1
    fi

    local ci_build_docker_ref_path=$ci_build_docker_ref
    local ci_build_docker_ref=$ci_build_docker_ref_path:$ci_build_docker_tag

    if ! log_exec docker build -t $ci_build_docker_ref $*; then
        logerror build failure.
        return 2
    fi
    
    if ! log_exec docker push $ci_build_docker_ref; then
        logerror uploading image $ci_build_docker_ref failure.
        return 3
    fi
    if [ "$ci_build_docker_tag" != "latest" ]; then
        log_exec docker tag $ci_build_docker_ref $ci_build_docker_ref_path:latest
        if ! log_exec docker push $ci_build_docker_ref_path:latest; then
            logerror uploading image $ci_build_docker_ref_path:latest failure.
            return 4
        fi
    fi
}

_ci_gitlab_runner_docker_build() {
    if [ -z "$GITLAB_CI" ]; then
        logerror Not a Gitlab CI environment.
        return 1
    fi
    loginfo Try to login to registry.
    if ! docker login -u gitlab-ci-token -p $CI_JOB_TOKEN $CI_REGISTRY; then
        logerror Cannot login to $CI_REGISTRY.
        return 2
    fi

    OPTIND=0
    while getopts ':t:r:' opt; do
        case $opt in
            t)
                local ci_build_docker_tag=$OPTARG
                ;;
            r)
                local ci_build_docker_ref=$OPTARG
                ;;
        esac
    done
    local opts=
    local -i idx=1
    while [ $idx -le $OPTIND ]; do
        eval "local opts=\"\$opts \$$idx\""
        local -i idx=idx+1
    done
    if [ -z "$ci_build_docker_tag" ]; then
        local ci_build_docker_tag=gitlab_ci_commit_hash
    fi
    if [ "$ci_build_docker_tag" = "gitlab_ci_commit_hash" ]; then
        local ci_build_docker_tag=${CI_COMMIT_SHA:0:10}
        local opts="$opts -t $ci_build_docker_tag"
    fi
    local -i shift_cnt=$OPTIND
    shift $shift_cnt
    log_exec _ci_docker_build $opts -r $CI_REGISTRY_IMAGE $*
    return $?
}

# 这里的逻辑很多是重复的，看能不能合并到一起。
_ci_gitlab_package_build() {
    if [ -z "$GITLAB_CI" ]; then
        logerror Not a Gitlab CI environment.
        return 1
    fi
    loginfo Try to login to registry.
    if ! docker login -u gitlab-ci-token -p $CI_JOB_TOKEN $CI_REGISTRY; then
        logerror Cannot login to $CI_REGISTRY.
        return 2
    fi

    OPTIND=0
    while getopts ':t:r:' opt; do
        case $opt in
            t)
                local ci_build_docker_tag=$OPTARG
                ;;
            r)
                local ci_build_docker_ref=$OPTARG
                ;;
        esac
    done
    local opts=
    local -i idx=1
    while [ $idx -le $OPTIND ]; do
        eval "local opts=\"\$opts \$$idx\""
        local -i idx=idx+1
    done
    if [ "$ci_build_docker_tag" = "gitlab_ci_commit_hash" ]; then
        local ci_build_docker_tag=${CI_COMMIT_SHA:0:10}
        local opts="$opts -t $ci_build_docker_tag"
    fi
    local -i shift_cnt=$OPTIND
    shift $shift_cnt
    log_exec _ci_build_package $opts -r $CI_REGISTRY_IMAGE $*
    return $?
}


_ci_build_package_generate_dockerfile() {
    local product_ref=$1
    local product_environment=$2
    local product_tag=$3
    local product_path=$4

    # TODO: 这里要防注入，对一些字符进行转义
    echo '
FROM '$PACKAGE_BASE_IMAGE'

RUN set -xe;\
    mkdir -p /_sar_package;\
    touch /_sar_package/meta;\
    echo PKG_REF='\\\'$product_ref\\\'' > /_sar_package/meta;\
    echo PKG_ENV='\\\'$product_environment\\\'' >> /_sar_package/meta;\
    echo PKG_TAG='\\\'$product_tag\\\'' >> /_sar_package/meta;\
    echo PKG_TYPE=package >> /_sar_package/meta;\
    mkdir -p /_sar_package/data;

COPY "'$product_path'" /package/data
'
    
}

_ci_build_package() {
    OPTIND=0
    while getopts 't:e:r:' opt; do
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
        esac
    done
    eval "local __=\$$OPTIND"
    local -i optind=$OPTIND
    if [ "$__" != "--" ] && [ ! -z "$__" ]; then
        local product_path=$__
    fi
    local -i optind=optind+1
    eval "local __=\$$optind"
    if [ "$__" = "--" ]; then
        local -i optind=optind+1
    fi
    if [ -z "$product_path" ]; then
        logerror output path not specified.
        return 1
    fi

    local -i shift_opt_cnt=optind-1
    shift $shift_opt_cnt

    local ci_package_ref=`_ci_build_generate_registry_path "$ci_package_prefix" "$ci_package_env_name"`
    if [ -z "$ci_package_ref" ]; then
        logerror Empty package ref.
        return 1
    fi
    local ci_package_ref=`path_join "$ci_package_ref" sar__package`
    loginfo build package with registry image: $ci_package_ref
    local ci_package_env_name=`_ci_get_env_value "$ci_package_env_name"`
    local ci_package_tag=`_ci_build_generate_registry_tag "$ci_package_tag"`

    # Generate dockerfile
    local dockerfile_path=/tmp/Dockerfile-PACKAGE-$RANDOM$RANDOM$RANDOM
    loginfo generate dockerfile: $dockerfile_path
    if ! log_exec _ci_build_package_generate_dockerfile "$ci_package_ref" "$ci_package_env_name" "$ci_package_tag" "$product_path" > "$dockerfile_path"; then 
        logerror generate dockerfile failure.
        return 1
    fi

    # build
    if ! log_exec docker build -t $ci_package_ref:$ci_package_tag -f "$dockerfile_path" $* .; then
        logerror build failure.
        return 2
    fi

    # upload
    if ! log_exec docker push $ci_package_ref:$ci_package_tag; then
        logerror uploading "image(package)" $ci_package_ref:$ci_package_tag failure.
        return 3
    fi
    if [ "$ci_package_tag" != "latest" ]; then
        log_exec docker tag "${ci_package_ref}:$ci_package_tag" "${ci_package_ref}:latest"
        if ! log_exec docker push "${ci_package_ref}:latest"; then
            logerror uploading "image(package)" ${ci_package_tag}:latest failure.
            return 4
        fi
    fi

    return 0
}

help_ci_build() {
    echo '
Project builder in CI Environment.

ci_build <mode> [options] -- [docker build options]

mode:
  gitlab-runner-docker
  docker
  package
  gitlab-package

options:
      -t <tag>                            Image tag / package tag (package mode)
                                          if `gitlab_ci_commit_hash` is specified in 
                                          `gitlab-runner-docker` mode, the tag will be substitute 
                                          with actually commit hash.
      -e <environment_variable_name>      Identify docker path by environment variable.
      -r <ref_prefix>                     Image reference prefix.

example:
      ci_build gitlab-runner-docker -t gitlab_ci_commit_hash -e ENV .
      ci_build gitlab-runner-docker -t gitlab_ci_commit_hash -e ENV -- --build-arg="myvar=1" .
      ci_build gitlab-runner-docker -t gitlab_ci_commit_hash -r registry.mine.cn/test/myimage -e ENV -- --build-arg="myvar=1" .
      ci_build docker -t stable_version -r registry.mine.cn/test/myimage -e ENV .
      ci_build package -t be/recruitment2019 -e ENV bin
      ci_build gitlab-package -t gitlab_ci_commit_hash -e ENV bin
'
}

ci_build() {
    local mode=$1
    shift 1
    case $mode in
        gitlab-runner-docker)
            _ci_gitlab_runner_docker_build $*
            return $?
            ;;
        docker)
            _ci_docker_build $*
            return $?
            ;;
        package)
            _ci_build_package $*
            return $?
            ;;
        gitlab-package)
            _ci_gitlab_package_build $*
            return $?
            ;;
        *)
            logerror unsupported ci type: $mode
            help_ci_build
            return 1
            ;;
    esac
}

# runtime_image -r be/recruitment2019 -e master -t ci_commit_hash -e master 
# runtime_image_base_image registry.stuhome.com/devops/php:7-1.0.1
# runtime_image_add_dependency -r be/recruitment2019 -t 3928ea19 -e master /app
# runtime_image_add_dependency -r fe/recruitment2019 -t 281919ea -e master /app/statics
# starconf_set_entry xxxxxx
# starconf_configure_root xxxx
# runtime_image_bootstrap_run /app/my_start_script.sh
# runtime_image_build_start
# deploy_runtime_image -r be/recruitment2019 -e master -t ci_commit_hash -e master

runtime_image_add_dependency() {
    return 0
}