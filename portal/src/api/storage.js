import request from '@/utils/request'
import { getToken } from '@/utils/auth'

export function getFiles(params) {
  return request({
    url: '/storage/list',
    method: 'get',
    params
  })
}

export function upload(item) {
  var form_data = new FormData()
  form_data.append('file', item.file)
  return request({
    url: '/storage/upload',
    method: 'post',
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    data: form_data,
    onUploadProgress: progressEvent => {
      const complete = (progressEvent.loaded / progressEvent.total * 100 | 0)
      item.onProgress({ percent: complete })
    }
  })
}

export function download(params) {
  return request({
    url: '/storage/download',
    method: 'get',
    params: {
      filename: params
    },
    responseType: 'blob'
  })
}

export function genDownloadLink(filename) {
  // TODO: insecure, should use temporary token
  return process.env.VUE_APP_BASE_API + '/storage/download?filename=' + filename + '&t=' + getToken()
}
