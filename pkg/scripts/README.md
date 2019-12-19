# Runtime scripts

基础运行时框架，目标是支持工程化配置，不再分散维护脚本。运行环境应该包含这些的脚本。

主要提供各类拓展的命令行工具，包括但不限于 CI、docker 拓展命令。符合 runtime 规范的节点具备一些额外的功能（比如服务发现）。

可以在里面添加更多的脚本，脚本里可以灵活的使用框架所实现的各种基础工具（如日志、静态 json 配置）。

目前主要是 bash scripts. 也支持集成二进制文件到整个框架内。



### 打包

支持将所有的脚本按照依赖关系打成一个脚本。也支持打包二进制（支持不同平台）到脚本内。

```bash
bin/sar_execute sar_bundle ./dist         # dist目录,默认入口点:main.sh
bin/sar_execute sar_bundle ./dist shim.sh # dist目录,入口点:shim.sh
```

#### 集成二进制文件

例子: 打包 yq (一个 golang 实现的 yaml 预处理器)

```bash
# yq darwin x86_64 版本，二进制文件在 libexec/yq_darwin_amd64
sar_register_binary yq darwin x86_64 libexec/yq_darwin_amd64
# --lazy-load 使得二进制只在被使用到的时候解包。这种方式能够提高加载速度。当然，没有打包就没有解包的说法，--lazy-load 只影响打包后行为。
sar_register_binary --lazy-load yq darwin x86_64 libexec/yq_darwin_amd64
# --no-bundled 使得二进制文件不会被打包
sar_register_binary --no-bundled yq darwin x86_64 libexec/yq_darwin_amd64
```

引用 yq

```bash
yq new .... # 像平常一样使用即可
```



### 安装

```bash
./install.sh /opt/sar
```



### 非安装的使用方式

#### 导入到整个环境

```bash
source bin/sar_runtime
```

#### 直接运行某个命令

```bash
bin/sar_execute <command> ...
```

#### 直接使用打包后的脚本

```bash
source dist/main.sh
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

不带任何参数或者带 "-h" 执行相关命令可以查看帮助。

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

sar_import 用于引入其他脚本的内容，相当于 source。但 sar_import 在背后做了比 source 更多的工作，比如分析依赖关系等。

sar_import 指令屏蔽了执行环境的差异，开发环境和安装后的运行环境是兼容的，可以不必关心 PATH 的问题。

例子：

```bash
#! /bin/bash
sar_import log.sh # 引入日志库
sar_import utils.sh # 引入工具库

loginfo 打一条日志 # loginfo 是由 log.sh 实现的日志打印函数
```

### 如何让 install.sh 安装新添加的文件

只需要在 settings/componment.sh 的 COMPONMENT 数组添加你自己的脚本(整个目录也是可以的)即可。

### 导出命令 (exported command)

导出的命令，指的是安装后，直接可以使用而不必首先 **source runtime_env** 的命令。

若需要导出命令，在 settings/componment.sh 的 EXPORT_COMMANDS 添加相应的命令即可。