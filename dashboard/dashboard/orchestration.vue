<template>
    <div class="orchestration">
        <div class="nav-bar">
            <el-breadcrumb class="left v-full" separator-class="el-icon-arrow-right">
                <el-breadcrumb-item>Dashboard</el-breadcrumb-item>
                <el-breadcrumb-item :to="{ name: 'orchestration' }">{{ $t('ui.orchestration.title') }}</el-breadcrumb-item>
            </el-breadcrumb>
        </div>
        <el-tabs v-model="active">
            <el-tab-pane :label="$t('ui.orchestration.my_application')" name="mine"></el-tab-pane>
            <el-tab-pane :label="$t('ui.orchestration.all_application')" name="all"></el-tab-pane>
        </el-tabs>
        <el-button class="right" type="primary" @click="onCreateApplication" size="mini">{{ $t('ui.orchestration.create_application') }}</el-button>
        <el-table :data="applicationList" style="width: 100%">
            <el-table-column prop="name" :label="$t('ui.orchestration.application_name')">
            </el-table-column>
            <el-table-column prop="state" :label="$t('ui.orchestration.application_state')" width="100">
                <template slot-scope="scope">
                    <el-tag type="info" size="small">{{ scope.row.state }}</el-tag>
                </template>
            </el-table-column>
            <el-table-column prop="version" :label="$t('ui.orchestration.version_number')" width="100"></el-table-column>
            <el-table-column prop="owner" :label="$t('ui.orchestration.owner')">
            </el-table-column>
            <el-table-column prop="resource" :label="$t('ui.orchestration.resource_usage')">
            </el-table-column>
            <el-table-column prop="operation" :label="$t('ui.orchestration.operation')" fixed="right" width="150">
                <template slot-scope="scope">
                    <el-dropdown>
                        <span class="el-dropdown-link">
                            {{ $t('ui.orchestration.operation') }}<i class="el-icon-arrow-down el-icon--right"></i>
                        </span>
                        <el-dropdown-menu slot="dropdown">
                            <el-dropdown-item><i class="el-icon-check" /><span>{{ $t('ui.orchestration.operations.deploy') }}</span></el-dropdown-item>
                            <el-dropdown-item><i class="el-icon-refresh-left" /><span>{{ $t('ui.orchestration.operations.rollback') }}</span></el-dropdown-item>
                            <el-dropdown-item><i class="el-icon-switch-button" /><span>{{ $t('ui.orchestration.operations.shutdown') }}</span></el-dropdown-item>
                            <el-dropdown-item><i class="el-icon-upload2" /><span>{{ $t('ui.orchestration.operations.upgrade') }}</span></el-dropdown-item>
                            <el-dropdown-item><i class="el-icon-remove-outline" /><span>{{ $t('ui.orchestration.operations.remove') }}</span></el-dropdown-item>
                        </el-dropdown-menu>
                    </el-dropdown>
                </template>
            </el-table-column>
        </el-table>
    </div>
</template>
<script>
export default {
    name: "orchestration-panel",
    data() {
        return {
            active: "mine",
            applicationList: [{
                name: "uestc-life-be",
                owner: "hzy",
                state: "待部署",
                resource: "资源",
                operation: "op",
                version: 1,
          }]
        }
    },
    methods: {
        onCreateApplication() {
            this.$router.push({name: "create-application"})
        }
    }
}
</script>
<style>
</style>