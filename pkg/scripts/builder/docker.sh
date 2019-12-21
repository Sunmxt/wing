sar_import builder/common.sh

_ci_docker_build() {
    # Parse normal options
    OPTIND=0
    while getopts 't:e:p:r:sfh:' opt; do
        case $opt in
            t)
                local ci_build_docker_tag=$OPTARG
                ;;
            e)
                local ci_build_docker_env_name=$OPTARG
                ;;
            p)
                local ci_project_path=$OPTARG
                ;;
            r)
                local ci_registry=$OPTARG
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

    # resolve path
    local ci_build_docker_ref=`_ci_build_generate_registry_prefix "$ci_registry" "$ci_project_path" "$ci_build_docker_env_name"`
    if [ -z "$ci_build_docker_ref" ]; then
        logerror Empty image reference.
        return 1
    fi
    local ci_build_docker_ref_path=$ci_build_docker_ref
    # resolve tag
    if [ -z "$ci_build_docker_tag" ]; then
        if [ ! -z "$hash_target_content" ]; then
            local ci_build_docker_tag=`hash_content_for_key $hash_target_content`
            if [ -z "$ci_build_docker_tag" ]; then
                logerror cannot generate hash by files.
                return 1
            fi
            local ci_build_docker_tag=${ci_build_docker_tag:0:10}
        else
            local ci_build_docker_tag=`_ci_build_generate_tag`
        fi 
    fi
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
        if [ "$ci_build_docker_tag" != "latest" ]; then
            log_exec docker tag $ci_build_docker_ref $ci_build_docker_ref_path:latest
            if ! log_exec docker push $ci_build_docker_ref_path:latest; then
                logerror uploading image $ci_build_docker_ref_path:latest failure.
                return 4
            fi
        fi
    fi
}

_ci_gitlab_runner_docker_build() {
    if [ ! -z "${GITLAB_CI+x}" ]; then
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
                local ci_build_docker_registry=$OPTARG
                ;;
            p)
                local ci_build_project_path=$OPTARG
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
    local -a opts=()
    local -i idx=1
    while [ $idx -lt $OPTIND ]; do
        eval "opts+=(\"\${$idx}\")"
        local -i idx=idx+1
    done
    local -i shift_cnt=$OPTIND
    shift $shift_cnt

    if [ "$ci_build_docker_tag" != "gitlab_ci_commit_hash" ]; then
        opts+=("-t" "$ci_build_docker_tag")
    fi
    if [ ! -z "$ci_build_docker_env_name" ]; then
        opts+=("-e" "$ci_build_docker_env_name")
    fi
    if [ ! -z "$ci_build_docker_registry" ]; then
        opt+=("-r" "$ci_build_docker_registry")
    fi
    if [ ! -z "$ci_build_project_path" ]; then
        opt+=("-p" "$ci_build_project_path")
    fi

    log_exec _ci_docker_build ${opts[@]} -r $CI_REGISTRY_IMAGE $has_docker_ext $*
}

_ci_auto_docker_build() {
    if [ -z "${GITLAB_CI+x}" ]; then
        _ci_gitlab_runner_docker_build $*
        return $?
    fi
    _ci_docker_build $*
}
