<template>
    <div v-on:load="onLoginShow">
        <p>login</p>
        <el-input v-model="input" placeholder="请输入内容"></el-input>
    </div>
</template>
<script>
export default {
    name: "Login",
    data(){
        return {
            input: ''
        }
    },
    methods: {
        onLoginShow(){
            console.log('show login')
        }
    },
    props: ['localetext']
}
</script>
<style>
</style>