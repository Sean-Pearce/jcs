<template>
  <div class="app-container">
    <el-form ref="form" :model="form" label-width="120px">
      <el-form-item label="存储后端">
        <el-checkbox-group v-model="form.sites">
          <el-checkbox v-for="site in sites" :key="site" :label="site">{{ site }}</el-checkbox>
        </el-checkbox-group>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="onSubmit">Create</el-button>
        <el-button @click="onCancel">Cancel</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>
import { getSites } from '@/api/storage'

export default {
  data() {
    return {
      form: {
        sites: []
      },
      sites: []
    }
  },
  created() {
    this.fetchData()
  },
  methods: {
    fetchData() {
      getSites().then(response => {
        this.sites = response.data.items
      })
    },
    onSubmit() {
      this.$message('submitted')
    },
    onCancel() {
      this.$message({
        message: 'cancelled',
        type: 'warning'
      })
    }
  }
}
</script>

<style scoped>
.line{
  text-align: center;
}
</style>

