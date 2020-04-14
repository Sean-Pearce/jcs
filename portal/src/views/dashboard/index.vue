<template>
  <div class="app-container">
    <el-upload
      action="foobar"
      :http-request="handleUpload"
    >
      <el-button type="primary"><svg-icon icon-class="upload" /> 上传文件</el-button>
    </el-upload>
    <el-table
      v-loading="listLoading"
      :data="files.filter(data => !search || data.filename.toLowerCase().includes(search.toLowerCase()))"
      fit
    >
      <el-table-column
        label="文件名"
        prop="filename"
      />
      <el-table-column
        label="大小"
        prop="size"
      />
      <el-table-column
        label="修改时间"
        prop="last_modified"
      />
      <el-table-column class-name="status-col" label="位置">
        <template slot-scope="scope">
          <el-tag v-for="loc in scope.row.location" :key="loc" type="info">{{ loc }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column
        align="right"
      >
        <template slot="header" slot-scope="scope">
          <el-input
            v-model="search"
            size="mini"
            placeholder="搜索文件"
          />
        </template>
        <template slot-scope="scope">
          <el-button
            size="mini"
            type="primary"
            @click="handleDownload(scope.row.filename)"
          >下载</el-button>
          <el-button
            size="mini"
            type="danger"
            @click="handleDelete(scope.row.filename)"
          >删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script>
import { getFiles, upload, download } from '@/api/storage'

export default {
  data() {
    return {
      files: [],
      search: ''
    }
  },
  created() {
    this.fetchData()
  },
  methods: {
    fetchData() {
      this.listLoading = true
      getFiles().then(response => {
        this.files = response.data.items
        this.listLoading = false
      })
    },
    handleUpload(req) {
      upload(req)
    },
    handleDownload(filename) {
      download(filename)
    },
    handleDelete(filename) {
      console.log(filename)
    }
  }
}
</script>

<style>
.el-tag+.el-tag{
  margin-left: 5px;
}
</style>
