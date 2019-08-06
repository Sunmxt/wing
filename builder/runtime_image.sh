sar_import ci.sh
sar_import settings/image.sh

_runtime_image_get_stash_dir() {
    local prefix=$1
    local env=$2
    local tag=$3
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
    while getopts 't:e:r:' opt; do
        case $opt in
            t)
                local ci_image_tag=$OPTARG
                ;;
            e)
                local ci_image_env_name=$OPTARG
                ;;
            r)
                local ci_image_prefix=$OPTARG
                ;;
            *)
                logerror "[runtime_image_builder]" unexcepted options -$opt.
                ;;
        esac
    done
    

    `_runtime_image_get_stash_dir $ci_image_tag $ci_image_env_name`
}