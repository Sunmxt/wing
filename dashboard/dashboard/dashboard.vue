<template>
    <div class="container wing">
        <el-container class="container">
            <el-aside class="nav" width="6cm">
                <div class="logo-box" @click="onLogoClicked">
                    <div>
                        <img class="logo" src="../res/logo.svg" /><div class="logo-text">Wing</div>
                    </div>
                </div>
                <el-menu :default-active="index" ref="menu" @select="onSelect" class="nav-menu">
                    <el-menu-item v-for="(settings, idx) in panels" :index="settings.name" :key="idx">
                        <i :class="settings.icon.ref" v-if="settings.icon.type == 'elem'" />
                        <img class="menu-icon" :src="settings.icon.ref" v-if="settings.icon.type == 'img'" />
                        <span>{{ $t(settings.title.main) }}</span><span style="font-size: 0.4em"> {{ $t(settings.title.sub) }}</span>
                    </el-menu-item>
                </el-menu>
            </el-aside>
            <el-container>
                <el-header class="nav-title" height="1.5cm">
                    <div class="board-title header-box left"><span>应用编排</span></div>
                    <div class="user-tag header-box right"><span>{{ userName }}<i class="el-icon-arrow-down el-icon--right"/></span></div>
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
import {init} from './proc.js'

const rawPanelSettings = {
    overview: {
        title: {
            main: 'ui.dashboard.nav.overview',
            sub: 'ui.dashboard.navSubtitle.overview'
        },
        icon: {
            type: 'elem',
            ref: 'el-icon-tickets'
        }
    },
    lb: {
        title: {
            main: 'ui.dashboard.nav.lb',
            sub: 'ui.dashboard.navSubtitle.lb'
        },
        icon: {
            type: 'img',
            ref: 'svg-move'
        }
    },
    orchestration: {
        title: {
            main: 'ui.dashboard.nav.orchestration',
            sub: 'ui.dashboard.navSubtitle.orchestration'
        },
        icon: {
            type: 'img',
            ref: 'svg-layers'
        }
    },
    cicd: {
        title: {
            main: 'ui.dashboard.nav.cicd',
            sub: 'ui.dashboard.navSubtitle.cicd'
        },
        icon: {
            type: 'img',
            ref: 'svg-anchor'
        }
    }
}

export default {
    name: "Dashboard",
    data(){
        return {
            index: "",
        }
    },
    methods: {
        onSelect(index, indexPath) {
            this.index = index
            router.push({name: index})
        },
        onLogoClicked() {
            router.push({name: "overview"})
        }
    },
    beforeRouteEnter(to, from, next){
        next(vm => {
            vm.index = to.matched[to.matched.length - 1].meta.navIndex
            init.call(vm)
        })
    },
    props: [],
    computed: {
        panels() {
            let panels = []
            let avaliables = this.$store.state.avaliablePanels
            for( let idx in avaliables ) {
                let name = avaliables[idx]
                if( name in this.panelSettings) {
                    panels.push({
                        name,
                        ...this.panelSettings[name]
                    })
                }
            }
            console.log(panels)
            return panels
        },
        imgs() {
            return {
                'svg-move': require('../res/move.svg'),
                'svg-layers': require('../res/layers.svg'),
                'svg-anchor': require('../res/anchor.svg'),
            }
        },
        panelSettings() {
            let settings = {}
            for(let key in rawPanelSettings) {
                settings[key] = Object.assign({}, rawPanelSettings[key])
                if(settings[key].icon.type == 'img'){
                    settings[key].icon.ref = this.imgs[settings[key].icon.ref]
                }
            }
            return settings
        },
        userName() {
            return this.$store.state.user.id
        }
    }
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