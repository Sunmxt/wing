import VueRouter from 'vue-router'
import Login from './login/login.vue'
import Dashboard from './dashboard/dashboard.vue'
import {onSwitched as onSwitchedLogin} from './login/proc.js'
import {onSwitched as onSwitchedDashboard} from './dashboard/proc.js'
import {UpdateDashboardLocaleMessage, message as localeMessage} from './locale.js'

const routes = [
    { 
        name: "login",
        path: "/login",
        component: Login,
        props: {
            localetext: localeMessage.login
        }
    },
    { 
        name: "dashboard",
        path: "/",
        component: Dashboard,
        props: {
            localetext: localeMessage.login
        }
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