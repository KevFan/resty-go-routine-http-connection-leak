package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/go-resty/resty/v2"
)

type Server1Response struct {
	Msg        string `json:"msg"`
	GoRoutines int    `json:"goRoutines"`
}

type Server2Response struct {
	GoRoutines      int             `json:"goRoutines"`
	Server1Response Server1Response `json:"server1Response"`
}

func main() {
	http.HandleFunc("/hello", getTestHandler)
	http.HandleFunc("/forward", getForwardHandler)

	bindAddress := os.Getenv("BIND_ADDRESS")
	if bindAddress == "" {
		bindAddress = ":8080"
	}

	log.Fatal(http.ListenAndServe(bindAddress, nil))
}

func getTestHandler(w http.ResponseWriter, r *http.Request) {
	resp := Server1Response{
		Msg:        "Hello World",
		GoRoutines: runtime.NumGoroutine(),
	}

	// Convert response object to JSON
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the response headers
	w.Header().Set("Content-Type", "application/json")
	// Request clients to close the connection after receiving a response
	//w.Header().Set("Connection", "close")
	w.WriteHeader(http.StatusOK)

	// Write the JSON response
	fmt.Fprintf(w, string(jsonData))

	// Log
	log.Println("response", string(jsonData))
}

func getForwardHandler(w http.ResponseWriter, r *http.Request) {
	// Create a new Resty client
	client := resty.New()

	// Set transport options
	// The default client transport has default fields set which prevents this leak - https://github.com/go-resty/resty/blob/master/transport.go#L17-L35
	transport := &http.Transport{
		//DisableKeepAlives: true, // This will disable keeping the connection alive in the client - this may not be ideal
		//IdleConnTimeout: time.Second, // Sets the max idle connection time before closing alive connections
	}
	client.SetTransport(transport)

	// Send a request to the /hello endpoint in server 1
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{}`).
		Post("http://localhost:8080/hello")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Unmarshal the response
	resp1 := Server1Response{}
	if err := json.Unmarshal(resp.Body(), &resp1); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Wrap the response with the number of goroutines in server 2
	resp2 := Server2Response{
		Server1Response: resp1,
		GoRoutines:      runtime.NumGoroutine(),
	}

	// Convert the response object to JSON
	jsonData, err := json.Marshal(resp2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Set the response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Write the JSON response
	fmt.Fprintf(w, string(jsonData))

	// Log
	log.Println("response", string(jsonData))
}
