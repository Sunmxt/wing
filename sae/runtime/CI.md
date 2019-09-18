# 接入 CI（持续集成） 

DevOps 组基于 Gitlab-CI 提供了一套构建工具，用于**构建产物打包📦、上传⏫**和**标准镜像的构建**。



- 打包规范（NET规则）
- 用 .gitlab-ci.yml 定制你自己的 CI Pipeline



## 打包规范 (NET规则)

我们用 **N**amespace（命名空间）、**E**nvironment（环境）、**T**ag（标签），来唯一确定一个构建所产生的产物。

### 命名空间 （Namespace）

我们以 Gitlab 项目的命名空间为准，一般是 ```<组名>/<项目名>``` 或```<用户名>/<项目名>```。以 DevOps 组的 Dockerepo 为例，Namespace 为 ```devops/dockerepo```

### 环境（Environment）

环境一般指线上环境、测试环境等。需要指定名称。

### 标签（Tag）

标签用于区分同一项目下同一环境下的多个不同产物，推荐用 commit hash 作为 Tag。



## 用 .gitlab-ci.yml 定制你自己的 CI Pipeline

最简单的 CI 只包括构建和打包的过程 (stage)，在 gitlab-ci.yml 中只有一个stage。以 一个 golang 项目为例：

```yaml
image: registry.stuhome.com/devops/dockerepo/golang:1.11-alpine # 请选择工作室提供的基础镜像，只有这些镜像内可直接使用 CI 拓展命令。

### 请包括这一部分
services:
    - docker:18.09.7-dind
variables:
    DOCKER_HOST: tcp://docker:2375/
    DOCKER_DRIVER: overlay2
    GIT_SUBMODULE_STRATEGY: recursive
### 


stages: # 定义你的构建阶段
    - build

package-nginx-base-conf:
    stage: build
    script:
    - make # 构建命令，需要替换为你的项目的构建脚本或命令
    - ci_build gitlab-package -e wing_$CI_COMMIT_REF_NAME ./bin # 上传产物, -e 指定了环境，Tag 和 Namespace 会自动确定，默认分别为 commit hash 的前10位 和 该项目的 gitlab namespace 名称。-e 也可以指定一个环境变量名，会自动替换成对应的值。比如 -e CI_COMMIT_REF_NAME，会替换为对应的分支名称。
```

