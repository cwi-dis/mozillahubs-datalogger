package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"time"
)

type successResponse struct {
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

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
			http.Error(writer, "{ \"status\": \"error\" }", http.StatusBadRequest)
			return
		}

		// TODO parse body and save to file

		msg, _ := json.Marshal(&successResponse{
			Status: "ok",
			Time:   getTimestamp(),
		})

		fmt.Fprintf(writer, string(msg))
	}
}

func main() {
	log.Println("Server listening on port 5000")

	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":5000", nil)
}
