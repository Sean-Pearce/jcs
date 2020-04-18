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
import { setStrategy, getStrategy } from '@/api/user'

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
      getStrategy().then(response => {
        this.sites = response.data.sites
        this.form.sites = response.data.strategy.sites
      })
    },
    onSubmit() {
      setStrategy(this.form).then(response => {
        if (response.code === 9200) {
          this.$message({
            message: 'submitted',
            type: 'success'
          })
        } else {
          this.$message({
            message: 'submitted',
            type: 'danger'
          })
        }
      })
    },
    onCancel() {
      this.fetchData()
    }
  }
}
</script>

<style scoped>
.line{
  text-align: center;
}
</style>

