# netdog

an immitation of netcat for fun and practice


## usage

```bash
go install  github.com/JackKCWong/netdog@latest
netdog --help

# examples
# send raw tcp payload to target
printf "GET /get HTTP/1.0\r\n\r\n" | netdog httpbin.org:80
printf "GET /get HTTP/1.0\r\n\r\n" | netdog --tls httpbin.org:443

# send raw websocket frames
netdog --tls ws.postman-echo.com:443 ws.http hello.bin close.bin

# test tcp / tls connection to target
netdog dial httpbin.org:80
netdog dial --tls httpbin.org:443
printf "httpbin.org:80\nhttpbin.org:443" | netdog dial # this only test for tcp connection, not tls

# lookup DNS and time the responses
netdog lookup httpbin.org
netdog lookup --name httpbin.org
printf "httpbin.org\nbaidu.com" | netdog lookup 
printf "tcp 127.0.0.1:8080\ntcp 127.0.0.2:80\n" | netdog lookup --grep
```


## TODOs

[x] support unix socket

[x] host:port scanning

[x] DNS lookup

## notes

* turning a hex string into binary: `xxd -r -p hex.txt out.bin`
