PACKAGE_BASE_IMAGE='alpine:3.7'

SAR_RUNTIME_PKG_PREFIX='registry.stuhome.com/devops/runtime'
SAR_RUNTIME_PKG_ENV='master'
SAR_RUNTIME_PKG_TAG='latest'

SAR_RUNTIME_ALPINE_APK_MIRROR='mirrors.tuna.tsinghua.edu.cn'
SAR_RUNTIME_ALPINE_DEPENDENCIES=(
    'jq' 'bash' 'supervisor' 'coreutils' 'procps' 'vim' 'net-tools'
    'bind-tools' 'tzdata' 'gettext' 'py2-pip'
)

SAR_RUNTIME_SYS_PYTHON_DEPENDENCIES=(
    'supervisor-logging' 'ipython'
)