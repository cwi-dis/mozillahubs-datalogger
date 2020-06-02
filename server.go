package main

import (
	"fmt"
	"net/http"
)

func handleRequest(writer http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		fmt.Fprintf(writer, "Hello World")
	}
}

func main() {
	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":5000", nil)
}
