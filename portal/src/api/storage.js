import request from '@/utils/request'

export function getFiles(params) {
  return request({
    url: '/storage/list',
    method: 'get',
    params
  })
}

export function upload(item) {
  var form_data = new FormData()
  console.log(item)
  form_data.append('file', item.file)
  return request({
    url: '/storage/upload',
    method: 'post',
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    data: form_data
  })
}

export function download(params) {
  return request({
    url: '/storage/download',
    method: 'get',
    params: {
      filename: params
    }
  })
}

export function getSites(params) {
  return request({
    url: '/user/site',
    method: 'get',
    params
  })
}
