
sar_import lib.sh

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
    eval `local __=\$$OPTIND`
    local -i optind=$OPTIND
    if [ "$__" = "--" ]; then
        local -i optind=optind+1
    fi
    local -i shift_opt_cnt=optind-1
    shift $shift_opt_cnt

    if ! [ -z "$ci_registry_image" ]; then
        local ci_registry_image=`echo $ci_registry_image | sed -E 's/(\/)*$//g'`
    fi
    local ci_build_docker_ref=$ci_registry_image
    if [ ! -z "$ci_build_docker_env_name" ]; then
        eval "local path_appended=\$$ci_build_docker_env_name"
        if [ ! -z "$path_appended" ]; then
            loginfo add path to image ref: $path_appended
            local ci_build_docker_ref=$ci_build_docker_ref/$path_appended
        fi
    fi
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
    log_exec _ci_docker_build $opts -r $CI_REGISTRY_IMAGE $*
    return $?
}

# ci_build <mode> [options] -- [docker build options]
# 
# mode:
#   gitlab-runner-docker
#   docker
#
# options:
#       -t <tag>                            Image tag. 
#                                           if `gitlab_ci_commit_hash` is specified in 
#                                           `gitlab-runner-docker` mode, the tag will be substitute 
#                                           with actually commit hash.
#       -e <environment_variable_name>      Identify docker path by environment variable.
#       -r <ref_prefix>
ci_build() {
    case $1 in
        gitlab-runner-docker)
            shift 1
            _ci_gitlab_runner_docker_build $*
            return $?
            ;;
        docker)
            shift 1
            _ci_docker_build $*
            return $?
            ;;
        *)
            logerror unsupported ci type: $1
            return 1
            ;;
    esac
}