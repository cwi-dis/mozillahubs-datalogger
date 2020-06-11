package main

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type inputData struct {
	Info []interface{}   `json:"info"`
	Data [][]interface{} `json:"data"`
}

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

func writeToFile(path string, body *inputData) error {
	start := time.Now().UnixNano()
	outFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0744)

	if err != nil {
		return err
	}

	defer outFile.Close()
	zipWriter, err := gzip.NewWriterLevel(outFile, gzip.BestCompression)

	if err != nil {
		return err
	}

	defer zipWriter.Close()
	infoPart := ""

	for i := 0; i < len(body.Info); i++ {
		if body.Info[i] == nil {
			infoPart += fmt.Sprintf(",")
		} else {
			infoPart += fmt.Sprintf("%v,", body.Info[i])
		}
	}

	for i := 0; i < len(body.Data); i++ {
		dataPart := ""

		for j := 0; j < len(body.Data[i]); j++ {
			if body.Data[i][j] == nil {
				dataPart += fmt.Sprintf(",")
			} else {
				dataPart += fmt.Sprintf("%v,", body.Data[i][j])
			}
		}

		zipWriter.Write([]byte(infoPart + dataPart[:len(dataPart)-1] + "\n"))
	}

	log.Printf("writeToFile %d", time.Now().UnixNano()-start)
	return nil
}

func parseRequestBody(req *http.Request) (*inputData, error) {
	start := time.Now().UnixNano()
	jsonDecoder := json.NewDecoder(req.Body)

	var bodyData *inputData = &inputData{}

	if err := jsonDecoder.Decode(bodyData); err != nil {
		return nil, err
	}

	if len(bodyData.Info) == 0 || len(bodyData.Data) == 0 {
		return nil, errors.New("Required fields are missing or invalid")
	}

	log.Printf("parseRequestBody %d", time.Now().UnixNano()-start)
	return bodyData, nil
}

func createHandlerWithPath(saveDir string) func(http.ResponseWriter, *http.Request) {
	var mutex = &sync.Mutex{}

	return func(writer http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			log.Println("Processing request")
			start := time.Now().UnixNano()

			bodyData, err := parseRequestBody(req)

			if err != nil {
				log.Println("Could not decode body data:", err)

				msg, _ := json.Marshal(&errorResponse{Status: "error"})
				http.Error(writer, string(msg), http.StatusBadRequest)

				return
			}

			saveName := path.Join(saveDir, time.Now().Format("datalog-2006-01-02.csv.gz"))
			mutex.Lock()

			if err := writeToFile(saveName, bodyData); err != nil {
				mutex.Unlock()
				log.Println("Could not save data to file:", err)

				msg, _ := json.Marshal(&errorResponse{Status: "error"})
				http.Error(writer, string(msg), http.StatusBadRequest)

				return
			}

			mutex.Unlock()

			msg, _ := json.Marshal(&successResponse{
				Status: "ok",
				Time:   getTimestamp(),
			})

			log.Printf("createHandlerWithPath %d", time.Now().UnixNano()-start)
			fmt.Fprintf(writer, string(msg))
		} else {
			http.NotFound(writer, req)
		}
	}
}

func startServer(saveDir string, port int) {
	portSpec := fmt.Sprintf(":%d", port)

	log.Printf("Server listening on port %d", port)

	http.HandleFunc("/mozillahubs", createHandlerWithPath(saveDir))
	err := http.ListenAndServe(portSpec, nil)

	if err != nil {
		log.Println(err)
	}
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
