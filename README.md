# netdog

an immitation of netcat for fun and practice


## usage

```bash
go get -u github.com/JackKCWong/go-netdog
netdog --help

# examples
printf "GET /get HTTP/1.0\r\n\r\n" | go-netdog httpbin.org:80
printf "httpbin.org:80\nhttpbin.org:443" | go-netdog dial 
printf "httpbin.org\nbaidu.com" | go-netdog lookup 
```


## TODOs

[x] support unix socket

[x] host:port scanning
