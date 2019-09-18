# StarStudio Runtime

工作室的基础运行时框架，目的是规范化容器打包，减轻心智负担。

主要提供各类拓展的命令行工具，包括但不限于 CI、docker 拓展命令。后面会集成服务注册、发现的 agent。符合 runtime 规范的容器具备一些额外的功能（比如服务发现）。

可以在里面添加更多的脚本，脚本里可以灵活的使用框架所实现的各种基础工具（如日志、静态 json 配置）。

#### 目前已有的组件
- log.sh: 日志库
- ci.sh: 各类 ci 工具，简化 CI 接入，规范镜像等。
- install.sh: 安装器，用于将 runtime 安装到容器或其他运行环境中。
- docker_utils.sh: 一些 docker 拓展命令（比如检查一个镜像是否存在）。
- utils.sh: 一些实用工具函数。

### 安装

```bash
./install.sh /opt/sar
```

### 使用

安装后就可以使用runtime的各种拓展命令了。目前已导出拓展命令如下，安装后即可使用：

- log.sh (日志)
  - loginfo
  - logerror
  - logwarn
- utils.sh (工具集合)
  - path_join
  - is_image_exists
  - docker_installed
- builder/ci.sh (持续集成、构建相关)
  - ci_build
  - ci_package_pull
  - ci_login
- builder/runtime_image.sh (镜像构建)
  - build_runtime_image
  - runtime_image_base_image
  - runtime_image_add_dependency
  - runtime_image_bootstrap_run
  - runtime_image_pre_build_run
  - runtime_image_post_build_run
  - runtime_image_pre_build_script
  - runtime_image_post_build_script
  - runtime_image_build_start

不带任何参数执行相关命令可以查看帮助。

如果要使用未导出的函数，任意目录执行

```bash
source runtime_env
```

## 开发

### 环境

```bash
source bin/sar_activate
```

然后直接调用你写的函数即可。若修改了代码，重新 source 一次或者使用 sar_import 一下相应的文件即可。

### sar_import

sar_import 用于引入其他脚本的内容，用于替代 source。提供这一语句，是为了不用关心 PATH 问题。

sar_import 指令屏蔽了执行环境的差异，开发环境和安装后的运行环境是兼容的，可以不必关心 PATH 的问题。

例子：

```bash
#! /bin/bash
sar_import log.sh # 引入日志库
sar_import utils.sh # 引入工具库

loginfo 打一条日志 # loginfo 是由 log.sh 实现的日志打印函数
```

### 如何让 install.sh 安装新添加的文件

只需要在 settings/componment.sh 的 COMPONMENT 数组添加你自己的脚本即可。

### 导出命令 (exported command)

导出的命令，指的是安装后，直接可以使用而不必首先 **source runtime_env** 的命令。

若需要导出命令，在 settings/componment.sh 的 EXPORT_COMMANDS 添加相应的命令即可。