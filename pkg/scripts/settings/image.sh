PACKAGE_BASE_IMAGE='registry.stuhome.com/sunmxt/wing/alpine/master:3.7'

SAR_CI_REGISTRY='registry.stuhome.com'

SAR_RUNTIME_PKG_PREFIX='registry.stuhome.com/devops/runtime'
SAR_RUNTIME_PKG_ENV='sae_runtime_master'
SAR_RUNTIME_PKG_TAG='latest'
SAR_RUNTIME_PKG_PROJECT_PATH='sunmxt/wing'

# alpine
SAR_RUNTIME_ALPINE_DEPENDENCIES=(
    'jq' 'bash' 'supervisor' 'coreutils' 'procps' 'vim' 'net-tools'
    'bind-tools' 'tzdata' 'gettext' 'py2-pip' 'curl'
)

# centos
SAR_RUNTIME_YUM_DEPENDENCIES=(
    'bash' 'supervisor' 'jq' 'vim' 'python' 'net-tools' 'tzdata'
    'gettext' 'python-pip' 'coreutils' 'procps' 'curl' 'bind-utils'
)

# apt
SAR_RUNTIME_APT_DEPENDENCIES=(
    'bash' 'supervisor' 'jq' 'vim' 'python' 'net-tools' 'tzdata'
    'gettext' 'python-pip' 'coreutils' 'procps' 'curl' 'dnsutils'
)

# python
SAR_RUNTIME_SYS_PYTHON_DEPENDENCIES=(
    'ipython'
)
SAR_PYTHON_MIRRORS='https://pypi.tuna.tsinghua.edu.cn/simple'

SAR_RUNTIME_APP_DEFAULT_WORKING_DIR='/'