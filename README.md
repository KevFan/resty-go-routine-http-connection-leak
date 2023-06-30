# resty-go-routine-http-connection-leak
This is an example project showcasing a Go routine leak that occurs when idle connections are kept alive using Resty with custom transport options. 
It highlights the importance of setting the correct transport options, as incorrect configurations can lead to Go routine leaks in both the 
server and dependent services.
## Pre-requisites
* [Go1.20](https://go.dev/dl/)

## Run Application
To showcase this behavior, follow these steps:

* Start server in one terminal:
```shell
go run ./main/main.go
```
* Start another server in a different terminal:
```shell
BIND_ADDRESS=:9090 go run ./main/main.go
```
* Send load to the `/forward` endpoint in server 2, which sends a POST request to server 1:
```shell
for i in {1..400}; do
    curl localhost:9090/forward
    echo ""
done
```
* The responses from server 2 show that the Go routines of server 2 and server 1 increase with each request in about a 2:1 ratio:
```shell
{"goRoutines":791,"Server1Response":{"msg":"Hello World","goRoutines":395}}
{"goRoutines":793,"Server1Response":{"msg":"Hello World","goRoutines":396}}
{"goRoutines":795,"Server1Response":{"msg":"Hello World","goRoutines":397}}
{"goRoutines":797,"Server1Response":{"msg":"Hello World","goRoutines":398}}
{"goRoutines":799,"Server1Response":{"msg":"Hello World","goRoutines":399}}
{"goRoutines":801,"Server1Response":{"msg":"Hello World","goRoutines":400}}
{"goRoutines":803,"Server1Response":{"msg":"Hello World","goRoutines":401}}
```
* Even when sending requests at a later time, there will be no decrease in Go routines since the HTTP connections are kept alive.
  * It's important to note that the ratio mentioned may not be maintained, as the Resty client itself is removed from memory. 
However, the number of Go routines in Server 2 will not go below the number of Go routines in Server 1.
* This can be prevented by either:
  * Setting `DisableKeepAlives: true` in `&http.Transport{}` for the Resty client in the `/forward` endpoint to disable keeping the connection alive in the client.
  * Setting `IdleConnTimeout: time.Second` in `&http.Transport{}` for the Resty client in the `/forward endpoint`. 
    This sets the maximum idle connection time before closing alive connections. If not defined, the default zero means no limit - https://pkg.go.dev/net/http#Transport.
  * Setting `w.Header().Set("Connection", "close")` in the response headers of the `/hello` endpoint. This requests clients to close the connection after receiving the response.