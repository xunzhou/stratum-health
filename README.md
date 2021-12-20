# stratum-health
Health monitor for multiple Stratum-Based mining pool

## Usage:
### Start server:
```
$ go build .
$ ./stratum-health 
2021/12/19 22:41:18 Listening on:3001
...
```
### Client: 
```
$ curl 127.0.0.1:3001/
OK

$ curl 127.0.0.1:3001/all
us1.ethermine.org:4444
5 packets transmitted, 5 received, 0% packet loss, time 5.21856889s
min/avg/max = 31.053059ms, 31.956277ms, 32.594635ms

us2.ethermine.org:4444
5 packets transmitted, 5 received, 0% packet loss, time 5.221494083s
min/avg/max = 31.6581ms, 32.238997ms, 32.885343ms

us2.ethermine.org:5555
5 packets transmitted, 5 received, 0% packet loss, time 5.529270909s
min/avg/max = 29.850313ms, 72.723995ms, 237.733442ms
```