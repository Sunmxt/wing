import {router} from '../route.js'

export function refreshUserInfo() {
    console.log('refresh user info.')
    let c = this
    axios.get('/api/login').then(function (resp) {
        if (!resp.data.data.login) {
            router.push({name: 'login'})
            return
        }
        c.$store.state.user.id = resp.data.data.id
        c.$store.state.user.login = resp.data.data.login
        c.$store.state.user.name = resp.data.data.name
    })
}

export function init() {
    console.log('switch to dashboard.')
    let c = this
    axios.post("/api/login", {}).then(function (resp) {
        if (!resp.data.success) {
            router.push({ name: 'login' })
        } else {
            refreshUserInfo.call(c)
        }
    })
}