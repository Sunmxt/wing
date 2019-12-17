COMPONENTS=(
    "lib.sh" "runtime.sh" "docker.sh" "log.sh" "utils.sh" "bundle.sh" "binary.sh"

    'settings/component.sh' 'settings/image.sh' 'settings/install.sh' 'settings/wing.sh'
    'settings/binary.sh' 'settings/bundle.sh' 'settings/wing.sh'

    'builder/ci.sh' 'builder/common.sh' 'builder/runtime_image.sh' 'builder/validate.sh'

    'libexec/yq_darwin_amd64' 'libexec/yq_linux_amd64'
)

EXPORT_COMMANDS=(
    "ci_build" "ci_package_pull" "ci_login"
    "runtime_image_post_build_script" "runtime_image_pre_build_script"
    "runtime_image_add_service" "runtime_image_add_dependency"
    "runtime_image_base_image" "build_runtime_image"
    "logerror" "logwarn" "loginfo"
    "path_join"
    "is_image_exists"
    "docker_installed"
    "yq"
)
