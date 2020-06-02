package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func getTimestamp() int64 {
	now := time.Now()
	return now.UnixNano()
}

func handleRequest(writer http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		log.Println("Processing request")
		_, err := ioutil.ReadAll(req.Body)

		if err != nil {
			log.Println("Could not read request body")

			msg, _ := json.Marshal(map[string]string{
				"status": "error",
			})

			http.Error(writer, string(msg), http.StatusBadRequest)
			return
		}

		// TODO parse body and save to file

		msg, _ := json.Marshal(map[string]string{
			"status": "OK",
			"type":   "HTTPD",
			"time":   string(getTimestamp()),
		})

		fmt.Fprintf(writer, string(msg))
	}
}

func main() {
	log.Println("Server listening on port 5000")

	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":5000", nil)
}
