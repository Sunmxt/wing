import {UpdateLoginLocaleMessage} from '../locale.js'
import {router} from '../route.js'
import { Form } from 'element-ui';

function onSwitched() {
    UpdateLoginLocaleMessage(false)
    axios.post("/api/login", {}).then(function (resp) {
        if (resp.data.success) {
            router.push('dashboard')
        }
    })
}

function verifyForm() {
    switch( this.activeTab ) {
    case "login":
        if( !this.username ) {
            this.$message(this.localetext.username_missing)
            this.$refs.loginUsernameInputbox.focus()
            return false
        } else if( !this.password ) {
            this.$message(this.localetext.password_missing)
            this.$refs.loginPasswordInputbox.focus()
            return false
        }
        break
    case "register":
        if( !this.username ) {
            this.$message(this.localetext.username_missing)
            this.$refs.registerUsernameInputbox.focus()
            return false
        } else if( !this.password ) {
            this.$message(this.localetext.password_missing)
            this.$refs.registerPasswordInputbox.focus()
            return false
        } else if( !this.passwordConfrim ) {
            this.$message(this.localetext.password_confrim_missing)
            this.$refs.registerPasswordConfrimBox.focus()
            return false
        } else if( this.password != this.passwordConfrim ) {
            this.$message(this.localetext.password_confrim_unmatched)
            this.$refs.registerPasswordConfrimBox.focus()
            return false
        }
    }
    return true
}

function login() {
    if(!verifyForm.call(this)) {
        return
    }
    let loginParams = new FormData()
    loginParams.set("username", this.username)
    loginParams.set("password", this.password)
    axios.post('/api/login', loginParams).then((resp) => {
        console.log(resp)
    })
}

function register() {
    if(!verifyForm.call(this)) {
        return
    }
    this.$message('暂时不支持注册')
}

export {onSwitched, login, register}