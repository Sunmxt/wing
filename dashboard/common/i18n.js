export default {
    cn: {
        ui: {
            login: {
                login: "登录",
                register: "注册",
                account: "账户",
                password: "密码",
                account_subname: "Account",
                password_subname: "Password",
                prompt: {
                    username: "用户名",
                    password: "密码",
                    password_confirm: "重复输入以确认密码"
                },
                password_confirm: "确认密码",
                password_confirm_subname: "Password confirm",
                username_missing: "用户名不能为空",
                password_missing: "密码不能为空",
                password_confirm_missing: "确认密码不能为空",
                password_confirm_unmatched: "确认密码不一致"
            },
            dashboard: {
                nav: {
                    overview: "概览",
                    orchestration: "应用编排"
                }
            },
            orchestration: {
                title: "应用编排",
                create_application: "创建应用",
                submit: "提交",
                form: {
                    basic_info: "基本信息",
                    application_name: "应用名",
                    application_description: "应用描述",
                    application_description_prompt: "在此添加应用说明",

                    runtime_info: "运行信息",
                    image: "基础镜像",
                    image_prompt: "如: registry.stuhome.com/dockerepo/alpine",
                    image_tag_prompt: "如: 1.0.1",
                    bootstrap_command: "启动命令",
                    bootstrap_command_prompt: "启动命令",
                    environment_variable: "环境变量",
                    variable_name: "变量名",
                    variable_value: "值",
                    add_variable: "添加",

                    resource_quota: "资源配置",
                    cpu_quota: "CPU",
                    cpu_quota_prompt: "期望CPU配额（如：0.1）",
                    memory_quota: "内存",
                    memory_quota_prompt: "期望内存配额（单位：MB，如：0.1）",
                },

                my_application: "我的应用",
                all_application: "所有应用",
                application_name: "应用名",
                application_state: "应用状态",
                version_number: "版本号",
                owner: "所有者",
                resource_usage: "资源使用",
                operation: "操作",
                operations: {
                    deploy: "部署",
                    rollback: "回滚",
                    shutdown: "停止",
                    upgrade: "升级",
                    remove: "删除",
                }
            }
        }
    },
    en: {
        ui: {
            login: {
                login: "Login",
                register: "Register",
                account: "Account",
                password: "Password",
                account_subname: "",
                password_subname: "",
                prompt: {
                    username: "Username",
                    password: "Password",
                    password_confirm: "Confirm password"
                },
                password_confirm: "Password confrim",
                password_confirm_subname: "",
                username_missing: "Username should not be empty.",
                password_missing: "Password should not be empty/",
                password_confirm_missing: "Confrim password should not be empty",
                password_confirm_unmatched: "Confrim password does not matched."
            },
            dashboard: {
                nav: {
                    overview: "Overview",
                    orchestration: "Orchestration"
                }
            },
            orchestration: {
                title: "Orchestration",
                create_application: "New",
                my_application: "Mine",
                all_application: "All",
                application_name: "Name",
                application_state: "State",
                version_number: "Version",
                owner: "Owner",
                resource_usage: "Resource usage",
                operation: "Modify",
                operations: {
                    deploy: "Deploy",
                    rollback: "Rollback",
                    shutdown: "Shutdown",
                    upgrade: "Upgrade",
                    remove: "Remove",
                },
                submit: "Submit",
                form: {
                    basic_info: "Basic Information",
                    application_name: "Name",
                    application_description: "Description",
                    application_description_prompt: "Add application description here.",

                    runtime_info: "Runtime settings",
                    image: "Image",
                    image_prompt: "example: registry.stuhome.com/dockerepo/alpine",
                    image_tag_prompt: "example: 1.0.1",
                    bootstrap_command: "command",
                    bootstrap_command_prompt: "command",
                    environment_variable: "Environment variables",
                    variable_name: "name",
                    variable_value: "value",
                    add_variable: "append",

                    resource_quota: "Resource",
                    cpu_quota: "CPU",
                    cpu_quota_prompt: "Required CPU quota（example：0.1）",
                    memory_quota: "memory",
                    memory_quota_prompt: "Required memory quota（unit：MB，example：0.1）",
                }
            }
        }
    }
}