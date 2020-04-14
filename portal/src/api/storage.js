import request from '@/utils/request'

export function getFiles(params) {
  return request({
    url: '/storage/list',
    method: 'get',
    params
  })
}

export function upload(params) {
  return request({
    url: '/storage/upload',
    method: 'post',
    params
  })
}

export function download(params) {
  return request({
    url: '/storage/download',
    method: 'get',
    params
  })
}

export function getSites(params) {
  return request({
    url: '/user/site',
    method: 'get',
    params
  })
}
