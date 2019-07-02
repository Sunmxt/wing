import VueRouter from 'vue-router'
import Login from './login/login.vue'
import Dashboard from './dashboard/dashboard.vue'
import OverviewPanel from './dashboard/overview.vue'
import OrchestrationPanel from './dashboard/orchestration.vue'
import NewApplicationPanel from './dashboard/new_application.vue'
import {init as initLogin} from './login/proc.js'

const routes = [
    { 
        name: "login",
        path: "/login",
        component: Login,
        beforeEnter: (to, from, next) => {
            initLogin()
            next()
        }
    },
    { 
        name: "dashboard",
        path: "/",
        component: Dashboard,
        children: [
            {
                name: "overview",
                path: "overview",
                component: OverviewPanel,
                meta: {
                    navIndex: "overview"
                }
            },
            {
                name: "orchestration",
                path: "orchestration",
                component: OrchestrationPanel,
                meta: {
                    navIndex: "orchestration"
                }
            },
            {
                name: "create-application",
                path: "application/create",
                component: NewApplicationPanel,
                meta: {
                    navIndex: "orchestration"
                }
            }
        ],
        redirect: {
            name: "overview"
        }
    }
]

let router = new VueRouter({
    mode: 'history',
    routes
})

export {router}