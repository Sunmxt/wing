function updateSettings() {
    let state = this.$store.state
    axios.get('/api/settings').then(resp => {
        state.avaliablePanels = resp.data.data.avaliable_panels
        state.signUpEnabled = resp.data.data.accept_registration
    })
}

export function init() {
    updateSettings.call(this)
}