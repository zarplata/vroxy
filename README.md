vroxy
=====

Proxy server for balancing requests to VK API to avoid rate limit exceeded.

Requests grouping to chunks and sends every second using `Execute`
method without the habit of rate limitations.

## Requests

Requests are the same as the original requests except for the hostname.

## Responses

The proxy response is different from the original one. Requests are queued and send 
asynchronously to VK, and therefore it is not possible to provide original responses back.

### OK

```
$ curl -iX POST http://localhost:8080/method/messages.send\?access_token\=[...] \
  -F user_id=2554441 \
  -F message=test

HTTP/1.1 100 Continue

HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Sat, 17 Mar 2018 18:07:28 GMT
Content-Length: 17

{"success":true}
```

### Bad request

```
curl -iX POST \
  http://localhost:8080/method/messages.send\?access_token\= \
  -F user_id=2554441 \
  -F message=test

HTTP/1.1 400 Bad Request
Content-Type: application/json; charset=utf-8
Date: Sat, 17 Mar 2018 18:21:34 GMT
Content-Length: 61
Connection: close

{"error":"query `access_token` is required","success":false}
```
