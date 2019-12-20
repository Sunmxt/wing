PACKAGE_BASE_IMAGE='alpine:3.7'

SAR_CI_REGISTRY='registry.stuhome.com'

SAR_RUNTIME_PKG_PREFIX='registry.stuhome.com/devops/runtime'
SAR_RUNTIME_PKG_ENV='sar-runtime-master'
SAR_RUNTIME_PKG_TAG='latest'
SAR_RUNTIME_PKG_PROJECT_PATH='sunmxt/wing'

# alpine
SAR_RUNTIME_ALPINE_APK_MIRROR='mirrors.tuna.tsinghua.edu.cn'
SAR_RUNTIME_ALPINE_DEPENDENCIES=(
    'jq' 'bash' 'supervisor' 'coreutils' 'procps' 'vim' 'net-tools'
    'bind-tools' 'tzdata' 'gettext' 'py2-pip'
)

# python
SAR_RUNTIME_SYS_PYTHON_DEPENDENCIES=(
    'supervisor-logging' 'ipython'
)
SAR_PYTHON_MIRRORS='https://pypi.tuna.tsinghua.edu.cn/simple'

SAR_RUNTIME_APP_DEFAULT_WORKING_DIR='/'