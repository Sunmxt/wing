sar_import builder/common.sh
sar_import settings/image.sh
sar_import utils.sh

_runtime_image_stash_prefix() {
    local prefix=$1
    local env=$2
    local tag=$3
    `hash_for_key "$prefix" "$env" "$tag"`
}

build_runtime_image_help() {
    echo '
Start build runtime image.

usage:
    build_runtime_image -t <tag> -e <environment_varaible_name> -r prefix

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
                local ci_image_env_name=`_ci_get_env_name "$OPTARG"`
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
    local stash_prefix=`_runtime_image_stash_prefix "$prefix" "$ci_image_env_name" "$ci_image_tag"`
    eval "_SAR_RT_BUILD_${context}_${stash_prefix}_PREFIX=$ci_image_prefix"
    eval "_SAR_RT_BUILD_${context}_${stash_prefix}_ENV=$ci_image_env_name"
    eval "_SAR_RT_BUILD_${context}_${stash_prefix}_TAG=$ci_image_tag"
    eval "_SAR_RT_BUILD_${context}_STASH_PREFIX=${stash_prefix}"
}

build_runtime_image_base_image_help() {
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

build_runtime_image_base_image() {
    OPTIND=0
    while getopts 'c:' opt; do
        case $opt in
            c)
                local context=$OPTARG
                ;;
            *)
                build_runtime_image_base_image_help
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
    eval "local stash_prefix=\$_SAR_RT_BUILD_${context}_STASH_PREFIX"
    eval "local _SAR_RT_BUILD_${context}_BASE_IMAGE=$base_image"
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
                local ci_package_env_name=`_ci_get_env_name "$OPTARG"`
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
}