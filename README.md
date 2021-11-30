# netdog

an immitation of netcat for fun and practice


## usage

```bash
go get -u github.com/JackKCWong/go-netdog
go-netdog --help

# examples
# send raw tcp payload to target
printf "GET /get HTTP/1.0\r\n\r\n" | go-netdog httpbin.org:80
printf "GET /get HTTP/1.0\r\n\r\n" | go-netdog --tls httpbin.org:443

# test tcp / tls connection to target
go-netdog dial httpbin.org:80
go-netdog dial --tls httpbin.org:443
printf "httpbin.org:80\nhttpbin.org:443" | go-netdog dial # this only test for tcp connection, not tls

# lookup DNS and time the responses
go-netdog lookup httpbin.org
go-netdog lookup --name httpbin.org
printf "httpbin.org\nbaidu.com" | go-netdog lookup 
```


## TODOs

[x] support unix socket

[x] host:port scanning

[x] DNS lookup
