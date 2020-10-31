<template>
  <div class="app-container">
    <el-form ref="form" :model="form" label-width="120px">
      <el-form-item label="原密码">
        <el-input placeholder="请输入原密码" v-model="oldpw" show-password style="width: 250px"></el-input>
      </el-form-item>
      <el-form-item label="新密码">
        <el-input placeholder="请输入新密码" v-model="newpw" show-password style="width: 250px"></el-input>
      </el-form-item>
      <el-form-item label="重复密码">
        <el-input placeholder="请重复新密码" v-model="repw" show-password style="width: 250px"></el-input>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="onSubmit">提交</el-button>
        <el-button @click="onCancel">清空</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script>
import { changePassword } from '@/api/user'

export default {
  data() {
    return {
      oldpw: '',
      newpw: '',
      repw: ''
    }
  },
  methods: {
    onSubmit() {
      if (this.newpw != this.repw) {
          this.$message({
            message: '两次输入的新密码不同',
            type: 'danger'
          })
          return
      }
      var that = this
      changePassword(this.oldpw, this.newpw).then(response => {
        if (response.code === 9200) {
          this.$message({
            message: '修改密码成功',
            type: 'success'
          })
          that.onCancel()
        } else {
          this.$message({
            message: '原密码错误',
            type: 'danger'
          })
        }
      })
    },
    onCancel() {
      this.oldpw = ''
      this.newpw = ''
      this.repw = ''
    }
  }
}
</script>