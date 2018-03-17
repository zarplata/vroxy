vroxy
=====

Proxy server for balancing a requests to VK API to avoid rate limit exceeded.

Requests grouping to chunks and sends every seconds using Execute
method without the habit of rate limitations.

## Requests

Requests are the same as the original requests except for the hostname.

## Responses

Because the queue is used to deliver a requests, the proxy responses are
very different from the original responses.

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
