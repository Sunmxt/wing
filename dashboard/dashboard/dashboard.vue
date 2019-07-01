<template>
    <div class="container wing">
        <el-container class="container">
            <el-aside class="nav" width="6cm">
                <div class="logo-box">
                    <div>
                        <img class="logo" src="../res/logo.svg" /><div class="logo-text">Wing</div>
                    </div>
                </div>
                <el-menu :default-active="index" ref="menu" @select="onSelect" class="nav-menu">
                    <el-menu-item index="overview">
                        <i class="el-icon-tickets"></i>
                        <span>概览</span><span style="font-size: 0.4em"> Overview</span>
                    </el-menu-item>
                    <el-menu-item index="lb">
                        <img class="menu-icon" src="../res/move.svg"/>
                        <span>负载均衡</span><span style="font-size: 0.4em"> Load Balance</span>
                    </el-menu-item>
                    <el-menu-item index="orchestration">
                        <img class="menu-icon" src="../res/layers.svg"/>
                        <span>应用编排</span><span style="font-size: 0.4em"> Orchestration</span>
                    </el-menu-item>
                    <el-menu-item index="cicd">
                        <img class="menu-icon" src="../res/anchor.svg" />
                        <span>持续集成</span><span style="font-size: 0.4em"> CI/CD</span>
                    </el-menu-item>
                </el-menu>
            </el-aside>
            <el-container>
                <el-header class="nav-title" height="1.5cm">
                    <div class="board-title"><span>应用编排</span></div>
                </el-header>
                <el-main>
                    <router-view></router-view>
                </el-main>
            </el-container>
        </el-container>
    </div>
</template>
<script>
import {router} from '../route.js'

export default {
    name: "Dashboard",
    data(){
        return {
            index: ""
        }
    },
    methods: {
        onSelect(index, indexPath) {
            this.index = index
            router.push({name: index})
        }
    },
    beforeRouteEnter(to, from, next){
        next(vm => {
            vm.index = to.matched[to.matched.length - 1].meta.navIndex
        })
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