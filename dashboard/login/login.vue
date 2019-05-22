<template>
    <div class='login-page-container-h'>
            <div class="login-box">
                <div class="login-title-box">
                    <div class="login-logo-box">
                        <img src="../res/logo_white.svg" class="login-logo" />
                    </div>
                </div>
                <div class="login-ui-container">
                    <el-tabs v-model="activeTab">
                        <el-tab-pane v-bind:label="localetext.login" name="login">{{ localetext.login }}</el-tab-pane>
                        <el-tab-pane v-bind:label="localetext.register" name="register">{{ localetext.register }}</el-tab-pane>
                    </el-tabs>
                    <div v-if="activeTab=='login'">
                        <div class="login-form-box">
                            <div class="login-form-item-title"><span>{{ localetext.account }}</span> <span class="login-small-title">{{ localetext.account_subname }}</span></div>
                            <el-input ref="loginUsernameInputbox" class="login-input-box" v-model="username" v-bind:placeholder="localetext.username_prompt" ></el-input>
                            <div class="login-form-item-title">{{ localetext.password }} <span class="login-small-title">{{ localetext.password_subname }}</span></div>
                            <el-input ref="loginPasswordInputbox" class="login-input-box" v-model="password" v-bind:placeholder="localetext.password_prompt" type="password"></el-input>
                        </div>
                        <div class="login-action-bar">
                            <p class="login-message-box">{{ errorMessage }}</p>
                            <el-button type="primary" style="float: right" v-on:click="login">{{ localetext.login }}</el-button>
                        </div>
                    </div>
                    <div v-if="activeTab=='register'">
                        <div class="login-form-box">
                            <div class="login-form-item-title"><span>{{ localetext.account }}</span> <span class="login-small-title">{{ localetext.account_subname }}</span></div>
                            <el-input ref="registerUsernameInputbox" class="login-input-box" v-model="username" v-bind:placeholder="localetext.username_prompt" ></el-input>
                            <div class="login-form-item-title">{{ localetext.password }} <span class="login-small-title">{{ localetext.password_subname }}</span></div>
                            <el-input ref="registerPasswordInputbox" class="login-input-box" v-model="password" v-bind:placeholder="localetext.password_prompt" type="password"></el-input>
                            <div class="login-form-item-title">{{ localetext.password_confrim }} <span class="login-small-title">{{ localetext.password_confrim_subname }}</span></div>
                            <el-input ref="registerPasswordConfrimBox" class="login-input-box" v-model="passwordConfrim" v-bind:placeholder="localetext.password_confrim_prompt" type="password"></el-input>
                        </div>
                        <div class="login-action-bar">
                            <p class="login-message-box">{{ errorMessage }}</p>
                            <el-button type="primary" style="float: right" v-on:click="register">{{ localetext.register }}</el-button>
                        </div>
                    </div>
                </div>
            </div>
    </div>
</template>
<script>
import {login, register} from './proc.js'

export default {
    data(){
        return {
            username: "",
            password: "",
            passwordConfrim: "",
            activeTab: "login",
            errorMessage: ""
        }
    },
    props: ["localetext"],
    methods: {
        login,
        register
    }
}
</script>
<style>
</style>
