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
                        <el-tab-pane v-bind:label="$t('ui.login.login')" name="login">{{ $t('ui.login.login') }}</el-tab-pane>
                        <el-tab-pane v-bind:label="$t('ui.login.register')" name="register">{{ $t('ui.login.register') }}</el-tab-pane>
                    </el-tabs>
                    <div v-if="activeTab=='login'">
                        <div class="login-form-box">
                            <div class="login-form-item-title"><span>{{ $t('ui.login.account') }}</span> <span class="login-small-title">{{ $t('ui.login.account_subname') }}</span></div>
                            <el-input ref="loginUsernameInputbox" class="login-input-box" v-model="username" v-bind:placeholder="$t('ui.login.prompt.username')" ></el-input>
                            <div class="login-form-item-title">{{ $t('ui.login.password') }} <span class="login-small-title">{{ $t('ui.login.password_subname') }}</span></div>
                            <el-input ref="loginPasswordInputbox" class="login-input-box" v-model="password" v-bind:placeholder="$t('ui.login.prompt.password')" type="password"></el-input>
                        </div>
                        <div class="login-action-bar">
                            <p class="login-message-box">{{ errorMessage }}</p>
                            <el-button type="primary" style="float: right" v-on:click="login">{{ $t('ui.login.login') }}</el-button>
                        </div>
                    </div>
                    <div v-if="activeTab=='register'">
                        <div class="login-form-box">
                            <div class="login-form-item-title"><span>{{ $t('ui.login.register') }}</span> <span class="login-small-title">{{ $t('ui.login.account_subname') }}</span></div>
                            <el-input ref="registerUsernameInputbox" class="login-input-box" v-model="username" v-bind:placeholder="$t('ui.login.prompt.username')" ></el-input>
                            <div class="login-form-item-title">{{ $t('ui.login.password') }} <span class="login-small-title">{{ $t('ui.login.password_subname') }}</span></div>
                            <el-input ref="registerPasswordInputbox" class="login-input-box" v-model="password" v-bind:placeholder="$t('ui.login.prompt.password')" type="password"></el-input>
                            <div class="login-form-item-title">{{ $t('ui.login.password_confirm') }} <span class="login-small-title">{{ $t('ui.login.password_confirm_subname') }}</span></div>
                            <el-input ref="registerPasswordConfrimBox" class="login-input-box" v-model="passwordConfrim" v-bind:placeholder="$t('ui.login.prompt.password_confirm')" type="password"></el-input>
                        </div>
                        <div class="login-action-bar">
                            <p class="login-message-box">{{ errorMessage }}</p>
                            <el-button type="primary" style="float: right" v-on:click="register">{{ $t('ui.login.register') }}</el-button>
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
    methods: {
        login,
        register
    }
}
</script>
<style>
</style>
