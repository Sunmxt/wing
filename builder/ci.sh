sar_import lib.sh
sar_import builder/common.sh
sar_import settings/image.sh
sar_import docker_utils.sh

enable_dockercli_experimentals

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
        else
            local registry_path=$registry_path/$env
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
    while getopts 't:e:r:sfh:' opt; do
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
            s)
                local ci_no_push=1
                ;;
            f)
                local force_to_build=1
                ;;
            h)
                local hash_target_content="$OPTARG"
                ;;
        esac
    done
    eval "local __=\${$OPTIND}"
    local -i optind=$OPTIND
    if [ "$__" = "--" ]; then
        local -i optind=optind+1
    fi
    local -i shift_opt_cnt=optind-1
    shift $shift_opt_cnt

    local ci_build_docker_ref=`_ci_build_generate_registry_path "$ci_registry_image" "$ci_build_docker_env_name"`

    if [ -z "$ci_build_docker_tag" ]; then
        if [ ! -z "$hash_target_content" ]; then
            local ci_build_docker_tag=`hash_content_for_key $hash_target_content`
            if [ -z "$ci_build_docker_tag" ]; then
                logerror cannot generate hash by files.
                return 1
            fi
            local ci_build_docker_tag=${ci_build_docker_tag:0:10}
        else
            local ci_build_docker_tag=latest
        fi 
    fi
    if [ -z "$ci_build_docker_ref" ]; then
        logerror Empty image reference.
        return 1
    fi

    local ci_build_docker_ref_path=$ci_build_docker_ref
    local ci_build_docker_ref=$ci_build_docker_ref_path:$ci_build_docker_tag

    if is_image_exists "$ci_build_docker_ref" && [ -z "$force_to_build" ]; then
        logwarn Skip image build: $ci_build_docker_ref
        return 0
    fi

    if ! log_exec docker build -t $ci_build_docker_ref $*; then
        logerror build failure.
        return 2
    fi
    if [ -z "$ci_no_push" ]; then
        if ! log_exec docker push $ci_build_docker_ref; then
            logerror uploading image $ci_build_docker_ref failure.
            return 3
        fi
    fi
    if [ "$ci_build_docker_tag" != "latest" ]; then
        log_exec docker tag $ci_build_docker_ref $ci_build_docker_ref_path:latest
        if [ -z "$ci_no_push" ]; then
            if ! log_exec docker push $ci_build_docker_ref_path:latest; then
                logerror uploading image $ci_build_docker_ref_path:latest failure.
                return 4
            fi
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
    while getopts 't:r:e:sfh:' opt; do
        case $opt in
            t)
                local ci_build_docker_tag=$OPTARG
                ;;
            r)
                local ci_build_docker_ref=$OPTARG
                ;;
            e)
                local ci_build_docker_env_name=$OPTARG
                ;;
            s)
                local ci_no_push=1
                ;;
            f)
                local force_to_build=1
                ;;
            h)
                local hash_target_content="$OPTARG"
                ;;
        esac
        eval "local opt=\${$OPTIND}"
        if [ "${opt:0:2}" = "--" ]; then
            local has_docker_ext="--"
            break
        fi
    done
    local opts=
    local -i idx=1
    while [ $idx -lt $OPTIND ]; do
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
    if [ ! -z "$ci_build_docker_env_name" ]; then
        local opts="$opts -e $ci_build_docker_env_name"
    fi
    local -i shift_cnt=$OPTIND
    shift $shift_cnt
    log_exec _ci_docker_build $opts -r $CI_REGISTRY_IMAGE $has_docker_ext $*
    return $?
}

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
    while getopts 't:r:e:sf' opt; do
        case $opt in
            t)
                local ci_build_docker_tag=$OPTARG
                ;;
            r)
                local ci_build_docker_ref=$OPTARG
                ;;
            e)
                local ci_build_docker_env_name=$OPTARG
                ;;
            s)
                local ci_no_push=1
                ;;
            f)
                local force_to_build=1
                ;;
        esac
        eval "local opt=\${$OPTIND}"
        if [ "${opt:0:2}" = "--" ]; then
            local has_docker_ext="--"
            break
        fi
    done
    local opts=
    local -i idx=1
    while [ $idx -lt $OPTIND ]; do
        eval "local opt=\${$idx}"
        local opt=${opt:1:1}
        if [ "${opt:1:1}" = "c" ]; then
            local -i idx=idx+2
            continue
        fi
        eval "local opts=\"\$opts \${$idx}\""
        local -i idx=idx+1
    done
    if [ "$ci_build_docker_tag" = "gitlab_ci_commit_hash" ]; then
        local ci_build_docker_tag=${CI_COMMIT_SHA:0:10}
        local opts="$opts -t $ci_build_docker_tag"
    fi
    if [ ! -z "$ci_build_docker_env_name" ]; then
        local opts="$opts -e $ci_build_docker_env_name"
    fi
    local -i shift_cnt=$OPTIND-1
    shift $shift_cnt
    log_exec _ci_build_package $opts -r $CI_REGISTRY_IMAGE $has_docker_ext $*
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

COPY "'$product_path'" /package/data

RUN set -xe;\
    mkdir -p /_sar_package;\
    touch /_sar_package/meta;\
    echo PKG_REF='\\\'$product_ref\\\'' > /_sar_package/meta;\
    echo PKG_ENV='\\\'$product_environment\\\'' >> /_sar_package/meta;\
    echo PKG_TAG='\\\'$product_tag\\\'' >> /_sar_package/meta;\
    echo PKG_TYPE=package >> /_sar_package/meta;\
    mkdir -p /_sar_package/data;\
    rm -rf /package/data/.git /package/data/.gitlab-ci.yml;
'
    
}

_ci_build_package() {
    OPTIND=0
    while getopts 't:e:r:sf' opt; do
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
            s)
                local ci_no_push=1
                ;;
            f)
                local force_to_build=1
                ;;
        esac
    done
    eval "local __=\${$OPTIND}"
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

    if [ -z "$ci_package_tag" ]; then
        local ci_package_tag=`hash_content_for_key "$product_path"`
        if [ -z "$ci_package_tag" ]; then
            logerror cannot generate hash by files.
            return 1
        fi
    fi

    local ci_package_ref=`path_join "$ci_package_ref" sar__package`
    local ci_package_env_name=`_ci_get_env_value "$ci_package_env_name"`
    local ci_package_tag=`_ci_build_generate_registry_tag "$ci_package_tag"`
    loginfo build package with registry image: $ci_package_ref:$ci_package_tag

    if is_image_exists "$ci_package_ref:$ci_package_tag" && [ -z "$force_to_build" ]; then
        logwarn Skip image build: "$ci_package_ref:$ci_package_tag"
        return 0
    fi

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
    if [ -z "$ci_no_push" ]; then
        if ! log_exec docker push $ci_package_ref:$ci_package_tag; then
            logerror uploading "image(package)" $ci_package_ref:$ci_package_tag failure.
            return 3
        fi
    fi
    if [ "$ci_package_tag" != "latest" ]; then
        log_exec docker tag "${ci_package_ref}:$ci_package_tag" "${ci_package_ref}:latest"
        if [ -z "$ci_no_push" ]; then
            if ! log_exec docker push "${ci_package_ref}:latest"; then
                logerror uploading "image(package)" ${ci_package_tag}:latest failure.
                return 4
            fi
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
      -s                                  Do not push image to regsitry.
      -h <path_to_hash>                   Use file(s) hash for tag.
      -f                                  force to build.

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

ci_login_help() {
    echo '
Sign up to ci environment.

usage:
    ci_login [-p password] <username>
'
}

ci_login() {
    if ! docker_installed ; then
        logerror [ci_login] docker not installed.
        return 1
    fi
    OPTIND=0
    while getopts 'p:' opt; do
        case $opt in
            p)
                local password=$OPTARG
                ;;
            *)
                logerror unsupported option: $opt.
                return 1
                ;;
        esac
    done
    local password
    if [ ! -z "$password" ]; then
        logwarn Attention! Specifing plain password is not recommanded.
    else
        while [ -z "$password" ]; do
            loginfo Empty password.
            echo -n 'Password: '
            read -s password
            echo
        done
    fi

    eval "local __=\${$OPTIND}"
    local username="`strip \"$__\"`"
    if [ -z "$username" ] ; then
        logerror username is empty.
        return 1
    fi

    docker login "$SAR_CI_REGISTRY" -u "$username" -p "$password"
}

ci_package_pull_help() {
    echo '
Pull package.

usage:
    ci_package_pull [options] <path>

options:
      -t <tag>                            Image tag / package tag (package mode)
      -e <environment>                    Identify docker path by environment variable.
      -r <ref_prefix>                     Image reference prefix.
      -f <full_reference>                 Full image reference.

example:
    ci_package_pull -r devops/runtime -e master ./my_dir
    ci_package_pull -f registry.stuhome.com/devops/runtime/master/sar__package ./my_dir
'
}

ci_package_pull() {
    if ! docker_installed ; then
        logerror [ci_package_pull] docker not installed.
        return 1
    fi

    OPTIND=0
    while getopts 't:r:e:f:' opt; do
        case $opt in
            t)
                local ci_build_docker_tag=$OPTARG
                ;;
            r)
                local ci_build_docker_prefix=$OPTARG
                ;;
            e)
                local ci_build_docker_env_name=$OPTARG
                ;;
            f)
                local ci_full_reference=$OPTARG
                ;;
        esac
    done

    eval "local __=\${$OPTIND}"
    local target_path="`strip \"$__\"`"
    if [ -z "$target_path" ] ; then
        logerror target path not specified.
        return 1
    fi

    if ! mkdir -p "$target_path"; then
        logerror cannot create directory.
        return 1
    fi
    local target_path=`full_path_of "$target_path"`

    if [ -z "$ci_full_reference" ]; then
        if [ -z "$ci_build_docker_tag" ]; then
            local ci_build_docker_tag=latest
        fi
        local ci_full_reference=`_ci_get_package_ref "$ci_build_docker_prefix" "$ci_build_docker_env_name" "$ci_build_docker_tag"`
        local ci_full_reference="$SAR_CI_REGISTRY/$ci_full_reference"
    fi
    if [ -z "$ci_full_reference" ]; then
        logerror empty image reference.
        return 1
    fi
    if ! log_exec docker pull "$ci_full_reference"; then
        logerror pull package image "$ci_full_reference" failure.
        return 1
    fi

    loginfo extract package to $target_path.
    if ! docker run --entrypoint='' -v "$target_path:/_sar_package_extract_mount" "$ci_full_reference" sh -c "cp -rv /package/data/* /_sar_package_extract_mount/"; then
        logerror extract package failure.
    fi
    loginfo package extracted to $target_path.
}