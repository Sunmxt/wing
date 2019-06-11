import {UpdateDashboardLocaleMessage} from '../locale.js'
import {router} from '../route.js'

export function onSwitched() {
    UpdateDashboardLocaleMessage(false)
    console.log('switch to dashboard.')
    axios.post("/api/login", {}).then(function (resp) {
        if (!resp.data.success) {
            router.push({
                name: 'login'
            })
        }
    })
}