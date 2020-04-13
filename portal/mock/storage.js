import Mock from 'mockjs'

const data = Mock.mock({
  'items|10': [{
    filename: '@word',
    size: '@integer(3, 50)' + ' MB',
    last_modified: '@datetime',
    'location|1': [['bj'], ['bj', 'sh'], ['bj', 'sh', 'gz']]
  }]
})

const site = Mock.mock({
  'items': ['bj', 'sh', 'gz']
})

export default [
  {
    url: '/vue-admin-template/file',
    type: 'get',
    response: config => {
      const items = data.items
      return {
        code: 20000,
        data: {
          total: items.length,
          items: items
        }
      }
    }
  },
  {
    url: '/vue-admin-template/site',
    type: 'get',
    response: config => {
      const items = site.items
      return {
        code: 20000,
        data: {
          total: items.length,
          items: items
        }
      }
    }
  }
]
