package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path"
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

func createHandlerWithPath(saveDir string) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			log.Println("Processing request")
			body, err := ioutil.ReadAll(req.Body)

			if err != nil {
				log.Println("Could not read request body:", err)

				msg, _ := json.Marshal(&errorResponse{Status: "error"})
				http.Error(writer, string(msg), http.StatusBadRequest)

				return
			}

			saveName := path.Join(saveDir, time.Now().Format("datalog-2006-01-02.json"))

			if err := writeToFile(saveName, string(body)); err != nil {
				log.Println("Could not save data to file:", err)

				msg, _ := json.Marshal(&errorResponse{Status: "error"})
				http.Error(writer, string(msg), http.StatusBadRequest)

				return
			}

			msg, _ := json.Marshal(&successResponse{
				Status: "ok",
				Time:   getTimestamp(),
			})

			fmt.Fprintf(writer, string(msg))
		}
	}
}

func startServer(saveDir string, port int) {
	portSpec := fmt.Sprintf(":%d", port)

	log.Printf("Server listening on port %d", port)

	http.HandleFunc("/mozillahubs", createHandlerWithPath(saveDir))
	http.ListenAndServe(portSpec, nil)
}

func main() {
	port := flag.Int("p", 6000, "Port to listen on")
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println("USAGE:", os.Args[0], "[-p port] saveDir")
		os.Exit(1)
	}

	startServer(flag.Arg(0), *port)
}
