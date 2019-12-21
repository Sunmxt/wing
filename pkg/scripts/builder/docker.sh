sar_import builder/common.sh

_ci_docker_build() {
    # Parse normal options
    OPTIND=0
    local ci_build_docker_tag=
    local ci_build_docker_env_name=
    local ci_project_path=
    local ci_registry=
    local ci_no_push=
    local force_to_build=
    local hash_target_content=

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
    if [ -z "${GITLAB_CI+x}" ]; then
        logerror Not a Gitlab CI environment.
        return 1
    fi
    loginfo Try to login to registry.
    if ! docker login -u gitlab-ci-token -p $CI_JOB_TOKEN $CI_REGISTRY; then
        logerror Cannot login to $CI_REGISTRY.
        return 2
    fi

    OPTIND=0
    local -a opts=()
    while getopts 't:r:e:p:sfh:' opt; do
        case $opt in
            t)
                local ci_build_docker_tag=$OPTARG
                [ "$ci_build_docker_tag" = "gitlab_ci_commit_hash" ] && continue
                ;;
            r)
                ;;
            p)
                ;;
            e)
                ;;
            s)
                opts+=("-s")
                continue
                ;;
            f)
                opts+=("-f")
                continue
                ;;
            h)
                ;;
        esac
        opts+=("-$opt" "$OPTARG")
    done

    local -i optind=$OPTIND-1
    eval "local __=\${$optind}"
    if [ "$__" == "--" ]; then
        local has_docker_ext="--"
    fi
    shift $optind

    log_exec _ci_docker_build ${opts[@]} $has_docker_ext $*
}

_ci_auto_docker_build() {
    if [ ! -z "${GITLAB_CI+x}" ]; then
        log_exec _ci_gitlab_runner_docker_build $*
        return $?
    fi
    log_exec _ci_docker_build $*
}
