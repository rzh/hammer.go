hammer.go
=========

Hammer.go, a performance testing tool in Go. Rewrote from earlier version. 


## Files:
hammer.go - the client hammer tool
server.go - a mockup server in Go, to provide HTML and JSON/REST response
server.py - a mockup server in Python, mainly used to verify authentication works with a different language
src/oauth/oauth.go - a modified version of Gary Burd's Go OAuth lib, added body hash for OAuth

## Usage:
1. If you test stateless API, simply define API and its distribution in a JSON file. 
```
go run hammer.go -rps 100 -profile profile/test_scenario_profile.json
```

2. For test with awareness of session, you have to create as a Go module. You can see example in the repo. A way to define this with JSON and Lua will come. 
```
go run hammer.go -rps 100 -type ws_session -size 100
```

> try `go run hammer.go --help` to get all cmd parameters


## To get binary for Hammer:
**You have to properly compile/update Go for Linux first**

1. `brew install go --HEAD --cross-compile-common` will resolve all
2. `GOOS=linux GOARCH=amd64 CGO_ENABLE=0 go build -o hammer.linux hammer.go`


## Make it yourself:
* none session scenario
 1. add or modify existed .json files in /profile, following current format

* session scenario
 1. create your .go file in /scenario
 2. follow any *_scenario.go file as example


## Sample usage

This is to test with the bundled server.go, we use profiles/profile.json as input. After start mockup server, run this command
```
go run hammer.go -rps 100 -profile profile/profile.json
```

the output will be
```
cpu number ->  2
rps ->  5
slow threshold ->  200 ms
profile type ->  default
Proxy ->  none
Import Call -> W: 20.000000 URL: http://127.0.0.1:9000/json  Method: GET
Import Call -> W: 30.000000 URL: http://127.0.0.1:9000/html  Method: GET
Import Call -> W: 40.000000 URL: http://127.0.0.1:9000/json  Method: GET
total:     5  req/s:     5  res/s:     4  avg:  0.0006  pending:  1  err: 0 | 0.00%  slow:  0.00%
total:    10  req/s:     5  res/s:     5  avg:  0.0005  pending:  1  err: 0 | 0.00%  slow:  0.00%
total:    15  req/s:     5  res/s:     5  avg:  0.0008  pending:  1  err: 0 | 0.00%  slow:  0.00%
total:    20  req/s:     5  res/s:     5  avg:  0.0008  pending:  1  err: 0 | 0.00%  slow:  0.00%
```

the stats will be print every 1 sec. It shows important stats about the test, for example
```
total:    20  req/s:     5  res/s:     5  avg:  0.0008  pending:  1  err: 0 | 0.00%  slow:  0.00%
```

means, the test send 20 request total, and we are seeing 5 request per second, and 5 response per second, which means the server is in good condition. The average response time is 0.0008 second, and there is 1 request in backlog.  Total error is 0, thus the error rate is also 0.00%. Hammer also measures slow response, which is defined by any reponse coming back longer than 200ms. The stat line print out slow response ratio at the last percentage number. 

## About JSON file format
```
{"Weight": 20, "URL":"http://127.0.0.1:9000/json", "Method":"GET", "Body":"", "Type":"REST"}
{"Weight": 30, "URL":"http://127.0.0.1:9000/html", "Method":"GET", "Body":"", "Type":"WWW"}
{"Weight": 40, "URL":"http://127.0.0.1:9000/json", "Method":"GET", "Body":"", "Type":"REST"}
```

Each line is an API, and you shall provide URL, HTTP Method, and BODY for the request. Type defines whether this is a regular HTTP call or RESTful API call. Weight is used to control distribution of the request. For example, in the above profile, API in line one comes with weight of 20, thus, there will be around 20/(20_30_40) = 22.2% requst is send to this URL. 

## Few todos, 
* Add pattern replacement for JSON profile back, which enable randomlize URL or BODY defined by JSON profile.
* Lua support for modify traffic profile, as well as inspect response need be added back
* WWW report interface will be added with graphics report
* maybe others.. :)
