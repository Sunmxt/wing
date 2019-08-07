
_ci_get_env_value() {
    local env=$1
    if [ ! -z "$env" ]; the
        eval "local value=\$$env"
        if [ ! -z "$value" ]; then
            local env=$value
        fi
    fi
    echo $env
}