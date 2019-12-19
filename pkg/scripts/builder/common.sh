sar_import utils.sh

_ci_get_package_ref() {
    local prefix=`_ci_build_generate_registry $1`
    local env=`_ci_build_generate_env_ref "$2"`
    local tag=`_ci_build_generate_tag "$3"`
    if [ ! -z "$env" ]; then
        echo -n $prefix/$env/sar__package:$tag
    else
        echo -n $prefix/sar__package:$tag
    fi
}

_ci_build_generate_registry() {
    local registry=`strip "$1"`
    registry=${registry:="$SAR_CI_REGISTRY"}
    registry=${registry:="docker.io"}
    echo "$prefix" | sed -E 's/(\/)*$//g'
}

_ci_build_generate_env_ref() {
    local env=`strip "$1"`

    if [ ! -z "$env" ]; then
        # from environment variables.
        eval "local path_appended=\$$env"
        
        if [ ! -z "$path_appended" ]; then
            env="$path_appended"
            local registry_path=$registry_path/$path_appended
        else
        fi
    else
        # use default path.
        if [ ! -z "$GITLAB_CI" ]; then
            # Gitlab CI.
            
        fi

        # fallback: from local refs.
        if [ -z "$env" ]; then
            # from local repo: use branch name.
            env=`git rev-parse --abbrev-ref HEAD | tr -d ' ' | tr '-[:lower:]' '_[:upper:]' | sed 's/^_*//g'`
        fi
    fi

    if [ -z "$env" ]; then
        logerror "cannot resolve environment."
        return 1
    fi
}

_ci_build_generate_tag() {
    local tag=`strip "$1"`
    if [ -z "$tag" ] && [ ! -z "$GITLAB_CI" ]; then
        # Gitlab CI.
        tag=${CI_COMMIT_SHA:0:10}
    fi
    # from local repo
    if [ -z "$tag" ]; then
        env=`git rev-parse HEAD | tr -d ' ' | tr '-[:lower:]' '_[:upper:]' | sed 's/^_*//g'`
        tag=${tag:0:10}
    fi
    tag=${tag:=latest}
    echo -n "$tag"
}

_ci_build_generate_registry_path() {
    local prefix=`__ci_build_generate_registry "$1"`
    local env=`_ci_build_generate_env_ref "$2"`
    if [ -z "$prefix" ] || [ -z "$env" ]; then
        return 1
    fi
    echo "$registry_path/$env"
}