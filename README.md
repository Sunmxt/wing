# StarStudio Runtime

工作室的基础运行时框架。

主要提供各类拓展的命令行工具，包括但不限于 CI、docker 拓展命令。

你可以在里面添加更多的脚本，脚本里可以灵活的使用框架所实现的各种基础工具（如日志）。



### 安装

```bash
./install.sh /opt/sar
```

### 使用

任意目录执行

```bash
source runtime_env
```

就可以愉快的使用各种拓展命令了。



## 开发

### 环境

```bash
source debug_env.sh
```

然后直接调用你写的函数即可。若修改了代码，重新 source 一次或者使用 sar_import 一下相应的文件即可。

### sar_import

sar_import 用于引入其他脚本的内容。提供这一语句，是为了不用关心 PATH 问题。

sar_import 指令在开发环境和安装后的运行环境是兼容的，你可以不必关心 PATH 的问题，大胆使用即可。

例子：

```bash
#! /bin/bash
sar_import log.sh # 引入日志库
sar_import utils.sh # 引入工具库

loginfo 打一条日志 # loginfo 是由 log.sh 实现的日志打印函数
```

### 如何让 install.sh 安装我添加的脚本到安装环境

只需要在 settings/componment.sh 的 COMPONMENT 数组添加你自己的脚本即可。

