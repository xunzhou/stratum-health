# stratum-health
Health monitor for multiple Stratum-Based mining pool

## Usage:
### Start server:
```
$ go build .
$ ./stratum-health 
2021/12/19 22:41:18 Listening on:3001

# Docker
$ docker run -it -v $(pwd)/stratum-health.yaml:/app/stratum-health.yaml -p80:3001 stratum-health
...
```
### Client: 
```
$ curl 127.0.0.1:3001
Unauthorized
$ curl user:password@127.0.0.1:3001
OK

$ curl -s user:password@127.0.0.1:3001/all | jq 
[
  {
    "Host": "us2.ethermine.org:4444",
    "Trans": 5,
    "Recev": 5,
    "Loss": 0,
    "Time": 5218481655,
    "Min": "30.728607ms",
    "Avg": "31.58873ms",
    "Max": "32.828504ms"
  },
  {
    "Host": "us1.ethermine.org:4444",
    "Trans": 5,
    "Recev": 5,
    "Loss": 0,
    "Time": 5229126465,
    "Min": "31.520419ms",
    "Avg": "33.345364ms",
    "Max": "39.09858ms"
  },
  {
    "Host": "us2.ethermine.org:5555",
    "Trans": 5,
    "Recev": 5,
    "Loss": 0,
    "Time": 5308782304,
    "Min": "29.364105ms",
    "Avg": "31.946676ms",
    "Max": "39.190233ms"
  }
]
```