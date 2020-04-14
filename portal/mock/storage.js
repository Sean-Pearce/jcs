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
  'items': ['bj', 'sh', 'gz'],
  'selected': ['bj', 'sh']
})

export default [
  {
    url: '/storage/list',
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
    url: '/user/site',
    type: 'get',
    response: config => {
      const items = site.items
      const selected = site.selected
      return {
        code: 20000,
        data: {
          total: items.length,
          items: items,
          selected: selected
        }
      }
    }
  }
]
