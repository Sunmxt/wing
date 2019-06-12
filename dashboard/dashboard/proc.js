import {router} from '../route.js'

export function onSwitched() {
    console.log('switch to dashboard.')
    axios.post("/api/login", {}).then(function (resp) {
        if (!resp.data.success) {
            router.push({
                name: 'login'
            })
        }
    })
}