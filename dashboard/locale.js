export let message = {
    login: {
        login: "",
        register: "",
        account: "",
        password: "",
        account_subname: "",
        password_subname: "",
        username_prompt: "",
        password_prompt: "",
        password_confrim_prompt: "",
        password_confrim: "",
        password_confrim_subname: "",
        username_missing: "",
        password_missing: "",
        password_confrim_missing: "",
        password_confrim_unmatched: "",
    },
    dashboard: {
    },
    state: {
        loginLoaded: false,
        dashboardLoaded: false
    }
}

export let UpdateLoginLocaleMessage = function (force) {
    return new Promise(function(resolve, reject) {
        axios.get('/api/locale/login/list').then(function(response){
            if(!force && message.state.loginLoaded) {
                resolve()
                return
            }
            let data = response.data.data
            message.login.login = data.login
            message.login.register = data.register
            message.login.account = data.account
            message.login.password = data.password
            message.login.account_subname = data.account_subname
            message.login.password_subname = data.password_subname
            message.login.username_prompt = data.username_prompt
            message.login.password_prompt = data.password_prompt
            message.login.password_confrim_prompt = data.password_confrim_prompt
            message.login.password_confrim = data.password_confrim
            message.login.password_confrim_subname = data.password_confrim_subname
            message.login.username_missing = data.username_missing
            message.login.password_missing = data.password_missing
            message.login.password_confrim_missing = data.password_confrim_missing
            message.login.password_confrim_unmatched = data.password_confrim_unmatched
            resolve()
        }).catch(function(error){
            reject(error)
        })
    })
}

export let UpdateDashboardLocaleMessage = function (force) {
}
