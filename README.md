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
# example outoput
# remote             tcp latency    ip                ip lookup
# httpbin.org:80    255.128834ms    44.195.190.188    ec2-44-195-190-188.compute-1.amazonaws.com.

netdog dial --tls httpbin.org:443
# example outoput
# remote            tcp latency     tls latency     total latency    ip                tls version   cipher                                  ip lookup
# httpbin.org:443    240.778292ms    458.478625ms    699.256917ms    44.195.190.188    TLS1.2        TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256    ec2-44-195-190-188.compute-1.amazonaws.com.

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
