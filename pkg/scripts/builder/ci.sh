sar_import lib.sh
sar_import docker.sh
sar_import builder/common.sh
sar_import builder/archifacts.sh
sar_import builder/docker.sh
sar_import settings/image.sh

enable_dockercli_experimentals


help_ci_build() {
    echo '
Project builder in CI Environment.

ci_build <mode> [options] -- [docker build options]

mode:
  docker                build and push to gitlab registry.
  package               upload archifacts in local development environment.

options:
      -t <tag>                            Image tag / package tag (package mode)
                                          if `gitlab_ci_commit_hash` is specified in 
                                          `gitlab-runner-docker` mode, the tag will be substitute 
                                          with actually commit hash.
      -e <environment_variable_name>      Identify docker path by environment variable.
      -r <host>                           registry.
      -p <project_path>                   project path.
      -s                                  Do not push image to regsitry.
      -h <path_to_hash>                   Use file(s) hash for tag.
      -f                                  force to build.

example:
      ci_build package -p be/recruitment2019 bin
      ci_build docker -p be/recruitment2019 .
      ci_build docker -e dev -- --build-arg="myvar=1" .
      ci_build docker -t gitlab_ci_commit_hash -r registry.mine.cn -p test/myimage -e dev -- --build-arg="myvar=1" .
      ci_build docker -t stable_version -r registry.mine.cn/test/myimage .
'
}

ci_build() {
    local mode=$1
    shift 1
    case $mode in
        docker)
            _ci_auto_docker_build $*
            return $?
            ;;
        package)
            _ci_auto_package $*
            return $?
            ;;
        wing-gitlab)
            _ci_wing_gitlab_package_build $*
            return $?
            ;;
        *)
            logerror unsupported ci type: $mode
            help_ci_build
            return 1
            ;;
    esac
}

ci_login_help() {
    echo '
Sign up to ci environment.

usage:
    ci_login [-p password] <username>
'
}

ci_login() {
    if ! docker_installed ; then
        logerror [ci_login] docker not installed.
        return 1
    fi
    OPTIND=0
    while getopts 'p:' opt; do
        case $opt in
            p)
                local password=$OPTARG
                ;;
            *)
                logerror unsupported option: $opt.
                return 1
                ;;
        esac
    done
    local password
    if [ ! -z "$password" ]; then
        logwarn Attention! Specifing plain password is not recommanded.
    else
        while [ -z "$password" ]; do
            loginfo Empty password.
            echo -n 'Password: '
            read -s password
            echo
        done
    fi

    eval "local __=\${$OPTIND}"
    local username="`strip \"$__\"`"
    if [ -z "$username" ] ; then
        logerror username is empty.
        return 1
    fi

    docker login "$SAR_CI_REGISTRY" -u "$username" -p "$password"
}