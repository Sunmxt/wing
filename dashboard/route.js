import VueRouter from 'vue-router'
import Login from './login/login.vue'
import Dashboard from './dashboard/dashboard.vue'
import {onSwitched as onSwitchedLogin} from './login/proc.js'
import {onSwitched as onSwitchedDashboard} from './dashboard/proc.js'

const routes = [
    { 
        name: "login",
        path: "/login",
        component: Login,
    },
    { 
        name: "dashboard",
        path: "/",
        component: Dashboard,
    }
]

let router = new VueRouter({
    mode: 'history',
    routes
})

router.afterEach((to, from) => {
    switch(to.name) {
    case "dashboard":
        onSwitchedDashboard()
        break
    case "login":
        onSwitchedLogin()
        break
    }
})

export {router}