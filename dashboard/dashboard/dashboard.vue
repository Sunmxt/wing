<template>
    <div class="container wing" v-on:load="onLoginShow">
        <el-container class="container">
            <el-aside class="nav" width="6cm">
                <div class="logo-box">
                    <div>
                        <img class="logo" src="../res/logo.svg" /><div class="logo-text">Wing</div>
                    </div>
                </div>
                <el-menu class="nav-menu">
                    <el-menu-item index="1">
                        <i class="el-icon-menu"></i>
                        <span>概览</span><span style="font-size: 0.4em"> Overview</span>
                    </el-menu-item>
                    <el-menu-item index="2">
                        <i class="el-icon-menu"></i>
                        <span>应用编排</span><span style="font-size: 0.4em"> Orchestration</span>
                    </el-menu-item>
                </el-menu>
            </el-aside>
            <el-container>
                <el-header class="title" height="1.5cm">
                    <div class="board-title"><span>应用编排</span></div>
                </el-header>
                <el-main>
                </el-main>
            </el-container>
        </el-container>
    </div>
</template>
<script>
export default {
    name: "Dashboard",
    data(){
        return {
            input: ''
        }
    },
    methods: {
        onLoginShow(){
            console.log('show login')
        }
    },
    props: []
}
</script>
<style>

.container, html, body{
    float: left;
    height: 100%;
    width: 100%;
}

body {
    margin: 0;
}
</style>