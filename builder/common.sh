sar_import utils.sh

_ci_get_env_value() {
    local env=$1
    if [ ! -z "$env" ]; then
        eval "local value=\$$env"
        if [ ! -z "$value" ]; then
            local env=$value
        fi
    fi
    echo $env
}

_ci_get_image_ref() {
    local prefix=`strip $1`
    local env=`_ci_get_env_value "$2" | xargs`
    local tag=`strip $3`
    path_join "$prefix" "$env" "$tag"
}