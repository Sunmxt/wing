import Vue from 'vue'
import Vuex from 'vuex'
import VueI18n from 'vue-i18n'
import Wing from './wing.vue'
import VueRouter from 'vue-router'
import ElementUI from 'element-ui'
import {router} from './route.js'
import imessage from './common/i18n.js'
import './css/theme.scss'

Vue.use(VueRouter)
Vue.use(ElementUI)
Vue.use(Vuex)
Vue.use(VueI18n)

new Vue({
    el: "#wing",
    template: "<Wing/>",
    components: {
        Wing
    },
    router,
    data: {
    },
    store: new Vuex.Store({
        state: {
            lang: "cn",
            user: {
                id: "",
                name: "",
                login: false
            }
        }
    }),
    i18n: new VueI18n({
        locale: 'cn',
        messages: imessage
    })
})
