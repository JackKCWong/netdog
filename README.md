# netdog

an immitation of netcat for fun and practice


## usage

```bash
go get -u github.com/JackKCWong/go-netdog
netdog --help

# examples
printf "GET /get HTTP/1.0\r\n\r\n" | go-netdog httpbin.org 80
```


## TODOs

[x] support unix socket

[ ] hosts scanning
