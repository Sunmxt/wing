import Vue from 'vue'
import Wing from './wing.vue'
import Login from './login.vue'
import Dashboard from './dashboard.vue'
import VueRouter from 'vue-router'
import ElementUI from 'element-ui'
import './css/theme.scss'

const routes = [
    { name: "login", path: "/login", component: Login },
    { name: "dashboard", path: "/", component: Dashboard }
]

const router = new VueRouter({
    mode: 'history',
    routes
})

router.afterEach((from, to) => {
    console.log({
        from,
        to
    })
})

Vue.use(VueRouter)
Vue.use(ElementUI)

new Vue({
    el: "#wing",
    template: "<Wing/>",
    components: {
        Wing
    },
    router
})
