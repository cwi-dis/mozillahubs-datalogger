package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

type successResponse struct {
	Status string  `json:"status"`
	Time   float64 `json:"time"`
}

type errorResponse struct {
	Status string `json:"status"`
}

func getTimestamp() float64 {
	now := time.Now()
	return float64(now.UnixNano()) / math.Pow10(9)
}

func writeToFile(path string, body string) error {
	outFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0744)

	if err != nil {
		return err
	}

	defer outFile.Close()

	if _, err := outFile.WriteString(body + "\n"); err != nil {
		return err
	}

	return nil
}

func handleRequest(rootDir string) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			log.Println("Processing request")
			body, err := ioutil.ReadAll(req.Body)

			if err != nil {
				log.Println("Could not read request body")

				msg, _ := json.Marshal(&errorResponse{Status: "error"})
				http.Error(writer, string(msg), http.StatusBadRequest)

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
}

func main() {
	log.Println("Server listening on port 5000")

	http.HandleFunc("/", handleRequest("/Users/tom/Desktop"))
	http.ListenAndServe(":5000", nil)
}
