sar_import utils.sh

_ci_get_package_ref() {
    local host=`_ci_build_generate_registry $1`
    local path=`_ci_build_generate_package_path $2`
    local env=`_ci_build_generate_env_ref "$3"`
    local tag=`_ci_build_generate_tag "$4"`
    if [ -z "$host" ] || [ -z "$path" ]; then
        return 1
    fi
    if [ ! -z "$env" ]; then
        echo -n $host/$path/$env/sar__package:$tag
    else
        echo -n $host/$path/sar__package:$tag
    fi
}

_ci_build_generate_registry() {
    local registry=`strip "$1"`
    registry=${registry:="$SAR_CI_REGISTRY"}
    registry=${registry:="docker.io"}
    echo "$registry" | sed -E 's/^(\/)*$//g'
}

_ci_build_generate_env_ref() {
    local env=`strip "$1"`

    if [ ! -z "$env" ]; then
        # from environment variables.
        if echo "$env" | grep -qE '^[[:alnum:]_]+$'; then
            eval "local path_appended=\${$env}"
        fi
        
        if [ ! -z "$path_appended" ]; then
            env="$path_appended"
        fi
    else
        # use default path.
        if [ ! -z "$GITLAB_CI" ]; then
            # Gitlab CI.
            env="$CI_COMMIT_REF_NAME"
        fi

        # fallback: from local refs.
        if [ -z "$env" ]; then
            # from local repo: use branch name.
            env=`git rev-parse --abbrev-ref HEAD`
        fi
    fi

    if [ -z "$env" ]; then
        logerror "cannot resolve environment."
        return 1
    fi
    echo -n "$env" | tr -d ' ' | tr '[:upper:]' '[:lower:]' | sed 's/^_*//g'
}

_ci_build_generate_package_path() {
    local path=`strip "$1"`
    if [ -z "$path" ] && [ ! -z "$GITLAB_CI" ]; then
        path="$CI_PROJECT_PATH"
    fi
    if [ -z "$path" ]; then
        logerror "project path not given."
        return 1
    fi
    echo -n "$path" | tr -d ' ' | tr '[:upper:]' '[:lower:]' | sed 's/^_*//g'
}

_ci_build_generate_tag() {
    local tag=`strip "$1"`
    if [ -z "$tag" ] && [ ! -z "$GITLAB_CI" ]; then
        # Gitlab CI.
        tag=${CI_COMMIT_SHA:0:10}
    fi
    # from local repo
    if [ -z "$tag" ]; then
        env=`git rev-parse HEAD `
        tag=${tag:0:10}
    fi
    tag=${tag:=latest}
    echo -n "$tag" | tr -d ' ' | tr '[:upper:]' '[:lower:]' | sed 's/^_*//g'
}

_ci_build_generate_registry_prefix() {
    local host=`_ci_build_generate_registry "$1"`
    local package_path=`_ci_build_generate_package_path "$2"`
    local env=`_ci_build_generate_env_ref "$3"`
    if [ -z "$host" ] || [ -z "$package_path" ] ||  [ -z "$env" ]; then
        return 1
    fi
    echo "$host/$package_path/$env"
}