package main

import (
	"encoding/json"
	"errors"
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

		outFile.WriteString(infoPart + dataPart[:len(dataPart)-1] + "\n")
	}

	log.Printf("writeToFile %d", time.Now().UnixNano() - start)
	return nil
}

func parseRequestBody(body *[]byte) (*inputData, error) {
  start := time.Now().UnixNano()
	var bodyData *inputData = &inputData{}

	if err := json.Unmarshal(*body, bodyData); err != nil {
		return nil, err
	}

	if len(bodyData.Info) == 0 || len(bodyData.Data) == 0 {
		return nil, errors.New("Required fields are missing or invalid")
	}

	log.Printf("parseRequestBody %d", time.Now().UnixNano() - start)
	return bodyData, nil
}

func createHandlerWithPath(saveDir string) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			log.Println("Processing request")
      start := time.Now().UnixNano()
			var err error

			timestamp := getTimestamp()
			body, err := ioutil.ReadAll(req.Body)

			if err != nil {
				log.Println("Could not read request body:", err)

				msg, _ := json.Marshal(&errorResponse{Status: "error"})
				http.Error(writer, string(msg), http.StatusBadRequest)

				return
			}

			bodyData, err := parseRequestBody(&body)

			if err != nil {
				log.Println("Could not decode body data:", err)

				msg, _ := json.Marshal(&errorResponse{Status: "error"})
				http.Error(writer, string(msg), http.StatusBadRequest)

				return
			}

			saveName := path.Join(saveDir, time.Now().Format("datalog-2006-01-02.csv"))

			if err := writeToFile(saveName, bodyData); err != nil {
				log.Println("Could not save data to file:", err)

				msg, _ := json.Marshal(&errorResponse{Status: "error"})
				http.Error(writer, string(msg), http.StatusBadRequest)

				return
			}

			msg, _ := json.Marshal(&successResponse{
				Status: "ok",
				Time:   timestamp,
			})

      log.Printf("createHandlerWithPath %d", time.Now().UnixNano() - start)
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
