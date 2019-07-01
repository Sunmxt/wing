<template>
    <div class="new-application">
        <div class="nav-bar">
            <el-breadcrumb class="left v-full" separator-class="el-icon-arrow-right">
                <el-breadcrumb-item>Dashboard</el-breadcrumb-item>
                <el-breadcrumb-item :to="{ name: 'orchestration' }">{{ $t('ui.orchestration.title') }}</el-breadcrumb-item>
                <el-breadcrumb-item>{{ $t('ui.orchestration.create_application') }}</el-breadcrumb-item>
            </el-breadcrumb>
            <el-button class="right" type="primary" @click="onSubmit" size="mini">{{ $t('ui.orchestration.submit') }}</el-button>
        </div>
        <div class="container">
            <el-form ref="form" :model="form" label-width="80px">
                <el-collapse v-model="activePanels">
                    <el-collapse-item name="1">
                        <template slot="title">
                            <span class="collapse-title"><i class="el-icon-s-order"></i><span class="collapse-title-text">{{ $t('ui.orchestration.form.basic_info') }}</span></span>
                        </template>
                        <el-form-item :label="$t('ui.orchestration.form.application_name')" size="mini">
                            <el-input v-model="form.name" :label="$t('ui.orchestration.form.application_name')"></el-input>
                        </el-form-item>
                        <el-form-item :label="$t('ui.orchestration.form.application_description')">
                            <el-input type="textarea" v-model="form.description" :rows="5" :placeholder="$t('ui.orchestration.form.application_description_prompt')"></el-input>
                        </el-form-item>
                    </el-collapse-item>
                    <el-collapse-item name="2">
                        <template slot="title" >
                            <span class="collapse-title"><i class="el-icon-cpu"></i><span class="collapse-title-text">{{ $t('ui.orchestration.form.runtime_info') }}</span></span>
                        </template>
                        <el-form-item :label="$t('ui.orchestration.form.image')" size="mini">
                            <el-row>
                                <el-col :span="11">
                                    <el-input v-model="form.image" :placeholder="$t('ui.orchestration.form.image_prompt')"></el-input>
                                </el-col>
                                <el-col class="center" :span="1"> : </el-col>
                                <el-col :span="3">
                                    <el-input v-model="form.imageVersion" :placeholder="$t('ui.orchestration.form.image_tag_prompt')"></el-input>
                                </el-col>
                            </el-row>
                        </el-form-item>
                        <el-form-item :label="$t('ui.orchestration.form.bootstrap_command')" size="mini">
                            <el-input v-model="form.command" :placeholder="$t('ui.orchestration.form.bootstrap_command_prompt')"></el-input>
                        </el-form-item>
                        <el-form-item :label="$t('ui.orchestration.form.environment_variable')" size="mini">

                            <el-row v-for="(item, index) in form.environmentVariables" :key="index">
                                <el-col :span="6">
                                    <el-input v-model="item.name" :placeholder="$t('ui.orchestration.form.variable_name')"></el-input>
                                </el-col>
                                <el-col class="center" :span="1"> = </el-col>
                                <el-col :span="5">
                                    <el-input v-model="item.value" :placeholder="$t('ui.orchestration.form.variable_value')"></el-input>
                                </el-col>
                                <el-col class="center" :span="1">
                                    <el-button type="text" @click="onDeleteEnvironmentVariable(index)" size="mini"><i class="el-icon-remove-outline"></i></el-button>
                                </el-col>
                            </el-row>
                            <el-row>
                                <el-button type="text" @click="onAddEnvironmentVariable" size="mini">{{ $t('ui.orchestration.form.add_variable') }} <i class="el-icon-circle-plus-outline"></i></el-button>
                            </el-row>
                        </el-form-item>
                    </el-collapse-item>
                    <el-collapse-item name="3">
                        <template slot="title">
                            <span class="collapse-title"><i class="el-icon-takeaway-box"></i><span class="collapse-title-text">资源配置</span></span>
                        </template>
                        <el-row >
                            <!-- <el-slider v-model="form.memory" max="65536" step="32" show-input> </el-slider> -->
                            <el-col :span="6">
                                <el-form-item :label="$t('ui.orchestration.form.cpu_quota')" size="mini">
                                    <el-input type="number" min="0.1" step="0.1" v-model="form.cpu" :placeholder="$t('ui.orchestration.form.cpu_quota_prompt')"></el-input>
                                </el-form-item>
                            </el-col>
                            <el-col :span="6">
                                <el-form-item :label="$t('ui.orchestration.form.memory_quota')" size="mini">
                                    <el-input type="number" min="128" step="128" v-model="form.memory" :placeholder="$t('ui.orchestration.form.memory_quota_prompt')"></el-input>
                                </el-form-item>
                            </el-col>
                        </el-row>
                    </el-collapse-item>
                </el-collapse>
            </el-form>
        </div>
    </div>
</template>
<script>
export default {
    name: "new-application",
    data() {
        return {
            form: {
                name: "",
                image: "",
                imageVersion: "",
                description: "",
                command: "",
                environmentVariablesIndex: 1,
                environmentVariables: {},
                memory: null,
                cpu: null
            },
            activePanels: ["1", "2", "3"]
        }
    },
    methods: {
        onSubmit() {
            console.log('submit')
        },
        onAddEnvironmentVariable() {
            console.log(this)
            let newIndex = this.form.environmentVariablesIndex
            this.form.environmentVariablesIndex ++
            this.$set(this.form.environmentVariables, newIndex, {
                name: "",
                value: ""
            })
        },
        onDeleteEnvironmentVariable(index) {
            this.$delete(this.form.environmentVariables, index)
        }
    }
}
</script>
<style>
</style>
