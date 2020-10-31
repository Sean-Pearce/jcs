#!/usr/bin/env python3

import requests

def login(jcs_url, user, password):
    if jcs_url.endswith('/'):
        jcs_url = jcs_url[:-1]
    print(f'signing in to {jcs_url}')
    header = {
        'Origin': jcs_url,
        'Referer': jcs_url + '/',
    }
    login_url = jcs_url+'/api/user/login'
    data = {
        'username': user,
        'password': password,
    }
    resp = requests.post(login_url, data=data, headers=header)
    print(f'Response: {resp.status_code}, {resp.text}')
    return resp.status_code

if __name__ == '__main__':
    try:
        code = login('http://localhost:8080/', 'admin', 'admin')
    except:
        print('failed to connect to server')
        code = -1
    finally:
        if code != 200:
            exit(-1)

