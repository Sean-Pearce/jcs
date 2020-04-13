import request from '@/utils/request'

export function getFiles(params) {
  return request({
    url: '/vue-admin-template/file',
    method: 'get',
    params
  })
}

export function upload(params) {
  return request({
    url: '/vue-admin-template/upload',
    method: 'post',
    params
  })
}

export function download(params) {
  return request({
    url: '/vue-admin-template/download',
    method: 'get',
    params
  })
}
