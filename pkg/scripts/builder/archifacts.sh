sar_import docker.sh
sar_import builder/common.sh
sar_import settings/wing.sh

ci_package_pull_help() {
    echo '
Pull archifacts.

usage:
    ci_package_pull [options] <path>

options:
      -t <tag>                            Image tag / package tag (package mode)
      -e <environment>                    Identify docker path by environment variable.
      -r <host>                           registry.
      -p <project_path>                   project path.
      -f <full_reference>                 Full reference.

example:
    ci_package_pull -p devops/runtime -e master ./my_dir
    ci_package_pull -f registry.stuhome.com/devops/runtime/master/sar__package ./my_dir
'
}

ci_package_pull() {
    if ! docker_installed ; then
        logerror [ci_package_pull] docker not installed.
        return 1
    fi

    OPTIND=0
    while getopts 't:r:e:f:p:' opt; do
        case $opt in
            t)
                local ci_build_tag=$OPTARG
                ;;
            r)
                local ci_build_host=$OPTARG
                ;;
            e)
                local ci_build_env_name=$OPTARG
                ;;
            p)
                local ci_build_package_path=$OPTARG
                ;;
            f)
                local ci_full_reference=$OPTARG
                ;;
            h)
                ci_package_pull_help
                return
                ;;
        esac
    done

    eval "local __=\${$OPTIND}"
    local target_path="`strip \"$__\"`"
    if [ -z "$target_path" ] ; then
        ci_package_pull_help
        logerror target path not specified.
        return 1
    fi

    if ! mkdir -p "$target_path"; then
        logerror cannot create directory.
        return 1
    fi
    local target_path=`full_path_of "$target_path"`

    if [ -z "$ci_full_reference" ]; then
        if [ -z "$ci_build_tag" ]; then
            local ci_build_tag=latest
        fi
        local ci_full_reference=`_ci_get_package_ref "$ci_build_host" "$ci_build_package_path" "$ci_build_env_name" "$ci_build_tag"`
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

_ci_auto_package() {
    if [ ! -z "$GITLAB_CI" ]; then
        _ci_gitlab_package_build $*
        return $?
    fi
    _ci_build_package $*
}

_ci_build_package() {
    OPTIND=0
    while getopts 't:e:r:p:sf' opt; do
        case $opt in
            t)
                local ci_package_tag=$OPTARG
                ;;
            e)
                local ci_package_env_name=$OPTARG
                ;;
            r)
                local ci_registry_host=$OPTARG
                ;;
            p)
                local ci_package_path=$OPTARG
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
    eval "local __=\${$optind}"
    if [ "$__" = "--" ]; then
        local -i optind=optind+1
    fi
    if [ -z "$product_path" ]; then
        logerror output path not specified.
        return 1
    fi
    local -i shift_opt_cnt=optind-1
    shift $shift_opt_cnt

    # backend: docker
    # resolve prefix
    local ci_package_ref=`_ci_build_generate_registry_prefix "$ci_registry_host" "$ci_package_path" "$ci_package_env_name"`
    if [ -z "$ci_package_ref" ]; then
        logerror Empty package ref.
        return 1
    fi
    local ci_package_ref=`path_join "$ci_package_ref" sar__package`
    # resolve default tag.
    if [ -z "$ci_package_tag" ]; then
        local ci_package_tag=`_ci_build_generate_tag "$ci_package_tag"`
    fi
    if [ -z "$ci_package_tag" ]; then
        local ci_package_tag=`hash_content_for_key "$product_path"`
        if [ -z "$ci_package_tag" ]; then
            logerror cannot generate hash by files.
            return 1
        fi
        local ci_package_tag=${ci_package_tag:0:10}
    fi
    # resolve environment
    local ci_package_env_name=`_ci_build_generate_env_ref "$ci_package_env_name"`
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
    while getopts 't:r:p:e:sf' opt; do
        case $opt in
            t)
                local ci_build_tag=$OPTARG
                ;;
            r)
                local ci_registry=$OPTARG
                ;;
            p)
                local ci_package_path=$OPTARG
                ;;
            e)
                local ci_build_env_name=$OPTARG
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
            local has_ext_opt="--"
            break
        fi
    done
    local opts=()
    local -i idx=1
    while [ $idx -lt $OPTIND ]; do
        eval "local opt=\${$idx}"
        local opt=${opt:1:1}
        if [ "${opt:1:1}" = "c" ]; then
            local -i idx=idx+2
            continue
        fi
        eval "opts+=(\"\${$idx}\")"
        local -i idx=idx+1
    done

    if [ ! -z "$ci_registry" ]; then
        opts+=("-r" "$ci_registry")
    fi
    if [ ! -z "$ci_build_env_name" ]; then
        opts+=("-e" "$ci_build_env_name")
    fi
    if [ "$ci_build_tag" = "gitlab_ci_commit_hash" ]; then
        local ci_build_tag=`_ci_build_generate_tag ""`
        opts+=("-t" "$ci_build_tag")
    fi

    local -i shift_cnt=$OPTIND-1
    shift $shift_cnt

    log_exec _ci_build_package ${opts[@]} $has_ext_opt $*
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

_ci_wing_report() {
    local params="$1"
    local saved=wing-report-`hash_for_key "$WING_REPORT_URL $WING_CI_TOKEN $params"`
    if ! curl -X POST -L -H "Wing-Auth-Token: $WING_CI_TOKEN" "$WING_REPORT_URL" -d "$params" > "$saved"; then
        return 1
    fi
    test `jq '.success' "$saved"` = "true"
}

_ci_wing_gitlab_package_build() {
    local product_path="$1"

    echo " _       __ _                  ______               _            ";
    echo "| |     / /(_)____   ____ _   / ____/____   ____ _ (_)____   ___ ";
    echo "| | /| / // // __ \ / __ \`/  / __/  / __ \ / __ \`// // __ \ / _ \ ";
    echo "| |/ |/ // // / / // /_/ /  / /___ / / / // /_/ // // / / //  __/";
    echo "|__/|__//_//_/ /_/ \__, /  /_____//_/ /_/ \__, //_//_/ /_/ \___/ ";
    echo "                  /____/                 /____/                  ";
    echo; echo;

    if [ -z "$WING_CI_TOKEN" ]; then
        logerror wing auth token is empty.
        return 1
    fi

    if [ -z "$WING_JOB_URL" ]; then
        logerror missing remote job.
        return 2
    fi

    if [ -z "$WING_REPORT_URL" ]; then
        logerror missing report endpoint.
        return 3
    fi

    # build
    local build_script=wing-build-`hash_for_key "$WING_JOB_URL"`.sh
    loginfo fetching build script from wing platform...
    if ! curl -L -H "Wing-Auth-Token: $WING_CI_TOKEN" "$WING_JOB_URL" > "$build_script"; then
        logerror fetch build script failure.
        return 1
    fi
    chmod a+x "$build_script"
    loginfo "save to: $build_script"

    # Sync state to wing server
    if ! _ci_wing_report "product_token=$WING_PRODUCT_TOKEN&type=$WING_REPORT_TYPE_START_BUILD_PACKAGE&succeed=true&namespace=$CI_REGISTRY_IMAGE&environment=$CI_COMMIT_REF_NAME&commit_hash=$CI_COMMIT_SHA&tag=$CI_COMMIT_SHORT_SHA"; then
        logerror cannot talk to wing server.
        return 1
    fi

    # start build
    if ! "./$build_script"; then
        logerror build failure.
        if ! _ci_wing_report "product_token=$WING_PRODUCT_TOKEN&reason=SCM.BuildProductFailure&type=$WING_REPORT_TYPE_FINISH_BUILD_PACKAGE&namespace=$CI_REGISTRY_IMAGE&environment=$CI_COMMIT_REF_NAME&commit_hash=$CI_COMMIT_SHA&tag=$CI_COMMIT_SHORT_SHA"; then
            logerror cannot talk to wing server.
        fi
        return 1
    fi
    # upload product.
    if ! ci_build gitlab-package -e $CI_COMMIT_REF_NAME "$product_path"; then
        logerror upload product failure.
        if ! _ci_wing_report "product_token=$WING_PRODUCT_TOKEN&reason=SCM.UploadProductFailure&type=$WING_REPORT_TYPE_FINISH_BUILD_PACKAGE&namespace=$CI_REGISTRY_IMAGE&environment=$CI_COMMIT_REF_NAME&commit_hash=$CI_COMMIT_SHA&tag=$CI_COMMIT_SHORT_SHA"; then
            logerror cannot talk to wing server.
        fi
        return 1
    fi

    # sync success.
    if ! _ci_wing_report "product_token=$WING_PRODUCT_TOKEN&type=$WING_REPORT_TYPE_FINISH_BUILD_PACKAGE&succeed=true&namespace=$CI_REGISTRY_IMAGE&environment=$CI_COMMIT_REF_NAME&commit_hash=$CI_COMMIT_SHA&tag=$CI_COMMIT_SHORT_SHA"; then
        logerror cannot talk to wing server.
        return 1
    fi
}