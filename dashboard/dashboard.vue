<template>
    <div class='login-page-container-h'>
            <div class="login-box">
                <div class="login-title-box">
                    <div class="login-logo-box">
                        <img src="./res/logo_white.svg" class="login-logo" />
                    </div>
                </div>
                <div class="login-ui-container">
                    <el-tabs v-model="activeTab">
                        <el-tab-pane label="登录" name="login">登录</el-tab-pane>
                        <el-tab-pane label="注册" name="register">注册</el-tab-pane>
                    </el-tabs>
                    <div v-if="activeTab=='login'">
                        <div class="login-form-box">
                            <div class="login-form-item-title"><span>账户</span> <span class="login-small-title">Account</span></div>
                            <el-input class="login-input-box" v-model="username" placeholder="用户名" ></el-input>
                            <div class="login-form-item-title">密码 <span class="login-small-title">Password</span></div>
                            <el-input class="login-input-box" v-model="password" placeholder="密码" ></el-input>
                        </div>
                        <div class="login-action-bar">
                            <p class="login-message-box">{{ errorMessage }}</p>
                            <el-button type="primary" style="float: right">登录</el-button>
                        </div>
                    </div>
                    <div v-if="activeTab=='register'">
                        <div class="login-form-box">
                            <div class="login-form-item-title"><span>账户</span> <span class="login-small-title">Account</span></div>
                            <el-input class="login-input-box" v-model="username" placeholder="用户名" ></el-input>
                            <div class="login-form-item-title">密码 <span class="login-small-title">Password</span></div>
                            <el-input class="login-input-box" v-model="password" placeholder="密码" ></el-input>
                            <div class="login-form-item-title">确认密码 <span class="login-small-title">Password confirm</span></div>
                            <el-input class="login-input-box" v-model="passwrodConfirm" placeholder="确认密码" ></el-input>
                        </div>
                        <div class="login-action-bar">
                            <p class="login-message-box">{{ errorMessage }}</p>
                            <el-button type="primary" style="float: right">注册</el-button>
                        </div>
                    </div>
                </div>
            </div>
    </div>
</template>
<script>
export default {
    data(){
        return {
            username: "",
            password: "",
            passwrodConfirm: "",
            activeTab: "login",
            errorMessage: ""
        }
    }
}
</script>
<style>
</style>
