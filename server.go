package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

func getTimestamp() float64 {
	now := time.Now()
	return float64(now.UnixNano()) / math.Pow10(9)
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
			"time":   strconv.FormatFloat(getTimestamp(), 'f', 7, 64),
		})

		fmt.Fprintf(writer, string(msg))
	}
}

func main() {
	log.Println("Server listening on port 5000")

	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":5000", nil)
}
