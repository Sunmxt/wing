import Vue from 'vue'
import Wing from './wing.vue'
import VueRouter from 'vue-router'
import ElementUI from 'element-ui'

import {router} from './route.js'
import {message as localeMessage} from './locale.js'
import './css/theme.scss'


Vue.use(VueRouter)
Vue.use(ElementUI)

new Vue({
    el: "#wing",
    template: "<Wing/>",
    components: {
        Wing
    },
    router,
    data: {
        messages: localeMessage
    }
})
