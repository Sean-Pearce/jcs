import request from '@/utils/request'

export function login(data) {
  const form_data = new FormData()
  for (var key in data) {
    form_data.append(key, data[key])
  }
  return request({
    url: '/user/login',
    method: 'post',
    data: form_data
  })
}

export function getInfo(token) {
  return request({
    url: '/user/info',
    method: 'get',
    params: { token }
  })
}

export function logout() {
  return request({
    url: '/user/logout',
    method: 'post'
  })
}

export function setStrategy(pref) {
  return request({
    url: '/user/strategy',
    method: 'post',
    data: pref
  })
}

export function getStrategy() {
  return request({
    url: '/user/strategy',
    method: 'get'
  })
}

export function changePassword(oldpw, newpw) {
  const form_data = new FormData()
  form_data.append('password', oldpw)
  form_data.append('new_password', newpw)
  return request({
    url: '/user/passwd',
    method: 'post',
    data: form_data
  })
}