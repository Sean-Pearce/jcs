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

export function setPreference(pref) {
  return request({
    url: '/user/preference',
    method: 'post',
    data: {
      'preference': pref
    }
  })
}
